package docker

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"syscall"
	"testing"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/mount"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz-lib/mock/net"
	"genaiz.com/genaiz/task"
)

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

func Test_handleContainerAttach(t *testing.T) {

}

func Test_handleContainerAttach_AttachError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: true,
		},
		containerAttachError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerAttachError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerAttachId)
}

func Test_handleContainerAttach_ChannelError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId", Names: []string{"name"}}
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: true,
		},
		containerAttachResponse: &types.HijackedResponse{
			Reader: bufio.NewReader(strings.NewReader("Will yield unrecognized input header")),
			Conn:   net.StubConn{},
		},
		containerWaitResponse: make(chan container.WaitResponse),
		containerWaitError:    make(chan error),
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerAttach(testParams, testState))
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_ChannelExitNonZero(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId", Names: []string{"name"}}
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: true,
		},
		containerAttachResponse: &types.HijackedResponse{
			Reader: bufio.NewReader(strings.NewReader("")),
			Conn:   net.StubConn{},
		},
		containerWaitResponse: make(chan container.WaitResponse),
		containerWaitError:    make(chan error),
	}

	defer installDockerClient(stubClient)()
	go func() {
		stubClient.containerWaitResponse <- container.WaitResponse{StatusCode: 1}
	}()
	time.Sleep(100 * time.Millisecond)
	assert.Error(t, handleContainerAttach(testParams, testState))
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_InspectError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
	}
	var stubClient = &stubDockerClient{
		containerInspectError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerInspectError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
}

func Test_handleContainerAttach_KillPanic(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId", Names: []string{"name"}}
	var testContext = context.WithoutCancel(t.Context())
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: testContext,
			},
		},
	}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: false,
		},
		containerAttachResponse: &types.HijackedResponse{
			Reader: bufio.NewReader(strings.NewReader("")),
			Conn:   net.StubConn{},
		},
		containerWaitResponse: make(chan container.WaitResponse),
		containerWaitError:    make(chan error),
		containerKillError:    errors.New("panic"),
	}
	var stubSignals = make(chan os.Signal, 128)

	defer installSignalProvider(func() chan os.Signal {
		return stubSignals
	})()
	defer installDockerClient(stubClient)()

	go func() {
		for {
			if stubClient.containerAttachId != "" {
				stubSignals <- syscall.SIGINT
				break
			}
		}
	}()

	assert.NoError(t, handleContainerAttach(testParams, testState))
}

func Test_handleContainerAttach_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerAttach(testParams, testState))
}

func Test_handleContainerAttach_StartError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId", Names: []string{"name"}}
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: true,
		},
		containerAttachResponse: &types.HijackedResponse{
			Reader: bufio.NewReader(strings.NewReader("response")),
			Conn:   net.StubConn{},
		},
		containerStartError:   errors.New("expected"),
		containerWaitResponse: make(chan container.WaitResponse),
		containerWaitError:    make(chan error),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_StartErrorDispose(t *testing.T) {
	var expectedContainer = container.Summary{ID: "expectedId", Names: []string{"name"}}
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Dispose: true,
			Env: task.Env{
				Context: t.Context(),
			},
		},
	}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerInspect: &expectedContainer,
		containerInspectConfig: &container.Config{
			Tty: true,
		},
		containerAttachResponse: &types.HijackedResponse{
			Reader: bufio.NewReader(strings.NewReader("response")),
			Conn:   net.StubConn{},
		},
		containerStartError:   errors.New("expected"),
		containerWaitResponse: make(chan container.WaitResponse),
		containerWaitError:    make(chan error),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerInspectId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerContext(t *testing.T) {
	var testContainer = container.Summary{ID: "id"}
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &ContainerParams{Name: "expectedName"}
	var stubClient = &stubDockerClient{
		containerList: []container.Summary{testContainer},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerContext(testParams, testState), stubClient.containerListError)
	assert.NotEmpty(t, stubClient.containerListFilter)
	assert.True(t, stubClient.containerListFilter.Match("name", testParams.Name))
	assert.NotEmpty(t, testState.Internal)
	assert.Equal(t, stubClient.containerList, NewClientState(testState).containers)
}

func Test_handleContainerContext_ContainerListError(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &ContainerParams{Prefix: "expectedPrefix"}
	var stubClient = &stubDockerClient{
		containerListError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerContext(testParams, testState), stubClient.containerListError)
	assert.NotEmpty(t, stubClient.containerListFilter)
	assert.True(t, stubClient.containerListFilter.Match("name", testParams.Prefix))
}

func Test_handleContainerContext_ContainerNotFound(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &ContainerParams{Name: "expectedName"}
	var stubClient = &stubDockerClient{
		containerList: []container.Summary{},
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerContext(testParams, testState))
	assert.NotEmpty(t, stubClient.containerListFilter)
	assert.True(t, stubClient.containerListFilter.Match("name", testParams.Name))
}

