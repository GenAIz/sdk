// Package ioz provides various Streams implementation to help streaming data across channels using various Readers and Writers
package ioz

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"os/signal"
	"syscall"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/signalz"
)

var (
	ErrorUntilAborted  = errors.New("process aborted")
	ErrorUntilNotFound = errors.New("until expired without a result")
)

type Fork interface {
	GetWaitError() error

	Run(context.Context) error

	WithPipeErr(func(string)) Fork

	WithPipeOut(func(string)) Fork

	WithStdErr(*os.File) Fork

	WithStdIn(*os.File) Fork

	WithStdOut(*os.File) Fork
}

type fork struct {
	cmd         *exec.Cmd
	pipeErrFunc func(string)
	pipeOutFunc func(string)
	waitError   error
}

func (f *fork) GetWaitError() error {
	return f.waitError
}

func (f *fork) Run(ctx context.Context) error {
	var channelShell = signalz.NewSignalChannel()
	var childCtx, cancel = context.WithCancel(ctx)
	var pipeErr, pipeOut io.ReadCloser
	var err error

	go signalz.ForwardTerminate(childCtx, channelShell, func(code string) {
		_ = f.cmd.Process.Kill()
	})

	if pipeErr, err = f.pipeErr(); err == nil {
		if pipeOut, err = f.pipeOut(); err == nil {

			if err = f.cmd.Start(); err == nil {
				defer func() {
					var e = f.cmd.Wait()
					var exitErr *exec.ExitError

					if errors.As(e, &exitErr) {
						if ws, ok := exitErr.Sys().(syscall.WaitStatus); ok {
							f.waitError = fmt.Errorf("child process exited with code %d", ws.ExitStatus())
						}
					}
				}()

				if pipeOut != nil {
					go Scan(pipeOut, f.pipeOutFunc)
				}

				if pipeErr != nil {
					go Scan(pipeErr, f.pipeErrFunc)
				}
			}
		}
	}

	defer cancel()
	return err
}

func (f *fork) WithPipeErr(pipeErrFunc func(string)) Fork {
	f.pipeErrFunc = pipeErrFunc
	return f
}

func (f *fork) WithPipeOut(pipeOutFunc func(string)) Fork {
	f.pipeOutFunc = pipeOutFunc
	return f
}

func (f *fork) WithStdErr(stdErr *os.File) Fork {
	f.cmd.Stderr = stdErr
	return f
}

func (f *fork) WithStdIn(stdIn *os.File) Fork {
	f.cmd.Stdin = stdIn
	return f
}

func (f *fork) WithStdOut(stdOut *os.File) Fork {
	f.cmd.Stdout = stdOut
	return f
}

func (f *fork) pipeErr() (io.ReadCloser, error) {
	if f.pipeErrFunc != nil {
		return f.cmd.StderrPipe()
	}

	return nil, nil
}

func (f *fork) pipeOut() (io.ReadCloser, error) {
	if f.pipeOutFunc != nil {
		return f.cmd.StdoutPipe()
	}

	return nil, nil
}

func NewFork(cmd *exec.Cmd) Fork {
	return &fork{cmd: cmd}
}

// Scan will read all input from an io.ReadCloser calling consumer on all strings scanned until EOF
func Scan(reader io.ReadCloser, consumer func(string)) {
	var scanner = bufio.NewScanner(reader)

	for scanner.Scan() {
		consumer(scanner.Text())
	}
}

type ProberProcessor[T any] func([]byte) (*T, error)

// Prober wraps a bufio.Scanner preventing soft interruption from the OS while scanning. Typically, this is guards against SIGINT, but not SIGTERM or SIGKILL
type Prober[T any] interface {
	Until(io.ReadCloser) (*T, error)
}

type prober[T any] struct {
	processor ProberProcessor[T]
}

func (p prober[T]) Until(reader io.ReadCloser) (*T, error) {
	var channel = make(chan os.Signal, 1)
	var scanner = bufio.NewScanner(reader)
	var err, cancelledErr error
	var result *T

	signal.Notify(channel, os.Interrupt)

	go func() {
		var code = <-channel

		if code == os.Interrupt {
			fmt.Println("User cancelled process...")
			cancelledErr = ErrorUntilAborted
		}
	}()

	defer filez.CloseSilently(reader)
	defer signalz.StopCatch(channel)

	for scanner.Scan() {
		if result, err = p.processor(scanner.Bytes()); err != nil {
			break
		}
	}

	if cancelledErr != nil {
		return result, cancelledErr
	}

	if result == nil {
		return nil, ErrorUntilNotFound
	}

	return result, err
}

func NewProber[T any](processor ProberProcessor[T]) Prober[T] {
	return &prober[T]{
		processor: processor,
	}
}
