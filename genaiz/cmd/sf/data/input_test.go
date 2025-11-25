package data

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

type stubExecutor struct {
}

func (se *stubExecutor) Add(string, string) error {
	return nil
}

func (se *stubExecutor) Init(string, string) (string, error) {
	return "", nil
}

func (se *stubExecutor) Remove(string, string) error {
	return nil
}

func TestNewDataInput(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = NewDataInput(testLedger, []config.Definer{}, func(command *cobra.Command) Executor {
		return nil
	})

	assert.Empty(t, testCmd.Run)
	assert.Equal(t, 2, len(testCmd.Commands()))
}

func Test_newInputAddExecutorFactory(t *testing.T) {
	var expectedExecutor = &stubExecutor{}
	var factory = newInputAddExecutorFactory(func(command *cobra.Command) Executor {
		return expectedExecutor
	})

	assert.Same(t, expectedExecutor, factory(&cobra.Command{}))
}

func Test_newInputRemoveExecutorFactory(t *testing.T) {
	var expectedExecutor = &stubExecutor{}
	var factory = newInputRemoveExecutorFactory(func(command *cobra.Command) Executor {
		return expectedExecutor
	})

	assert.Same(t, expectedExecutor, factory(&cobra.Command{}))
}
