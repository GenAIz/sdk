package cli

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
)

func TestArgsFolderValidator(t *testing.T) {
	var validator = ArgsFolderValidator("test", config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{"valid"}))
}

func TestArgsFolderValidator_InvalidDir(t *testing.T) {
	if fd, err := os.CreateTemp("", "genaiz-args-invalid-file*.yaml"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		var validator = ArgsFolderValidator("test", config.Validation.Handle)

		err = validator(nil, []string{fd.Name()})
		assert.Error(t, err)
		assert.Contains(t, err.Error(), fd.Name())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestArgsFolderValidator_InvalidName(t *testing.T) {
	var expectedType = "testType"
	var expectedInvalid = "..invalid"
	var validator = ArgsFolderValidator(expectedType, config.Validation.Handle)
	var err = validator(nil, []string{expectedInvalid})

	assert.Error(t, err)
	assert.Contains(t, err.Error(), expectedType)
	assert.Contains(t, err.Error(), expectedInvalid)
}

func TestArgsFolderValidator_NoArgs(t *testing.T) {
	var expectedType = "testType"
	var validator = ArgsFolderValidator(expectedType, config.Validation.Handle)

	assert.NoError(t, validator(nil, []string{}))
}
