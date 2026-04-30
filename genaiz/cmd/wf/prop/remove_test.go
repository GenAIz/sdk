package prop

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	removeError    error
	initError      error
	initNodeHandle string

	argWorkflow string
	argNode     string
	argKey      string
	argInit     string
}

func (s *stubRemoveExecutor) Remove(workflowArg string, nodeArg string, key string) error {
	s.argWorkflow = workflowArg
	s.argNode = nodeArg
	s.argKey = key
	return s.removeError
}

func (s *stubRemoveExecutor) Init(nodeArg string) (string, error) {
	s.argInit = nodeArg
	return s.initNodeHandle, s.initError
}

func TestNewRemoveProp(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var expectedWorkflowArg = "wfArg"
	var expectedKey = "key"
	var expectedInitArg = "initNode"
	var testExecutor = &stubRemoveExecutor{
		initNodeHandle: expectedNodeArg,
	}
	var testFactory = newTestRemovePropFactory(testExecutor)
	var testCmd = NewRemoveProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkflowArg, expectedInitArg, expectedKey})
	assert.Equal(t, expectedInitArg, testExecutor.argInit)
	assert.Equal(t, expectedKey, testExecutor.argKey)
	assert.Equal(t, expectedNodeArg, testExecutor.argNode)
	assert.Equal(t, expectedWorkflowArg, testExecutor.argWorkflow)
	assert.False(t, patch.Called)
}

func TestNewRemoveProp_RemoveError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var expectedWorkflowArg = "wfArg"
	var expectedKey = "key"
	var expectedInitArg = "initNode"
	var testExecutor = &stubRemoveExecutor{
		removeError:    errors.New("expected"),
		initNodeHandle: expectedNodeArg,
	}
	var testFactory = newTestRemovePropFactory(testExecutor)
	var testCmd = NewRemoveProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkflowArg, expectedInitArg, expectedKey})
	assert.Equal(t, expectedInitArg, testExecutor.argInit)
	assert.Equal(t, expectedKey, testExecutor.argKey)
	assert.Equal(t, expectedNodeArg, testExecutor.argNode)
	assert.Equal(t, expectedWorkflowArg, testExecutor.argWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewRemoveProp_InitError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var testExecutor = &stubRemoveExecutor{
		initError: errors.New("expected"),
	}
	var testFactory = newTestRemovePropFactory(testExecutor)
	var testCmd = NewRemoveProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{"wfArg", expectedNodeArg, "key"})
	assert.Equal(t, expectedNodeArg, testExecutor.argInit)
	assert.Empty(t, testExecutor.argKey)
	assert.Empty(t, testExecutor.argNode)
	assert.Empty(t, testExecutor.argWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newTestRemovePropFactory(executor *stubRemoveExecutor) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return executor
	}
}
