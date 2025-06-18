package mapping

import (
	"github.com/spf13/cobra"
)

func NewReset(wiremockUrl *string) *cobra.Command {
	var reset = &cobra.Command{
		Use:     "reset",
		Short:   "Invokes wiremock mapping reset",
		Long:    "Invokes wiremock mapping reset on wiremock/__admin/mappings/reset",
		Example: "genaiz-it wiremock mapping reset",
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewMappingExecutor(*wiremockUrl)

			cobra.CheckErr(exec.Reset())
		},
	}

	return reset
}
