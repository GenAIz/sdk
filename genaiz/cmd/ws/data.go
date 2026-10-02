package ws

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type WorkspaceNodeResolveTaskFactory func() *task.Task[broker.WorkspaceNodeResolveParams]

type DataInstanceExecutor struct {
	BaseExecutor
	*DataInstanceOptions

	accountParams          config.AccountParametric
	printerParams          cli.PrinterParametric
	nodeResolveTaskFactory WorkspaceNodeResolveTaskFactory
}

type DataInstanceOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (dio DataInstanceOptions) allDefiners() []config.Definer {
	return []config.Definer{
		dio.optionAccount,
		dio.optionJsonPrinter,
	}
}

func NewData(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var dataCmd = &cobra.Command{
		Use:     "data",
		Aliases: []string{"dt"},
		Short:   "Manages data sources, stores and sets for nodes under workspace flows",
	}

	dataCmd.AddCommand(NewDataSource(ledger, wsCli))
	dataCmd.AddCommand(NewDataStore(ledger, wsCli))
	return dataCmd
}
