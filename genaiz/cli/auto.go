package cli

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/accounts"
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
	Accounts  func() AutoRegistering
	Solutions func() AutoRegistering
}

type bridgeAccounts struct {
	providerFactory accounts.UserAccountsFactory
}

func (ba bridgeAccounts) Arguments(cmd *cobra.Command, ledger *config.Ledger) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return ba.bridge(ledger, toComplete)
	}
}

func (ba bridgeAccounts) Complete(provider accounts.UserAccountsFactory, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var uaProvider = provider.WithFilter(toComplete).Provider()

	if userAccounts, err := uaProvider.Get(); err == nil {
		if len(userAccounts) > 0 {
			var results []cobra.Completion

			for _, a := range userAccounts {
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
	var factory = ba.providerFactory.
		WithLogger(ledger.Logger).
		WithFilter(toComplete).
		WithParams(&broker.AccountListParams{
			Broker: broker.Broker{
				AuthFile: ledger.AuthFile,
			},
		})

	return ba.Complete(factory, toComplete)
}

type bridgeSolutions struct {
	facadeProvider func() mgmt.UserSolutionFacade
}

func (bs bridgeSolutions) Arguments(cmd *cobra.Command, ledger *config.Ledger) {
	cmd.ValidArgsFunction = func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
		return bs.bridge(ledger, toComplete)
	}
}

func (bs bridgeSolutions) Complete(facade mgmt.UserSolutionFacade, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	if solutions, err := facade.Filtering(toComplete).Get(); err == nil {
		if len(solutions) > 0 {
			var results []cobra.Completion

			for _, s := range solutions {
				results = append(results, s.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (bs bridgeSolutions) Option(cmd *cobra.Command, ledger *config.Ledger, option *config.StringOption) {
	var err = cmd.RegisterFlagCompletionFunc(option.Param,
		func(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
			return bs.bridge(ledger, toComplete)
		})

	panicz.PanicIfError(err)
}

func (bs bridgeSolutions) bridge(ledger *config.Ledger, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var oem, _, _ = broker.ParseFqdnVersion(toComplete)
	var facade = bs.facadeProvider().
		WithLogger(ledger.Logger).
		WithParams(&broker.SolutionListParams{
			Broker: broker.Broker{
				AuthFile: ledger.AuthFile,
			},
			Oem: oem,
		})

	return bs.Complete(facade, toComplete)
}

func init() {
	AutoBridge = &autoBridge{
		Accounts: func() AutoRegistering {
			return &bridgeAccounts{
				providerFactory: accounts.NewUserAccountsFactory(),
			}
		},
		Solutions: func() AutoRegistering {
			return &bridgeSolutions{
				facadeProvider: mgmt.NewUserSolutionFacade,
			}
		},
	}
}
