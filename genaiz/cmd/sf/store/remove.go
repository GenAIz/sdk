package store

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
)

type RemoveExecutor interface {
	Remove(string)
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveStore(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm [OEM/][HANDLE][:VERSION]",
		Short:   "Removes a data store from a Smart Function",
		Long:    "Removes a data store by full or partial qualified domain name and version from a Smart Function",
		Example: "genaiz sf data str rm my-datastore:1.1.0 --oem=com.genaiz.dev",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var arg = cli.ArgsOptionalSingle(args)

			exec.Remove(arg)
		},
	}

	return rmCmd
}
