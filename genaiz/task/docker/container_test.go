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

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz-lib/mock/net"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker/adapt"
	"genaiz.com/genaiz/task/shared"
)

func newContainerParams(ctx context.Context) *ContainerParams {
	return &ContainerParams{
		Env: task.Env{
			Context: ctx,
		},
	}
}

func TestContainerTasks_CreateTaskFactory(t *testing.T) {
	var testTask = ContainerTasks.CreateTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.Empty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestContainerTasks_DisposeTaskFactory(t *testing.T) {
	var testTask = ContainerTasks.DisposeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestContainerTasks_StartTaskFactory(t *testing.T) {
	var testTask = ContainerTasks.StartTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestContainerTasks_StopTaskFactory(t *testing.T) {
	var testTask = ContainerTasks.StopTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleContainerAttach(t *testing.T) {

}

func Test_handleContainerAttach_AttachError(t *testing.T) {
	var expectedId = "expectedId"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID: expectedId,
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: true,
				},
			},
		},

		containerAttachError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerAttachError)
	assert.Equal(t, expectedId, stubClient.containerInspectId)
	assert.Equal(t, expectedId, stubClient.containerAttachId)
}

func Test_handleContainerAttach_ChannelError(t *testing.T) {
	var expectedId = "expectedId"
	var expectedName = "name"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:    expectedId,
					Names: []string{expectedName},
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: true,
				},
			},
		},
		containerAttachResult: client.ContainerAttachResult{
			HijackedResponse: client.HijackedResponse{
				Reader: bufio.NewReader(strings.NewReader("Will yield unrecognized input header")),
				Conn:   net.StubConn{},
			},
		},
		containerWaitResult: client.ContainerWaitResult{
			Result: make(chan container.WaitResponse),
			Error:  make(chan error),
		},
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerAttach(testParams, testState))
	assert.Equal(t, expectedId, stubClient.containerInspectId)
	assert.Equal(t, expectedId, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_ChannelExitNonZero(t *testing.T) {
	var expectedId = "expectedId"
	var expectedName = "name"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:    expectedId,
					Names: []string{expectedName},
				},
			},
		},
		Logger: testLogger,
	}
	var testChannel = make(chan container.WaitResponse)
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: true,
				},
			},
		},
		containerAttachResult: client.ContainerAttachResult{
			HijackedResponse: client.HijackedResponse{
				Reader: bufio.NewReader(strings.NewReader("")),
				Conn:   net.StubConn{},
			},
		},
		containerWaitResult: client.ContainerWaitResult{
			Result: testChannel,
			Error:  make(chan error),
		},
	}

	defer installDockerClient(stubClient)()
	go func() {
		testChannel <- container.WaitResponse{StatusCode: 1}
	}()
	time.Sleep(300 * time.Millisecond)
	assert.Error(t, handleContainerAttach(testParams, testState))
	assert.Equal(t, expectedId, stubClient.containerInspectId)
	assert.Equal(t, expectedId, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_InspectError(t *testing.T) {
	var expectedId = "expectedId"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID: expectedId,
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerInspectError)
	assert.Equal(t, expectedId, stubClient.containerInspectId)
}

func Test_handleContainerAttach_KillPanic(t *testing.T) {
	var expectedId = "expectedId"
	var expectedName = "name"
	var testContext = context.WithoutCancel(t.Context())
	var testParams = newContainerParams(testContext)
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:    expectedId,
					Names: []string{expectedName},
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: false,
				},
			},
		},
		containerAttachResult: client.ContainerAttachResult{
			HijackedResponse: client.HijackedResponse{
				Reader: bufio.NewReader(strings.NewReader("")),
				Conn:   net.StubConn{},
			},
		},
		containerWaitResult: client.ContainerWaitResult{
			Result: make(chan container.WaitResponse),
			Error:  make(chan error),
		},
		containerKillError: errors.New("panic"),
	}
	var stubSignals = make(chan os.Signal, 128)
	var gate = false

	defer installSignalProvider(func() chan os.Signal {
		return stubSignals
	})()
	defer installDockerClient(stubClient)()

	go func() {
		for {
			if stubClient.containerAttachId != "" && !gate {
				stubSignals <- syscall.SIGINT
				gate = true
				break
			}
		}
	}()

	assert.NoError(t, handleContainerAttach(testParams, testState))

	for {
		if gate {
			break
		}
	}
}

func Test_handleContainerAttach_NoContainers(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{}

	assert.Error(t, handleContainerAttach(testParams, testState))
}

