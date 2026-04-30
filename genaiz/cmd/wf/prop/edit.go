package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type EditExecutor interface {
	Edit(string, string, string, string) error

	Init(string) (string, error)
}

type EditExecutorFactory func(command *cobra.Command) EditExecutor

func NewEditProp(factory EditExecutorFactory) *cobra.Command {
	var editCmd = &cobra.Command{
		Use:           "edit WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY VALUE",
		Short:         "Edits the property value of a Workflow Node",
		Long:          "Edits the property value of a Workflow Node, if it exists",
		Example:       "genaiz wf prop edit my-workflow my-node MY_PROP 0.2",
		Args:          cobra.ExactArgs(4),
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var handle string
			var err error

			if handle, err = executor.Init(args[1]); err == nil {
				err = executor.Edit(args[0], handle, args[2], args[3])
			}

			lang.HandleExit(err)
		},
	}

	return editCmd
}
