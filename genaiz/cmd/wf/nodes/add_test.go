package nodes

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

var (
	errorInvalid = errors.New("test error")
)

type stubAddExecutor struct {
	addError      error
	addNodeHandle *string
	addWorkflow   *string
	initPath      string
	initHandle    string
	initError     error
}

func (sa *stubAddExecutor) Add(wf string, nodeHandle string) error {
	*sa.addWorkflow = wf
	*sa.addNodeHandle = nodeHandle
	return sa.addError
}

func (sa *stubAddExecutor) Init(path string) (string, error) {
	sa.initPath = path
	return sa.initHandle, sa.initError
}

func TestNewAddNodes(t *testing.T) {
	var actualWorkflow string
	var actualNodeHandle string
	var expectedWorkflow = "_workflow"
	var expectedNode = "_add-handle"
	var testAddNodes = NewAddNodes(
		func(command *cobra.Command) AddExecutor {
			return &stubAddExecutor{
				addWorkflow:   &actualWorkflow,
				addNodeHandle: &actualNodeHandle,
				initHandle:    expectedNode,
			}
		},
		newAddValidator(true))

	testAddNodes.SetArgs([]string{expectedWorkflow, expectedNode})
	assert.NoError(t, testAddNodes.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedNode, actualNodeHandle)
}

func TestNewAddNodesInvalidArgs(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testAddNodes = NewAddNodes(
		newAddFactory(nil, nil),
		newAddValidator(false))

	defer patch.Unpatch()
	testAddNodes.SetArgs([]string{"_workflow", "_add-handle"})
	assert.NoError(t, testAddNodes.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newAddFactory(actualWorkflow *string, actualNodeHandle *string) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addWorkflow:   actualWorkflow,
			addNodeHandle: actualNodeHandle,
		}
	}
}

func newAddValidator(valid bool) AddValidator {
	return func(args ...string) error {
		if valid {
			return nil
		} else {
			return errorInvalid
		}
	}
}
