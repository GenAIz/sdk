package store

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type ListExecutor interface {
	List(string) error
}

type ListExecutorFactory func(command *cobra.Command) ListExecutor

func NewListStore(factory ListExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "list [OEM[/HANDLE][:VERSION][-rc-N]]",
		Short:   "Lists data stores available to an account",
		Long:    "Lists data stores available to an account matching the optional argument filter on DataLink coordinates",
		Example: "genaiz dt str list com.genaiz.dev/my-datalink:1.0.0",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var arg = cli.ArgsOptionalSingle(args)

			lang.HandleExit(executor.List(arg))
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return addCmd
}
