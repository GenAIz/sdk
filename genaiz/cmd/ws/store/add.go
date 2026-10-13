package store

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string, string, string) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddStore(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE]|FLOW_ID [SF_HANDLE]|NODE_ID [DATA_STORE_NAME]|DATA_STORE_ID",
		Short:   "Adds a data store to a workspace flow node",
		Long:    "Adds a data store to a workspace flow node using workspace, workflow handle and node handle or their ids",
		Example: "genaiz ws node data str add my-workspace my-workflow my-node myLockerStr",
		Args:    cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)

			lang.HandleExit(exec.Add(args[0], args[1], args[2], args[3]))
		},
	}

	return addCmd
}
