package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, string, string) error

	Init(string) (string, error)
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveProp(factory RemoveExecutorFactory) *cobra.Command {
	var editCmd = &cobra.Command{
		Use:           "rm WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY",
		Short:         "Removes a property by key from a Workflow Node",
		Long:          "Removes a property by key from a Workflow Node, if it exists",
		Example:       "genaiz wf prop rm my-workflow my-node MY_PROP",
		Args:          cobra.ExactArgs(3),
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var handle string
			var err error

			if handle, err = executor.Init(args[1]); err == nil {
				err = executor.Remove(args[0], handle, args[2])
			}

			lang.HandleExit(err)
		},
	}

	return editCmd
}