func Test_handleContainerContext_HasContainers(t *testing.T) {
	var testContainer = container.Summary{ID: "id"}
	var testState = &task.State{Internal: clientTracking{containers: []container.Summary{testContainer}}}
	var testParams = &ContainerParams{}

	assert.NoError(t, handleContainerContext(testParams, testState))
}

func Test_handleContainerContext_NoNaming(t *testing.T) {
	var testState = &task.State{}
	var testParams = &ContainerParams{}

	assert.Panics(t, func() { _ = handleContainerContext(testParams, testState) })
}

func Test_handleContainerCreate(t *testing.T) {
	var testDir = t.TempDir()
	var expectedImage = "expectedImage"
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedImage,
	}
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Dispose: true,
		},
		MountInput:  testDir,
		MountLog:    testDir,
		MountOutput: testDir,
		MountVar:    testDir,
		Prefix:      "expectedPrefix",
	}
	var expectedContainer = &container.Summary{ID: "0123456789ABC"}
	var stubClient = &stubDockerClient{
		containerCreate:         expectedContainer,
		containerCreateWarnings: []string{"testWarning"},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleContainerCreate(testParams, testState))
	assert.NotEmpty(t, stubClient.containerCreateConfig)
	assert.False(t, stubClient.containerCreateConfig.Tty)
	assert.Equal(t, expectedImage, stubClient.containerCreateConfig.Image)
	assert.Equal(t, expectedContainer.ID[0:12], testState.Output)
	assert.Equal(t, fmt.Sprintf("%s-0", testParams.Prefix), stubClient.containerCreateName)
	assert.NotEmpty(t, stubClient.containerHostConfig)
	assert.Equal(t, 4, len(stubClient.containerHostConfig.Mounts))
	assert.True(t, slices.ContainsFunc(stubClient.containerHostConfig.Mounts, func(m mount.Mount) bool {
		return m.Source == testDir && m.Target == InputMount.MountPath
	}))
	assert.True(t, slices.ContainsFunc(stubClient.containerHostConfig.Mounts, func(m mount.Mount) bool {
		return m.Source == testDir && m.Target == LogMount.MountPath
	}))
	assert.True(t, slices.ContainsFunc(stubClient.containerHostConfig.Mounts, func(m mount.Mount) bool {
		return m.Source == testDir && m.Target == OutputMount.MountPath
	}))
	assert.True(t, slices.ContainsFunc(stubClient.containerHostConfig.Mounts, func(m mount.Mount) bool {
		return m.Source == testDir && m.Target == VarMount.MountPath
	}))
	assert.Equal(t, 7, len(stubClient.containerCreateConfig.Env))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s", InputMount.EnvVar, InputMount.MountPath))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s", LogMount.EnvVar, LogMount.MountPath))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s", OutputMount.EnvVar, OutputMount.MountPath))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s", VarMount.EnvVar, VarMount.MountPath))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s/%s", envProgressFile, VarMount.MountPath, "progress"))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s/%s", envResultFile, VarMount.MountPath, "result"))
	assert.Contains(t, stubClient.containerCreateConfig.Env, fmt.Sprintf("%s=%s/%s", envStatusFile, VarMount.MountPath, "status"))
}

func Test_handleContainerCreate_CreateError(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Dispose: true,
		},
		DockerImage: "expectedImage",
		Prefix:      "expectedPrefix",
	}
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerCreate(testParams, testState))
	assert.Empty(t, testState.Output)
	assert.NotEmpty(t, stubClient.containerCreateConfig)
	assert.False(t, stubClient.containerCreateConfig.Tty)
	assert.Equal(t, testParams.DockerImage, stubClient.containerCreateConfig.Image)
}

func Test_handleContainerCreate_NamingError(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerCreate(testParams, testState))
}

func Test_handleContainerCreatePretend(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &ContainerParams{
		DockerImage: "expectedImage",
		Prefix:      "expectedPrefix",
		MountInput:  testDir,
		MountLog:    testDir,
		MountOutput: testDir,
		MountVar:    testDir,
	}

	defer patch.Unpatch()
	assert.NoError(t, handleContainerCreatePretend(testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	actual := cast.ToStringSlice(patch.CalledWith)
	assert.Contains(t, actual, testParams.DockerImage)
	assert.Contains(t, actual, fmt.Sprintf("%s-0", testParams.Prefix))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", InputMount.EnvVar, InputMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", LogMount.EnvVar, LogMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", OutputMount.EnvVar, OutputMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", VarMount.EnvVar, VarMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envProgressFile, VarMount.MountPath, "progress"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envResultFile, VarMount.MountPath, "result"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envStatusFile, VarMount.MountPath, "status"))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, InputMount.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, LogMount.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, OutputMount.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, VarMount.MountPath))
}

