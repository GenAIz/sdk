package cli

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task/broker"
)

var (
	AutoBridge *autoBridge
)

type AutoRegistering interface {
	Arguments(*cobra.Command, *config.Ledger)

	Option(*cobra.Command, *config.Ledger, *config.StringOption)
}

type autoBridge struct {
	Accounts func() AutoRegistering
}

type bridgeAccounts struct {
	facadeProvider func() mgmt.UserAccountFacade
}

func (ba bridgeAccounts) Arguments(cmd *cobra.Command, ledger *config.Ledger) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return ba.bridge(ledger, toComplete)
	}
}

func (ba bridgeAccounts) Complete(facade mgmt.UserAccountFacade, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if accounts, err := facade.Filtering(toComplete).Get(); err == nil {
		if len(accounts) > 0 {
			var results []cobra.Completion

			for _, a := range accounts {
				results = append(results, a.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (ba bridgeAccounts) Option(cmd *cobra.Command, ledger *config.Ledger, option *config.StringOption) {
	var err = cmd.RegisterFlagCompletionFunc(option.Param,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return ba.bridge(ledger, toComplete)
		})

	panicz.PanicIfError(err)
}

func (ba bridgeAccounts) bridge(ledger *config.Ledger, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var facade = ba.facadeProvider().
		WithLogger(ledger.Logger).
		WithParams(&broker.AuthParams{
			Broker: &broker.Broker{
				AuthFile: ledger.AuthFile,
			},
		})

	return ba.Complete(facade, toComplete)
}

func init() {
	AutoBridge = &autoBridge{
		Accounts: func() AutoRegistering {
			return &bridgeAccounts{
				facadeProvider: mgmt.NewUserAccountFacade,
			}
		},
	}
}
