// Package ioz provides various Streams implementation to help streaming data across channels using various Readers and Writers
package ioz

import (
	"bufio"
	"context"
	"io"
	"os"

	"github.com/docker/docker/pkg/stdcopy"
)

type Streams interface {
	Stream(context.Context) error
}

type HiJackedIoStreamer struct {
	outStream io.Writer
	errStream io.Writer
	reader    *bufio.Reader
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

func NewHiJackedChannel(ctx context.Context, streamer Streams) chan error {
	var result = make(chan error, 1)

	go func() {
		result <- func() error {
			if err := streamer.Stream(ctx); err != nil {
				return err
			}

			return nil
		}()
	}()
	return result
}

func NewHiJackedStreamer(reader *bufio.Reader, out, err io.Writer) *HiJackedIoStreamer {
	return &HiJackedIoStreamer{
		outStream: out,
		errStream: err,
		reader:    reader,
	}
}

func NewHiJackedStreamerStd(reader *bufio.Reader) *HiJackedIoStreamer {
	return NewHiJackedStreamer(reader, os.Stdout, os.Stderr)
}
