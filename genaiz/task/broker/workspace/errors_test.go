package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWorkspaceErrors_WorkflowHandleRequired(t *testing.T) {
	var testError = Errors.WorkflowHandleRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkflowHandleRequired(), testError)
}

func TestWorkspaceErrors_WorkflowIdKnown(t *testing.T) {
	var testError = Errors.WorkflowIdKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkflowIdKnown(), testError)
}

func TestWorkspaceErrors_WorkflowIdRequired(t *testing.T) {
	var testError = Errors.WorkflowIdRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkflowIdRequired(), testError)
}

func TestWorkspaceErrors_WorkflowOemRequired(t *testing.T) {
	var testError = Errors.WorkflowOemRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkflowOemRequired(), testError)
}

func TestWorkspaceErrors_WorkflowVersionRequired(t *testing.T) {
	var testError = Errors.WorkflowVersionRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkflowVersionRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceConflict(t *testing.T) {
	var testError = Errors.WorkspaceConflict()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceConflict(), testError)
}

func TestWorkspaceErrors_WorkspaceDefInvalid(t *testing.T) {
	var testError = Errors.WorkspaceDefInvalid()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceDefInvalid(), testError)
}

func TestWorkspaceErrors_WorkspaceFlowConflict(t *testing.T) {
	var testError = Errors.WorkspaceFlowConflict()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceFlowConflict(), testError)
}

func TestWorkspaceErrors_WorkspaceFlowDefInvalid(t *testing.T) {
	var testError = Errors.WorkspaceFlowDefInvalid()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceFlowDefInvalid(), testError)
}

func TestWorkspaceErrors_WorkspaceFlowRequired(t *testing.T) {
	var testError = Errors.WorkspaceFlowRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceFlowRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceFlowUnresolved(t *testing.T) {
	var testError = Errors.WorkspaceFlowUnresolved()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceFlowUnresolved(), testError)
}

func TestWorkspaceErrors_WorkspaceIdKnown(t *testing.T) {
	var testError = Errors.WorkspaceIdKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceIdKnown(), testError)
}

func TestWorkspaceErrors_WorkspaceIdRequired(t *testing.T) {
	var testError = Errors.WorkspaceIdRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceIdRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceInvalidNco(t *testing.T) {
	var testError = Errors.WorkspaceInvalidNco()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceInvalidNco(), testError)
}

func TestWorkspaceErrors_WorkspaceInvalidOwner(t *testing.T) {
	var testError = Errors.WorkspaceInvalidOwner()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceInvalidOwner(), testError)
}

func TestWorkspaceErrors_WorkspaceNameRequired(t *testing.T) {
	var testError = Errors.WorkspaceNameRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNameRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeFnHandleRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeFnHandleRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeFnHandleRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeFnPortNotFound(t *testing.T) {
	var testError = Errors.WorkspaceNodeFnPortNotFound()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeFnPortNotFound(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeFnPortRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeFnPortRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeFnPortRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeKnown(t *testing.T) {
	var testError = Errors.WorkspaceNodeKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeKnown(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeSetKnown(t *testing.T) {
	var testError = Errors.WorkspaceNodeSetKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeSetKnown(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeSetRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeSetRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeSetRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeSourceRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeSourceRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeSourceRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNodeStoreRequired(t *testing.T) {
	var testError = Errors.WorkspaceNodeStoreRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNodeStoreRequired(), testError)
}

func TestWorkspaceErrors_WorkspaceNotFound(t *testing.T) {
	var testError = Errors.WorkspaceNotFound()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceNotFound(), testError)
}

func TestWorkspaceErrors_WorkspaceVisibilityRequired(t *testing.T) {
	var testError = Errors.WorkspaceVisibilityRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.WorkspaceVisibilityRequired(), testError)
}
