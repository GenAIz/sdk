package stdz

import (
	"bufio"
	"io"
	"os"
	"sync/atomic"
	"time"

	"genaiz.com/genaiz-lib/lang/filez"
)

type Input interface {
	io.Closer
	Poll(time.Duration, string, func())
}

type input struct {
	channel chan string
	closed  atomic.Bool
	paused  atomic.Bool
	reader  io.ReadCloser
}

func (i *input) drain() {
	for {
		select {
		case <-i.channel:
		default:
			return
		}
	}
}

func (i *input) Close() error {
	i.closed.Store(true)
	close(i.channel)
	return i.reader.Close()
}

func (i *input) Poll(timeout time.Duration, polled string, handler func()) {
	i.drain()
	i.paused.Store(false)
	defer i.paused.Store(true)

	select {
	case msg := <-i.channel:
		if msg == polled {
			handler()
		}
	case <-time.After(timeout):
		handler()
	}
}

func NewInput(reader io.ReadCloser) Input {
	var in = &input{
		channel: make(chan string, 1),
		reader:  reader,
	}

	go func() {
		var scanner = bufio.NewScanner(reader)

		for scanner.Scan() {
			if in.paused.Load() || in.closed.Load() {
				continue
			}

			select {
			case in.channel <- scanner.Text():
			default:
			}
		}
	}()
	return in
}

type SecretHandler func() ([]byte, error)

func NewDeviceHandler(device string, deviceHandler func(int) ([]byte, error)) SecretHandler {
	return func() ([]byte, error) {
		var tty *os.File
		var err error

		if tty, err = os.OpenFile(device, os.O_RDWR, 0); err == nil {
			defer filez.CloseSilently(tty)

			return deviceHandler(int(tty.Fd()))
		}

		return nil, err
	}
}
