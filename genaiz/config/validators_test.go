package config

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/layout"
)

func TestAllOf(t *testing.T) {
	var testFailAllOf = AllOf(func(value any) bool {
		return true
	}, func(value any) bool {
		return false
	})

	assert.False(t, testFailAllOf("test"))
	assert.True(t, AllOf(func(value any) bool {
		return true
	})("test"))
}

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

func TestStringMaxLength(t *testing.T) {
	var testValidates = stringMaxLength(4)

	assert.True(t, testValidates("test"))
	assert.False(t, testValidates("testing"))
}

func TestStringMinLength(t *testing.T) {
	var testValidates = stringMinLength(5)

	assert.False(t, testValidates("test"))
	assert.True(t, testValidates("testing"))
}

func TestValidateDirCreated(t *testing.T) {
	assert.True(t, validateDirCreated("/tmp"))
	assert.True(t, validateDirCreated("/tmp/genait-a/b"))
	assert.NoError(t, os.RemoveAll("/tmp/genait-a"))
}

func TestValidateDirExists(t *testing.T) {
	if fd, err := os.CreateTemp("/tmp", "genaiz.config*"); err == nil {
		defer filez.RemoveSilently(fd.Name())

		assert.False(t, validateDirExists(fd.Name()))
		assert.True(t, validateDirExists("/tmp"))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestValidateFileExists(t *testing.T) {
	if fd, err := os.CreateTemp("/tmp", "genaiz.config*"); err == nil {
		defer filez.RemoveSilently(fd.Name())

		assert.True(t, validateFileExists(fd.Name()))
		assert.False(t, validateFileExists("/tmp"))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestValidateVersion(t *testing.T) {
	assert.False(t, validateVersion(""))
	assert.False(t, validateVersion("0.1"))
	assert.False(t, validateVersion("0.1.0.1"))
	assert.False(t, validateVersion("0.1a.0"))
	assert.False(t, validateVersion("03.0.0"))
	assert.True(t, validateVersion("0.0.0"))
}
