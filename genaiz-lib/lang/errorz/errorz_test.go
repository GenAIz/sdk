package errorz

import (
	"bufio"
	"bytes"
	"errors"
	"genaiz.com/genaiz-lib/mock"
	"github.com/stretchr/testify/assert"
	"io"
	"os"
	"testing"
)

func TestDeferOnExit_AllNil(t *testing.T) {
	var errBytes bytes.Buffer
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var fn = deferOnExit(nil, bufio.NewWriter(&errBytes), nil)

	defer patch.Unpatch()
	fn()
	assert.False(t, patch.Called)
	assert.Empty(t, errBytes.Bytes())
}

func TestDeferOnExit_NilErrorPointer(t *testing.T) {
	var actual string
	var errBytes bytes.Buffer
	var err error
	var expected = "expected"
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var fn = deferOnExit(&err, bufio.NewWriter(&errBytes), func() {
		actual = expected
	})

	defer patch.Unpatch()
	fn()
	assert.False(t, patch.Called)
	assert.Empty(t, errBytes.Bytes())
	assert.Equal(t, expected, actual)
}

func TestDeferOnExit(t *testing.T) {
	var expected = "expected"
	var err = errors.New(expected)
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var r, w, _ = os.Pipe()
	os.Stderr = w
	var fn = DeferOnExit(&err, nil)
	var stderrRestore = os.Stderr

	defer patch.Unpatch()
	defer func() {
		os.Stderr = stderrRestore
	}()

	fn()
	_ = w.Close()
	b, _ := io.ReadAll(r)

	assert.True(t, patch.Called)
	assert.Contains(t, string(b), expected)
}
