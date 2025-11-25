package output

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, string) error

	Init(string, string) (string, error)
}

type RemoveExecutorFactory func(*cobra.Command) RemoveExecutor

func NewRemoveOutput(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm PATH|HANDLE",
		Short:   "Removes an output data port from a Smart Function",
		Long:    "Removes an output data port from a Smart Function, if it exists",
		Example: "genaiz sf data output rm port",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var dataType = "output"
			var exec = factory(cmd)
			var handle string
			var err error

			if handle, err = exec.Init(dataType, args[0]); err == nil {
				err = exec.Remove(dataType, handle)
			}

			lang.HandleExit(err)
		},
	}

	return rmCmd
}
