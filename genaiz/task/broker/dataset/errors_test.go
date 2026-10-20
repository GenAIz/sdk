package dataset

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDataSetErrors_DataSetFileExists(t *testing.T) {
	var testError = Errors.DataSetFileExists()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetFileExists(), testError)
}

func TestDataSetErrors_DataSetFileUnknown(t *testing.T) {
	var testError = Errors.DataSetFileUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetFileUnknown(), testError)
}

func TestDataSetErrors_DataSetHandleRequired(t *testing.T) {
	var testError = Errors.DataSetHandleRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetHandleRequired(), testError)
}

func TestDataSetErrors_DataSetIdKnown(t *testing.T) {
	var testError = Errors.DataSetIdKnown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetIdKnown(), testError)
}

func TestDataSetErrors_DataSetIdUnknown(t *testing.T) {
	var testError = Errors.DataSetIdUnknown()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetIdUnknown(), testError)
}

func TestDataSetErrors_DataSetWorkspaceRequired(t *testing.T) {
	var testError = Errors.DataSetWorkspaceRequired()

	assert.NotNil(t, testError)
	assert.ErrorIs(t, Errors.DataSetWorkspaceRequired(), testError)
}
