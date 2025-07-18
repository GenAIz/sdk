package task

import (
	"errors"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

var (
	testLogger = &logrus.Logger{}
)

func TestState_GetContainersSize(t *testing.T) {
	var testState = &State{Containers: &[]container.Summary{}}

	assert.EqualValues(t, 0, testState.GetContainersSize())
}

func TestState_GetContainersSizeUninitialized(t *testing.T) {
	var testState = &State{Containers: nil}

	assert.EqualValues(t, 0, testState.GetContainersSize())
}

func TestState_HasContainers(t *testing.T) {
	var testState = &State{Containers: &[]container.Summary{
		{
			ID: "Test",
		},
	}}

	assert.True(t, testState.HasContainers())
}

func TestTask_ExecuteOnPrepareFailure(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return expectedError
		},
	}
	var testState = testTask.Execute(&testParam, testLogger)

	assert.True(t, testState.Completed)
	assert.EqualValues(t, expectedError, testState.Error)
}

func TestTask_ExecuteOnPrepareIncompleteFailure(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return errors.New("notExpected")
		},
		OnIncomplete: func(params *string, state *State) error {
			return expectedError
		},
	}
	var testState = testTask.Execute(&testParam, testLogger)

	assert.True(t, testState.Completed)
	assert.EqualValues(t, expectedError, testState.Error)
}

func TestTask_ExecuteOnPrepareCompleteFailure(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return errors.New("notExpected")
		},
		OnIncomplete: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			return expectedError
		},
	}
	var testState = testTask.Execute(&testParam, testLogger)

	assert.False(t, testState.Completed)
	assert.EqualValues(t, expectedError, testState.Error)
}

func TestTask_ExecuteOnPrepareComplete(t *testing.T) {
	var expectedOutput = "output"
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			state.Output = expectedOutput
			return nil
		},
	}
	var testState = testTask.Execute(&testParam, testLogger)

	assert.True(t, testState.Completed)
	assert.EqualValues(t, expectedOutput, testState.Output)
}

func TestTask_PretendOnPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnPretend: func(params *string, state *State) error {
			return nil
		},
	}

	defer patch.Unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.False(t, patch.Called)
}

func TestTask_PretendOnPretendFailure(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnPretend: func(params *string, state *State) error {
			return expectedError
		},
	}

	defer patch.Unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestTask_PretendOnPretendNil(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
	}

	defer patch.Unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.False(t, patch.Called)
}
