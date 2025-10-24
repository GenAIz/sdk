package docker

import (
	"errors"
	"strings"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/task"
)

func TestRunParams_WaitCondition(t *testing.T) {
	var testParams = &RunParams{}

	assert.Equal(t, container.WaitConditionNextExit, testParams.WaitCondition())
	testParams.Dispose = true
	assert.Equal(t, container.WaitConditionRemoved, testParams.WaitCondition())
}

func TestNewRunTask(t *testing.T) {
	var testTask = NewRunTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewTestTask(t *testing.T) {
	var testTask = NewTestTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleRunCompletion(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var expectedId = "atLeast12CharactersLong"
	var expectedImage = "expectedImageId"
	var stubClient = &stubDockerClient{
		containerCreate: &container.Summary{
			ID: expectedId,
		},
	}
	var actual string

	defer installDockerClient(stubClient)()
	testParams.DockerImage = expectedImage
	assert.ErrorIs(t, handleRunCompletion(testParams, testState), stubClient.containerStartError)
	assert.Equal(t, expectedId[0:12], testState.Output)
	actual = strings.Join(testState.Reports, "\n")
	assert.Contains(t, actual, expectedId[0:12])
	assert.Contains(t, actual, expectedImage)
}

func Test_handleRunCompletion_CreateError(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunCompletion(testParams, testState), stubClient.containerCreateError)
}

func Test_handleRunCompletion_StartError(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var expectedId = "atLeast12CharactersLong"
	var stubClient = &stubDockerClient{
		containerCreate: &container.Summary{
			ID: expectedId,
		},
		containerStartError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunCompletion(testParams, testState), stubClient.containerStartError)
}

func Test_handleRunContext(t *testing.T) {
	var testParams = &ContainerParams{DockerImage: "image"}
	var testState = &task.State{Logger: logrus.New()}
	var expectedId = "sha256:123456789ABCD"
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{
			{
				ID: expectedId,
			},
		},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleRunContext(testParams, testState), ErrorNoImage)
	assert.Equal(t, expectedId, testState.Output)
}

func Test_handleRunContext_empty(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunContext(testParams, testState), ErrorNoImage)
}

func Test_handleRunContext_error(t *testing.T) {
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{
		imageListError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleRunContext(testParams, testState), stubClient.imageListError)
}

func Test_handleRunPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testParams = &ContainerParams{}
	var expectedImageId = "imageId"
	var testState = &task.State{
		Logger: logrus.New(),
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
	var testParams = &ContainerParams{}
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{
		containerCreateError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleTestCompletion(testParams, testState), stubClient.containerCreateError)
}

func Test_handleTestCompletion_AttachError(t *testing.T) {
	var testParams = &ContainerParams{RunParams: RunParams{Env: task.Env{Context: t.Context()}}}
	var testState = &task.State{Logger: logrus.New()}
	var expectedId = "atLeast12CharactersLong"
	var stubClient = &stubDockerClient{
		containerCreate: &container.Summary{
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
