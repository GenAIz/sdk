package flow

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type CreateExecutor interface {
	Create(string, string, string) error
}

type CreateExecutorFactory func(command *cobra.Command) CreateExecutor

type CreateOptions struct {
	OptionAccount     *config.StringOption
	OptionDescription *config.StringOption
	OptionJsonPrinter *config.BoolOption
	OptionName        *config.StringOption
}

func (co CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.OptionAccount,
		co.OptionDescription,
		co.OptionJsonPrinter,
		co.OptionName,
	}
}

type CreateAutoBridge struct {
	ledger *config.Ledger

	solutionFacadeProvider  func() mgmt.UserSolutionFacade
	workflowFacadeProvider  func() mgmt.UserWorkflowFacade
	workspaceFacadeProvider func() mgmt.UserWorkspacesFacade
}

func (cab CreateAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = cab.bridgeWorkspaces(toComplete)
	} else if argsCount == 1 {
		// In the case where the toComplete string refers to a workflowId, we can not autocomplete as the orchestrator
		// does not provide a list of workflows, without solution coordinates
		results, directive = cab.bridgeSolutions(toComplete)
	} else if argsCount == 2 {
		results, directive = cab.bridgeWorkflows(args[1], toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

func (cab CreateAutoBridge) bridgeSolutions(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var oem, _, _ = broker.ParseFqdnVersion(toComplete)
	var params = &broker.SolutionListParams{
		Broker: broker.Broker{
			AuthFile: cab.ledger.AuthFile,
		},
		Oem: oem,
	}
	var facade = cab.solutionFacadeProvider().
		WithParams(params).
		WithLogger(cab.ledger.Logger).
		Filtering(toComplete)

	if solutions, err := facade.Get(); err == nil {
		if len(solutions) > 0 {
			var results []cobra.Completion

			for _, w := range solutions {
				results = append(results, w.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (cab CreateAutoBridge) bridgeWorkflows(fqdnVersion string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var oem, handle, ver = broker.ParseFqdnVersion(fqdnVersion)
	var params = &broker.WorkflowListParams{
		Broker: broker.Broker{
			AuthFile: cab.ledger.AuthFile,
		},
		Oem:     oem,
		Handle:  handle,
		Version: ver,
	}
	var facade = cab.workflowFacadeProvider().
		WithParams(params).
		WithLogger(cab.ledger.Logger).
		Filtering(toComplete)

	if workflows, err := facade.Get(); err == nil {
		if len(workflows) > 0 {
			var results []cobra.Completion

			for _, w := range workflows {
				results = append(results, w.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func (cab CreateAutoBridge) bridgeWorkspaces(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var params = &broker.WorkspaceListParams{
		Broker: broker.Broker{
			AuthFile: cab.ledger.AuthFile,
		},
		RcEnabled: true,
	}
	var facade = cab.workspaceFacadeProvider().
		WithParams(params).
		WithLogger(cab.ledger.Logger).
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

func NewCreate(ledger *config.Ledger, options *CreateOptions, factory CreateExecutorFactory) *cobra.Command {
	var createAuto = NewCreateAuto(ledger)
	var createCmd = &cobra.Command{
		Use:     "create [WORKSPACE_NAME]|WORKSPACE_ID [SOLUTION_FQDN_VERSION WORKFLOW_HANDLE]|WORKFLOW_ID",
		Short:   "Creates a Workspace Flow container",
		Long:    "Creates a Workspace Flow container definition for a Solution Workflow",
		Example: "genaiz create my-workspace com.genaiz.dev/my-function:1.0.0 my-workflow --json",
		Args:    cobra.MatchAll(cobra.MinimumNArgs(2), cobra.MaximumNArgs(3)),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var err error

			if len(args) == 2 {
				// case where we did not auto-complete on solutions and we knew which workflow id to use
				err = exec.Create(args[0], "", args[1])
			} else {
				// case where we did autocomplete and the last argument may be a handle or a workflow id, so we still need the solution
				err = exec.Create(args[0], args[1], args[2])
			}

			lang.HandleExit(err)
		},
	}

	ledger.Register(createCmd, options.allDefiners()...)
	// auto-completion is a composite for this command
	createCmd.ValidArgsFunction = createAuto.bridgeArguments
	return createCmd
}

func NewCreateAuto(ledger *config.Ledger) *CreateAutoBridge {
	return &CreateAutoBridge{
		ledger: ledger,

		solutionFacadeProvider:  mgmt.NewUserSolutionFacade,
		workflowFacadeProvider:  mgmt.NewUserWorkflowFacade,
		workspaceFacadeProvider: mgmt.NewUserWorkspacesFacade,
	}
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		OptionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Flow.Create.Account).
			BuildStringOption(),
		OptionDescription: cli.Options.Workspaces.FlowDescription().
			WithKeys(&schema.Genaiz.Workspace.Flow.Create.Description).
			BuildStringOption(),
		OptionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Flow.Create.Printer).
			BuildBoolOption(),
		OptionName: cli.Options.Workspaces.FlowName().
			WithKeys(&schema.Genaiz.Workspace.Flow.Create.Name).
			BuildStringOption(),
	}
}
