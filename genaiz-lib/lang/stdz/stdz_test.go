package stdz

import (
	"os"
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
