package store

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	removeError error
	dataStore   string
	flow        string
	node        string
	workspace   string
}

func (s *stubRemoveExecutor) Remove(ws, fl, nd, ds string) error {
	s.workspace = ws
	s.flow = fl
	s.node = nd
	s.dataStore = ds
	return s.removeError
}

func TestNewRemoveStore(t *testing.T) {
	var stubExecutor = &stubRemoveExecutor{}
	var testCmd = NewRemoveStore(func(command *cobra.Command) RemoveExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataStore",
	}

	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataStore)
}

func TestNewRemoveStore_RemoveError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var stubExecutor = &stubRemoveExecutor{
		removeError: expectedError,
	}
	var testCmd = NewRemoveStore(func(command *cobra.Command) RemoveExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataStore",
	}

	defer patch.Unpatch()
	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataStore)
	assert.True(t, patch.Called)
	assert.Equal(t, 1, patch.CalledWith)
}
