package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
)

func TestIdentity_HasIdentifier(t *testing.T) {
	var testIdentity = &Identity{}

	assert.False(t, testIdentity.HasIdentifier())
	testIdentity.Id = "37"
	assert.True(t, testIdentity.HasIdentifier())
}

func TestIdentity_HasRepoIdentifier(t *testing.T) {
	var testIdentity = &Identity{}

	assert.False(t, testIdentity.HasRepoIdentifier())
	testIdentity.Id = "37"
	assert.False(t, testIdentity.HasRepoIdentifier())
	testIdentity.Hash = "hash"
	assert.True(t, testIdentity.HasRepoIdentifier())
}

func TestIdentity_IsFlagSet(t *testing.T) {
	var testIdentity = &Identity{
		Flags: 1 << 0,
	}

	assert.True(t, testIdentity.IsFlagSet(1<<0))
}

func TestConfigParams_GetConfigFile(t *testing.T) {
	var expectedName = "name"
	var expectedFolder = "folder"
	var expected = expectedFolder + "/" + expectedName + "." + ConfigTypeJson
	var testConfigParams = &ConfigParams{ConfigName: expectedName, ConfigType: lang.Ref(ConfigTypeJson)}

	assert.Equal(t, expected, testConfigParams.GetConfigFile(expectedFolder))
}

func TestConfigParams_IsConfigTypeNone(t *testing.T) {
	var testConfigParams = &ConfigParams{}

	assert.True(t, testConfigParams.IsConfigTypeNone())
	testConfigParams.ConfigType = lang.Ref(ConfigTypeNone)
	assert.True(t, testConfigParams.IsConfigTypeNone())
	testConfigParams.ConfigType = lang.Ref(ConfigTypeJson)
	assert.False(t, testConfigParams.IsConfigTypeNone())
}

func TestConfigParams_ResolveConfigPath(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigType:   lang.Ref(ConfigTypeYaml),
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}
	var actual, err = testParams.ResolveConfigPath()

	assert.NoError(t, err)
	assert.Equal(t, testParams.GetConfigPath(), actual)
}

func TestConfigParams_ResolveConfigPath_ConfigFileExists(t *testing.T) {
	var testDir = t.TempDir()
	var configName = "configParams_ResolveConfigPath"
	var actual string

	if fd, err := os.CreateTemp(testDir, configName+"*."+ConfigTypeYaml); err == nil {
		defer filez.RemoveSilently(fd.Name())
		var base = filepath.Base(fd.Name())
		var testParams = &ConfigParams{
			ConfigType:   lang.Ref(ConfigTypeYaml),
			ConfigName:   base[0:strings.Index(base, ".")],
			ConfigFolder: testDir,
		}

		actual, err = testParams.ResolveConfigPath()
		assert.ErrorIs(t, err, ErrorConfigFileExists)
		assert.Equal(t, fd.Name(), actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestConfigParams_ResolveConfigPath_ConfigFileInvalid(t *testing.T) {
	var testDir = t.TempDir()
	var invalidPath = filepath.Join(testDir, "configParams.Invalid")
	var invalidFile = invalidPath + "." + ConfigTypeYaml
	var actual string

	if err := os.MkdirAll(invalidFile, 0750); err == nil {
		var testParams = &ConfigParams{
			ConfigType:   lang.Ref(ConfigTypeYaml),
			ConfigName:   "configParams.Invalid",
			ConfigFolder: testDir,
		}

		actual, err = testParams.ResolveConfigPath()
		assert.Error(t, err)
		assert.Empty(t, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestConfigParams_ResolveConfigPath_ConfigTypeNone(t *testing.T) {
	var testDir = t.TempDir()
	var actual string
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}

	if fd, err := os.Create(filepath.Join(testDir, "testName.yaml")); err == nil {
		defer filez.CloseSilently(fd)

		actual, err = testParams.ResolveConfigPath()
		assert.Error(t, err)
		assert.Equal(t, fd.Name(), actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestConfigParams_ResolveConfigPath_ConfigTypeNone_FileNotFound(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}
	var actual, err = testParams.ResolveConfigPath()

	assert.Error(t, err)
	assert.Empty(t, actual)
}

func TestConfigParams_ResolveConfigPath_ConfigTypeNone_FolderNotFound(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), "/_not_found")
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}
	var actual, err = testParams.ResolveConfigPath()

	assert.Error(t, err)
	assert.Empty(t, actual)
}
