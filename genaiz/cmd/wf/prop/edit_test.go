package prop

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubEditExecutor struct {
	editError      error
	initError      error
	initNodeHandle string

	argWorkflow string
	argNode     string
	argKey      string
	argValue    string
	argInit     string
}

func (s *stubEditExecutor) Edit(wfArg string, nodeArg string, key string, value string) error {
	s.argWorkflow = wfArg
	s.argNode = nodeArg
	s.argKey = key
	s.argValue = value
	return s.editError
}

func (s *stubEditExecutor) Init(nodeArg string) (string, error) {
	s.argInit = nodeArg
	return s.initNodeHandle, s.initError
}

func TestNewEditProp(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var expectedWorkflowArg = "wfArg"
	var expectedValue = "value"
	var expectedKey = "key"
	var expectedInitArg = "initNode"
	var testExecutor = &stubEditExecutor{
		initNodeHandle: expectedNodeArg,
	}
	var testFactory = newTestEditPropFactory(testExecutor)
	var testCmd = NewEditProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkflowArg, expectedInitArg, expectedKey, expectedValue})
	assert.Equal(t, expectedInitArg, testExecutor.argInit)
	assert.Equal(t, expectedValue, testExecutor.argValue)
	assert.Equal(t, expectedKey, testExecutor.argKey)
	assert.Equal(t, expectedNodeArg, testExecutor.argNode)
	assert.Equal(t, expectedWorkflowArg, testExecutor.argWorkflow)
	assert.False(t, patch.Called)
}

func TestNewEditProp_EditError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var expectedWorkflowArg = "wfArg"
	var expectedValue = "value"
	var expectedKey = "key"
	var expectedInitArg = "initNode"
	var testExecutor = &stubEditExecutor{
		editError:      errors.New("expected"),
		initNodeHandle: expectedNodeArg,
	}
	var testFactory = newTestEditPropFactory(testExecutor)
	var testCmd = NewEditProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkflowArg, expectedInitArg, expectedKey, expectedValue})
	assert.Equal(t, expectedInitArg, testExecutor.argInit)
	assert.Equal(t, expectedValue, testExecutor.argValue)
	assert.Equal(t, expectedKey, testExecutor.argKey)
	assert.Equal(t, expectedNodeArg, testExecutor.argNode)
	assert.Equal(t, expectedWorkflowArg, testExecutor.argWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewEditProp_InitError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedNodeArg = "nodeArg"
	var testExecutor = &stubEditExecutor{
		initError: errors.New("expected"),
	}
	var testFactory = newTestEditPropFactory(testExecutor)
	var testCmd = NewEditProp(testFactory)

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{"wfArg", expectedNodeArg, "key", "value"})
	assert.Equal(t, expectedNodeArg, testExecutor.argInit)
	assert.Empty(t, testExecutor.argValue)
	assert.Empty(t, testExecutor.argKey)
	assert.Empty(t, testExecutor.argNode)
	assert.Empty(t, testExecutor.argWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newTestEditPropFactory(executor *stubEditExecutor) EditExecutorFactory {
	return func(command *cobra.Command) EditExecutor {
		return executor
	}
}
