package source

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, string, string, string) error
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveSource(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE]|FLOW_ID [NODE_HANDLE]|NODE_ID [DATA_SOURCE_NAME]|DATASOURCE_ID",
		Short:   "Removes a data source from a workspace flow node",
		Long:    "Removes a data source from a workspace flow node using workspace, workflow handle and node handle or their ids",
		Example: "genaiz ws node data src rm my-workspace my-workflow my-node myLockerSrc",
		Args:    cobra.ExactArgs(4),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)

			lang.HandleExit(exec.Remove(args[0], args[1], args[2], args[3]))
		},
	}

	return rmCmd
}
