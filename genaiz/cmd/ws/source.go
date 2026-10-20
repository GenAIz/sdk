package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/args"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/cmd/ws/source"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type DataSourceResolveTaskFactory func() *task.Task[broker.DataSourceResolveParams]

type DataSourceAutoBridge struct {
	workspaces     auto.Bridge
	workspaceFlows auto.WorkspaceBridge
	workspaceNodes auto.WorkspaceFlowBridge
	dataSources    auto.Bridge
}

func (sab DataSourceAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
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
	} else if argsCount == 3 {
		results, directive = sab.dataSources.Bridge(toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

type DataSourceExecutor struct {
	DataInstanceExecutor

	addDataSource string
	rmDataSource  string

	dataSourceResolveTaskFactory DataSourceResolveTaskFactory
	nodeSourceAddTaskFactory     broker.WorkspaceNodeSourceAddTaskFactory
	nodeSourceRemoveTaskFactory  broker.WorkspaceNodeSourceRemoveTaskFactory
}

func (dse *DataSourceExecutor) Add(ws, fl, nd, ds string) error {
	dse.nodeArguments = args.NewNodeResolution(ws, fl, nd)
	dse.addDataSource = ds
	dse.Cli.Exec(dse.Ledger, dse)
	return nil
}

func (dse *DataSourceExecutor) Display() {
	var optionMap = dse.nodeArguments.MakeOptionMap()

	if dse.addDataSource != "" {
		optionMap["add data source"] = dse.addDataSource
	} else if dse.rmDataSource != "" {
		optionMap["rm data source"] = dse.rmDataSource
	}

	dse.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&dse.optionAccount.Option,
	)
}

func (dse *DataSourceExecutor) Pretend() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var workers []task.Worker
	var plan *task.Plan

	if dse.addDataSource != "" {
		plan = task.NewPlan("NodeSourceAdd", dse.Ledger.Logger)
	} else {
		plan = task.NewPlan("NodeSourceRm", dse.Ledger.Logger)
	}

	workers = append(workers, task.NewPretender(resolveParams, dse.nodeResolveTaskFactory()))

	if dse.addDataSource != "" {
		var dsParams = dse.newDataSourceResolveParams(brokerParams, dse.addDataSource)
		var sourceAddParams = dse.newNodeSourceParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewPretender(dsParams, dse.dataSourceResolveTaskFactory()))
		workers = append(workers, task.NewPretender(sourceAddParams, dse.nodeSourceAddTaskFactory()))
	} else if dse.rmDataSource != "" {
		var dsParams = dse.newDataSourceResolveParams(brokerParams, dse.rmDataSource)
		var sourceRmParams = dse.newNodeSourceParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewPretender(dsParams, dse.dataSourceResolveTaskFactory()))
		workers = append(workers, task.NewPretender(sourceRmParams, dse.nodeSourceRemoveTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (dse *DataSourceExecutor) Proceed() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var workers []task.Worker
	var plan *task.Plan

	if dse.printerParams.IsDefault() {
		var planName = "NodeSourceAdd"

		if dse.rmDataSource != "" {
			planName = "NodeSourceRm"
		}

		plan = task.NewPlan(planName, dse.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = dse.printerParams.Printer()

		plan = task.NewPlanBuilder(dse.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	workers = append(workers, task.NewWorker(resolveParams, dse.nodeResolveTaskFactory()))

	if dse.addDataSource != "" {
		var dsParams = dse.newDataSourceResolveParams(brokerParams, dse.addDataSource)
		var sourceAddParams = dse.newNodeSourceParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewWorker(dsParams, dse.dataSourceResolveTaskFactory()))
		workers = append(workers, task.NewWorker(sourceAddParams, dse.nodeSourceAddTaskFactory()))
	} else if dse.rmDataSource != "" {
		var dsParams = dse.newDataSourceResolveParams(brokerParams, dse.rmDataSource)
		var sourceRmParams = dse.newNodeSourceParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewWorker(dsParams, dse.dataSourceResolveTaskFactory()))
		workers = append(workers, task.NewWorker(sourceRmParams, dse.nodeSourceRemoveTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (dse *DataSourceExecutor) Remove(ws, fl, nd, ds string) error {
	dse.nodeArguments = args.NewNodeResolution(ws, fl, nd)
	dse.rmDataSource = ds
	dse.Cli.Exec(dse.Ledger, dse)
	return nil
}

func (dse *DataSourceExecutor) newDataSourceResolveParams(brokerParams *broker.Broker, dataSource string) *broker.DataSourceResolveParams {
	var result = &broker.DataSourceResolveParams{
		Broker: *brokerParams,
	}

	if dsId, err := strconv.ParseInt(dataSource, 10, 64); err == nil {
		result.DataSourceId = new(dsId)
	} else {
		result.DataSourceName = dataSource
	}

	return result
}

func (dse *DataSourceExecutor) newNodeSourceParams(dsParams *broker.DataSourceResolveParams, nodeId *int64) *broker.WorkspaceNodeSourceParams {
	return &broker.WorkspaceNodeSourceParams{
		DataSourceResolveParams: dsParams,
		NodeId:                  nodeId,
	}
}

func NewDataSource(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var sourceAuto = NewDataSourceAutoBridge(ledger)
	var addOptions = NewDataSourceAddOptions()
	var rmOptions = NewDataSourceRmOptions()
	var addCommand = source.NewAddSource(newSourceAddFactory(ledger, wsCli, addOptions))
	var rmCommand = source.NewRemoveSource(newSourceRemoveFactory(ledger, wsCli, rmOptions))
	var srcCmd = &cobra.Command{
		Use:     "source",
		Aliases: []string{"src"},
		Short:   "Manages data sources for nodes under workspace flows",
	}

	srcCmd.AddCommand(addCommand)
	srcCmd.AddCommand(rmCommand)
	ledger.Register(addCommand, addOptions.allDefiners()...)
	ledger.Register(rmCommand, rmOptions.allDefiners()...)
	addCommand.ValidArgsFunction = sourceAuto.bridgeArguments
	rmCommand.ValidArgsFunction = sourceAuto.bridgeArguments
	cli.AutoBridge.Accounts().Option(addCommand, ledger, addOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(rmCommand, ledger, rmOptions.optionAccount)
	return srcCmd
}

func NewDataSourceExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *DataInstanceOptions) *DataSourceExecutor {
	return &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Cli:     wsCli,
				Context: ctx,
				Ledger:  ledger,
			},
			DataInstanceOptions: options,

			accountParams:          config.NewAccountParams(ledger, options.optionAccount),
			printerParams:          cli.NewPrinterParam(ledger, options.optionJsonPrinter),
			nodeResolveTaskFactory: broker.WorkspaceTasks.ResolveNodeTaskFactory,
		},

		dataSourceResolveTaskFactory: broker.NewDataSourceResolveTask,
		nodeSourceAddTaskFactory:     broker.WorkspaceTasks.AddSourceNodeTaskFactory,
		nodeSourceRemoveTaskFactory:  broker.WorkspaceTasks.RemoveSourceNodeTaskFactory,
	}
}

func NewDataSourceAutoBridge(ledger *config.Ledger) *DataSourceAutoBridge {
	return &DataSourceAutoBridge{
		dataSources:    auto.NewDataSourceAutoBridge(ledger),
		workspaces:     auto.NewWorkspaceBridge(ledger),
		workspaceFlows: auto.NewWorkspaceFlowBridge(ledger),
		workspaceNodes: auto.NewWorkspaceNodeAutoBridge(ledger),
	}
}

func NewDataSourceAddOptions() *DataInstanceOptions {
	return &DataInstanceOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Node.Source.Add.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Node.Source.Add.Printer).
			BuildBoolOption(),
	}
}

func NewDataSourceRmOptions() *DataInstanceOptions {
	return &DataInstanceOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Node.Source.Remove.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Node.Source.Remove.Printer).
			BuildBoolOption(),
	}
}

func newSourceAddFactory(ledger *config.Ledger, wsCli *Cli, addOptions *DataInstanceOptions) source.AddExecutorFactory {
	return func(cmd *cobra.Command) source.AddExecutor {
		return NewDataSourceExecutor(cmd.Context(), ledger, wsCli, addOptions)
	}
}

func newSourceRemoveFactory(ledger *config.Ledger, wsCli *Cli, rmOptions *DataInstanceOptions) source.RemoveExecutorFactory {
	return func(cmd *cobra.Command) source.RemoveExecutor {
		return NewDataSourceExecutor(cmd.Context(), ledger, wsCli, rmOptions)
	}
}
