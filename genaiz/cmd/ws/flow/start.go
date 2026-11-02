package flow

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/run"
	"genaiz.com/genaiz/task/broker/workspace"
)

type StartExecutor interface {
	Start(string, string) error
}

type StartExecutorFactory func(command *cobra.Command) StartExecutor

type StartAutoBridge struct {
	workspaces     auto.Bridge
	workspaceFlows auto.WorkspaceBridge
}

func (sab StartAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = sab.workspaces.Bridge(toComplete)
	} else if argsCount == 1 {
		// In the case where the toComplete string refers to a workflowId, we can not autocomplete as the orchestrator
		// does not provide a list of workflows, without solution coordinates
		results, directive = sab.workspaceFlows.WithWorkspace(args[0]).
			Bridge(toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

type StartOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
	optionName        *config.StringOption
}

func (so StartOptions) AccountParams(ledger *config.Ledger) config.AccountParametric {
	return config.NewAccountParams(ledger, so.optionAccount)
}

func (so StartOptions) Options() []*config.Option {
	return []*config.Option{
		&so.optionName.Option,
		&so.optionAccount.Option,
		&so.optionJsonPrinter.Option,
	}
}

func (so StartOptions) PrinterParams(ledger *config.Ledger) cli.PrinterParametric {
	return cli.NewPrinterParam(ledger, so.optionJsonPrinter)
}

func (so StartOptions) StartParams(ledger *config.Ledger, brokerParams *broker.Broker,
	readParams *workspace.FlowReadParams) *broker.RunStartParams {
	var name = ledger.GetString(so.optionName)

	return &broker.RunStartParams{
		Broker:         *brokerParams,
		FlowReadParams: readParams,
		StartParams:    &run.StartParams{},
		Name:           name,
	}
}

func (so StartOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionAccount,
		so.optionJsonPrinter,
		so.optionName,
	}
}

func NewStart(ledger *config.Ledger, options *StartOptions, factory StartExecutorFactory) *cobra.Command {
	var startAuto = NewStartAuto(ledger)
	var startCmd = &cobra.Command{
		Use:     "start [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE]|FLOW_ID",
		Short:   "Starts a Run for a Workspace Flow",
		Long:    "Starts a Run for Workspace Flow creating a Run if the Flow is ready",
		Example: "genaiz ws flow start my-workspace my-workflow --json",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var err = exec.Start(args[0], args[1])

			lang.HandleExit(err)
		},
	}

	ledger.Register(startCmd, options.allDefiners()...)
	// auto-completion is a composite for this command
	startCmd.ValidArgsFunction = startAuto.bridgeArguments
	return startCmd
}

func NewStartAuto(ledger *config.Ledger) *StartAutoBridge {
	return &StartAutoBridge{
		workspaces:     auto.NewWorkspaceBridge(ledger),
		workspaceFlows: auto.NewWorkspaceFlowBridge(ledger),
	}
}

func NewStartOptions() *StartOptions {
	return &StartOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Flow.Start.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Flow.Start.Printer).
			BuildBoolOption(),
		optionName: cli.Options.Runs.Name().
			WithKeys(&schema.Genaiz.Workspace.Flow.Start.Name).
			BuildStringOption(),
	}
}
