package auto

import (
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task/broker"
)

type WorkspaceBridge interface {
	Bridge

	WithReadyOnly(*config.BoolOption) WorkspaceBridge

	WithWorkspace(string) WorkspaceBridge
}

type WorkspaceFlowBridge interface {
	Bridge

	WithWorkspaceFlow(string, string) WorkspaceFlowBridge
}

type WorkspaceAutoBridge struct {
	ledger *config.Ledger

	workspaceFacadeProvider func() mgmt.UserWorkspacesFacade
}

func (wab WorkspaceAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var params = &broker.WorkspaceListParams{
		Broker: broker.Broker{
			AuthFile: wab.ledger.AuthFile,
		},
		RcEnabled: true,
	}
	var facade = wab.workspaceFacadeProvider().
		WithParams(params).
		WithLogger(wab.ledger.Logger).
		Filtering(toComplete)

	if workspaces, err := facade.Get(); err == nil {
		if len(workspaces) > 0 {
			var results []cobra.Completion

			for _, w := range workspaces {
				results = append(results, w.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func NewWorkspaceBridge(ledger *config.Ledger) *WorkspaceAutoBridge {
	return &WorkspaceAutoBridge{
		ledger: ledger,

		workspaceFacadeProvider: mgmt.NewUserWorkspacesFacade,
	}
}

type WorkspaceFlowAutoBridge struct {
	ledger          *config.Ledger
	readyOnlyOption *config.BoolOption
	workspace       string

	workspaceFlowFacadeProvider func() mgmt.UserWorkspaceFlowsFacade
}

func (fab WorkspaceFlowAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var resolveParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			Broker: &broker.Broker{
				AuthFile: fab.ledger.AuthFile,
			},
		},
	}
	var params = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: resolveParams,
	}
	var facade = fab.workspaceFlowFacadeProvider().
		WithParams(params).
		WithLogger(fab.ledger.Logger).
		Filtering(toComplete)

	if fab.readyOnlyOption != nil {
		params.ReadyOnly = fab.ledger.GetBool(fab.readyOnlyOption)
	}

	if id, err := strconv.Atoi(fab.workspace); err == nil {
		resolveParams.WorkspaceId = new(int64(id))
	} else {
		resolveParams.WorkspaceName = fab.workspace
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

func (fab WorkspaceFlowAutoBridge) WithReadyOnly(readyOnlyOption *config.BoolOption) WorkspaceBridge {
	return &WorkspaceFlowAutoBridge{
		ledger:                      fab.ledger,
		readyOnlyOption:             readyOnlyOption,
		workspace:                   fab.workspace,
		workspaceFlowFacadeProvider: fab.workspaceFlowFacadeProvider,
	}
}

func (fab WorkspaceFlowAutoBridge) WithWorkspace(workspace string) WorkspaceBridge {
	return &WorkspaceFlowAutoBridge{
		ledger:                      fab.ledger,
		readyOnlyOption:             fab.readyOnlyOption,
		workspace:                   workspace,
		workspaceFlowFacadeProvider: fab.workspaceFlowFacadeProvider,
	}
}

func NewWorkspaceFlowBridge(ledger *config.Ledger) *WorkspaceFlowAutoBridge {
	return &WorkspaceFlowAutoBridge{
		ledger: ledger,

		workspaceFlowFacadeProvider: mgmt.NewUserWorkspaceFlowsFacade,
	}
}

type WorkspaceNodeAutoBridge struct {
	ledger    *config.Ledger
	flow      string
	workspace string

	workspaceNodeFacadeProvider func() mgmt.UserWorkspaceNodesFacade
}

func (nab WorkspaceNodeAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var nodeParams = &broker.WorkspaceNodeListParams{
		Broker: &broker.Broker{
			AuthFile: nab.ledger.AuthFile,
		},
	}
	var facade = nab.workspaceNodeFacadeProvider().
		WithParams(nodeParams).
		WithLogger(nab.ledger.Logger).
		Filtering(toComplete)

	if id, err := strconv.ParseInt(nab.workspace, 10, 64); err == nil {
		nodeParams.WorkspaceId = new(id)
	} else {
		nodeParams.WorkspaceName = nab.workspace
	}

	if id, err := strconv.ParseInt(nab.flow, 10, 64); err == nil {
		nodeParams.WorkspaceFlowId = new(id)
	} else {
		nodeParams.WorkflowHandle = nab.flow
	}

	if nodes, err := facade.Get(); err == nil {
		if len(nodes) > 0 {
			var results []cobra.Completion

			for _, nd := range nodes {
				results = append(results, nd.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (nab WorkspaceNodeAutoBridge) WithWorkspaceFlow(workspace, flow string) WorkspaceFlowBridge {
	return &WorkspaceNodeAutoBridge{
		ledger:    nab.ledger,
		flow:      flow,
		workspace: workspace,

		workspaceNodeFacadeProvider: nab.workspaceNodeFacadeProvider,
	}
}

func NewWorkspaceNodeAutoBridge(ledger *config.Ledger) *WorkspaceNodeAutoBridge {
	return &WorkspaceNodeAutoBridge{
		ledger: ledger,

		workspaceNodeFacadeProvider: mgmt.NewUserWorkspaceNodesFacade,
	}
}
