package docker

import (
	"fmt"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/stretchr/testify/assert"
)

func TestContainerMountPoint_GetEnvKeyValuePair(t *testing.T) {
	var testMountPoint = &ContainerMountPoint{
		EnvVar:    "env",
		MountPath: "path",
	}

	assert.Equal(t, fmt.Sprintf("%s=%s", testMountPoint.EnvVar, testMountPoint.MountPath),
		testMountPoint.GetEnvKeyValuePair())
}

func TestContainerMountPoint_MakeMount(t *testing.T) {
	var testDir = t.TempDir()
	var testMountPoint = &ContainerMountPoint{MountPath: "/mnt/path"}
	var actual = testMountPoint.MakeMount(testDir)

	assert.Equal(t, testDir, actual.Source)
	assert.Equal(t, testMountPoint.MountPath, actual.Target)
	assert.False(t, actual.ReadOnly)
	assert.Equal(t, actual.BindOptions, &mount.BindOptions{CreateMountpoint: true})
}

func TestContainerParams_GetName(t *testing.T) {
	var testParams = &ContainerParams{}
	var expectedName = "name"
	var expectedPrefix = "prefix"
	var actual, err = testParams.GetName([]container.Summary{})

	assert.Empty(t, actual)
	assert.Error(t, err)
	testParams.Prefix = expectedPrefix
	actual, err = testParams.GetName([]container.Summary{})
	assert.Equal(t, expectedPrefix+"-0", actual)
	testParams.Name = expectedName
	actual, err = testParams.GetName([]container.Summary{})
	assert.Equal(t, expectedName, actual)
}

func TestContainerParams_GetContainerMountBinds(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &ContainerParams{}
	var actual []ContainerMountBind

	assert.Empty(t, testParams.GetContainerMountBinds())
	testParams.MountInput = filepath.Join(testDir, "input")
	actual = testParams.GetContainerMountBinds()
	assert.Equal(t, 1, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testParams.MountInput
	}))
	testParams.MountOutput = filepath.Join(testDir, "output")
	actual = testParams.GetContainerMountBinds()
	assert.Equal(t, 2, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testParams.MountOutput
	}))
	testParams.MountLog = filepath.Join(testDir, "log")
	actual = testParams.GetContainerMountBinds()
	assert.Equal(t, 3, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testParams.MountLog
	}))
	testParams.MountVar = filepath.Join(testDir, "var")
	actual = testParams.GetContainerMountBinds()
	assert.Equal(t, 4, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testParams.MountVar
	}))
}

func TestContainerParams_MakeDisposableName(t *testing.T) {
	var testParams = &ContainerParams{}
	var expectedPrefix = "prefix"

	assert.True(t, strings.HasPrefix(testParams.MakeDisposableName(), "container"))
	testParams.Prefix = expectedPrefix
	assert.True(t, strings.HasPrefix(testParams.MakeDisposableName(), expectedPrefix))
	testParams.Prefix = expectedPrefix + "-10"
	assert.True(t, strings.HasPrefix(testParams.MakeDisposableName(), expectedPrefix))
	testParams.Name = "name"
	assert.True(t, strings.HasPrefix(testParams.MakeDisposableName(), testParams.Name))
}

func TestContainerParams_fmtArgs(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &ContainerParams{
		MountInput:  filepath.Join(testDir, "input"),
		MountOutput: filepath.Join(testDir, "output"),
		MountLog:    filepath.Join(testDir, "log"),
		MountVar:    filepath.Join(testDir, "var"),
	}
	var actual = testParams.fmtArgs()

	assert.Contains(t, actual, "src="+testParams.MountInput)
	assert.Contains(t, actual, "src="+testParams.MountOutput)
	assert.Contains(t, actual, "src="+testParams.MountLog)
	assert.Contains(t, actual, "src="+testParams.MountVar)
	assert.Contains(t, actual, InputMount.EnvVar+"="+InputMount.MountPath)
	assert.Contains(t, actual, OutputMount.EnvVar+"="+OutputMount.MountPath)
	assert.Contains(t, actual, LogMount.EnvVar+"="+LogMount.MountPath)
	assert.Contains(t, actual, VarMount.EnvVar+"="+VarMount.MountPath)
}

func TestContainerParams_fmtMountBindArg(t *testing.T) {
	var testParams = &ContainerParams{}

	assert.Empty(t, testParams.fmtMountBindArg(&InputMount, ""))
}

func TestContainerParams_fmtMountEnvArg(t *testing.T) {
	var testParams = &ContainerParams{}

	assert.Empty(t, testParams.fmtMountEnvArg(&InputMount, ""))
}

func TestNewCreateTask(t *testing.T) {
	var testTask = NewCreateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.Empty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewDisposeTask(t *testing.T) {
	var testTask = NewDisposeTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewStartTask(t *testing.T) {
	var testTask = NewStartTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewStopTask(t *testing.T) {
	var testTask = NewStopTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}
