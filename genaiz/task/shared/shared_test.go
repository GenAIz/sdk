package shared

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
)

type testSpec struct {
	Description string
	Key         string
	Value       string
	secret      bool
}

func (t testSpec) GetDefaultValue() string {
	return t.Value
}

func (t testSpec) GetDescription() string {
	return t.Description
}

func (t testSpec) GetKey() string {
	return t.Key
}

func (t testSpec) IsSecret() bool {
	return t.secret
}

func (t testSpec) Validate(value any) error {
	_ = value
	return nil
}

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

func TestConfigParams_EnsureConfigPath(t *testing.T) {
	var testFolder = t.TempDir()
	var testFile = filepath.Join(testFolder, "Test.yaml")
	var testParams = &ConfigParams{
		ConfigFolder: testFolder,
		ConfigName:   "Test",
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var actual string

		defer filez.CloseSilently(fd)
		actual, err = testParams.EnsureConfigPath()
		assert.NoError(t, err)
		assert.Equal(t, fd.Name(), actual)
		return
	}

	assert.Error(t, err)
}

func TestConfigParams_EnsureConfigPath_Exists(t *testing.T) {
	var testFolder = t.TempDir()
	var testParams = &ConfigParams{
		ConfigFolder: testFolder,
		ConfigName:   "Test",
		ConfigType:   new(ConfigTypeYaml),
	}
	var testFile = filepath.Join(testFolder, testParams.ConfigName+"."+ConfigTypeYaml)
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var actual string

		filez.CloseSilently(fd)
		actual, err = testParams.EnsureConfigPath()
		assert.NoError(t, err)
		assert.Equal(t, testFile, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestConfigParams_EnsureConfigPath_PathNotExist(t *testing.T) {
	var testFolder = t.TempDir()
	var testParams = &ConfigParams{
		ConfigFolder: testFolder,
		ConfigName:   "Test",
		ConfigType:   new(ConfigTypeYaml),
	}
	var actual string
	var err error

	actual, err = testParams.EnsureConfigPath()
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestConfigParams_GetConfigFile(t *testing.T) {
	var expectedName = "name"
	var expectedFolder = "folder"
	var expected = expectedFolder + "/" + expectedName + "." + ConfigTypeJson
	var testConfigParams = &ConfigParams{ConfigName: expectedName, ConfigType: new(ConfigTypeJson)}

	assert.Equal(t, expected, testConfigParams.GetConfigFile(expectedFolder))
}

func TestConfigParams_IsConfigTypeNone(t *testing.T) {
	var testConfigParams = &ConfigParams{}

	assert.True(t, testConfigParams.IsConfigTypeNone())
	testConfigParams.ConfigType = new(ConfigTypeNone)
	assert.True(t, testConfigParams.IsConfigTypeNone())
	testConfigParams.ConfigType = new(ConfigTypeJson)
	assert.False(t, testConfigParams.IsConfigTypeNone())
}

func TestConfigParams_ResolveConfigPath(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigType:   new(ConfigTypeYaml),
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
			ConfigType:   new(ConfigTypeYaml),
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
			ConfigType:   new(ConfigTypeYaml),
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

func TestConfigParams_ResolveOptionalType(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigType:   new(ConfigTypeYaml),
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}
	var actual, err = testParams.ResolveOptionalType(ConfigTypeYaml)

	assert.NoError(t, err)
	assert.Equal(t, testParams.GetConfigPath(), actual)
}

func TestConfigParams_ResolveOptionalType_NoConfigNameFound(t *testing.T) {
	var testDir = t.TempDir()
	var expectedName = "testName"
	var testParams = &ConfigParams{
		ConfigName:   expectedName,
		ConfigFolder: testDir,
	}
	var actual, err = testParams.ResolveOptionalType(ConfigTypeJson)

	assert.NoError(t, err)
	assert.Equal(t, testParams.GetConfigPath()+"."+ConfigTypeJson, actual)
}

func TestFindVarSpec(t *testing.T) {
	var expectedVarSpec = &testSpec{
		Key: "expectedKey",
	}

	assert.Equal(t, expectedVarSpec, FindVarSpec([]VarSpec{expectedVarSpec}, expectedVarSpec.Key))
}

func TestFindVarSpec_NoResults(t *testing.T) {
	var testVarSpec = &testSpec{
		Key: "testKey",
	}

	assert.Nil(t, FindVarSpec([]VarSpec{}, "key"))
	assert.Nil(t, FindVarSpec([]VarSpec{testVarSpec}, "key"))
}

func TestNewVarSpecState(t *testing.T) {
	var expectedPropSpec = testSpec{
		Key:   "propKey",
		Value: "propValue",
	}
	var testState = &task.State{
		Internal: VarSpecTracking{
			VarSpecs: []VarSpec{
				expectedPropSpec,
			},
		},
	}
	var testVarSpecState = NewVarSpecState(testState)

	assert.Equal(t, expectedPropSpec.Key, testVarSpecState.VarSpecs[0].GetKey())
	assert.Equal(t, expectedPropSpec.Value, testVarSpecState.VarSpecs[0].GetDefaultValue())
}

func TestVarSpecState_AddSpecs(t *testing.T) {
	var expectedPropSpec = testSpec{
		Key:   "propKey",
		Value: "propValue",
	}
	var testState = &task.State{}
	var testVarSpecState = NewVarSpecState(testState)

	testVarSpecState.AddSpecs([]VarSpec{expectedPropSpec})
	assert.Equal(t, expectedPropSpec.Key, testVarSpecState.VarSpecs[0].GetKey())
	assert.Equal(t, expectedPropSpec.Value, testVarSpecState.VarSpecs[0].GetDefaultValue())
	actual := testState.Internal.(VarSpecTracking)
	assert.Equal(t, expectedPropSpec.Key, actual.VarSpecs[0].GetKey())
	assert.Equal(t, expectedPropSpec.Value, actual.VarSpecs[0].GetDefaultValue())
}

func TestVarSpecState_HasSpec(t *testing.T) {
	var expectedKey = "key"
	var testTracking = &VarSpecTracking{
		VarSpecs: []VarSpec{testSpec{
			Key: expectedKey,
		}},
	}

	assert.True(t, testTracking.HasSpec(expectedKey))
	assert.False(t, testTracking.HasSpec("notKey"))
}

func TestVarSpecTracking_MergeSpecs(t *testing.T) {
	var expectedPropSpec = testSpec{
		Key:   "propKey",
		Value: "propValue",
	}
	var expectedPropSpec2 = testSpec{
		Key:   "propKey2",
		Value: "propValue2",
	}
	var testState = &task.State{
		Internal: VarSpecTracking{
			VarSpecs: []VarSpec{
				expectedPropSpec,
			},
		},
	}
	var testVarSpecState = NewVarSpecState(testState)

	assert.NoError(t, testVarSpecState.MergeSpecs([]VarSpec{expectedPropSpec2}))
	assert.Equal(t, expectedPropSpec.Key, testVarSpecState.VarSpecs[0].GetKey())
	assert.Equal(t, expectedPropSpec.Value, testVarSpecState.VarSpecs[0].GetDefaultValue())
	assert.Equal(t, expectedPropSpec2.Key, testVarSpecState.VarSpecs[1].GetKey())
	assert.Equal(t, expectedPropSpec2.Value, testVarSpecState.VarSpecs[1].GetDefaultValue())
}

func TestVarSpecTracking_MergeSpecs_Duplicate(t *testing.T) {
	var expectedPropSpec = testSpec{
		Key:   "propKey",
		Value: "propValue",
	}
	var testState = &task.State{
		Internal: VarSpecTracking{
			VarSpecs: []VarSpec{
				expectedPropSpec,
			},
		},
	}
	var testVarSpecState = NewVarSpecState(testState)

	assert.Error(t, testVarSpecState.MergeSpecs([]VarSpec{expectedPropSpec}))
}
