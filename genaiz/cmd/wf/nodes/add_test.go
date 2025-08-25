package nodes

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var (
	errorInvalid = errors.New("test error")
)

type stubAddExecutor struct {
	actualWorkflow   *string
	actualNodeHandle *string
}

func (sa *stubAddExecutor) Add(wf string, nodeHandle string) {
	*sa.actualWorkflow = wf
	*sa.actualNodeHandle = nodeHandle
}

func TestNewAddNodes(t *testing.T) {
	var actualWorkflow string
	var actualNodeHandle string
	var expectedWorkflow = "_workflow"
	var expectedNode = "_add-handle"
	var testAddNodes = NewAddNodes(
		newAddFactory(&actualWorkflow, &actualNodeHandle),
		newAddValidator(true))

	testAddNodes.SetArgs([]string{expectedWorkflow, expectedNode})
	assert.NoError(t, testAddNodes.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedNode, actualNodeHandle)
}

func TestNewAddNodesInvalidArgs(t *testing.T) {
	var testAddNodes = NewAddNodes(nil, newAddValidator(false))

	testAddNodes.SetArgs([]string{"_workflow", "_add-handle"})
	assert.ErrorIs(t, testAddNodes.Execute(), errorInvalid)
}

func newAddFactory(actualWorkflow *string, actualNodeHandle *string) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			actualWorkflow:   actualWorkflow,
			actualNodeHandle: actualNodeHandle,
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
