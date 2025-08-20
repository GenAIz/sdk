package links

import (
	"github.com/spf13/cobra"
)

type AddExecutor interface {
	Add(string, []string)
}

type AddExecutorFactory func(*cobra.Command) AddExecutor

type AddValidator func([]string) error

func NewAddLinks(factory AddExecutorFactory, validator AddValidator) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add WORKFLOW_HANDLE NODE_HANDLE_LEFT[PORT_LEFT]:NODE_HANDLE_RIGHT[PORT_RIGHT] [NODE_HANDLE...]",
		Short:   "Adds links to an existing workflow",
		Long:    "Adds links to an existing workflow, unknown links are added as DataSet handles",
		Example: "genaiz wf ln add workflow-1 smart-function-1:smart-function-2[port2]",
		Args: cobra.MatchAll(cobra.MinimumNArgs(2), func(cmd *cobra.Command, args []string) error {
			return validator(args[1:])
		}),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			executor.Add(args[0], args[1:])
		},
	}

	return addCmd
}