func Test_handleContainerAttach_StartError(t *testing.T) {
	var expectedId = "expectedId"
	var expectedName = "name"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:    expectedId,
					Names: []string{expectedName},
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: true,
				},
			},
		},
		containerAttachResult: client.ContainerAttachResult{
			HijackedResponse: client.HijackedResponse{
				Reader: bufio.NewReader(strings.NewReader("response")),
				Conn:   net.StubConn{},
			},
		},
		containerStartError: errors.New("expected"),
		containerWaitResult: client.ContainerWaitResult{
			Result: make(chan container.WaitResponse),
			Error:  make(chan error),
		},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedId, stubClient.containerInspectId)
	assert.Equal(t, expectedId, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerAttach_StartErrorDispose(t *testing.T) {
	var expectedId = "expectedId"
	var expectedName = "name"
	var testParams = newContainerParams(t.Context())
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:    expectedId,
					Names: []string{expectedName},
				},
			},
		},
		Logger: testLogger,
	}
	var stubClient = &stubDockerClient{
		containerInspectResult: client.ContainerInspectResult{
			Container: container.InspectResponse{
				ID: expectedId,
				Config: &container.Config{
					Tty: true,
				},
			},
		},
		containerAttachResult: client.ContainerAttachResult{
			HijackedResponse: client.HijackedResponse{
				Reader: bufio.NewReader(strings.NewReader("response")),
				Conn:   net.StubConn{},
			},
		},
		containerStartError: errors.New("expected"),
		containerWaitResult: client.ContainerWaitResult{
			Result: make(chan container.WaitResponse),
			Error:  make(chan error),
		},
	}

	defer installDockerClient(stubClient)()
	testParams.Dispose = true
	assert.ErrorIs(t, handleContainerAttach(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedId, stubClient.containerInspectId)
	assert.Equal(t, expectedId, stubClient.containerAttachId)
	assert.NotEmpty(t, stubClient.containerAttachOptions)
	assert.True(t, stubClient.containerAttachOptions.Stdout)
	assert.True(t, stubClient.containerAttachOptions.Stderr)
	assert.True(t, stubClient.containerAttachOptions.Stream)
	assert.False(t, stubClient.containerAttachOptions.Stdin)
}

func Test_handleContainerContext(t *testing.T) {
	var testContainer = container.Summary{ID: "id"}
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Name: "myName",
		},
	}
	var stubClient = &stubDockerClient{
		containerListResult: client.ContainerListResult{
			Items: []container.Summary{testContainer},
		},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerContext(testParams, testState), stubClient.containerListError)
	assert.NotEmpty(t, stubClient.containerListOptions.Filters)
	assert.True(t, stubClient.containerListOptions.Filters["name"][testParams.Name])
	assert.NotEmpty(t, testState.Internal)
	assert.Equal(t, stubClient.containerListResult.Items, NewClientState(testState).containers)
}

func Test_handleContainerContext_ContainerListError(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Prefix: "expectedPrefix",
		},
	}
	var stubClient = &stubDockerClient{
		containerListError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleContainerContext(testParams, testState), stubClient.containerListError)
	assert.NotEmpty(t, stubClient.containerListOptions.Filters)
	assert.True(t, stubClient.containerListOptions.Filters["name"][testParams.Prefix])
}

func Test_handleContainerContext_ContainerNotFound(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Name: "expectedName",
		},
	}
	var stubClient = &stubDockerClient{
		containerListResult: client.ContainerListResult{
			Items: []container.Summary{},
		},
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerContext(testParams, testState))
	assert.NotEmpty(t, stubClient.containerListOptions.Filters)
	assert.True(t, stubClient.containerListOptions.Filters["name"][testParams.Name])
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

	assert.ErrorIs(t, handleContainerContext(testParams, testState), Errors.ContainerNameRequired())
}

