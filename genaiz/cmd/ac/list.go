package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type ListExecutor struct {
	*ListOption
	ledger *config.Ledger

	cliPrinterProvider        func() cli.Printer
	userAccountFacadeProvider func() mgmt.UserAccountFacade
}

func (le ListExecutor) List(filter string) error {
	var accountList = le.userAccountFacadeProvider()
	var printer = le.getPrinter()
	var accounts, err = accountList.WithParams(le.newAuthParams()).
		WithLogger(le.ledger.Logger).
		Filtering(filter).
		Get()

	if err == nil {
		return printer.Print(accounts)
	}

	return printer.Error(err)
}

func (le ListExecutor) getPrinter() cli.Printer {
	if le.cliPrinterProvider == nil {
		return cli.StdPrinter.JsonOrConsole(le.ledger, le.optionJsonPrinter)
	}

	return le.cliPrinterProvider()
}

func (le ListExecutor) newAuthParams() *broker.AuthParams {
	return &broker.AuthParams{
		Broker: &broker.Broker{
			AuthFile: le.ledger.AuthFile,
		},
		// Always show the expired accounts on list
		Expired: true,
	}
}

type ListOption struct {
	optionJsonPrinter *config.BoolOption
}

func (lo ListOption) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionJsonPrinter,
	}
}

func NewList(ledger *config.Ledger) *cobra.Command {
	var exec = NewListExecutor(ledger)
	var login = &cobra.Command{
		Use:     "list [USER_STRING][@][HOST_STRING]",
		Aliases: []string{"ls"},
		Short:   "Lists currently known account session, filtering by username and host",
		Long:    "Lists currently known account session, filtering by username and host, providing expiry, hostname and user",
		Example: "genaiz account list",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var filter = cli.ArgsOptionalSingle(args)
			var err = exec.List(filter)

			lang.HandleExit(err)
		},
	}

	ledger.Register(login, exec.allDefiners()...)
	return login
}

func NewListExecutor(ledger *config.Ledger) *ListExecutor {
	return &ListExecutor{
		ListOption:                NewListOptions(),
		ledger:                    ledger,
		userAccountFacadeProvider: mgmt.NewUserAccountFacade,
	}
}

func NewListOptions() *ListOption {
	return &ListOption{
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Account.List.Printer).
			BuildBoolOption(),
	}
}
