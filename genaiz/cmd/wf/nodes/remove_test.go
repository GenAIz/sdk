package nodes

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	actualWorkflow *string
	actualNodes    *[]string
	findError      error
	findHandle     string
	findPath       string
}

func (sr *stubRemoveExecutor) Find(path string) (string, error) {
	sr.findPath = path
	return sr.findHandle, sr.findError
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
		func(command *cobra.Command) RemoveExecutor {
			return &stubRemoveExecutor{
				actualWorkflow: &actualWorkflow,
				actualNodes:    &actualNodes,
				findHandle:     expectedNodes,
			}
		},
		newRemoveValidator(true))

	testRmNodes.SetArgs([]string{expectedWorkflow, expectedNodes})
	assert.NoError(t, testRmNodes.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedNodes, actualNodes[0])
}

func TestNewRemoveNodesInvalidArgs(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testRmNodes = NewRemoveNodes(
		newRemoveFactory(nil, nil),
		newRemoveValidator(false))

	defer patch.Unpatch()
	testRmNodes.SetArgs([]string{"_workflow", "_rm-handle"})
	assert.NoError(t, testRmNodes.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
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
