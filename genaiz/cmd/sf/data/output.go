package data

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/sf/data/output"
	"genaiz.com/genaiz/config"
)

func NewDataOutput(ledger *config.Ledger, optionDefiners []config.Definer, factory ExecutorFactory) *cobra.Command {
	var outputAddFactory = newOutputAddExecutorFactory(factory)
	var outputRmFactory = newOutputRemoveExecutorFactory(factory)
	var outputAddCmd = output.NewAddOutput(outputAddFactory)
	var outputRmCmd = output.NewRemoveOutput(outputRmFactory)
	var outputCmd = &cobra.Command{
		Use:     "output",
		Aliases: []string{"out"},
		Short:   "Manages data output port configurations for Smart Functions",
	}

	outputCmd.AddCommand(outputAddCmd)
	outputCmd.AddCommand(outputRmCmd)
	ledger.Register(outputAddCmd, optionDefiners...)
	return outputCmd
}

func newOutputAddExecutorFactory(factory ExecutorFactory) output.AddExecutorFactory {
	return func(cmd *cobra.Command) output.AddExecutor {
		return factory(cmd)
	}
}

func newOutputRemoveExecutorFactory(factory ExecutorFactory) output.RemoveExecutorFactory {
	return func(cmd *cobra.Command) output.RemoveExecutor {
		return factory(cmd)
	}
}
