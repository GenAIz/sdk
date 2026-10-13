package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/cmd/ws/store"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type DataStoreResolveTaskFactory func() *task.Task[broker.DataStoreResolveParams]
type WorkspaceNodeStoreAddTaskFactory func() *task.Task[broker.WorkspaceNodeStoreParams]
type WorkspaceNodeStoreRemoveTaskFactory func() *task.Task[broker.WorkspaceNodeStoreParams]

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
	flow         string
	node         string
	workspace    string

	dataStoreResolveTaskFactory DataStoreResolveTaskFactory
	nodeStoreAddTaskFactory     WorkspaceNodeStoreAddTaskFactory
	nodeStoreRemoveTaskFactory  WorkspaceNodeStoreRemoveTaskFactory
}

func (dse *DataStoreExecutor) Add(workspaceArg, flowArg, nodeArg, dsArg string) error {
	dse.addDataStore = dsArg
	dse.flow = flowArg
	dse.node = nodeArg
	dse.workspace = workspaceArg
	dse.Cli.Exec(dse.Ledger, dse)
	return nil
}

func (dse *DataStoreExecutor) Display() {
	var optionMap = make(map[string]string)
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.newNodeResolveParams(brokerParams)

	if dse.addDataStore != "" {
		optionMap["add data store"] = dse.addDataStore
	} else if dse.rmDataStore != "" {
		optionMap["rm data store"] = dse.rmDataStore
	}

	if resolveParams.WorkspaceId == nil {
		optionMap["workspace name"] = cast.ToString(resolveParams.WorkspaceName)
	} else {
		optionMap["workspace id"] = cast.ToString(resolveParams.WorkspaceId)
	}

	if resolveParams.FlowId == nil {
		optionMap["workflow handle"] = resolveParams.WorkflowHandle
	} else {
		optionMap["workspace flow id"] = cast.ToString(resolveParams.FlowId)
	}

	if resolveParams.NodeId == nil {
		optionMap["smart function handle"] = resolveParams.FnHandle
	} else {
		optionMap["node id"] = cast.ToString(resolveParams.NodeId)
	}

	dse.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&dse.optionAccount.Option,
	)
}

func (dse *DataStoreExecutor) Pretend() {
	var brokerParams = dse.accountParams.BrokerParams()
	var resolveParams = dse.newNodeResolveParams(brokerParams)
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
	var resolveParams = dse.newNodeResolveParams(brokerParams)
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

func (dse *DataStoreExecutor) Remove(workspaceArg, flowArg, nodeArg, dsArg string) error {
	dse.rmDataStore = dsArg
	dse.flow = flowArg
	dse.node = nodeArg
	dse.workspace = workspaceArg
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

func (dse *DataStoreExecutor) newNodeResolveParams(brokerParams *broker.Broker) *broker.WorkspaceNodeResolveParams {
	var result = &broker.WorkspaceNodeResolveParams{
		Broker: *brokerParams,
	}

	if wsId, err := strconv.ParseInt(dse.workspace, 10, 64); err == nil {
		result.WorkspaceId = new(wsId)
	} else {
		result.WorkspaceName = dse.workspace
	}

	if flId, err := strconv.ParseInt(dse.flow, 10, 64); err == nil {
		result.FlowId = new(flId)
	} else {
		result.WorkflowHandle = dse.flow
	}

	if ndId, err := strconv.ParseInt(dse.node, 10, 64); err == nil {
		result.NodeId = new(ndId)
	} else {
		result.FnHandle = dse.node
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
			nodeResolveTaskFactory: broker.NewWorkspaceNodeResolveTask,
		},

		dataStoreResolveTaskFactory: broker.NewDataStoreResolveTask,
		nodeStoreAddTaskFactory:     broker.NewWorkspaceNodeStoreAddTask,
		nodeStoreRemoveTaskFactory:  broker.NewWorkspaceNodeStoreRemoveTask,
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
