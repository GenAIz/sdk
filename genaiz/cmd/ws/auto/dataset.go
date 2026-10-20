package auto

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/ws/args"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt/workspaces"
	"genaiz.com/genaiz/task/broker"
)

type DataSetAutoBridge interface {
	Bridge
	WithWorkspaceNode(string, string, string) DataSetAutoBridge
}

type dataSetAutoBridge struct {
	ledger    *config.Ledger
	flow      string
	node      string
	workspace string

	workspaceNodeDataSetsFactory workspaces.UserWorkspaceDataSetsFactory
}

func (dab dataSetAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var nodeArguments = args.NewNodeResolution(dab.workspace, dab.flow, dab.node)
	var resolveNodeParams = nodeArguments.ResolveNodeParams(&broker.Broker{
		AuthFile: dab.ledger.AuthFile,
	})
	var facade = dab.workspaceNodeDataSetsFactory.
		WithLogger(dab.ledger.Logger).
		WithParams(resolveNodeParams).
		WithFilter(toComplete).
		Provider()

	if dataSets, err := facade.Get(); err == nil {
		if len(dataSets) > 0 {
			var results []cobra.Completion

			for _, ds := range dataSets {
				results = append(results, ds.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (dab dataSetAutoBridge) WithWorkspaceNode(workspace, flow, node string) DataSetAutoBridge {
	return &dataSetAutoBridge{
		ledger:    dab.ledger,
		flow:      flow,
		node:      node,
		workspace: workspace,

		workspaceNodeDataSetsFactory: dab.workspaceNodeDataSetsFactory,
	}
}

func NewDataSetAutoBridge(ledger *config.Ledger) DataSetAutoBridge {
	return &dataSetAutoBridge{
		ledger: ledger,

		workspaceNodeDataSetsFactory: workspaces.NewUserWorkspaceDataSetsFactory(),
	}
}
