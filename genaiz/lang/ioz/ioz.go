package ioz

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/docker/docker/pkg/stdcopy"
)

type HiJackedIoStreamer struct {
	outStream io.Writer
	errStream io.Writer
	reader    *bufio.Reader
}

func NewHiJackedStdStreamer(reader *bufio.Reader) *HiJackedIoStreamer {
	return &HiJackedIoStreamer{
		outStream: os.Stdout,
		errStream: os.Stderr,
		reader:    reader,
	}
}

func (s *HiJackedIoStreamer) begin() <-chan error {
	outputDone := make(chan error)

	go func() {
		var _, err = stdcopy.StdCopy(s.outStream, s.errStream, s.reader)

		outputDone <- err
	}()

	return outputDone
}

func (s *HiJackedIoStreamer) Stream(ctx context.Context) error {
	var outputDone = s.begin()

	select {
	case err := <-outputDone:
		return err
	case <-ctx.Done():
		return ctx.Err()
	}
}

func NewHiJackedChannel(ctx context.Context, reader *bufio.Reader) chan error {
	var result = make(chan error, 1)

	go func() {
		result <- func() error {
			var streamer = NewHiJackedStdStreamer(reader)

			if err := streamer.Stream(ctx); err != nil {
				return err
			}

			return nil
		}()
	}()
	return result
}
