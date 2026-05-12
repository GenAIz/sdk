package config

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/shared"
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
	var testValidates = AllFromEnumerated(shared.ArchTypes)

	assert.True(t, testValidates([]string{"x86", "arm"}))
	assert.False(t, testValidates([]string{"x86_64", "test"}))
}

func TestAnyOfEnumerated(t *testing.T) {
	var testValidates = AnyOfEnumerated(shared.FunctionTypes)

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
	var testDir = t.TempDir()

	assert.True(t, validateDirCreated(testDir))
	assert.True(t, validateDirCreated(filepath.Join(testDir, "genait-a", "b")))
}

func TestValidateDirExists(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := os.CreateTemp(testDir, "genaiz.config*"); err == nil {
		defer filez.CloseSilently(fd)

		assert.False(t, validateDirExists(fd.Name()))
		assert.True(t, validateDirExists(testDir))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestValidateFileExists(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := os.CreateTemp(testDir, "genaiz.config*"); err == nil {
		defer filez.RemoveSilently(fd.Name())

		assert.True(t, validateFileExists(fd.Name()))
		assert.False(t, validateFileExists(testDir))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestValidateEnvKeyPairs(t *testing.T) {
	assert.False(t, validateEnvKeyPairs("invalid"))
	assert.False(t, validateEnvKeyPairs([]string{"key=valid", "invalid"}))
	assert.True(t, validateEnvKeyPairs("key_1=valid"))
	assert.True(t, validateEnvKeyPairs([]string{"key_2=value,value", "MY_KEY=$VALUE"}))
}

func TestValidateEnvKeyValue(t *testing.T) {
	assert.False(t, validateEnvKeyValue("invalid"))
	assert.True(t, validateEnvKeyValue("key_1=valid"))
}

func TestValidateRepository(t *testing.T) {
	assert.False(t, validateRepository(""))
	assert.False(t, validateRepository("abcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyzabcdefghijklmnopqrstuvwxyz"))
	assert.False(t, validateRepository("namespace/invalid.."))
	assert.True(t, validateRepository("handle"))
	assert.True(t, validateRepository("oem/handle"))
	assert.True(t, validateRepository("oem/namespace/handle"))
}

func TestValidateVersion(t *testing.T) {
	assert.False(t, validateVersion(""))
	assert.False(t, validateVersion("0.1"))
	assert.False(t, validateVersion("0.1.0.1"))
	assert.False(t, validateVersion("0.1a.0"))
	assert.False(t, validateVersion("03.0.0"))
	assert.True(t, validateVersion("0.0.0"))
}

func TestValidators_ValidateEnvKey(t *testing.T) {
	assert.Error(t, Validation.ValidateEnvKey("..Not/valid"))
	assert.NoError(t, Validation.ValidateEnvKey("A_VALID_KEY"))
	assert.Error(t, Validation.ValidateEnvKey("small_caps_are_not_valid"))
	assert.NoError(t, Validation.ValidateEnvKey("_ANOTHER_VALID_KEY"))
}
