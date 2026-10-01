package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	removeError error
	dataSource  string
	flow        string
	node        string
	workspace   string
}

func (s *stubRemoveExecutor) Remove(ws, fl, nd, ds string) error {
	s.workspace = ws
	s.flow = fl
	s.node = nd
	s.dataSource = ds
	return s.removeError
}

func TestNewRemoveSource(t *testing.T) {
	var stubExecutor = &stubRemoveExecutor{}
	var testCmd = NewRemoveSource(func(command *cobra.Command) RemoveExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataSource",
	}

	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataSource)
}

func TestNewRemoveSource_RemoveError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var stubExecutor = &stubRemoveExecutor{
		removeError: expectedError,
	}
	var testCmd = NewRemoveSource(func(command *cobra.Command) RemoveExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataSource",
	}

	defer patch.Unpatch()
	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataSource)
	assert.True(t, patch.Called)
	assert.Equal(t, 1, patch.CalledWith)
}
