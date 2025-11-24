package output

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string) error

	Init(string, string) (string, error)
}

type AddExecutorFactory func(*cobra.Command) AddExecutor

func NewAddOutput(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add PATH|HANDLE",
		Short:   "Adds an output data port to a Smart Function",
		Long:    "Adds an output data port to a Smart Function, if it is not already present",
		Example: "genaiz sf data output add folder/run/output/port --name='My Port' --description='An example output port",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var dataType = "output"
			var exec = factory(cmd)
			var handle string
			var err error

			if handle, err = exec.Init(dataType, args[0]); err == nil {
				err = exec.Add(dataType, handle)
			}

			lang.HandleExit(err)
		},
	}

	return addCmd
}
