package adapt

import (
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/stretchr/testify/assert"
)

func TestContainerArguments_GetContainerMountBinds(t *testing.T) {
	var testDir = t.TempDir()
	var testArgs = &ContainerArguments{}
	var actual []ContainerMountBind

	assert.Empty(t, testArgs.GetMountBinds())
	testArgs.MountInput = filepath.Join(testDir, "input")
	actual = testArgs.GetMountBinds()
	assert.Equal(t, 1, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testArgs.MountInput
	}))
	testArgs.MountOutput = filepath.Join(testDir, "output")
	actual = testArgs.GetMountBinds()
	assert.Equal(t, 2, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testArgs.MountOutput
	}))
	testArgs.MountLog = filepath.Join(testDir, "log")
	actual = testArgs.GetMountBinds()
	assert.Equal(t, 3, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testArgs.MountLog
	}))
	testArgs.MountVar = filepath.Join(testDir, "var")
	actual = testArgs.GetMountBinds()
	assert.Equal(t, 4, len(actual))
	assert.True(t, slices.ContainsFunc(actual, func(b ContainerMountBind) bool {
		return b.HostPath == testArgs.MountVar
	}))
}

func TestContainerArguments_GetEnvironment(t *testing.T) {
	var expectedEnv = map[string]string{
		"key": "value",
	}
	var testArgs = &ContainerArguments{
		EnvVars: expectedEnv,
	}

	assert.Equal(t, expectedEnv, testArgs.GetEnvironment())
}

func TestContainerArguments_GetEnvironment_Empty(t *testing.T) {
	var testArgs = &ContainerArguments{}

	assert.Empty(t, testArgs.GetEnvironment())
}

func TestContainerArguments_GetName(t *testing.T) {
	var testArgs = &ContainerArguments{}
	var expectedName = "name"
	var expectedPrefix = "prefix"
	var actual = testArgs.GetName([]container.Summary{})

	assert.Empty(t, actual)
	testArgs.Prefix = expectedPrefix
	actual = testArgs.GetName([]container.Summary{})
	assert.Equal(t, expectedPrefix+"-0", actual)
	testArgs.Name = expectedName
	actual = testArgs.GetName([]container.Summary{})
	assert.Equal(t, expectedName, actual)
}

func TestContainerArguments_MakeDisposableName(t *testing.T) {
	var testArgs = &ContainerArguments{}
	var expectedPrefix = "prefix"

	assert.True(t, strings.HasPrefix(testArgs.MakeDisposableName(), "container"))
	testArgs.Prefix = expectedPrefix
	assert.True(t, strings.HasPrefix(testArgs.MakeDisposableName(), expectedPrefix))
	testArgs.Prefix = expectedPrefix + "-10"
	assert.True(t, strings.HasPrefix(testArgs.MakeDisposableName(), expectedPrefix))
	testArgs.Name = "name"
	assert.True(t, strings.HasPrefix(testArgs.MakeDisposableName(), testArgs.Name))
}

func TestContainerArguments_MountArguments(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &ContainerArguments{
		MountInput:  filepath.Join(testDir, "input"),
		MountOutput: filepath.Join(testDir, "output"),
		MountLog:    filepath.Join(testDir, "log"),
		MountVar:    filepath.Join(testDir, "var"),
	}
	var actual = testParams.MountArguments()

	assert.Contains(t, actual, "src="+testParams.MountInput)
	assert.Contains(t, actual, "src="+testParams.MountOutput)
	assert.Contains(t, actual, "src="+testParams.MountLog)
	assert.Contains(t, actual, "src="+testParams.MountVar)
	assert.Contains(t, actual, Mounts.Input.EnvVar+"="+Mounts.Input.MountPath)
	assert.Contains(t, actual, Mounts.Output.EnvVar+"="+Mounts.Output.MountPath)
	assert.Contains(t, actual, Mounts.Log.EnvVar+"="+Mounts.Log.MountPath)
	assert.Contains(t, actual, Mounts.Var.EnvVar+"="+Mounts.Var.MountPath)
}

func TestContainerArguments_fmtMountBindArg_Empty(t *testing.T) {
	var testParams = &ContainerArguments{}

	assert.Empty(t, testParams.fmtMountBindArg(&Mounts.Input, ""))
}

func TestContainerArguments_fmtMountEnvArg_Empty(t *testing.T) {
	var testParams = &ContainerArguments{}

	assert.Empty(t, testParams.fmtMountEnvArg(&Mounts.Input, ""))
}

func TestContainerMountPoint_GetEnvKeyValue(t *testing.T) {
	var testPoint = &ContainerMountPoint{
		EnvVar:    "VAR",
		MountPath: t.TempDir(),
	}
	var actualKey, actualPath = testPoint.GetEnvKeyValue()

	assert.Equal(t, testPoint.EnvVar, actualKey)
	assert.Equal(t, testPoint.MountPath, actualPath)
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
