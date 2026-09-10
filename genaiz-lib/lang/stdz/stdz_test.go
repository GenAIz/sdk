package stdz

import (
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
)

func TestInput_Poll_Immediate(t *testing.T) {
	var expectedInput = "t"
	var testInput Input
	var called bool

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	testInput = NewInput(r)
	defer filez.CloseSilently(testInput)
	_, err := w.Write([]byte(expectedInput))
	testInput.Poll(1*time.Second, expectedInput, func() {
		called = true
	})
	assert.NoError(t, err)

	// Wait for the input to be consumed
	time.Sleep(100 * time.Millisecond)

	assert.True(t, called)
}

func TestInput_Poll_Timeout(t *testing.T) {
	var expectedInput = "t"
	var testInput Input
	var called bool

	oldStdin := os.Stdin
	defer func() { os.Stdin = oldStdin }()

	r, w, _ := os.Pipe()
	os.Stdin = r
	testInput = NewInput(r)
	defer filez.CloseSilently(testInput)
	_, err := w.Write([]byte(expectedInput))
	testInput.Poll(0*time.Second, expectedInput, func() {
		called = true
	})
	assert.NoError(t, err)
	assert.True(t, called)
}

func TestNewDeviceHandler(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "tmpTty")
	var testHandler = NewDeviceHandler(testFile, func(i int) ([]byte, error) {
		var fd *os.File
		var result []byte
		var err error

		if fd = os.NewFile(uintptr(i), testFile); fd == nil {
			return nil, os.ErrInvalid
		}

		if result, err = os.ReadFile(fd.Name()); err != nil {
			return nil, err
		}

		return result, nil
	})
	var expectedSecret = []byte("secret")
	var out *os.File
	var err error

	if out, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(out)

		if _, err = out.Write(expectedSecret); err == nil {
			var actual []byte

			if actual, err = testHandler(); err == nil {
				assert.Equal(t, expectedSecret, actual)
				return
			}
		}
	}

	assert.Fail(t, err.Error())
}
