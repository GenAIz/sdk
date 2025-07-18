package registry

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
)

func NewStop() *cobra.Command {
	var stopCmd = &cobra.Command{
		Use:   "stop",
		Short: "Stops the CNCF Distribution registry",
		Long:  "Stops the CNCF Distribution registry if it is running",
		Run: func(cmd *cobra.Command, args []string) {
			var service = compose.NewRegistryService(cmd.Context())

			cobra.CheckErr(service.Stop())
		},
	}

	return stopCmd
}
