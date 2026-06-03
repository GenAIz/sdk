package ws

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type WorkspaceCreateTaskFactory func() *task.Task[broker.WorkspaceCreateParams]

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions

	accountParams              config.AccountParametric
	printerParams              cli.PrinterParametric
	workspaceName              string
	workspaceCreateTaskFactory WorkspaceCreateTaskFactory
}

func (ce CreateExecutor) Display() {
	var accountName = ce.Ledger.GetString(ce.optionAccount)
	var optionMap = map[string]string{
		"name": ce.workspaceName,
	}

	if accountName != "" {
		optionMap["account"] = accountName
	}

	ce.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&ce.optionDescription.Option,
		&ce.optionDisallowRc.Option,
		&ce.optionJsonPrinter.Option,
		&ce.optionVisibility.Option,
	)
}

func (ce CreateExecutor) Pretend() {
	var brokerParams = ce.accountParams.BrokerParams()
	var createParams = ce.newCreateParams(brokerParams)

	ce.workspaceCreateTaskFactory().Pretend(createParams, ce.Ledger.Logger)
}

func (ce CreateExecutor) Proceed() {
	var brokerParams = ce.accountParams.BrokerParams()
	var createParams = ce.newCreateParams(brokerParams)
	var plan *task.Plan

	if ce.printerParams.IsDefault() {
		plan = task.NewPlan("WorkspaceCreate", ce.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = ce.printerParams.Printer()

		plan = task.NewPlanBuilder(ce.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	task.Single(plan, createParams, ce.workspaceCreateTaskFactory())
}

func (ce CreateExecutor) newCreateParams(brokerParams *broker.Broker) *broker.WorkspaceCreateParams {
	panicz.RequiresNotNil("brokerParams", brokerParams)
	return &broker.WorkspaceCreateParams{
		Broker: *brokerParams,
		Workspace: &broker.Workspace{
			Name:        ce.workspaceName,
			Description: ce.Ledger.GetString(ce.optionDescription),
			RcEnabled:   !ce.Ledger.GetBool(ce.optionDisallowRc),
			Visibility:  ce.Ledger.GetString(ce.optionVisibility),
		},
	}
}

type CreateOptions struct {
	optionAccount     *config.StringOption
	optionDescription *config.StringOption
	optionDisallowRc  *config.BoolOption
	optionJsonPrinter *config.BoolOption
	optionVisibility  *config.StringOption
}

func (co CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionAccount,
		co.optionDescription,
		co.optionDisallowRc,
		co.optionJsonPrinter,
		co.optionVisibility,
	}
}

func NewCreate(ledger *config.Ledger, wsCli *Cli, wsValidation *Validation) *cobra.Command {
	var createOptions = NewCreateOptions()
	var createCmd = &cobra.Command{
		Use:     "create NAME",
		Short:   "Creates a Workspace",
		Long:    "Creates a Workspace on the currently active account or the specified account",
		Example: "genaiz ws create workspaceName --account=dev.genaiz.com --description=desc",
		Args:    cobra.MatchAll(cobra.ExactArgs(1), wsValidation.ArgsWorkspaceName(0)),
		Run: func(cmd *cobra.Command, args []string) {
			wsCli.Exec(ledger, NewCreateExecutor(cmd, ledger, wsCli, args[0], createOptions))
		},
	}

	ledger.Register(createCmd, createOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Register(createCmd, ledger, createOptions.optionAccount)
	return createCmd
}

func NewCreateExecutor(cmd *cobra.Command, ledger *config.Ledger, wsCli *Cli, nameArg string, options *CreateOptions) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Ledger:  ledger,
			Context: cmd.Context(),
		},
		CreateOptions: options,

		accountParams:              config.NewAccountParams(ledger, options.optionAccount),
		printerParams:              cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		workspaceName:              nameArg,
		workspaceCreateTaskFactory: broker.NewWorkspaceCreateTask,
	}
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Create.Account).
			BuildStringOption(),
		optionDescription: cli.Options.Workspaces.Description().
			WithKeys(&schema.Genaiz.Workspace.Create.Description).
			BuildStringOption(),
		optionDisallowRc: cli.Options.Workspaces.DisallowRc().
			WithKeys(&schema.Genaiz.Workspace.Create.DisallowsRc).
			BuildBoolOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Create.Printer).
			BuildBoolOption(),
		optionVisibility: cli.Options.Workspaces.Visibility().
			WithKeys(&schema.Genaiz.Workspace.Create.Visibility).
			WithDefaultValue(broker.WorkspaceVisibilityPrivate).
			BuildStringOption(),
	}
}
