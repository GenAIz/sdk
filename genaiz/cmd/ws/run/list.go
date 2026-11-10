package run

import (
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type ListExecutor interface {
	List(string) error
}

type ListExecutorFactory func(command *cobra.Command) ListExecutor

type ListOptions struct {
	optionAccount     *config.StringOption
	optionDateMonthly *config.BoolOption
	optionDateToday   *config.BoolOption
	optionDateWeekly  *config.BoolOption
	optionJsonPrinter *config.BoolOption
}

func (o ListOptions) AccountParams(ledger *config.Ledger) config.AccountParametric {
	return config.NewAccountParams(ledger, o.optionAccount)
}

func (o ListOptions) ListParams(ledger *config.Ledger, brokerParams *broker.Broker,
	readParams *workspace.FlowReadParams) *broker.RunListParams {
	var listParams = &broker.RunListParams{
		Broker:         *brokerParams,
		FlowReadParams: readParams,
	}

	if ledger.GetBool(o.optionDateMonthly) {
		listParams.Since = timez.NewMonthTime()
	} else if ledger.GetBool(o.optionDateWeekly) {
		listParams.Since = timez.NewWeekTime()
	} else {
		listParams.Since = timez.NewTodayTime()
	}

	return listParams
}

func (o ListOptions) Options() []*config.Option {
	return []*config.Option{
		&o.optionAccount.Option,
		&o.optionDateMonthly.Option,
		&o.optionDateToday.Option,
		&o.optionDateWeekly.Option,
		&o.optionJsonPrinter.Option,
	}
}

func (o ListOptions) PrinterParams(ledger *config.Ledger) cli.PrinterParametric {
	return cli.NewPrinterParam(ledger, o.optionJsonPrinter)
}

func (o ListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		o.optionAccount,
		o.optionDateMonthly,
		o.optionDateToday,
		o.optionDateWeekly,
		o.optionJsonPrinter,
	}
}

type ListAutoBridge struct {
	workspaces auto.Bridge
}

func (b ListAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = b.workspaces.Bridge(toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

func NewList(ledger *config.Ledger, options *ListOptions, factory ListExecutorFactory) *cobra.Command {
	var listAuto = NewListAuto(ledger)
	var listCmd = &cobra.Command{
		Use:     "list [WORKSPACE_NAME]|WORKSPACE_ID",
		Short:   "Lists workspace runs",
		Long:    "Lists workspace runs daily or since a specified date",
		Example: "genaiz ws run list my-workspace --json",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var err = exec.List(args[0])

			lang.HandleExit(err)
		},
	}

	ledger.Register(listCmd, options.allDefiners()...)
	cli.AutoBridge.Accounts().Option(listCmd, ledger, options.optionAccount)
	listCmd.ValidArgsFunction = listAuto.bridgeArguments
	return listCmd
}

func NewListAuto(ledger *config.Ledger) *ListAutoBridge {
	return &ListAutoBridge{
		workspaces: auto.NewWorkspaceBridge(ledger),
	}
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Run.List.Account).
			BuildStringOption(),
		optionDateMonthly: cli.Options.Runs.DateMonthly().
			WithKeys(&schema.Genaiz.Workspace.Run.List.DateMonthly).
			BuildBoolOption(),
		optionDateToday: cli.Options.Runs.DateToday().
			WithKeys(&schema.Genaiz.Workspace.Run.List.DateToday).
			WithDefaultValue(cast.ToString(true)).
			BuildBoolOption(),
		optionDateWeekly: cli.Options.Runs.DateWeekly().
			WithKeys(&schema.Genaiz.Workspace.Run.List.DateWeekly).
			BuildBoolOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Run.List.Printer).
			BuildBoolOption(),
	}
}
