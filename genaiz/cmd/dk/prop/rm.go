package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, string) error
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveSpec(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm KEY",
		Short:   "Removes a property spec from a Data Link",
		Long:    "Removes a property spec from a Data Link, if it exists",
		Example: "genaiz dk prop rm com.genaiz/datalink-1:1.0.0 MY_PROP_KEY",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Remove(args[0], args[1])

			lang.HandleExit(err)
		},
	}

	return rmCmd
}
