package input

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	initError        error
	initPathOrHandle string
	initResult       string
	initType         string
	rmError          error
	rmHandle         string
	rmType           string
}

func (sre *stubRemoveExecutor) Init(initType, initPathOrHandle string) (string, error) {
	sre.initType = initType
	sre.initPathOrHandle = initPathOrHandle
	return sre.initResult, sre.initError
}

func (sre *stubRemoveExecutor) Remove(rmType, rmHandle string) error {
	sre.rmType = rmType
	sre.rmHandle = rmHandle
	return sre.rmError
}

func TestNewRemoveInput(t *testing.T) {
	var expectedHandle = "handle"
	var testExecutor = &stubRemoveExecutor{
		initResult: expectedHandle,
	}
	var testCmd = NewRemoveInput(func(command *cobra.Command) RemoveExecutor {
		return testExecutor
	})

	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, "input", testExecutor.rmType)
	assert.Equal(t, expectedHandle, testExecutor.rmHandle)
}

func TestNewRemoveInput_RemoveError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHandle = "handle"
	var testExecutor = &stubRemoveExecutor{
		rmError:    errors.New("expected"),
		initResult: expectedHandle,
	}
	var testCmd = NewRemoveInput(func(command *cobra.Command) RemoveExecutor {
		return testExecutor
	})

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, "input", testExecutor.initType)
	assert.Equal(t, expectedHandle, testExecutor.initPathOrHandle)
	assert.Equal(t, "input", testExecutor.rmType)
	assert.Equal(t, expectedHandle, testExecutor.rmHandle)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewRemoveInput_InitError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHandle = "handle"
	var testExecutor = &stubRemoveExecutor{
		initError: errors.New("expected"),
	}
	var testCmd = NewRemoveInput(func(command *cobra.Command) RemoveExecutor {
		return testExecutor
	})

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}
