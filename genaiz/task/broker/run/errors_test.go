package run

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRunErrors_RunDuplicated(t *testing.T) {
	var testError = Errors.RunDuplicated(37)

	assert.NotNil(t, testError)
	assert.ErrorIs(t, testError, Errors.runDuplicated)
	assert.True(t, Errors.IsRunDuplicated(testError))
}

func TestRunErrors_RunFlowUnknown(t *testing.T) {
	var testError = Errors.RunFlowUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunFlowUnknown(), testError)
}

func TestRunErrors_RunIdKnown(t *testing.T) {
	var testError = Errors.RunIdKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunIdKnown(), testError)
}

func TestRunErrors_RunIdRequired(t *testing.T) {
	var testError = Errors.RunIdRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunIdRequired(), testError)
}

func TestRunErrors_RunListUnlimited(t *testing.T) {
	var testError = Errors.RunListUnlimited()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunListUnlimited(), testError)
}

func TestRunErrors_RunSolutionUnknown(t *testing.T) {
	var testError = Errors.RunSolutionUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunSolutionUnknown(), testError)
}

func TestRunErrors_RunUnprepared(t *testing.T) {
	var testError = Errors.RunUnprepared()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunUnprepared(), testError)
}

func TestRunErrors_RunWorkflowUnknown(t *testing.T) {
	var testError = Errors.RunWorkflowUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunWorkflowUnknown(), testError)
}

func TestRunErrors_RunWorkspaceUnknown(t *testing.T) {
	var testError = Errors.RunWorkspaceUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.RunWorkspaceUnknown(), testError)
}
