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

func TestPlan_Sequence_Continue(t *testing.T) {
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

func TestPlan_Sequence_Error(t *testing.T) {
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

func TestPlan_Sequence_NoSuccess(t *testing.T) {
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

func TestPlan_Sequence_OnFailure(t *testing.T) {
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

func TestPlan_Sequence_OnSuccess(t *testing.T) {
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

func TestPlan_Sequence_OnPrintReportsOnly(t *testing.T) {
	var actualOutput interface{}
	var expectedOutput = "output"
	var expectedReport = "report"
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testParam = "param"
	var testPlan = &Plan{
		Logger: testLogger,
		OnSuccess: func(i interface{}) {
			actualOutput = i
		},
		PrintReportsOnly: true,
	}
	var testWorker = NewWorker(&testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			state.Output = expectedOutput
			state.Report(expectedReport)
			return nil
		},
	})

	defer patch.Unpatch()
	testPlan.Sequence(testWorker)
	assert.False(t, patch.Called)
	assert.NotEqualValues(t, expectedOutput, actualOutput)
	assert.Equal(t, 1, len(actualOutput.([]string)))
	assert.Equal(t, expectedReport, actualOutput.([]string)[0])
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

func TestAttempt_Complete(t *testing.T) {
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

func TestConditional_OnFailure(t *testing.T) {
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

func TestConditional_OnSuccess(t *testing.T) {
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

func TestHandleFlag(t *testing.T) {
	var falseFlag, trueFlag bool

	HandleFlag(&falseFlag, false)("test")
	assert.False(t, falseFlag)
	HandleFlag(&trueFlag, true)("test")
	assert.True(t, trueFlag)
}

func TestHandleString(t *testing.T) {
	var expectedValue = "value"
	var stringValue string

	HandleString(&stringValue)(expectedValue)
	assert.Equal(t, expectedValue, stringValue)
}

func TestSingle_Complete(t *testing.T) {
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

func TestSingle_OnPrintReportsOnly(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedReport = "report"
	var testParam = "param"
	var actual interface{}
	var testPlan = &Plan{
		Logger: testLogger,
		OnSuccess: func(i interface{}) {
			actual = i
		},
		PrintReportsOnly: true,
	}

	defer patch.Unpatch()
	Single(testPlan, &testParam, &Task[string]{
		OnPrepare: func(params *string, state *State) error {
			return nil
		},
		OnComplete: func(params *string, state *State) error {
			state.Report(expectedReport)
			return nil
		},
	})
	assert.Equal(t, 1, len(actual.([]string)))
	assert.Equal(t, expectedReport, actual.([]string)[0])
	assert.False(t, patch.Called)
}

func Test_successWriter_Slice(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedOutput = "output"

	defer patch.Unpatch()
	successWriter([]string{expectedOutput})
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedOutput)
}
