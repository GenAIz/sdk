package auto

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task/broker"
)

type Bridge interface {
	Bridge(string) ([]cobra.Completion, cobra.ShellCompDirective)
}

type WorkspaceAutoBridge struct {
	Ledger *config.Ledger

	workspaceFacadeProvider func() mgmt.UserWorkspacesFacade
}

func (wab WorkspaceAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var params = &broker.WorkspaceListParams{
		Broker: broker.Broker{
			AuthFile: wab.Ledger.AuthFile,
		},
		RcEnabled: true,
	}
	var facade = wab.workspaceFacadeProvider().
		WithParams(params).
		WithLogger(wab.Ledger.Logger).
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
		Ledger: ledger,

		workspaceFacadeProvider: mgmt.NewUserWorkspacesFacade,
	}
}
