// Package signalz exists to complement os.signal without having to pull dependencies on moby/sys/signal and moby/term
package signalz

import (
	"context"
	"os"
	"os/signal"
	"syscall"
)

var (
	terminations = map[syscall.Signal]string{
		syscall.SIGABRT: "ABRT",
		syscall.SIGBUS:  "BUS",
		syscall.SIGFPE:  "FPE",
		syscall.SIGHUP:  "HUP",
		syscall.SIGILL:  "ILL",
		syscall.SIGINT:  "INT",
		syscall.SIGKILL: "KILL",
		syscall.SIGPIPE: "PIPE",
		syscall.SIGQUIT: "QUIT",
		syscall.SIGSEGV: "SEGV",
		syscall.SIGSYS:  "SYS",
		syscall.SIGTERM: "TERM",
		syscall.SIGXCPU: "XCPU",
		syscall.SIGXFSZ: "XFSZ",
	}
)

type EscapeError struct{}

func (EscapeError) Error() string {
	return "read escape sequence"
}

func ForwardTerminate(ctx context.Context, channel <-chan os.Signal, terminate func(code string)) {
	var s os.Signal
	var ok bool

	for {
		select {
		case s, ok = <-channel:
			if !ok {
				return
			}
		case <-ctx.Done():
			return
		}

		for term, value := range terminations {
			if s == term {
				terminate(value)
				return
			}
		}
	}
}

func NewSignalChannel() chan os.Signal {
	var channel = make(chan os.Signal, 128)

	signal.Notify(channel)
	return channel
}

func StopCatch(channel chan os.Signal) {
	signal.Stop(channel)
	close(channel)
}
