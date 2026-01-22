package prop

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

type AddValidator func(string) error

func NewAddSpec(factory AddExecutorFactory, validator AddValidator) *cobra.Command {
	var andCmd = &cobra.Command{
		Use:     "add [OEM/]HANDLE[:VERSION] KEY",
		Short:   "Adds a property spec to a Data Link",
		Long:    "Adds a property spec to a Data Link, if it is not already present",
		Example: "genaiz dk prop add com.genaiz/datalink-1:1.0.0 MY_PROP_KEY --name=NAME --type=double --default-value=0.1",
		Args: cobra.MatchAll(cobra.ExactArgs(2), func(cmd *cobra.Command, args []string) error {
			return validator(args[1])
		}),
		SilenceErrors: true,
		SilenceUsage:  true,
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)
			var err = executor.Add(args[0], args[1])

			lang.HandleExit(err)
		},
	}

	return andCmd
}
