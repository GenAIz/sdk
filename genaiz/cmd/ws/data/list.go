package data

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type ListAutoBridge struct {
	workspaces     auto.Bridge
	workspaceFlows auto.WorkspaceBridge
	workspaceNodes auto.WorkspaceFlowBridge
}

func (sab ListAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = sab.workspaces.Bridge(toComplete)
	} else if argsCount == 1 {
		results, directive = sab.workspaceFlows.WithWorkspace(args[0]).Bridge(toComplete)
	} else if argsCount == 2 {
		results, directive = sab.workspaceNodes.WithWorkspaceFlow(args[0], args[1]).Bridge(toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

type ListExecutor interface {
	List(string, string, string) error
}

type ListExecutorFactory func(command *cobra.Command) ListExecutor

type ListOptions struct {
	baseOptions
}

func (lo ListOptions) ListParams(brokerParams *broker.Broker,
	nodeResolveParams *broker.WorkspaceNodeResolveParams) *broker.WorkspaceNodeSetListParams {
	return &broker.WorkspaceNodeSetListParams{
		Broker:            *brokerParams,
		NodeResolveParams: &nodeResolveParams.NodeResolveParams,
	}
}

func NewDataList(ledger *config.Ledger, options *ListOptions, factory ListExecutorFactory) *cobra.Command {
	var setAutoBridge = NewDataListAutoBridge(ledger)
	var listCmd = &cobra.Command{
		Use:     "list [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE]|FLOW_ID [SF_HANDLE]|NODE_ID",
		Aliases: []string{"ls"},
		Short:   "Lists the data sets assigned to a workspace node",
		Long:    "Lists the data sets assigned to a workspace node, using workspace, flow handle and node handle or their ids",
		Example: "genaiz ws node data list my-workspace my-flow my-node",
		Args:    cobra.ExactArgs(3),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)

			lang.HandleExit(exec.List(args[0], args[1], args[2]))
		},
	}

	ledger.Register(listCmd, options.Definers()...)
	listCmd.ValidArgsFunction = setAutoBridge.bridgeArguments
	cli.AutoBridge.Accounts().Option(listCmd, ledger, options.optionAccount)
	return listCmd
}

func NewDataListAutoBridge(ledger *config.Ledger) *ListAutoBridge {
	return &ListAutoBridge{
		workspaces:     auto.NewWorkspaceBridge(ledger),
		workspaceFlows: auto.NewWorkspaceFlowBridge(ledger),
		workspaceNodes: auto.NewWorkspaceNodeAutoBridge(ledger),
	}
}

func NewDataListOptions() *ListOptions {
	return &ListOptions{
		baseOptions: baseOptions{
			optionAccount: cli.Options.Workspaces.Account().
				WithKeys(&schema.Genaiz.Workspace.Node.Data.List.Account).
				BuildStringOption(),
			optionJsonPrinter: cli.Options.Printer.JsonPrinter().
				WithKeys(&schema.Genaiz.Workspace.Node.Data.List.Printer).
				BuildBoolOption(),
		},
	}
}
