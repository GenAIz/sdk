package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type EditExecutor interface {
	Edit(string) error
}

type EditExecutorFactory func(command *cobra.Command) EditExecutor

func NewEditSpec(factory EditExecutorFactory) *cobra.Command {
	var editCmd = &cobra.Command{
		Use:     "edit KEY",
		Short:   "Edits a property spec of a Smart Function",
		Long:    "Edits a property spec of a Smart Function, if it exists",
		Example: "genaiz sf prop edit MY_PROP",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Edit(args[0])

			lang.HandleExit(err)
		},
	}

	return editCmd
}
