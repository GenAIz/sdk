package task

import (
	"bytes"
	"errors"
	"io"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/lang"
)

func TestPlan_SequenceContinue(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var actualCall = false
	var expectedError = errors.New("expected")
	var actualError error
	var testParam = "param"
	var testPlan = &Plan{
		Logger:            testLogger,
		ContinueOnFailure: true,
		OnFailure: func(i interface{}) {
			actualError = i.(error)
		}}
	var testPretender = NewPretender(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return errors.New("not expected")
		},
		OnPretend: func(params *string, state *State) error {
			actualCall = true
			return nil
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testPretender, NewPretender(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnPretend: func(params *string, state *State) error {
			state.Error = expectedError
			return expectedError
		},
	}))
	assert.True(t, actualCall)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
	assert.ErrorIs(t, actualError, expectedError)
}

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

func TestNewPlan_Failure(t *testing.T) {
	var output = new(bytes.Buffer)
	var expectedName = "Failed"
	var expectedFailure = "failure"
	var ioLogger = logrus.New()
	var testPlan = NewPlan(expectedName, ioLogger)

	ioLogger.Out = io.Writer(output)
	testPlan.OnFailure(expectedFailure)
	assert.Contains(t, output.String(), expectedFailure)
}

func TestNewPlan_Success(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedOutput = "output"
	var testPlan = NewPlan("test", nil)

	defer patch.Unpatch()
	testPlan.OnSuccess(expectedOutput)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedOutput)
}

func TestAttemptComplete(t *testing.T) {
	var expectedError = errors.New("test")
	var actualError error
	var testPlan = &Plan{Logger: testLogger, OnFailure: func(i interface{}) {
		actualError = i.(error)
	}}
	var testParam = "param"

	Attempt(testPlan, &testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return expectedError
		},
	})
	assert.ErrorIs(t, actualError, expectedError)
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
