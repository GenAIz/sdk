package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string) error
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveSpec(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm KEY",
		Short:   "Removes a property spec from a Smart Function",
		Long:    "Removes a property spec from a Smart Function, if it exists",
		Example: "genaiz sf prop rm MY_PROP",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Remove(args[0])

			lang.HandleExit(err)
		},
	}

	return rmCmd
}
