package task

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
)

var (
	testLogger, _ = test.NewNullLogger()
)

func TestEnv_ErrorStream(t *testing.T) {
	var testEnv = &Env{}
	var testFile = filepath.Join(t.TempDir(), "testErr")

	assert.Equal(t, os.Stderr, testEnv.ErrorStream())

	if fd, err := os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		testEnv.Err = fd
		assert.Equal(t, fd, testEnv.ErrorStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestEnv_InputStream(t *testing.T) {
	var testEnv = &Env{}
	var testFile = filepath.Join(t.TempDir(), "testInput")

	assert.Equal(t, os.Stdin, testEnv.InputStream())

	if fd, err := os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		testEnv.In = fd
		assert.Equal(t, fd, testEnv.InputStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestEnv_OutputStream(t *testing.T) {
	var testEnv = &Env{}
	var testFile = filepath.Join(t.TempDir(), "testOutput")

	assert.Equal(t, os.Stdout, testEnv.OutputStream())

	if fd, err := os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		testEnv.Out = fd
		assert.Equal(t, fd, testEnv.OutputStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestState_Report(t *testing.T) {
	var testState = &State{}
	var expectedReport = "report"

	testState.Report(expectedReport)
	assert.Contains(t, testState.Reports, expectedReport)
}

func TestState_Reportf(t *testing.T) {
	var testState = &State{}
	var expectedReport = "report"
	var expectedParam = 37

	testState.Reportf("%s %d", expectedReport, expectedParam)
	assert.Contains(t, testState.Reports, expectedReport+" "+cast.ToString(expectedParam))
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
