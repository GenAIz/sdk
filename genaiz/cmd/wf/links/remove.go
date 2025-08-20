package links

import (
	"github.com/spf13/cobra"
)

type RemoveExecutor interface {
	Remove(string, []string)
}

type RemoveExecutorFactory func(*cobra.Command) RemoveExecutor

type RemoveValidator func([]string) error

func NewRemoveLinks(factory RemoveExecutorFactory, validator RemoveValidator) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "remove WORKFLOW_HANDLE NODE_HANDLE_LEFT[PORT_LEFT]:NODE_HANDLE_RIGHT[PORT_RIGHT] [NODE_HANDLE...]",
		Aliases: []string{"rm"},
		Short:   "Removes links from an existing workflow",
		Long:    "Removes links from an existing workflow under the current workdir",
		Example: "genaiz wf ln rm workflow-1 smart-function-1[port1]:smart-function-2",
		Args: cobra.MatchAll(cobra.MinimumNArgs(2), func(cmd *cobra.Command, args []string) error {
			return validator(args[1:])
		}),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			executor.Remove(args[0], args[1:])
		},
	}

	return rmCmd
}
