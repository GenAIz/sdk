package streamz

import (
	"bufio"
	"bytes"
	"context"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/docker/docker/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
)

type StubStreamer struct {
	err error
}

func (ss StubStreamer) Stream(ctx context.Context) error {
	if ss.err != nil {
		return ss.err
	}

	return nil
}

func TestHiJackedStreamer_ends(t *testing.T) {
	var testTransfer = new(bytes.Buffer)
	var testWriter = stdcopy.NewStdWriter(testTransfer, stdcopy.Stdout)
	var expectedValue = "value"
	var n, err = testWriter.Write([]byte(expectedValue))
	var testStreamer *HiJackedIoStreamer
	var ctx, cancel = context.WithCancel(context.Background())

	assert.NoError(t, err)
	assert.EqualValues(t, len(expectedValue), n)
	testStreamer = NewHiJackedStreamer(bufio.NewReader(testTransfer), io.Discard, io.Discard)
	cancel()
	assert.Error(t, testStreamer.Stream(ctx))
}

func TestHiJackedIoStreamer_Stream(t *testing.T) {
	var testTransfer = new(bytes.Buffer)
	var testWriter = stdcopy.NewStdWriter(testTransfer, stdcopy.Stdout)
	var expectedValue = "value"
	var n, err = testWriter.Write([]byte(expectedValue))
	var testOut, testErr bytes.Buffer
	var testBufOut = bufio.NewWriter(&testOut)
	var testStreamer *HiJackedIoStreamer

	assert.NoError(t, err)
	assert.EqualValues(t, len(expectedValue), n)
	testStreamer = NewHiJackedStreamer(bufio.NewReader(testTransfer), testBufOut, bufio.NewWriter(&testErr))
	assert.NoError(t, testStreamer.Stream(context.Background()))
	assert.NoError(t, testBufOut.Flush())
	assert.EqualValues(t, expectedValue, testOut.String())
	assert.Empty(t, testErr.String())
}

func TestNewHiJackedChannel(t *testing.T) {
	var errChan = NewHiJackedChannel(context.Background(), &StubStreamer{})
	var testError = <-errChan

	assert.NoError(t, testError)
}

func TestNewHiJackedChannelError(t *testing.T) {
	var expectedError = errors.New("expected")
	var errChan = NewHiJackedChannel(context.Background(), &StubStreamer{
		err: expectedError,
	})
	var testError = <-errChan

	assert.EqualValues(t, expectedError, testError)
}

func TestNewHiJackedStreamerStd(t *testing.T) {
	var expectedReader = bufio.NewReader(strings.NewReader("input"))
	var testStreamer = NewHiJackedStreamerStd(expectedReader)

	assert.Same(t, expectedReader, testStreamer.reader)
	assert.Same(t, os.Stdout, testStreamer.outStream)
	assert.Same(t, os.Stderr, testStreamer.errStream)
}
