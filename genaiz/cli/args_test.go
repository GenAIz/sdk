package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
)

func TestArgsOptionalFolder(t *testing.T) {
	var validator = ArgsOptionalFolder("test", 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{"valid"}))
}

func TestArgsOptionalFolder_InvalidDir(t *testing.T) {
	if fd, err := os.CreateTemp("", "genaiz-args-invalid-file*.yaml"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		var validator = ArgsOptionalFolder("test", 1, config.Validation.Handle)

		err = validator(nil, []string{fd.Name()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fd.Name())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestArgsOptionalFolder_InvalidName(t *testing.T) {
	var expectedType = "testType"
	var expectedInvalid = "..invalid"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)
	var err = validator(nil, []string{expectedInvalid})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedType)
	assert.Contains(t, err.Error(), expectedInvalid)
}

func TestArgsFolderValidator_MultiArgs(t *testing.T) {
	var expectedType = "testType"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{"valid", "not--a-folder-argument"}))
}

func TestArgsFolderValidator_NoArgs(t *testing.T) {
	var expectedType = "testType"
	var validator = ArgsOptionalFolder(expectedType, 1, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{}))
}
