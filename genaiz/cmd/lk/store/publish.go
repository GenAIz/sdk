package store

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type PublishExecutor interface {
	Publish(string) error
}

type PublishExecutorFactory func(command *cobra.Command) PublishExecutor

func NewPublishStore(factory PublishExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "publish HANDLE",
		Short:   "Publishes a data store with its property map to an account",
		Long:    "Publishes a data store with its property map or creates a new data store on an account",
		Example: "genaiz lk str publish myHandle --name='My Name'",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var executor = factory(cmd)

			lang.HandleExit(executor.Publish(args[0]))
		},
		SilenceErrors: true,
		SilenceUsage:  true,
	}

	return addCmd
}