func Test_handleContainerCreatePretend_NoMounts(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testParams = &ContainerParams{
		DockerImage: "expectedImage",
		Prefix:      "expectedPrefix",
	}
	var testState = &task.State{Logger: logrus.New()}

	defer patch.Unpatch()
	assert.NoError(t, handleContainerCreatePretend(testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	actual := cast.ToStringSlice(patch.CalledWith)
	assert.Contains(t, actual, testParams.DockerImage)
	assert.Contains(t, actual, fmt.Sprintf("%s-0", testParams.Prefix))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", InputMount.EnvVar, InputMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", LogMount.EnvVar, LogMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", OutputMount.EnvVar, OutputMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", VarMount.EnvVar, VarMount.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envProgressFile, VarMount.MountPath, "progress"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envResultFile, VarMount.MountPath, "result"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", envStatusFile, VarMount.MountPath, "status"))
}

func Test_handleContainerCreatePretend_NamingError(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerCreatePretend(testParams, testState))
}

func Test_handleContainerDisposal(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{Force: true}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleContainerDisposal(testParams, testState))
	assert.Empty(t, testState.Internal)
	assert.Equal(t, expectedContainer.ID, stubClient.containerRemoveId)
	assert.NotEmpty(t, stubClient.containerRemoveOptions)
	assert.True(t, stubClient.containerRemoveOptions.Force)
	assert.True(t, stubClient.containerRemoveOptions.RemoveVolumes)
}

func Test_handleContainerDisposal_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleContainerDisposal(testParams, testState))
}

func Test_handleContainerDisposal_RemoveError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerRemoveError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerDisposal(testParams, testState), stubClient.containerRemoveError)
}

func Test_handleContainerDisposalPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainers = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainers},
		},
		Logger: logrus.New(),
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerDisposalPretend(testParams, testState))
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedContainers.ID)
}

func Test_handleContainerDisposalPretend_Force(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{Force: true}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerDisposalPretend(testParams, testState))
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedContainer.ID)
	assert.Contains(t, actual, "-f")
}

func Test_handleContainerDisposalPretend_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.NoError(t, handleContainerDisposalPretend(testParams, testState))
}

func Test_handleContainerStart(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   logrus.New(),
	}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleContainerStart(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerStartId)
	assert.Empty(t, stubClient.containerStartOptions.CheckpointID)
	assert.Empty(t, stubClient.containerStartOptions.CheckpointDir)
}

func Test_handleContainerStart_Error(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerStartError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerStart(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedContainer.ID, stubClient.containerStartId)
	assert.Empty(t, stubClient.containerStartOptions.CheckpointID)
	assert.Empty(t, stubClient.containerStartOptions.CheckpointDir)
}

func Test_handleContainerStart_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerStart(testParams, testState))
}

func Test_handleContainerStartPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{RunParams: RunParams{Attached: true}}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   logrus.New(),
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerStartPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Equal(t, 4, len(actual))
	assert.Equal(t, "--attach ", actual[1])
	assert.Empty(t, actual[2])
	assert.Equal(t, expectedContainer.ID, actual[3])
}

func Test_handleContainerStartPretend_Interactive(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{RunParams: RunParams{Interactive: true}}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   logrus.New(),
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerStartPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Equal(t, 4, len(actual))
	assert.Empty(t, actual[1])
	assert.Equal(t, "--interactive ", actual[2])
	assert.Equal(t, expectedContainer.ID, actual[3])
}

func Test_handleContainerStartPretend_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerStartPretend(testParams, testState))
}

func Test_handleContainerStop(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var expectedSecondContainer = container.Summary{ID: "DCBA9876543210"}
	var testParams = &ContainerParams{Force: true}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer, expectedSecondContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleContainerStop(testParams, testState), stubClient.containerStopError)
	assert.NotEmpty(t, stubClient.containerStopId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerStopId[0])
	assert.Equal(t, expectedSecondContainer.ID, stubClient.containerStopId[1])
	assert.Empty(t, stubClient.containerStopOptions)
}

func Test_handleContainerStop_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.NoError(t, handleContainerStop(testParams, testState))
}

func Test_handleContainerStop_StopError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		containerStopError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerStop(testParams, testState), stubClient.containerStopError)
	assert.NotEmpty(t, stubClient.containerStopId)
	assert.Equal(t, expectedContainer.ID, stubClient.containerStopId[0])
	assert.NotEmpty(t, stubClient.containerStopOptions)
	assert.Equal(t, -1, *stubClient.containerStopOptions.Timeout)
}

func Test_handleContainerStopPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerStopPretend(testParams, testState))
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedContainer.ID)
}

