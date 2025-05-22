package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func NewLogout(repo *config.Repo) *cobra.Command {
	return &cobra.Command{
		Use:     "logout",
		Short:   "Removes any previously acquired session",
		Long:    "Removes any previously acquired session tokens held for the current user",
		Example: "genaiz ac logout",
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
}
