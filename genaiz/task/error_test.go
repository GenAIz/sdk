package task

import (
	"errors"
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

func TestNewFailure(t *testing.T) {
	var expectedError = NewError("expected")
	var testError = NewFailure(expectedError)

	assert.Equal(t, testError.Error(), expectedError.Error())
}

func TestNewFailure_error(t *testing.T) {
	var expectedError = errors.New("expected")
	var testError = NewFailure(expectedError)

	assert.Equal(t, testError.Error(), expectedError.Error())
}

func TestNewFailure_unknown(t *testing.T) {
	var expectedObj = struct{ A string }{A: "Expected"}
	var testError = NewFailure(expectedObj)

	assert.NotEqual(t, testError.Error(), expectedObj.A)
	assert.NotEmpty(t, testError.Error())
}

func TestNewRequestError(t *testing.T) {
	var expectedMessage = "expectedMessage"
	var testError = NewRequestError(expectedMessage, 400)

	assert.Equal(t, expectedMessage, testError.Error())
	assert.True(t, testError.IsRequestError())
	assert.False(t, testError.IsFieldError())
	assert.False(t, testError.HasAllowedValues())
}
