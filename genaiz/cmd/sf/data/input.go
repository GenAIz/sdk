package data

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/sf/data/input"
	"genaiz.com/genaiz/config"
)

func NewDataInput(ledger *config.Ledger, optionDefiners []config.Definer, factory ExecutorFactory) *cobra.Command {
	var inputAddFactory = newInputAddExecutorFactory(factory)
	var inputRmFactory = newInputRemoveExecutorFactory(factory)
	var inputAddCmd = input.NewAddInput(inputAddFactory)
	var inputRmCmd = input.NewRemoveInput(inputRmFactory)
	var inputCmd = &cobra.Command{
		Use:     "input",
		Aliases: []string{"in"},
		Short:   "Manages data input port configurations for Smart Functions",
	}

	inputCmd.AddCommand(inputAddCmd)
	inputCmd.AddCommand(inputRmCmd)
	ledger.Register(inputAddCmd, optionDefiners...)
	return inputCmd
}

func newInputAddExecutorFactory(factory ExecutorFactory) input.AddExecutorFactory {
	return func(cmd *cobra.Command) input.AddExecutor {
		return factory(cmd)
	}
}

func newInputRemoveExecutorFactory(factory ExecutorFactory) input.RemoveExecutorFactory {
	return func(cmd *cobra.Command) input.RemoveExecutor {
		return factory(cmd)
	}
}
