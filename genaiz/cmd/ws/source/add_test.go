package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubAddExecutor struct {
	addError   error
	dataSource string
	flow       string
	node       string
	workspace  string
}

func (s *stubAddExecutor) Add(ws, fl, nd, ds string) error {
	s.workspace = ws
	s.flow = fl
	s.node = nd
	s.dataSource = ds
	return s.addError
}

func TestNewAddSource(t *testing.T) {
	var stubExecutor = &stubAddExecutor{}
	var testCmd = NewAddSource(func(command *cobra.Command) AddExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataSource",
	}

	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataSource)
}

func TestNewAddSource_AddError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var stubExecutor = &stubAddExecutor{
		addError: expectedError,
	}
	var testCmd = NewAddSource(func(command *cobra.Command) AddExecutor {
		return stubExecutor
	})
	var testArgs = []string{
		"workspace",
		"flow",
		"node",
		"dataSource",
	}

	defer patch.Unpatch()
	testCmd.SetArgs(testArgs)
	assert.NoError(t, testCmd.Execute())
	assert.Equal(t, testArgs[0], stubExecutor.workspace)
	assert.Equal(t, testArgs[1], stubExecutor.flow)
	assert.Equal(t, testArgs[2], stubExecutor.node)
	assert.Equal(t, testArgs[3], stubExecutor.dataSource)
	assert.True(t, patch.Called)
	assert.Equal(t, 1, patch.CalledWith)
}
