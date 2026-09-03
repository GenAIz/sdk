package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubUpdateExecutor struct {
	updateError  error
	updateHandle string
	updateKey    string
	updateValue  string
}

func (sae *stubUpdateExecutor) Update(handle, key, value string) error {
	sae.updateHandle = handle
	sae.updateKey = key
	sae.updateValue = value
	return sae.updateError
}

func TestNewUpdateSource(t *testing.T) {
	var testExecutor = &stubUpdateExecutor{}
	var testFactory = newStubUpdateExecutorFactory(testExecutor)
	var testCmd = NewUpdateSource(testFactory)
	var expectedHandle = "handle"
	var expectedKey = "key"
	var expectedValue = "value"

	testCmd.SetArgs([]string{expectedHandle, expectedKey, expectedValue})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, expectedHandle, testExecutor.updateHandle)
	assert.Equal(t, expectedKey, testExecutor.updateKey)
	assert.Equal(t, expectedValue, testExecutor.updateValue)
}

func TestNewUpdateSource_PipeError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var testFactory = newErrorUpdateExecutorFactory(expectedError)
	var testCmd = NewUpdateSource(testFactory)
	var expectedHandle = "handle"
	var expectedKey = "key"

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle, expectedKey})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newErrorUpdateExecutorFactory(expected error) UpdateExecutorFactory {
	return func(*cobra.Command) UpdateExecutor {
		return &stubUpdateExecutor{
			updateError: expected,
		}
	}
}

func newStubUpdateExecutorFactory(executor *stubUpdateExecutor) UpdateExecutorFactory {
	return func(*cobra.Command) UpdateExecutor {
		return executor
	}
}
