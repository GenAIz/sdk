package dt

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dt/source"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type SourceListExecutor struct {
	BaseListExecutor
	*SourceListOptions
	userDataSourceFacadeProvider func() mgmt.UserDataSourceFacade
}

func (sle SourceListExecutor) List(arg string) error {
	var listParams *broker.DataInstanceListParams
	var err error

	if listParams, err = sle.newDataInstanceListParams(arg); err == nil {
		var dataSourceList = sle.userDataSourceFacadeProvider()
		var printer = sle.printerParams.Printer()
		var sources []mgmt.UserLinkInstance

		sources, err = dataSourceList.WithParams(listParams).
			WithLogger(sle.ledger.Logger).
			Provider().
			Get()

		if err == nil {
			return printer.Print(sources)
		}

		return printer.Error(err)
	}

	return err
}

type SourceListOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (slo SourceListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		slo.optionAccount,
		slo.optionJsonPrinter,
	}
}

func NewSource(ledger *config.Ledger) *cobra.Command {
	var listOptions = NewSourceListOptions()
	var listCmd = source.NewListSource(newSourceListExecutorFactory(ledger, listOptions))
	var srcCmd = &cobra.Command{
		Use:     "source",
		Aliases: []string{"src"},
		Short:   "Manages data sources for an account",
	}

	srcCmd.AddCommand(listCmd)
	ledger.Register(listCmd, listOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(listCmd, ledger, listOptions.optionAccount)
	return srcCmd
}

func NewSourceListExecutor(ledger *config.Ledger, options *SourceListOptions) *SourceListExecutor {
	return &SourceListExecutor{
		BaseListExecutor: BaseListExecutor{
			ledger:        ledger,
			accountParams: config.NewAccountParams(ledger, options.optionAccount),
			printerParams: cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		},
		SourceListOptions: options,

		userDataSourceFacadeProvider: mgmt.NewUserDataSourceFacade,
	}
}

func NewSourceListOptions() *SourceListOptions {
	return &SourceListOptions{
		optionAccount: cli.Options.Sources.Account().
			WithKeys(&schema.Genaiz.Data.Source.List.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Data.Source.List.Printer).
			BuildBoolOption(),
	}
}

func newSourceListExecutorFactory(ledger *config.Ledger, options *SourceListOptions) source.ListExecutorFactory {
	return func(command *cobra.Command) source.ListExecutor {
		return NewSourceListExecutor(ledger, options)
	}
}
