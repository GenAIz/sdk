package task

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-it/mock"
	"genaiz.com/genaiz/lang"
)

func TestPlan_SequenceError(t *testing.T) {
	var expectedError = errors.New("expected")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{Logger: testLogger}
	var testWorker = NewWorker(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return expectedError
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testWorker)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPlan_SequenceNoSuccess(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{Logger: testLogger}
	var testWorker = NewWorker(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testWorker)
	assert.False(t, patch.Called)
}

func TestPlan_SequenceOnFailure(t *testing.T) {
	var actualError interface{}
	var expectedError = errors.New("error")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{
		Logger: testLogger,
		OnFailure: func(i interface{}) {
			actualError = i
		},
	}
	var testWorker = NewWorker(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return errors.New("not expected error")
		},
		OnIncomplete: func(params *string, state *State) error {
			return expectedError
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testWorker)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
	assert.EqualValues(t, expectedError, actualError)

}

func TestPlan_SequenceOnSuccess(t *testing.T) {
	var actualOutput interface{}
	var expectedOutput = "output"
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{
		Logger: testLogger,
		OnSuccess: func(i interface{}) {
			actualOutput = i
		},
	}
	var testWorker = NewWorker(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			state.Output = expectedOutput
			return nil
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testWorker)
	assert.False(t, patch.Called)
	assert.EqualValues(t, expectedOutput, actualOutput)
}

func TestConditionalOnFailure(t *testing.T) {
	var actualError interface{}
	var expectedError = errors.New("error")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{
		Logger: testLogger,
		OnFailure: func(i interface{}) {
			actualError = i
		},
	}
	var testIfTask = lang.Supplier(&Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return expectedError
		},
	})
	var testElseTask = lang.Supplier(&Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return errors.New("not the right error")
		},
	})

	defer patch.Unpatch()
	Conditional(testPlan, true, &testParam, testIfTask, testElseTask)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
	assert.Same(t, expectedError, actualError)
}

func TestConditionalOnSuccess(t *testing.T) {
	var actualOutput interface{}
	var expectedOutput = "output"
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{
		Logger: testLogger,
		OnSuccess: func(i interface{}) {
			actualOutput = i
		},
	}
	var testIfTask = lang.Supplier(&Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
	})
	var testElseTask = lang.Supplier(&Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			state.Output = expectedOutput
			return nil
		},
	})

	defer patch.Unpatch()
	Conditional(testPlan, false, &testParam, testIfTask, testElseTask)
	assert.False(t, patch.Called)
	assert.EqualValues(t, expectedOutput, actualOutput)
}

func TestSingleComplete(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{Logger: testLogger}

	defer patch.Unpatch()
	Single(testPlan, &testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
	})
	assert.False(t, patch.Called)
}
