package store

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddStore(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add [OEM/][HANDLE][:VERSION]",
		Short:   "Adds a data store to a Smart Function",
		Long:    "Adds a data store by full or partial qualified domain name and version to a Smart Function",
		Example: "genaiz sf data str add my-datastore:1.1.0 --oem=com.genaiz.dev",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var arg = cli.ArgsOptionalSingle(args)

			lang.HandleExit(exec.Add(arg))
		},
	}

	return addCmd
}