func Test_handleContainerCreate(t *testing.T) {
	var testDir = t.TempDir()
	var expectedImage = "expectedImage"
	var testState = &task.State{
		Logger: testLogger,
		Output: expectedImage,
	}
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Dispose: true,
		},
		ContainerArguments: adapt.ContainerArguments{
			MountInput:  testDir,
			MountLog:    testDir,
			MountOutput: testDir,
			MountVar:    testDir,
			Prefix:      "expectedPrefix",
		},
	}
	var expectedId = "0123456789ABC"
	var stubClient = &stubDockerClient{
		containerCreateResult: client.ContainerCreateResult{
			ID: expectedId,
		},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleContainerCreate(testParams, testState))
	assert.Equal(t, expectedId[0:12], testState.Output)
	assert.Equal(t, fmt.Sprintf("%s-0", testParams.Prefix), stubClient.containerCreateOptions.Name)

	if actualConfig := stubClient.containerCreateOptions.Config; actualConfig == nil {
		assert.Fail(t, "expected ContainerCreateOptions.Config to be set")
	} else {
		assert.False(t, actualConfig.Tty)
		assert.Equal(t, expectedImage, actualConfig.Image)
		assert.Equal(t, 7, len(actualConfig.Env))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Input.EnvVar, adapt.Mounts.Input.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Log.EnvVar, adapt.Mounts.Log.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Output.EnvVar, adapt.Mounts.Output.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Var.EnvVar, adapt.Mounts.Var.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvProgressFile, adapt.Mounts.Var.MountPath, "progress"))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvResultFile, adapt.Mounts.Var.MountPath, "result"))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvStatusFile, adapt.Mounts.Var.MountPath, "status"))
	}

	if hostConfig := stubClient.containerCreateOptions.HostConfig; hostConfig == nil {
		assert.Fail(t, "expected ContainerCreateOptions.HostConfig to be set")
	} else {
		assert.Equal(t, 4, len(hostConfig.Mounts))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Input.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Log.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Output.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Var.MountPath
		}))
	}
}

func Test_handleContainerCreate_CreateError(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Dispose: true,
		},
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImage",
			Prefix:      "expectedPrefix",
		},
	}
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handleContainerCreate(testParams, testState))
	assert.Empty(t, testState.Output)

	if actualConfig := stubClient.containerCreateOptions.Config; actualConfig == nil {
		assert.Fail(t, "expected ContainerCreateOptions.Config to be set")
	} else {
		assert.False(t, actualConfig.Tty)
		assert.Equal(t, testParams.DockerImage, actualConfig.Image)
	}
}

func Test_handleContainerCreate_NamingError(t *testing.T) {
	var testState = newTestState()

	assert.Error(t, handleContainerCreate(&ContainerParams{}, testState))
}

func Test_handleContainerCreate_UserWarning(t *testing.T) {
	var testDir = t.TempDir()
	var patchedUid = mock.Patches{T: t}.OsGetuid(-1)
	var patchedGid = mock.Patches{T: t}.OsGetgid(-1)
	var expectedImage = "expectedImage"
	var testState = &task.State{
		Logger: testLogger,
		Output: expectedImage,
	}
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Dispose: true,
		},
		ContainerArguments: adapt.ContainerArguments{
			MountInput:  testDir,
			MountLog:    testDir,
			MountOutput: testDir,
			MountVar:    testDir,
			Prefix:      "expectedPrefix",
		},
	}
	var expectedId = "0123456789ABC"
	var expectedWarning = "testWarning"
	var stubClient = &stubDockerClient{
		containerCreateResult: client.ContainerCreateResult{
			ID:       expectedId,
			Warnings: []string{expectedWarning},
		},
	}

	defer patchedUid.Unpatch()
	defer patchedGid.Unpatch()
	defer testHook.Reset()
	defer installDockerClient(stubClient)()
	testHook.Reset()
	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleContainerCreate(testParams, testState))
	assert.Equal(t, expectedId[0:12], testState.Output)
	assert.Equal(t, fmt.Sprintf("%s-0", testParams.Prefix), stubClient.containerCreateOptions.Name)

	if actualConfig := stubClient.containerCreateOptions.Config; actualConfig == nil {
		assert.Fail(t, "expected ContainerCreateOptions.Config to be set")
	} else {
		assert.False(t, actualConfig.Tty)
		assert.Equal(t, expectedImage, actualConfig.Image)
		assert.Equal(t, 7, len(actualConfig.Env))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Input.EnvVar, adapt.Mounts.Input.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Log.EnvVar, adapt.Mounts.Log.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Output.EnvVar, adapt.Mounts.Output.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s", adapt.Mounts.Var.EnvVar, adapt.Mounts.Var.MountPath))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvProgressFile, adapt.Mounts.Var.MountPath, "progress"))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvResultFile, adapt.Mounts.Var.MountPath, "result"))
		assert.Contains(t, actualConfig.Env, fmt.Sprintf("%s=%s/%s", adapt.EnvStatusFile, adapt.Mounts.Var.MountPath, "status"))
	}

	if hostConfig := stubClient.containerCreateOptions.HostConfig; hostConfig == nil {
		assert.Fail(t, "expected ContainerCreateOptions.HostConfig to be set")
	} else {
		assert.Equal(t, 4, len(hostConfig.Mounts))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Input.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Log.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Output.MountPath
		}))
		assert.True(t, slices.ContainsFunc(hostConfig.Mounts, func(m mount.Mount) bool {
			return m.Source == testDir && m.Target == adapt.Mounts.Var.MountPath
		}))
	}

	assert.True(t, slices.ContainsFunc(testHook.Entries, func(entry logrus.Entry) bool {
		return strings.Contains(entry.Message, "[-1:-1]")
	}))
	assert.True(t, slices.ContainsFunc(testHook.Entries, func(entry logrus.Entry) bool {
		return entry.Message == expectedWarning
	}))
}

