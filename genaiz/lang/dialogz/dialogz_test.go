package dialogz

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

type errorWriter struct {
	writeError error
}

func (ew errorWriter) Write(p []byte) (int, error) {
	_ = p
	return -1, ew.writeError
}

func TestConfirmYes_Error(t *testing.T) {
	var testInput = bytes.NewReader([]byte("Y\n"))
	var testOutput = &errorWriter{
		writeError: errors.New("expected"),
	}

	assert.False(t, ConfirmYes(testOutput, testInput, "message"))
}

func TestConfirmYes_InvalidButNo(t *testing.T) {
	var testInput = bytes.NewReader([]byte("\ntest\nno\n"))
	var testOutput = io.Writer(new(bytes.Buffer{}))

	assert.False(t, ConfirmYes(testOutput, testInput, "message"))
}

func TestConfirmYes_No(t *testing.T) {
	var testOutput = io.Writer(new(bytes.Buffer{}))

	assert.False(t, ConfirmYes(testOutput, bytes.NewReader([]byte("n\n")), "message"))
	assert.False(t, ConfirmYes(testOutput, bytes.NewReader([]byte("NO\n")), "message"))
}

func TestConfirmYes_Yes(t *testing.T) {
	var testOutput = io.Writer(new(bytes.Buffer{}))

	assert.True(t, ConfirmYes(testOutput, bytes.NewReader([]byte("Y\n")), "message"))
	assert.True(t, ConfirmYes(testOutput, bytes.NewReader([]byte("Yes\n")), "message"))
}
