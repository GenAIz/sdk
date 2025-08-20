package shared

import (
	"testing"

	"github.com/stretchr/testify/assert"

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
