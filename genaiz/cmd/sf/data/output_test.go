package data

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewDataOutput(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testCmd = NewDataOutput(testLedger, []config.Definer{}, func(command *cobra.Command) Executor {
		return nil
	})

	assert.Empty(t, testCmd.Run)
	assert.Equal(t, 2, len(testCmd.Commands()))
}

func Test_newOutputAddExecutorFactory(t *testing.T) {
	var expectedExecutor = &stubExecutor{}
	var factory = newOutputAddExecutorFactory(func(command *cobra.Command) Executor {
		return expectedExecutor
	})

	assert.Same(t, expectedExecutor, factory(&cobra.Command{}))
}

func Test_newOutputRemoveExecutorFactory(t *testing.T) {
	var expectedExecutor = &stubExecutor{}
	var factory = newOutputRemoveExecutorFactory(func(command *cobra.Command) Executor {
		return expectedExecutor
	})

	assert.Same(t, expectedExecutor, factory(&cobra.Command{}))
}