func Test_handleContainerCreate_ValidationError(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Dispose: true,
		},
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImage",
			Prefix:      "expectedPrefix",
			EnvVars: map[string]string{
				"NOT_LEGAL": "value",
			},
		},
		VarSpecs: []shared.VarSpec{
			broker.PropSpec{
				Key: "LEGAL",
			},
		},
	}

	assert.Error(t, handleContainerCreate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleContainerCreatePretend(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImage",
			Prefix:      "expectedPrefix",
			MountInput:  testDir,
			MountLog:    testDir,
			MountOutput: testDir,
			MountVar:    testDir,
		},
	}

	defer patch.Unpatch()
	assert.NoError(t, handleContainerCreatePretend(testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	actual := cast.ToStringSlice(patch.CalledWith)
	assert.Contains(t, actual, testParams.DockerImage)
	assert.Contains(t, actual, fmt.Sprintf("%s-0", testParams.Prefix))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Input.EnvVar, adapt.Mounts.Input.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Log.EnvVar, adapt.Mounts.Log.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Output.EnvVar, adapt.Mounts.Output.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Var.EnvVar, adapt.Mounts.Var.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvProgressFile, adapt.Mounts.Var.MountPath, "progress"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvResultFile, adapt.Mounts.Var.MountPath, "result"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvStatusFile, adapt.Mounts.Var.MountPath, "status"))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, adapt.Mounts.Input.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, adapt.Mounts.Log.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, adapt.Mounts.Output.MountPath))
	assert.Contains(t, actual[2], fmt.Sprintf("src=%s,dst=%s,type=bind", testDir, adapt.Mounts.Var.MountPath))
}

func Test_handleContainerCreatePretend_NoMounts(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImage",
			Prefix:      "expectedPrefix",
		},
	}

	defer patch.Unpatch()
	assert.NoError(t, handleContainerCreatePretend(testParams, testState))
	assert.NotEmpty(t, patch.CalledWith)
	actual := cast.ToStringSlice(patch.CalledWith)
	assert.Contains(t, actual, testParams.DockerImage)
	assert.Contains(t, actual, fmt.Sprintf("%s-0", testParams.Prefix))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Input.EnvVar, adapt.Mounts.Input.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Log.EnvVar, adapt.Mounts.Log.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Output.EnvVar, adapt.Mounts.Output.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s", adapt.Mounts.Var.EnvVar, adapt.Mounts.Var.MountPath))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvProgressFile, adapt.Mounts.Var.MountPath, "progress"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvResultFile, adapt.Mounts.Var.MountPath, "result"))
	assert.Contains(t, actual[1], fmt.Sprintf("%s=%s/%s", adapt.EnvStatusFile, adapt.Mounts.Var.MountPath, "status"))
}

func Test_handleContainerCreatePretend_ValidationError(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Dispose: true,
		},
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImage",
			Prefix:      "expectedPrefix",
			EnvVars: map[string]string{
				"LEGAL": "value",
			},
		},
		VarSpecs: []shared.VarSpec{
			&broker.PropSpec{
				Key:  "LEGAL",
				Type: broker.PropSpecTypeInt,
			},
		},
	}

	assert.Error(t, handleContainerCreate(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleContainerDisposal(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Force: true,
		},
	}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: testLogger,
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
	var testState = newTestState()

	assert.NoError(t, handleContainerDisposal(&ContainerParams{}, testState))
}

func Test_handleContainerDisposal_RemoveError(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: testLogger,
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
		Logger: testLogger,
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
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Force: true,
		},
	}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer},
		},
		Logger: testLogger,
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
	assert.NoError(t, handleContainerDisposalPretend(&ContainerParams{}, &task.State{}))
}

func Test_handleContainerStart(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var testParams = &ContainerParams{}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   testLogger,
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
		Logger:   testLogger,
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
	assert.Error(t, handleContainerStart(&ContainerParams{}, &task.State{}))
}

