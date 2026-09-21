package dt

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dt/store"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type StoreListExecutor struct {
	BaseListExecutor
	*StoreListOptions

	userDataStoreFacadeProvider func() mgmt.UserDataStoreFacade
}

func (sle StoreListExecutor) List(arg string) error {
	var listParams *broker.DataInstanceListParams
	var err error

	if listParams, err = sle.newDataInstanceListParams(arg); err == nil {
		var dataStoreList = sle.userDataStoreFacadeProvider()
		var printer = sle.printerParams.Printer()
		var stores []mgmt.UserLinkInstance

		stores, err = dataStoreList.WithParams(listParams).
			WithLogger(sle.ledger.Logger).
			Provider().
			Get()

		if err == nil {
			return printer.Print(stores)
		}

		return printer.Error(err)
	}

	return err
}

type StoreListOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (slo StoreListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		slo.optionAccount,
		slo.optionJsonPrinter,
	}
}

func NewStore(ledger *config.Ledger) *cobra.Command {
	var listOptions = NewStoreListOptions()
	var listCmd = store.NewListStore(newStoreListExecutorFactory(ledger, listOptions))
	var strCmd = &cobra.Command{
		Use:     "store",
		Aliases: []string{"str"},
		Short:   "Manages data stores for an account",
	}

	strCmd.AddCommand(listCmd)
	ledger.Register(listCmd, listOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(listCmd, ledger, listOptions.optionAccount)
	return strCmd
}

func NewStoreListExecutor(ledger *config.Ledger, options *StoreListOptions) *StoreListExecutor {
	return &StoreListExecutor{
		BaseListExecutor: BaseListExecutor{
			ledger:        ledger,
			accountParams: config.NewAccountParams(ledger, options.optionAccount),
			printerParams: cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		},
		StoreListOptions: options,

		userDataStoreFacadeProvider: mgmt.NewUserDataStoreFacade,
	}
}

func NewStoreListOptions() *StoreListOptions {
	return &StoreListOptions{
		optionAccount: cli.Options.Stores.Account().
			WithKeys(&schema.Genaiz.Data.Store.List.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Data.Store.List.Printer).
			BuildBoolOption(),
	}
}

func newStoreListExecutorFactory(ledger *config.Ledger, options *StoreListOptions) store.ListExecutorFactory {
	return func(command *cobra.Command) store.ListExecutor {
		return NewStoreListExecutor(ledger, options)
	}
}