func Test_handleContainerStopPretend_Force(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleContainerStopPretend(testParams, testState))
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedContainer.ID)
	assert.Contains(t, actual, "-t -1 ")
}

func Test_handleContainerStopPretend_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.NoError(t, handleContainerStopPretend(testParams, testState))
}

func Test_makeContainerChannel_ChannelError(t *testing.T) {
	var parent = t.Context()
	var testCtx = context.WithoutCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: parent,
			},
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResponse: expectedResponseChannel,
		containerWaitError:    expectedErrorChannel,
	}

	defer installDockerClient(stubClient)()
	var actual = makeContainerChannel(testCtx, testParams, "id")

	expectedErrorChannel <- errors.New("not canceled")
	value, ok := <-actual
	assert.True(t, ok)
	assert.Equal(t, 125, value)
}

func Test_makeContainerChannel_ChannelError_ContextCanceled(t *testing.T) {
	var parent = t.Context()
	var testCtx = context.WithoutCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: parent,
			},
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResponse: expectedResponseChannel,
		containerWaitError:    expectedErrorChannel,
	}

	defer installDockerClient(stubClient)()
	var actual = makeContainerChannel(testCtx, testParams, "id")

	expectedErrorChannel <- context.Canceled
	_, ok := <-actual
	assert.False(t, ok)
}

func Test_makeContainerChannel_Done(t *testing.T) {
	var parent = t.Context()
	var testCtx, cancel = context.WithCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: parent,
			},
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResponse: expectedResponseChannel,
		containerWaitError:    expectedErrorChannel,
	}

	defer installDockerClient(stubClient)()
	var actual = makeContainerChannel(testCtx, testParams, "id")

	cancel()
	_, ok := <-actual
	assert.False(t, ok)
}

func Test_makeContainerChannel_ResultError(t *testing.T) {
	var parent = t.Context()
	var testCtx, cancel = context.WithCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: parent,
			},
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResponse: expectedResponseChannel,
		containerWaitError:    expectedErrorChannel,
	}

	defer cancel()
	defer installDockerClient(stubClient)()
	var actual = makeContainerChannel(testCtx, testParams, "id")

	expectedResponseChannel <- container.WaitResponse{
		Error: &container.WaitExitError{Message: "test"},
	}
	value, ok := <-actual
	assert.True(t, ok)
	assert.Equal(t, 125, value)
}

func Test_makeContainerChannel_ResultStatusCode(t *testing.T) {
	var parent = t.Context()
	var testCtx, cancel = context.WithCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var expectedCode = int64(37)
	var testParams = &ContainerParams{
		RunParams: RunParams{
			Env: task.Env{
				Context: parent,
			},
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResponse: expectedResponseChannel,
		containerWaitError:    expectedErrorChannel,
	}

	defer cancel()
	defer installDockerClient(stubClient)()
	var actual = makeContainerChannel(testCtx, testParams, "id")

	expectedResponseChannel <- container.WaitResponse{
		StatusCode: expectedCode,
	}
	value, ok := <-actual
	assert.True(t, ok)
	assert.Equal(t, int(expectedCode), value)
}

func Test_makeEnvironmentValues(t *testing.T) {
	var testEnvMap = make(map[string]string)
	var expectedForeignKey = "FOREIGN_KEY"
	var expectedForeignValue = "foreign.value"
	var expectedOutputOverride = "OutputOverride"
	var expectedProgressOverride = "ProgressOverride"
	var expectedNotResolvedKey = "RESOLUTION_KEY"
	var expectedNotResolvedValue = "$VALUE_EXT_DEFINED"
	var actual []string

	testEnvMap[expectedForeignKey] = expectedForeignValue
	testEnvMap[OutputMount.EnvVar] = expectedOutputOverride
	testEnvMap[envProgressFile] = expectedProgressOverride
	testEnvMap[expectedNotResolvedKey] = expectedNotResolvedValue
	actual = makeEnvironmentValues(testEnvMap)
	assert.Equal(t, 9, len(actual))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s/%s", envResultFile, VarMount.MountPath, filepath.Base(FileMap[envResultFile])))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s/%s", envStatusFile, VarMount.MountPath, filepath.Base(FileMap[envStatusFile])))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", envProgressFile, expectedProgressOverride))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", VarMount.EnvVar, VarMount.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", LogMount.EnvVar, LogMount.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", InputMount.EnvVar, InputMount.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", OutputMount.EnvVar, expectedOutputOverride))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", expectedForeignKey, expectedForeignValue))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", expectedNotResolvedKey, expectedNotResolvedValue))
}

func installSignalProvider(provider func() chan os.Signal) func() {
	var current = signalProvider
	var result = func() {
		signalProvider = current
	}

	signalProvider = provider
	return result
}
