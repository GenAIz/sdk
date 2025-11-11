package layout

import (
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
)

func TestNewInitLayout(t *testing.T) {
	var expectedIn = "testIn"
	var expectedOut = "path"
	var testLayout = NewInitLayout(expectedIn, expectedOut)

	assert.Equal(t, expectedIn, testLayout.DirInput)
	assert.Equal(t, filepath.Join(expectedOut, "log"), testLayout.DirLog)
	assert.Equal(t, expectedOut, testLayout.DirOutput)
	assert.Equal(t, filepath.Join(expectedOut, "var"), testLayout.DirVar)
}

func TestNewInitLayout_Timestamp(t *testing.T) {
	var expectedIn = "testIn"
	var expectedOut = "path/{timestamp}"
	var testLayout = NewInitLayout(expectedIn, expectedOut)

	assert.Equal(t, expectedIn, testLayout.DirInput)
	assert.Equal(t, filepath.Join(expectedOut, "log"), testLayout.DirLog)
	assert.Equal(t, filepath.Join(expectedOut, "out"), testLayout.DirOutput)
	assert.Equal(t, filepath.Join(expectedOut, "var"), testLayout.DirVar)
}

func TestNewRunLayout(t *testing.T) {
	var testLayout = NewRunLayout()

	assert.NotEmpty(t, testLayout.DirInput)
	assert.NotEmpty(t, testLayout.DirLog)
	assert.NotEmpty(t, testLayout.DirOutput)
	assert.NotEmpty(t, testLayout.DirVar)
}

func TestInitState_AddParams(t *testing.T) {
	var testState = &task.State{}
	var testInitState = NewInitState(testState)
	var expectedMap = map[string]string{"key": "value"}

	testInitState.AddParams(expectedMap)
	assert.Equal(t, expectedMap, testInitState.params)
	actual, ok := testState.Internal.(initTracking)
	assert.True(t, ok)
	assert.Equal(t, expectedMap, actual.params)
}

func TestInitState_DefaultInput(t *testing.T) {
	var testDir = t.TempDir()
	var testInitState = NewInitState(&task.State{})
	var expectedMap = map[string]string{"input": testDir}
	var expectedOverride = "override"

	testInitState.AddParams(expectedMap)
	assert.Equal(t, testDir, testInitState.DefaultInput(""))
	assert.Equal(t, expectedOverride, testInitState.DefaultInput(expectedOverride))
}

func TestInitState_DefaultOutput(t *testing.T) {
	var testDir = t.TempDir()
	var testInitState = NewInitState(&task.State{})
	var expectedMap = map[string]string{"output": testDir}
	var expectedOverride = "override"

	testInitState.AddParams(expectedMap)
	assert.Equal(t, testDir, testInitState.DefaultOutput(""))
	assert.Equal(t, expectedOverride, testInitState.DefaultInput(expectedOverride))
}

func TestInitState_GetConfigFile(t *testing.T) {
	var testState = NewInitState(&task.State{})

	assert.Empty(t, testState.GetConfigFile())
}

func TestInitState_SetConfigFile(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "testFile")
	var testState = NewInitState(&task.State{})

	testState.SetConfigFile(testFile)
	assert.Equal(t, testFile, testState.GetConfigFile())
}

func TestNewInitState(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "testFile")
	var expectedMap = map[string]string{"expectedKey": "expectedValue"}
	var testState = &task.State{
		Internal: initTracking{
			configFile: testFile,
			params:     expectedMap,
		},
	}
	var actual = NewInitState(testState)

	assert.Equal(t, expectedMap, actual.params)
	assert.Equal(t, testFile, actual.configFile)
}
