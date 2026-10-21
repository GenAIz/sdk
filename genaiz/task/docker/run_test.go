package docker

import (
	"errors"
	"strings"
	"testing"

	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker/adapt"
)

func TestRunTasks_StandardTaskFactory(t *testing.T) {
	var testTask = RunTasks.StandardTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestRunTasks_TestTaskFactory(t *testing.T) {
	var testTask = RunTasks.TestTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleRunCompletion(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "expectedImageId",
		},
	}
	var expectedId = "atLeast12CharactersLong"
	var stubClient = &stubDockerClient{
		containerCreateResult: client.ContainerCreateResult{
			ID: expectedId,
		},
	}
	var actual string

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunCompletion(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedId[0:12], testState.Output)
	actual = strings.Join(testState.Reports, "\n")
	assert.Contains(t, actual, expectedId[0:12])
	assert.Contains(t, actual, testParams.DockerImage)
}

func Test_handleRunCompletion_CreateError(t *testing.T) {
	var testState = newTestState()
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunCompletion(&ContainerParams{}, testState), stubClient.containerCreateError)
}

func Test_handleRunCompletion_StartError(t *testing.T) {
	var testState = newTestState()
	var expectedId = "atLeast12CharactersLong"
	var stubClient = &stubDockerClient{
		containerCreateResult: client.ContainerCreateResult{
			ID: expectedId,
		},
		containerStartError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunCompletion(&ContainerParams{}, testState), stubClient.containerStartError)
}

func Test_handleRunContext(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		ContainerArguments: adapt.ContainerArguments{
			DockerImage: "image",
		},
	}
	var expectedId = "sha256:123456789ABCD"
	var stubClient = &stubDockerClient{
		imageListResult: client.ImageListResult{
			Items: []image.Summary{
				{
					ID: expectedId,
				},
			},
		},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleRunContext(testParams, testState), Errors.RunImageNotFound())
	assert.Equal(t, expectedId, testState.Output)
}

func Test_handleRunContext_empty(t *testing.T) {
	var testState = newTestState()
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunContext(&ContainerParams{}, testState), Errors.RunImageNotFound())
}

func Test_handleRunContext_error(t *testing.T) {
	var testState = newTestState()
	var stubClient = &stubDockerClient{
		imageListError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunContext(&ContainerParams{}, testState), stubClient.imageListError)
}

func Test_handleRunPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testParams = &ContainerParams{}
	var expectedImageId = "imageId"
	var testState = &task.State{
		Logger: testLogger,
		Output: expectedImageId,
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleRunPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Output)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, "-d ")
	assert.Contains(t, actual, expectedImageId)
	testParams.Attached = true
	testParams.Dispose = true
	testParams.DockerImage = expectedImageId
	patch.Called = false
	assert.NoError(t, handleRunPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Output)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.NotContains(t, actual, "-d ")
	assert.Contains(t, actual, "--rm ")
	assert.Contains(t, actual, expectedImageId)
}

func Test_handleTestCompletion_CreateError(t *testing.T) {
	var testState = newTestState()
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleTestCompletion(&ContainerParams{}, testState), stubClient.containerCreateError)
}

func Test_handleTestCompletion_AttachError(t *testing.T) {
	var testState = newTestState()
	var testParams = &ContainerParams{
		Env: task.Env{
			Context: t.Context(),
		},
	}
	var expectedId = "atLeast12CharactersLong"
	var stubClient = &stubDockerClient{
		containerCreateResult: client.ContainerCreateResult{
			ID: expectedId,
		},
		containerInspectError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleTestCompletion(testParams, testState), stubClient.containerInspectError)
}

func Test_handleTestPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testParams = &ContainerParams{}
	var expectedImageId = "imageId"
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedImageId,
	}
	var actual []string

	defer patch.Unpatch()
	assert.NoError(t, handleTestPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Output)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedImageId)
	testParams.Dispose = true
	testParams.DockerImage = expectedImageId
	patch.Called = false
	assert.NoError(t, handleTestPretend(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Output)
	assert.True(t, patch.Called)
	actual = patch.CalledWith.([]string)
	assert.Contains(t, actual, expectedImageId)
	assert.Contains(t, actual, "--rm ")
}
