package nodes

import (
	"github.com/spf13/cobra"
)

type AddExecutor interface {
	Add(string, string)
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

type AddValidator func(...string) error

func NewAddNodes(factory AddExecutorFactory, validator AddValidator) *cobra.Command {
	var andCmd = &cobra.Command{
		Use:     "add WORKFLOW NODE_HANDLE",
		Short:   "Adds nodes to an existing workflow",
		Long:    "Adds nodes to an existing workflow, creating a smart function if necessary",
		Example: "genaiz wf nd add workflow-1 node_handle --name=\"Node One\" --sf.handle=smart-function-1 --sf.version=0.0.1",
		Args: cobra.MatchAll(cobra.ExactArgs(2), func(cmd *cobra.Command, args []string) error {
			return validator(args[1])
		}),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			executor.Add(args[0], args[1])
		},
	}

	return andCmd
}
