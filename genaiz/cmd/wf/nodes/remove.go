package nodes

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Find(string) (string, error)

	Remove(string, ...string)
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

type RemoveValidator func(...string) error

func NewRemoveNodes(factory RemoveExecutorFactory, validator RemoveValidator) *cobra.Command {
	var andCmd = &cobra.Command{
		Use:     "remove WORKFLOW_HANDLE NODE_HANDLE|FUNCTION_PATH [NODE_HANDLE|FUNCTION_PATH...]]",
		Aliases: []string{"rm"},
		Short:   "Removes nodes from an existing workflow",
		Long:    "Removes nodes from an existing workflow under the current workdir",
		Example: "genaiz wf nd rm node-1 node-2...",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(2)),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var handles []string
			var err error

			for _, arg := range args[1:] {
				var handle string

				if handle, err = executor.Find(arg); err == nil {
					if err = validator(handle); err == nil {
						handles = append(handles, handle)
					}
				}

				if err != nil {
					break
				}
			}

			if err == nil {
				executor.Remove(args[0], handles...)
			} else {
				lang.HandleExit(err)
			}
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return andCmd
}
