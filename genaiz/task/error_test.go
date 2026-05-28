package task

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewError(t *testing.T) {
	var expectedMessage = "message"
	var testError = NewError(expectedMessage)

	assert.Error(t, testError)
	assert.False(t, testError.HasAllowedValues())
	assert.False(t, testError.IsFieldError())
	assert.False(t, testError.IsRequestError())
	assert.Equal(t, expectedMessage, testError.Error())
}

func TestNewErrorBuilder(t *testing.T) {
	var expectedMessage = "message"
	var testError = NewErrorBuilder().
		Allowed("allowed1", "allowed2").
		Field("field").
		Status(501).
		Build(expectedMessage)

	assert.Error(t, testError)
	assert.True(t, testError.HasAllowedValues())
	assert.True(t, testError.IsFieldError())
	assert.True(t, testError.IsRequestError())
	assert.Equal(t, expectedMessage, testError.Error())
}
