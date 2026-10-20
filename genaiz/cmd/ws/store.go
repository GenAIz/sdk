package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/args"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/cmd/ws/store"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type DataStoreResolveTaskFactory func() *task.Task[broker.DataStoreResolveParams]

type DataStoreAutoBridge struct {
	workspaces     auto.Bridge
	workspaceFlows auto.WorkspaceBridge
	workspaceNodes auto.WorkspaceFlowBridge
	dataStores     auto.Bridge
}

func (sab DataStoreAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
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
		results, directive = sab.dataStores.Bridge(toComplete)
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

type DataStoreExecutor struct {
	DataInstanceExecutor

	addDataStore string
	rmDataStore  string

	dataStoreResolveTaskFactory DataStoreResolveTaskFactory
	nodeStoreAddTaskFactory     broker.WorkspaceNodeStoreAddTaskFactory
	nodeStoreRemoveTaskFactory  broker.WorkspaceNodeStoreRemoveTaskFactory
}

func (dse *DataStoreExecutor) Add(ws, fl, nd, dsArg string) error {
	dse.nodeArguments = args.NewNodeResolution(ws, fl, nd)
	dse.addDataStore = dsArg
	dse.Cli.Exec(dse.Ledger, dse)
	return nil
}

func (dse *DataStoreExecutor) Display() {
	var optionMap = dse.nodeArguments.MakeOptionMap()

	if dse.addDataStore != "" {
		optionMap["add data store"] = dse.addDataStore
	} else if dse.rmDataStore != "" {
		optionMap["rm data store"] = dse.rmDataStore
	}

	dse.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&dse.optionAccount.Option,
	)
}

