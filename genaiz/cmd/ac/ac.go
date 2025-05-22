// Package ac provides commands for mapping Genaiz Broker Accounts.
// Broker Account commands include login, logout and policy for organization-wide configurations
//
// See: genaiz ac --help
package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func NewAc(repo *config.Repo) *cobra.Command {
	var ac = &cobra.Command{
		Use:     "account",
		Aliases: []string{"ac"},
		Short:   "Genaiz Account Toolkit",
	}

	ac.AddCommand(
		NewLogin(repo),
		NewLogout(repo),
		NewPolicy(repo))
	return ac
}
