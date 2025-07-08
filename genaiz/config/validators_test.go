package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/layout"
)

func TestAllFromEnumerated(t *testing.T) {
	var testValidates = AllFromEnumerated(layout.ArchTypes)

	assert.True(t, testValidates([]string{"x86", "arm"}))
	assert.False(t, testValidates([]string{"x86_64", "test"}))
}

func TestAnyOfEnumerated(t *testing.T) {
	var testValidates = AnyOfEnumerated(layout.FunctionTypes)

	assert.True(t, testValidates("trigger"))
	assert.False(t, testValidates("notValid"))
}

func TestOptionally(t *testing.T) {
	var invalidate = func(value any) bool {
		return false
	}

	assert.True(t, Optionally(invalidate)(nil))
	assert.True(t, Optionally(invalidate)(""))
	assert.False(t, Optionally(invalidate)(2))
}

func TestStringMatches(t *testing.T) {
	var testValidates = stringMatches("[a-z]")

	assert.True(t, testValidates("value"))
	assert.False(t, testValidates("VALUE"))
}

func TestValidateDirCreated(t *testing.T) {
	assert.True(t, validateDirCreated("/tmp"))
	assert.True(t, validateDirCreated("/tmp/genait-a/b"))
	assert.NoError(t, os.RemoveAll("/tmp/genait-a"))
}

func TestValidateDirExists(t *testing.T) {
	var file, err = os.CreateTemp("/tmp", "genait-*")

	assert.NoError(t, err)
	assert.False(t, validateDirExists(file.Name()))
	assert.True(t, validateDirExists("/tmp"))
	assert.NoError(t, os.Remove(file.Name()))
}

func TestValidateFileExists(t *testing.T) {
	var file, err = os.CreateTemp("/tmp", "genait-*")

	assert.NoError(t, err)
	assert.True(t, validateFileExists(file.Name()))
	assert.False(t, validateFileExists("/tmp"))
	assert.NoError(t, os.Remove(file.Name()))
}

func TestValidateVersion(t *testing.T) {
	assert.False(t, validateVersion(""))
	assert.False(t, validateVersion("0.1"))
	assert.False(t, validateVersion("0.1.0.1"))
	assert.False(t, validateVersion("0.1a.0"))
	assert.False(t, validateVersion("03.0.0"))
	assert.True(t, validateVersion("0.0.0"))
}