func (dse *DataStoreExecutor) Pretend() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var workers []task.Worker
	var plan *task.Plan

	if dse.addDataStore != "" {
		plan = task.NewPlan("NodeStoreAdd", dse.Ledger.Logger)
	} else {
		plan = task.NewPlan("NodeStoreRm", dse.Ledger.Logger)
	}

	workers = append(workers, task.NewPretender(resolveParams, dse.nodeResolveTaskFactory()))

	if dse.addDataStore != "" {
		var dsParams = dse.newDataStoreResolveParams(brokerParams, dse.addDataStore)
		var storeAddParams = dse.newNodeStoreParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewPretender(dsParams, dse.dataStoreResolveTaskFactory()))
		workers = append(workers, task.NewPretender(storeAddParams, dse.nodeStoreAddTaskFactory()))
	} else if dse.rmDataStore != "" {
		var dsParams = dse.newDataStoreResolveParams(brokerParams, dse.rmDataStore)
		var storeRmParams = dse.newNodeStoreParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewPretender(dsParams, dse.dataStoreResolveTaskFactory()))
		workers = append(workers, task.NewPretender(storeRmParams, dse.nodeStoreRemoveTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (dse *DataStoreExecutor) Proceed() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.nodeArguments.ResolveNodeParams(brokerParams)
	var workers []task.Worker
	var plan *task.Plan

	if dse.printerParams.IsDefault() {
		var planName = "NodeStoreAdd"

		if dse.rmDataStore != "" {
			planName = "NodeStoreRm"
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

	if dse.addDataStore != "" {
		var dsParams = dse.newDataStoreResolveParams(brokerParams, dse.addDataStore)
		var storeAddParams = dse.newNodeStoreParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewWorker(dsParams, dse.dataStoreResolveTaskFactory()))
		workers = append(workers, task.NewWorker(storeAddParams, dse.nodeStoreAddTaskFactory()))
	} else if dse.rmDataStore != "" {
		var dsParams = dse.newDataStoreResolveParams(brokerParams, dse.rmDataStore)
		var storeRmParams = dse.newNodeStoreParams(dsParams, resolveParams.NodeId)

		workers = append(workers, task.NewWorker(dsParams, dse.dataStoreResolveTaskFactory()))
		workers = append(workers, task.NewWorker(storeRmParams, dse.nodeStoreRemoveTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (dse *DataStoreExecutor) Remove(ws, fl, nd, ds string) error {
	dse.nodeArguments = args.NewNodeResolution(ws, fl, nd)
	dse.rmDataStore = ds
	dse.Cli.Exec(dse.Ledger, dse)
	return nil
}

func (dse *DataStoreExecutor) newDataStoreResolveParams(brokerParams *broker.Broker, dataStore string) *broker.DataStoreResolveParams {
	var result = &broker.DataStoreResolveParams{
		Broker: *brokerParams,
	}

	if dsId, err := strconv.ParseInt(dataStore, 10, 64); err == nil {
		result.DataStoreId = new(dsId)
	} else {
		result.DataStoreName = dataStore
	}

	return result
}

func (dse *DataStoreExecutor) newNodeStoreParams(dsParams *broker.DataStoreResolveParams, nodeId *int64) *broker.WorkspaceNodeStoreParams {
	return &broker.WorkspaceNodeStoreParams{
		DataStoreResolveParams: dsParams,
		NodeId:                 nodeId,
	}
}

func NewDataStore(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var storeAuto = NewDataStoreAutoBridge(ledger)
	var addOptions = NewDataStoreAddOptions()
	var rmOptions = NewDataStoreRmOptions()
	var addCommand = store.NewAddStore(newStoreAddFactory(ledger, wsCli, addOptions))
	var rmCommand = store.NewRemoveStore(newStoreRemoveFactory(ledger, wsCli, rmOptions))
	var strCmd = &cobra.Command{
		Use:     "store",
		Aliases: []string{"str"},
		Short:   "Manages data stores for nodes under workspace flows",
	}

	strCmd.AddCommand(addCommand)
	strCmd.AddCommand(rmCommand)
	ledger.Register(addCommand, addOptions.allDefiners()...)
	ledger.Register(rmCommand, rmOptions.allDefiners()...)
	addCommand.ValidArgsFunction = storeAuto.bridgeArguments
	rmCommand.ValidArgsFunction = storeAuto.bridgeArguments
	cli.AutoBridge.Accounts().Option(addCommand, ledger, addOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(rmCommand, ledger, rmOptions.optionAccount)
	return strCmd
}

func NewDataStoreExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *DataInstanceOptions) *DataStoreExecutor {
	return &DataStoreExecutor{
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

		dataStoreResolveTaskFactory: broker.NewDataStoreResolveTask,
		nodeStoreAddTaskFactory:     broker.WorkspaceTasks.AddStoreNodeTaskFactory,
		nodeStoreRemoveTaskFactory:  broker.WorkspaceTasks.RemoveStoreNodeTaskFactory,
	}
}

func NewDataStoreAutoBridge(ledger *config.Ledger) *DataStoreAutoBridge {
	return &DataStoreAutoBridge{
		dataStores:     auto.NewDataStoreAutoBridge(ledger),
		workspaces:     auto.NewWorkspaceBridge(ledger),
		workspaceFlows: auto.NewWorkspaceFlowBridge(ledger),
		workspaceNodes: auto.NewWorkspaceNodeAutoBridge(ledger),
	}
}

func NewDataStoreAddOptions() *DataInstanceOptions {
	return &DataInstanceOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Node.Store.Add.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Node.Store.Add.Printer).
			BuildBoolOption(),
	}
}

func NewDataStoreRmOptions() *DataInstanceOptions {
	return &DataInstanceOptions{
		optionAccount: cli.Options.Workspaces.Account().
			WithKeys(&schema.Genaiz.Workspace.Node.Store.Remove.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workspace.Node.Store.Remove.Printer).
			BuildBoolOption(),
	}
}

func newStoreAddFactory(ledger *config.Ledger, wsCli *Cli, addOptions *DataInstanceOptions) store.AddExecutorFactory {
	return func(cmd *cobra.Command) store.AddExecutor {
		return NewDataStoreExecutor(cmd.Context(), ledger, wsCli, addOptions)
	}
}

func newStoreRemoveFactory(ledger *config.Ledger, wsCli *Cli, rmOptions *DataInstanceOptions) store.RemoveExecutorFactory {
	return func(cmd *cobra.Command) store.RemoveExecutor {
		return NewDataStoreExecutor(cmd.Context(), ledger, wsCli, rmOptions)
	}
}