func Test_handleContainerStartPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedContainer = container.Summary{ID: "expectedId"}
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Attached: true,
		},
	}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   testLogger,
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
	var testParams = &ContainerParams{
		RunArguments: adapt.RunArguments{
			Interactive: true,
		},
	}
	var testState = &task.State{
		Internal: clientTracking{containers: []container.Summary{expectedContainer}},
		Logger:   testLogger,
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
	assert.Error(t, handleContainerStartPretend(&ContainerParams{}, &task.State{}))
}

func Test_handleContainerStop(t *testing.T) {
	var expectedContainer = container.Summary{ID: "0123456789ABCD"}
	var expectedSecondContainer = container.Summary{ID: "DCBA9876543210"}
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			Force: true,
		},
	}
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{expectedContainer, expectedSecondContainer},
		},
		Logger: testLogger,
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
		Logger: testLogger,
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
	assert.NoError(t, handleContainerStopPretend(&ContainerParams{}, &task.State{}))
}

func Test_makeContainerChannel_ChannelError(t *testing.T) {
	var parent = t.Context()
	var testCtx = context.WithoutCancel(t.Context())
	var expectedResponseChannel = make(chan container.WaitResponse)
	var expectedErrorChannel = make(chan error)
	var testParams = &ContainerParams{
		Env: task.Env{
			Context: parent,
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResult: client.ContainerWaitResult{
			Result: expectedResponseChannel,
			Error:  expectedErrorChannel,
		},
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
		Env: task.Env{
			Context: parent,
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResult: client.ContainerWaitResult{
			Result: expectedResponseChannel,
			Error:  expectedErrorChannel,
		},
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
		Env: task.Env{
			Context: parent,
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResult: client.ContainerWaitResult{
			Result: expectedResponseChannel,
			Error:  expectedErrorChannel,
		},
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
		Env: task.Env{
			Context: parent,
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResult: client.ContainerWaitResult{
			Result: expectedResponseChannel,
			Error:  expectedErrorChannel,
		},
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
		Env: task.Env{
			Context: parent,
		},
	}
	var stubClient = &stubDockerClient{
		containerWaitResult: client.ContainerWaitResult{
			Result: expectedResponseChannel,
			Error:  expectedErrorChannel,
		},
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
	var expectedForeignValue = "foreign.value"
	var expectedOutputOverride = "OutputOverride"
	var expectedProgressOverride = "ProgressOverride"
	var expectedNotResolvedValue = "$VALUE_EXT_DEFINED"
	var expectedSpec = broker.PropSpec{Key: "SPEC_KEY", Value: "expectedDefault"}
	var foreignSpec = broker.PropSpec{Key: "FOREIGN_KEY"}
	var notExpectedSpec = broker.PropSpec{Key: "NOT_KEY"}
	var notResolvedSpec = broker.PropSpec{Key: "RESOLUTION_KEY"}
	var actual []string

	testEnvMap[foreignSpec.Key] = expectedForeignValue
	testEnvMap[adapt.Mounts.Output.EnvVar] = expectedOutputOverride
	testEnvMap[adapt.EnvProgressFile] = expectedProgressOverride
	testEnvMap[notResolvedSpec.Key] = expectedNotResolvedValue
	actual = makeEnvironmentValues(testEnvMap, []shared.VarSpec{expectedSpec, notExpectedSpec, foreignSpec, notResolvedSpec})
	assert.Equal(t, 10, len(actual))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s/%s",
		adapt.EnvResultFile, adapt.Mounts.Var.MountPath, filepath.Base(adapt.EnvFileMap[adapt.EnvResultFile])))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s/%s",
		adapt.EnvStatusFile, adapt.Mounts.Var.MountPath, filepath.Base(adapt.EnvFileMap[adapt.EnvStatusFile])))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", adapt.EnvProgressFile, expectedProgressOverride))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", adapt.Mounts.Var.EnvVar, adapt.Mounts.Var.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", adapt.Mounts.Log.EnvVar, adapt.Mounts.Log.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", adapt.Mounts.Input.EnvVar, adapt.Mounts.Input.MountPath))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", adapt.Mounts.Output.EnvVar, expectedOutputOverride))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", foreignSpec.Key, expectedForeignValue))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", notResolvedSpec.Key, expectedNotResolvedValue))
	assert.Contains(t, actual, fmt.Sprintf("%s=%s", expectedSpec.Key, expectedSpec.Value))
	assert.NotContains(t, actual, notExpectedSpec.Key)
}

func installSignalProvider(provider func() chan os.Signal) func() {
	var current = signalProvider
	var result = func() {
		signalProvider = current
	}

	signalProvider = provider
	return result
}
