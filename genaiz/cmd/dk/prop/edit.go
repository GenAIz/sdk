package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type EditExecutor interface {
	Edit(string, string) error
}

type EditExecutorFactory func(command *cobra.Command) EditExecutor

func NewEditSpec(factory EditExecutorFactory) *cobra.Command {
	var editCmd = &cobra.Command{
		Use:     "edit [OEM/]HANDLE[:VERSION] KEY",
		Short:   "Edits a property spec of a Data Link",
		Long:    "Edits a property spec of a Data Link, if it exists",
		Example: "genaiz dk prop edit com.genaiz/datalink-1:1.0.0 MY_PROP_KEY --name='Example'",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Edit(args[0], args[1])

			lang.HandleExit(err)
		},
	}

	return editCmd
}
