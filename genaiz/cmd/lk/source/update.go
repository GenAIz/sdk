package source

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type UpdateExecutor interface {
	Update(string, string, string) error
}

type UpdateExecutorFactory func(command *cobra.Command) UpdateExecutor

func NewUpdateSource(factory UpdateExecutorFactory) *cobra.Command {
	var updateCmd = &cobra.Command{
		Use:     "update HANDLE KEY [VALUE]",
		Short:   "Updates a data source property map from a locker",
		Long:    "Updates a data source property map from a locker, expecting secret strings on STDIN",
		Example: "genaiz lk src update myHandle KEY 'a value'",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(2), cobra.MaximumNArgs(3)),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var valueArg string

			if len(args) == 3 {
				valueArg = args[2]
			}

			lang.HandleExit(executor.Update(args[0], args[1], valueArg))
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return updateCmd
}
