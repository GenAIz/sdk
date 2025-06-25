package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func NewPolicy(ledger *config.Ledger) *cobra.Command {
	return &cobra.Command{
		Use:     "policy",
		Short:   "Displays the list of configuration policies",
		Long:    "Displays the list of configuration policies for the current user with the current broker",
		Example: "genaiz ac logout",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}
