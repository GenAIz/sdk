package signalz

import (
	"context"
	"os"
	"syscall"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestEscapeError_Error(t *testing.T) {
	var testError = &EscapeError{}

	assert.NotEmpty(t, testError.Error())
}

func TestForwardTerminate(t *testing.T) {
	var testChannel = NewSignalChannel()

	go ForwardTerminate(context.Background(), testChannel, func(code string) {
		assert.EqualValues(t, code, terminations[syscall.SIGINT])
	})
	defer StopCatch(testChannel)

	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}

func TestForwardTerminateDone(t *testing.T) {
	var testChannel = NewSignalChannel()
	var ctx, cancel = context.WithCancel(context.Background())

	go ForwardTerminate(ctx, testChannel, func(code string) {
		assert.Fail(t, "should have not terminated")
	})
	defer StopCatch(testChannel)

	cancel()
	// An attempt to correct coverage flakiness as sometimes the testChannel is stopped before it can return on the cancel
	time.Sleep(100 * time.Millisecond)
}

func TestForwardTerminateNotOk(t *testing.T) {
	var testChannel = NewSignalChannel()
	var ctx = context.Background()

	go ForwardTerminate(ctx, testChannel, func(code string) {
		assert.Fail(t, "should have not terminated")
	})
	defer StopCatch(testChannel)

	// duplicates the Done channel, therefor not ok
	ctx.Done()
}

func TestNewSignalChannel(t *testing.T) {
	var testChannel = NewSignalChannel()
	var s os.Signal

	go func() {
		s = <-testChannel
		assert.EqualValues(t, syscall.SIGINT, s)
	}()
	defer StopCatch(testChannel)

	_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
}
