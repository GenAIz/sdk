package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(workflowArg string, nodeArg string, key string, value string) error

	Init(nodeArg string) (string, error)
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddProp(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:           "add WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH KEY VALUE",
		Short:         "Adds a property value to a Workflow Node",
		Long:          "Adds a property value to a Workflow Node, if it is not already present",
		Example:       "genaiz wf prop add my-workflow my-node MY_PROP 0.1",
		Args:          cobra.ExactArgs(4),
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var handle string
			var err error

			if handle, err = executor.Init(args[1]); err == nil {
				err = executor.Add(args[0], handle, args[2], args[3])
			}

			lang.HandleExit(err)
		},
	}

	return addCmd
}
