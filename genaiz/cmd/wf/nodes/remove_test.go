package nodes

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	actualWorkflow *string
	actualNodes    *[]string
}

func (sr *stubRemoveExecutor) Remove(wf string, nodes ...string) {
	*sr.actualWorkflow = wf
	*sr.actualNodes = nodes
}

func TestRemoveAddNodes(t *testing.T) {
	var actualWorkflow string
	var actualNodes []string
	var expectedWorkflow = "_workflow"
	var expectedNodes = "_add-handle"
	var testRmNodes = NewRemoveNodes(
		newRemoveFactory(&actualWorkflow, &actualNodes),
		newRemoveValidator(true))

	testRmNodes.SetArgs([]string{expectedWorkflow, expectedNodes})
	assert.NoError(t, testRmNodes.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedNodes, actualNodes[0])
}

func TestNewRemoveNodesInvalidArgs(t *testing.T) {
	var testRmNodes = NewRemoveNodes(nil, newRemoveValidator(false))

	testRmNodes.SetArgs([]string{"_workflow", "_rm-handle"})
	assert.ErrorIs(t, testRmNodes.Execute(), errorInvalid)
}

func newRemoveFactory(actualWorkflow *string, actualNodes *[]string) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			actualWorkflow: actualWorkflow,
			actualNodes:    actualNodes,
		}
	}
}

func newRemoveValidator(valid bool) RemoveValidator {
	return func(args ...string) error {
		if valid {
			return nil
		} else {
			return errorInvalid
		}
	}
}
