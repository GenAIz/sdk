package ws

import (
	"time"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

type ListExecutor struct {
	BaseExecutor
	*ListOptions

	accountParams            config.AccountParametric
	printerParams            cli.PrinterParametric
	workspaceFacadeProvider  func() mgmt.UserWorkspacesFacade
	workspaceListTaskFactory mgmt.WorkspaceListTaskFactory
}

func (le ListExecutor) Display() {
	var optionMap = map[string]string{}

	if name := le.Ledger.GetString(le.optionAccount); name != "" {
		optionMap["account"] = name
	}

	if dateFilter := le.getFromDate(); dateFilter != nil {
		optionMap["from"] = dateFilter.Format(time.DateTime)
	}

	le.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&le.optionJsonPrinter.Option,
		&le.optionOwnerOnly.Option,
		&le.optionRcEnabled.Option,
	)
}

func (le ListExecutor) Pretend() {
	var brokerParams = le.accountParams.BrokerParams()
	var listParams = le.newListParams(brokerParams)

	le.workspaceListTaskFactory().Pretend(listParams, le.Ledger.Logger)
}

func (le ListExecutor) Proceed() {
	var brokerParams = le.accountParams.BrokerParams()
	var listParams = le.newListParams(brokerParams)
	var printer = le.printerParams.Printer()
	var workspaceList, err = le.workspaceFacadeProvider().
		WithLogger(le.Ledger.Logger).
		WithParams(listParams).
		Provider().
		Get()

	if err == nil {
		lang.HandleExit(printer.Print(workspaceList))
		return
	}

	lang.HandleExit(printer.Error(err))
}

func (le ListExecutor) getFromDate() *time.Time {
	var result *time.Time

	if le.Ledger.GetBool(le.optionDateMonthly) {
		result = timez.NewMonthTime()
	} else if le.Ledger.GetBool(le.optionDateToday) {
		result = timez.NewTodayTime()
	} else if le.Ledger.GetBool(le.optionDateWeekly) {
		result = timez.NewWeekTime()
	}

	return result
}

func (le ListExecutor) newListParams(brokerParams *broker.Broker) *broker.WorkspaceListParams {
	panicz.RequiresNotNil("brokerParams", brokerParams)
	return &broker.WorkspaceListParams{
		Broker:    *brokerParams,
		FromDate:  le.getFromDate(),
		OwnerOnly: le.Ledger.GetBool(le.optionOwnerOnly),
		RcEnabled: le.Ledger.GetBool(le.optionRcEnabled),
	}
}

type ListOptions struct {
	optionAccount     *config.StringOption
	optionDateMonthly *config.BoolOption
	optionDateToday   *config.BoolOption
	optionDateWeekly  *config.BoolOption
	optionJsonPrinter *config.BoolOption
	optionOwnerOnly   *config.BoolOption
	optionRcEnabled   *config.BoolOption
}

func (lo ListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionAccount,
		lo.optionDateMonthly,
		lo.optionDateToday,
		lo.optionDateWeekly,
		lo.optionJsonPrinter,
		lo.optionOwnerOnly,
		lo.optionRcEnabled,
	}
}

func NewList(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var listOptions = NewListOptions()
	var listCmd = &cobra.Command{
		Use:     "list",
		Short:   "Lists available Workspaces",
		Long:    "Lists available Workspaces for the currently active account or a specified account",
		Example: "genaiz ws list --account=dev.genaiz.com --owner-only",
		Args:    cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			wsCli.Exec(ledger, NewListExecutor(cmd, ledger, wsCli, listOptions))
		},
	}

	ledger.Register(listCmd, listOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Register(listCmd, ledger, listOptions.optionAccount)
	return listCmd
}

func NewListExecutor(cmd *cobra.Command, ledger *config.Ledger, wsCli *Cli, options *ListOptions) *ListExecutor {
	return &ListExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: cmd.Context(),
			Ledger:  ledger,
		},
		ListOptions: options,

		accountParams:            config.NewAccountParams(ledger, options.optionAccount),
		printerParams:            cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		workspaceFacadeProvider:  mgmt.NewUserWorkspacesFacade,
		workspaceListTaskFactory: broker.NewWorkspaceListTask,
	}
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.List.Account).
			BuildStringOption(),
		optionDateMonthly: cli.Options.Workspaces.DateMonthly().
			WithKeys(&schema.Genaiz.Workspace.List.DateMonthly).
			BuildBoolOption(),
		optionDateToday: cli.Options.Workspaces.DateToday().
			WithKeys(&schema.Genaiz.Workspace.List.DateToday).
			BuildBoolOption(),
		optionDateWeekly: cli.Options.Workspaces.DateWeekly().
			WithKeys(&schema.Genaiz.Workspace.List.DateWeekly).
			BuildBoolOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.List.Printer).
			BuildBoolOption(),
		optionOwnerOnly: cli.Options.Workspaces.OwnerOnly().
			WithKeys(&schema.Genaiz.Workspace.List.OwnerOnly).
			BuildBoolOption(),
		optionRcEnabled: cli.Options.Workspaces.RcEnabled().
			WithKeys(&schema.Genaiz.Workspace.List.RcEnabled).
			BuildBoolOption(),
	}
}
