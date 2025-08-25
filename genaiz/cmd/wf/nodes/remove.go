package nodes

import (
	"github.com/spf13/cobra"
)

type RemoveExecutor interface {
	Remove(string, ...string)
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

type RemoveValidator func(...string) error

func NewRemoveNodes(factory RemoveExecutorFactory, validator RemoveValidator) *cobra.Command {
	var andCmd = &cobra.Command{
		Use:     "remove NODE_HANDLE [NODE_HANDLE...]",
		Aliases: []string{"rm"},
		Short:   "Removes nodes from an existing workflow",
		Long:    "Removes nodes from an existing workflow under the current workdir",
		Example: "genaiz wf nd rm node-1 node-2...",
		Args: cobra.MatchAll(cobra.MinimumNArgs(2), func(cmd *cobra.Command, args []string) error {
			return validator(args[1:]...)
		}),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			executor.Remove(args[0], args[1:]...)
		},
	}

	return andCmd
}
