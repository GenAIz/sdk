package nodes

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string) error

	Init(string) (string, error)
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

type AddValidator func(...string) error

func NewAddNodes(factory AddExecutorFactory, validator AddValidator) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH",
		Short:   "Adds nodes to an existing workflow",
		Long:    "Adds nodes to an existing workflow, creating a smart function if necessary",
		Example: "genaiz wf nd add workflow-1 node_handle --name=\"Node One\" --sf.handle=smart-function-1 --sf.version=0.0.1",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var handle string
			var err error

			if handle, err = executor.Init(args[1]); err == nil {
				if err = validator(handle); err == nil {
					err = executor.Add(args[0], handle)
				}
			}

			lang.HandleExit(err)
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return addCmd
}
