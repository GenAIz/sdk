// Package ac provides commands for mapping Genaiz Broker Accounts.
// Broker Account commands include login, logout and policy for organization-wide configurations
//
// See: genaiz ac --help
package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func NewAc(ledger *config.Ledger) *cobra.Command {
	var ac = &cobra.Command{
		Use:     "account",
		Aliases: []string{"ac"},
		Short:   "GenAIz Account Toolkit",
	}

	ac.AddCommand(NewActivate(ledger))
	ac.AddCommand(NewInspect(ledger))
	ac.AddCommand(NewList(ledger))
	ac.AddCommand(NewLogin(ledger))
	ac.AddCommand(NewLogout(ledger))
	return ac
}
