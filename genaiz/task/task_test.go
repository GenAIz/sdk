package task

import (
	"errors"
	"os"
	"testing"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/undefinedlabs/go-mpatch"
)

type PatchedExit struct {
	called     bool
	calledWith int
	patchFunc  *mpatch.Patch
}

func (p *PatchedExit) unpatch() {
	_ = p.patchFunc.Unpatch()
}

func newPatchedExit(t *testing.T, impl func(int)) *PatchedExit {
	patchedExit := &PatchedExit{called: false}

	patchFunc, err := mpatch.PatchMethod(os.Exit, func(code int) {
		patchedExit.called = true
		patchedExit.calledWith = code
		impl(code)
	})

	if err != nil {
		t.Errorf("Failed to patch os.Exit due to an error: %v", err)
		return nil
	}

	patchedExit.patchFunc = patchFunc
	return patchedExit
}

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
	var mockExit = func(int) {}
	var patch = newPatchedExit(t, mockExit)
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnPretend: func(params *string, state *State) error {
			return nil
		},
	}

	defer patch.unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.False(t, patch.called)
}

func TestTask_PretendOnPrepareFailure(t *testing.T) {
	var mockExit = func(int) {}
	var patch = newPatchedExit(t, mockExit)
	var expectedError = errors.New("expected")
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return expectedError
		},
	}

	defer patch.unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.True(t, patch.called)
	assert.EqualValues(t, 1, patch.calledWith)
}

func TestTask_PretendOnPretendFailure(t *testing.T) {
	var mockExit = func(int) {}
	var patch = newPatchedExit(t, mockExit)
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

	defer patch.unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.True(t, patch.called)
	assert.EqualValues(t, 1, patch.calledWith)
}

func TestTask_PretendOnPretendNil(t *testing.T) {
	var mockExit = func(int) {}
	var patch = newPatchedExit(t, mockExit)
	var testParam = "test"
	var testTask = &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
	}

	defer patch.unpatch()
	testTask.Pretend(&testParam, testLogger)
	assert.False(t, patch.called)
}
