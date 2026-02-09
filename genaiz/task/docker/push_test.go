package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"slices"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestNewPushTask(t *testing.T) {
	var testTask = NewPushTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handlePushComplete(t *testing.T) {
	var expectedAuth = "auth"
	var expectedOutput = "outputId"
	var expectedPath = "remotePath"
	var expectedVersion = "version"
	var expectedDigest = "digest"
	var testParams = &PushParams{}
	var testState = &task.State{
		Internal: &shared.Identity{
			Auth:    expectedAuth,
			Path:    expectedPath,
			Version: expectedVersion,
		},
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var testPushStrings = []string{
		makeEmptyStatusPushResponse(),
		makeEmptyIdPushResponse(),
		makeProgressEmptyPushResponse(),
		makeProgressDetailPushResponse(),
		makeAuxPushResponse(expectedDigest, expectedVersion),
	}
	var stubClient = &stubDockerClient{
		imagePushReader: io.NopCloser(strings.NewReader(strings.Join(testPushStrings, "\n"))),
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handlePushComplete(testParams, testState))
	assert.Equal(t, expectedOutput, testState.Output)
	assert.Equal(t, expectedOutput, stubClient.imageTagId)
	assert.Equal(t, expectedPath, stubClient.imageTagPath)
	assert.NotEmpty(t, stubClient.imagePushOptions.RegistryAuth)
	assert.Equal(t, expectedPath, stubClient.imagePushPath)
}

func Test_handlePushComplete_auxMissing(t *testing.T) {
	var expectedAuth = "auth"
	var expectedOutput = "outputId"
	var expectedPath = "remotePath"
	var testParams = &PushParams{}
	var testState = &task.State{
		Internal: &shared.Identity{
			Auth: expectedAuth,
			Path: expectedPath,
		},
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var stubClient = &stubDockerClient{
		imagePushReader: io.NopCloser(strings.NewReader("{}")),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handlePushComplete(testParams, testState), errImproperPushResponse)
	assert.Empty(t, testState.Output)
	assert.Equal(t, expectedOutput, stubClient.imageTagId)
	assert.Equal(t, expectedPath, stubClient.imageTagPath)
	assert.NotEmpty(t, stubClient.imagePushOptions.RegistryAuth)
	assert.Equal(t, expectedPath, stubClient.imagePushPath)
}

func Test_handlePushComplete_jsonError(t *testing.T) {
	var expectedAuth = "auth"
	var expectedOutput = "outputId"
	var expectedPath = "remotePath"
	var testParams = &PushParams{}
	var testState = &task.State{
		Internal: &shared.Identity{
			Auth: expectedAuth,
			Path: expectedPath,
		},
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var stubClient = &stubDockerClient{
		imagePushReader: io.NopCloser(strings.NewReader("not valid json")),
	}

	defer installDockerClient(stubClient)()
	assert.Error(t, handlePushComplete(testParams, testState))
	assert.Empty(t, testState.Output)
	assert.Equal(t, expectedOutput, stubClient.imageTagId)
	assert.Equal(t, expectedPath, stubClient.imageTagPath)
	assert.NotEmpty(t, stubClient.imagePushOptions.RegistryAuth)
	assert.Equal(t, expectedPath, stubClient.imagePushPath)
}

func Test_handlePushComplete_provisionMissing(t *testing.T) {
	var testState = &task.State{}

	assert.ErrorIs(t, handlePushComplete(&PushParams{}, testState), errorNoProvision)
	testState.Output = "output"
	assert.ErrorIs(t, handlePushComplete(&PushParams{}, testState), errorNoProvision)
	testState.Output = ""
	testState.Internal = &shared.Identity{}
	assert.ErrorIs(t, handlePushComplete(&PushParams{}, testState), errorNoProvision)
}

func Test_handlePushComplete_pushError(t *testing.T) {
	var expectedAuth = "auth"
	var expectedOutput = "outputId"
	var expectedPath = "remotePath"
	var testParams = &PushParams{}
	var testState = &task.State{
		Internal: &shared.Identity{
			Auth: expectedAuth,
			Path: expectedPath,
		},
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var stubClient = &stubDockerClient{
		imagePushError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handlePushComplete(testParams, testState), stubClient.imagePushError)
	assert.Empty(t, testState.Output)
	assert.Equal(t, expectedOutput, stubClient.imageTagId)
	assert.Equal(t, expectedPath, stubClient.imageTagPath)
	assert.NotEmpty(t, stubClient.imagePushOptions.RegistryAuth)
	assert.Equal(t, expectedPath, stubClient.imagePushPath)
}

func Test_handlePushComplete_tagError(t *testing.T) {
	var expectedOutput = "outputId"
	var expectedPath = "remotePath"
	var testParams = &PushParams{}
	var testState = &task.State{
		Internal: &shared.Identity{
			Path: expectedPath,
		},
		Logger: logrus.New(),
		Output: expectedOutput,
	}
	var stubClient = &stubDockerClient{
		imageTagError: errors.New("expected"),
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handlePushComplete(testParams, testState), stubClient.imageTagError)
	assert.Empty(t, testState.Output)
	assert.Equal(t, expectedOutput, stubClient.imageTagId)
	assert.Equal(t, expectedPath, stubClient.imageTagPath)
}

func Test_handlePushContext(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "some garbage",
		Internal: &shared.Identity{
			Auth:  "auth",
			Flags: broker.FunctionFlags.Provisioning,
		},
	}
	var testParams = &PushParams{}

	assert.NoError(t, handlePushContext(testParams, testState))
}

func Test_handlePushContext_buildMissing(t *testing.T) {
	assert.ErrorIs(t, handlePushContext(&PushParams{}, &task.State{}), errorNoBuild)
}

func Test_handlePushContext_conflictError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "some garbage",
		Internal: &shared.Identity{
			Auth:  "auth",
			Flags: broker.FunctionFlags.Active,
		},
	}
	var testParams = &PushParams{}

	assert.ErrorIs(t, handlePushContext(testParams, testState), errorConflictPush)
}

func Test_handlePushContext_inactiveError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "some garbage",
		Internal: &shared.Identity{
			Auth:  "",
			Flags: broker.FunctionFlags.Released,
		},
	}
	var testParams = &PushParams{}

	assert.ErrorIs(t, handlePushContext(testParams, testState), errorIllegalPush)
}

