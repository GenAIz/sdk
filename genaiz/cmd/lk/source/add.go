package source

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddSource(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add HANDLE FQDN:VERSION[-rc-N]",
		Short:   "Adds a data source property map to a locker",
		Long:    "Adds a data source property map for the specified Datalink to a locker",
		Example: "genaiz lk src add myHandle com.genaiz.dev/my-datalink:1.0.0",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			lang.HandleExit(executor.Add(args[0], args[1]))
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return addCmd
}
