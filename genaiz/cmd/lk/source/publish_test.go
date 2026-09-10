package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubPublishExecutor struct {
	publishError  error
	publishHandle string
}

func (sae *stubPublishExecutor) Publish(handle string) error {
	sae.publishHandle = handle
	return sae.publishError
}

func TestNewPublishSource(t *testing.T) {
	var testExecutor = &stubPublishExecutor{}
	var testFactory = newStubPublishExecutorFactory(testExecutor)
	var testCmd = NewPublishSource(testFactory)
	var expectedHandle = "handle"

	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, expectedHandle, testExecutor.publishHandle)
}

func TestNewPublishSource_Error(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var testFactory = newErrorPublishExecutorFactory(expectedError)
	var testCmd = NewPublishSource(testFactory)
	var expectedHandle = "handle"

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newErrorPublishExecutorFactory(expected error) PublishExecutorFactory {
	return func(*cobra.Command) PublishExecutor {
		return &stubPublishExecutor{
			publishError: expected,
		}
	}
}

func newStubPublishExecutorFactory(executor *stubPublishExecutor) PublishExecutorFactory {
	return func(*cobra.Command) PublishExecutor {
		return executor
	}
}
