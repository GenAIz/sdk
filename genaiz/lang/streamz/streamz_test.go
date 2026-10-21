package streamz

import (
	"bufio"
	"bytes"
	"context"
	"encoding/binary"
	"errors"
	"io"
	"os"
	"strings"
	"sync"
	"testing"

	"github.com/moby/moby/api/pkg/stdcopy"
	"github.com/stretchr/testify/assert"
)

const (
	stubWriterPrefixLen = 8
)

var bufPool = &sync.Pool{New: func() interface{} { return bytes.NewBuffer(nil) }}

type stubWriter struct {
	io.Writer
	prefix byte
}

func (w *stubWriter) Write(p []byte) (n int, err error) {
	if w == nil || w.Writer == nil {
		return 0, errors.New("writer not instantiated")
	}
	if p == nil {
		return 0, nil
	}

	header := [stubWriterPrefixLen]byte{0: w.prefix}
	binary.BigEndian.PutUint32(header[4:], uint32(len(p)))
	buf := bufPool.Get().(*bytes.Buffer)
	buf.Write(header[:])
	buf.Write(p)

	n, err = w.Writer.Write(buf.Bytes())
	n -= stubWriterPrefixLen
	if n < 0 {
		n = 0
	}

	buf.Reset()
	bufPool.Put(buf)
	return
}

type StubStreamer struct {
	err error
}

func (ss StubStreamer) Stream(context.Context) error {
	if ss.err != nil {
		return ss.err
	}

	return nil
}

func TestHiJackedStreamer_ends(t *testing.T) {
	var testTransfer = new(bytes.Buffer)
	var testWriter = &stubWriter{
		Writer: testTransfer,
		prefix: byte(stdcopy.Stdout),
	}
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
	var testWriter = &stubWriter{
		Writer: testTransfer,
		prefix: byte(stdcopy.Stdout),
	}
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
