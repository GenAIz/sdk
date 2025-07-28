package wiremock

import (
	"os/user"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
)

func NewStop() *cobra.Command {
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops the Wiremock Wiremock service",
		Long:  "Stops the Wiremock Wiremock service if it is running",
		Run: func(cmd *cobra.Command, args []string) {
			var currentUser, _ = user.Current()
			var service = compose.NewWiremockService(cmd.Context(), currentUser)

			cobra.CheckErr(service.Stop())
		},
	}

	return stopCmd
}
