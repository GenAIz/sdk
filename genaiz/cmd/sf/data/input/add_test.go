package input

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubAddExecutor struct {
	addError         error
	addHandle        string
	addType          string
	initError        error
	initPathOrHandle string
	initResult       string
	initType         string
}

func (sae *stubAddExecutor) Add(addType, addHandle string) error {
	sae.addType = addType
	sae.addHandle = addHandle
	return sae.addError
}

func (sae *stubAddExecutor) Init(initType, initPathOrHandle string) (string, error) {
	sae.initType = initType
	sae.initPathOrHandle = initPathOrHandle
	return sae.initResult, sae.initError
}

func TestNewAddInput(t *testing.T) {
	var expectedHandle = "handle"
	var testExecutor = &stubAddExecutor{
		initResult: expectedHandle,
	}
	var testCmd = NewAddInput(func(command *cobra.Command) AddExecutor {
		return testExecutor
	})

	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, "input", testExecutor.addType)
	assert.Equal(t, expectedHandle, testExecutor.addHandle)
}

func TestNewAddInput_AddError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHandle = "handle"
	var testExecutor = &stubAddExecutor{
		addError:   errors.New("expected"),
		initResult: expectedHandle,
	}
	var testCmd = NewAddInput(func(command *cobra.Command) AddExecutor {
		return testExecutor
	})

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, "input", testExecutor.initType)
	assert.Equal(t, expectedHandle, testExecutor.initPathOrHandle)
	assert.Equal(t, "input", testExecutor.addType)
	assert.Equal(t, expectedHandle, testExecutor.addHandle)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewAddInput_InitError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHandle = "handle"
	var testExecutor = &stubAddExecutor{
		initError: errors.New("expected"),
	}
	var testCmd = NewAddInput(func(command *cobra.Command) AddExecutor {
		return testExecutor
	})

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}
