package node

import (
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type ListExecutor interface {
	List(string, string) error
}

type ListExecutorFactory func(command *cobra.Command) ListExecutor

type ListOptions struct {
	OptionAccount     *config.StringOption
	OptionReadyOnly   *config.BoolOption
	OptionJsonPrinter *config.BoolOption
}

func (lo ListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.OptionAccount,
		lo.OptionReadyOnly,
		lo.OptionJsonPrinter,
	}
}

type ListAutoBridge struct {
	ledger      *config.Ledger
	readyOption *config.BoolOption
	workspaces  auto.Bridge

	workspaceFlowFacadeProvider func() mgmt.UserWorkspaceFlowsFacade
}

func (lab ListAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = lab.workspaces.Bridge(toComplete)
	} else if argsCount == 1 {
		// In the case where the toComplete string refers to a workflowId or workflowHandle, we still use workspace flows
		results, directive = lab.bridgeFlows(args[0], toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

func (lab ListAutoBridge) bridgeFlows(workspaceArg string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var resolveParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			Broker: &broker.Broker{
				AuthFile: lab.ledger.AuthFile,
			},
		},
	}
	var params = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: resolveParams,
		ReadyOnly:                  lab.ledger.GetBool(lab.readyOption),
	}
	var facade = lab.workspaceFlowFacadeProvider().
		WithParams(params).
		WithLogger(lab.ledger.Logger).
		Filtering(toComplete)

	if id, err := strconv.Atoi(workspaceArg); err == nil {
		resolveParams.WorkspaceId = new(int64(id))
	} else {
		resolveParams.WorkspaceName = workspaceArg
	}

	if flows, err := facade.Get(); err == nil {
		if len(flows) > 0 {
			var results []cobra.Completion

			for _, fl := range flows {
				results = append(results, fl.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func NewList(ledger *config.Ledger, options *ListOptions, factory ListExecutorFactory) *cobra.Command {
	var listAuto = NewListAuto(ledger, options.OptionReadyOnly)
	var listCmd = &cobra.Command{
		Use:     "list [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE|FLOW_ID",
		Short:   "Lists workspace flow nodes",
		Long:    "Lists workspace flow nodes under the specified flow",
		Example: "genaiz ws List my-workspace my-workflow --json",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var err = exec.List(args[0], args[1])

			lang.HandleExit(err)
		},
	}

	ledger.Register(listCmd, options.allDefiners()...)
	listCmd.ValidArgsFunction = listAuto.bridgeArguments
	return listCmd
}

func NewListAuto(ledger *config.Ledger, readyOption *config.BoolOption) *ListAutoBridge {
	return &ListAutoBridge{
		readyOption: readyOption,
		ledger:      ledger,
		workspaces:  auto.NewWorkspaceBridge(ledger),

		workspaceFlowFacadeProvider: mgmt.NewUserWorkspaceFlowsFacade,
	}
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		OptionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Node.List.Account).
			BuildStringOption(),
		OptionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Node.List.Printer).
			BuildBoolOption(),
		OptionReadyOnly: cli.Options.Workspaces.FlowReadyOnly().
			WithKeys(&schema.Genaiz.Workspace.Node.List.ReadyOnly).
			BuildBoolOption(),
	}
}