func Test_handlePushContext_neutralError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "some garbage",
		Internal: &shared.Identity{
			Auth:  "",
			Flags: broker.FunctionFlags.Active,
		},
	}
	var testParams = &PushParams{}

	assert.ErrorIs(t, handlePushContext(testParams, testState), errorNeutralPush)
}

func Test_handlePushContext_provisioningMissing(t *testing.T) {
	var testState = &task.State{Output: "some garbage"}

	assert.ErrorIs(t, handlePushContext(&PushParams{}, testState), errorNoProvision)
}

func Test_handlePushIncomplete(t *testing.T) {
	var testState = &task.State{
		Error: errorNeutralPush,
		Internal: &shared.Identity{
			Hash: "hash",
		},
	}

	assert.NoError(t, handlePushIncomplete(&PushParams{}, testState))
	assert.True(t, testState.Completed)
}

func Test_handlePushIncomplete_neutralError(t *testing.T) {
	var testState = &task.State{
		Error: errorNeutralPush,
	}

	assert.ErrorIs(t, handlePushIncomplete(&PushParams{}, testState), errorNeutralPush)
	assert.True(t, testState.Completed)
}

func Test_handlePushIncomplete_synchronizedError(t *testing.T) {
	var testState = &task.State{
		Error:    errorNeutralPush,
		Internal: &shared.Identity{},
	}

	assert.ErrorIs(t, handlePushIncomplete(&PushParams{}, testState), errorSynchronizedPush)
	assert.True(t, testState.Completed)
}

func Test_handlePushIncomplete_unhandledError(t *testing.T) {
	var expectedError = errors.New("expectedError")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handlePushIncomplete(&PushParams{}, testState), expectedError)
	assert.True(t, testState.Completed)
}

func Test_handlePushPretend(t *testing.T) {
	var lines []string
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {
		lines = append(lines, fmt.Sprintf(format, a...))
	})
	var expectedPath = "path"
	var expectedHash = "hash"
	var testState = &task.State{
		Internal: &shared.Identity{
			Hash: expectedHash,
			Path: expectedPath,
		},
		Logger: logrus.New(),
	}

	defer patch.Unpatch()
	assert.NoError(t, handlePushPretend(&PushParams{}, testState))
	assert.True(t, slices.ContainsFunc(lines, func(s string) bool {
		return strings.Contains(s, expectedHash)
	}))
	assert.True(t, slices.ContainsFunc(lines, func(s string) bool {
		return strings.Contains(s, expectedPath)
	}))
}

func Test_handlePushPretend_noProvision(t *testing.T) {
	assert.ErrorIs(t, handlePushPretend(&PushParams{}, &task.State{}), errorNoProvision)
}

func makeAuxPushResponse(digest, version string) string {
	var status = PushStatus{
		Status: version,
		Aux: &PushStatusAux{
			Digest: digest,
		},
	}

	resultBytes, _ := json.Marshal(status)
	return string(resultBytes)
}

func makeEmptyIdPushResponse() string {
	var status = PushStatus{
		Status: "status",
	}

	resultBytes, _ := json.Marshal(status)
	return string(resultBytes)
}

func makeEmptyStatusPushResponse() string {
	var status = PushStatus{}

	resultBytes, _ := json.Marshal(status)
	return string(resultBytes)
}

func makeProgressDetailPushResponse() string {
	var status = PushStatus{
		Id:       "id",
		Progress: "progress",
		ProgressDetail: &PushStatusProgressDetail{
			Current: 0,
			Total:   0,
		},
		Status: "status",
	}

	resultBytes, _ := json.Marshal(status)
	return string(resultBytes)
}

func makeProgressEmptyPushResponse() string {
	var status = PushStatus{
		Id:       "id",
		Progress: "no progress",
		Status:   "status",
	}

	resultBytes, _ := json.Marshal(status)
	return string(resultBytes)
}
