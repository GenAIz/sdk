package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

type AddValidator func(string) error

func NewAddSpec(factory AddExecutorFactory, validator AddValidator) *cobra.Command {
	var andCmd = &cobra.Command{
		Use:     "add KEY",
		Short:   "Adds a property spec to a Smart Function",
		Long:    "Adds a property spec to a Smart Function, if it is not already present",
		Example: "genaiz sf prop add MY_PROP --name=NAME --type=double --default-value=0.1",
		Args: cobra.MatchAll(cobra.ExactArgs(1), func(cmd *cobra.Command, args []string) error {
			return validator(args[0])
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Add(args[0])

			lang.HandleExit(err)
		},
	}

	return andCmd
}
