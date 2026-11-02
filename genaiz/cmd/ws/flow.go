package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/flow"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/run"
	"genaiz.com/genaiz/task/broker/workspace"
)

type FlowCreateExecutor struct {
	BaseExecutor
	*flow.CreateOptions

	solutionArg  string
	workflowArg  string
	workspaceArg string

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	flowCreateTaskFactory       broker.WorkspaceFlowCreateTaskFactory
	flowResolveTaskFactory      broker.WorkspaceFlowResolveTaskFactory
	workspaceResolveTaskFactory broker.WorkspaceResolveTaskFactory
}

func (e *FlowCreateExecutor) Create(workspaceArg, solutionArg, workflowArg string) error {
	e.solutionArg = solutionArg
	e.workflowArg = workflowArg
	e.workspaceArg = workspaceArg
	e.Cli.Exec(e.Ledger, e)
	return nil
}

func (e *FlowCreateExecutor) Display() {
	var optionMap = e.makeOptionMap()

	e.Ledger.DisplayOptionsWithMap(&optionMap, e.Options()...)
}

func (e *FlowCreateExecutor) Pretend() {
	var brokerParams = e.accountParams.BrokerParams()
	var createParams = e.CreateParams(e.Ledger)
	var resolveParams = e.newResolveParams(brokerParams, createParams)
	var flowParams = &broker.WorkspaceFlowCreateParams{
		Broker:           *brokerParams,
		FlowCreateParams: createParams,
	}
	var plan = task.NewPlan("WorkspaceFlowCreate", e.Ledger.Logger)
	var workers []task.Worker

	if !resolveParams.HasWorkspaceId() {
		workers = append(workers, task.NewPretender(resolveParams, e.workspaceResolveTaskFactory()))
	}

	if resolveParams.WorkflowId == nil {
		workers = append(workers, task.NewPretender(resolveParams, e.flowResolveTaskFactory()))
	}

	workers = append(workers, task.NewPretender(flowParams, e.flowCreateTaskFactory()))
	plan.Sequence(workers...)
}

func (e *FlowCreateExecutor) Proceed() {
	var brokerParams = e.accountParams.BrokerParams()
	var createParams = e.CreateParams(e.Ledger)
	var resolveParams = e.newResolveParams(brokerParams, createParams)
	var flowParams = &broker.WorkspaceFlowCreateParams{
		Broker:           *brokerParams,
		FlowCreateParams: createParams,
	}
	var workers []task.Worker
	var plan *task.Plan

	if e.printerParams.IsDefault() {
		plan = task.NewPlan("WorkspaceFlowCreate", e.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = e.printerParams.Printer()

		plan = task.NewPlanBuilder(e.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	if !resolveParams.HasWorkspaceId() {
		workers = append(workers, task.NewWorker(resolveParams, e.workspaceResolveTaskFactory()))
	}

	if resolveParams.WorkflowId == nil {
		workers = append(workers, task.NewWorker(resolveParams, e.flowResolveTaskFactory()))
	}

	workers = append(workers, task.NewWorker(flowParams, e.flowCreateTaskFactory()))
	plan.Sequence(workers...)
}

func (e *FlowCreateExecutor) makeOptionMap() map[string]string {
	var result = make(map[string]string)

	if _, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result["workspace id"] = e.workspaceArg
	} else {
		result["workspace name"] = e.workspaceArg
	}

	if _, err := strconv.ParseInt(e.solutionArg, 10, 64); err != nil {
		var oem, handle, ver = broker.ParseFqdnVersion(e.solutionArg)

		result["solution oem"] = oem
		result["solution handle"] = handle
		result["solution version"] = ver
	}

	if _, err := strconv.ParseInt(e.workflowArg, 10, 64); err == nil {
		result["workflow id"] = e.workflowArg
	} else {
		result["workflow handle"] = e.workflowArg
	}

	return result
}

func (e *FlowCreateExecutor) newResolveParams(brokerParams *broker.Broker, createParams *workspace.FlowCreateParams) *broker.WorkspaceFlowResolveParams {
	var result = &broker.WorkspaceFlowResolveParams{
		Broker:           *brokerParams,
		FlowCreateParams: createParams,
		// kludge: temporarily hardcoded to avoid explosion of test cases
		RcEnabled: true,
	}

	if id, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result.SetWorkspaceId(id)
	} else {
		result.WorkspaceName = e.workspaceArg
	}

	if _, err := strconv.ParseInt(e.solutionArg, 10, 64); err != nil {
		var oem, handle, ver = broker.ParseFqdnVersion(e.solutionArg)

		result.SolutionOem = oem
		result.SolutionHandle = handle
		result.SolutionVersion = ver
	}

	if id, err := strconv.ParseInt(e.workflowArg, 10, 64); err == nil {
		result.WorkflowId = new(id)
	} else {
		result.WorkflowHandle = e.workflowArg
	}

	return result
}

type FlowStartExecutor struct {
	BaseExecutor
	*flow.StartOptions

	flowArg      string
	workspaceArg string

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	flowPrepareTaskFactory      broker.WorkspaceFlowPrepareTaskFactory
	runResolveTaskFactory       broker.RunResolveTaskFactory
	runStartTaskFactory         broker.RunStartTaskFactory
	workspaceResolveTaskFactory broker.WorkspaceResolveTaskFactory
}

func (e FlowStartExecutor) Display() {
	var optionMap = e.makeOptionMap()

	e.Ledger.DisplayOptionsWithMap(&optionMap, e.Options()...)
}

func (e FlowStartExecutor) Pretend() {
	var brokerParams = e.accountParams.BrokerParams()
	var flowParams = e.newFlowParams(brokerParams)
	var startParams = &run.StartParams{}
	var runStartParams = &broker.RunStartParams{
		Broker:         *brokerParams,
		FlowReadParams: flowParams.FlowReadParams,
		StartParams:    startParams,
	}
	var plan = task.NewPlan("WorkspaceFlowStart", e.Ledger.Logger)
	var workers []task.Worker

	if runStartParams.FlowId == nil {
		var workspaceParams = e.newWorkspaceParams(brokerParams, flowParams.FlowReadParams)
		var runResolveParams = e.newRunResolveParams(brokerParams, flowParams.FlowReadParams, startParams)

		if !workspaceParams.HasWorkspaceId() {
			workers = append(workers, task.NewPretender(workspaceParams, e.workspaceResolveTaskFactory()))
		}

		workers = append(workers, task.NewPretender(runResolveParams, e.runResolveTaskFactory()))
	}

	workers = append(workers, task.NewPretender(flowParams, e.flowPrepareTaskFactory()))
	workers = append(workers, task.NewPretender(runStartParams, e.runStartTaskFactory()))
	plan.Sequence(workers...)
}

func (e FlowStartExecutor) Proceed() {
	var brokerParams = e.accountParams.BrokerParams()
	var flowParams = e.newFlowParams(brokerParams)
	var runStartParams = e.StartParams(e.Ledger, brokerParams, flowParams.FlowReadParams)
	var workers []task.Worker
	var plan *task.Plan

	if e.printerParams.IsDefault() {
		plan = task.NewPlan("WorkspaceFlowStart", e.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = e.printerParams.Printer()

		plan = task.NewPlanBuilder(e.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	if runStartParams.FlowId == nil {
		var workspaceParams = e.newWorkspaceParams(brokerParams, flowParams.FlowReadParams)
		var resolveParams = e.newRunResolveParams(brokerParams, flowParams.FlowReadParams, runStartParams.StartParams)

		// We need to call prepare, so even if we already have a RunId, we need a run resolution
		if !workspaceParams.HasWorkspaceId() {
			workers = append(workers, task.NewWorker(workspaceParams, e.workspaceResolveTaskFactory()))
		}

		workers = append(workers, task.NewWorker(resolveParams, e.runResolveTaskFactory()))
	}

	workers = append(workers, task.NewWorker(flowParams, e.flowPrepareTaskFactory()))
	// Note that RunId is not supported by the command at this point, so a start task should always check if there's a run already prepared
	workers = append(workers, task.NewWorker(runStartParams, e.runStartTaskFactory()))
	plan.Sequence(workers...)
}

func (e FlowStartExecutor) makeOptionMap() map[string]string {
	var result = make(map[string]string)

	if _, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result["workspace id"] = e.workspaceArg
	} else {
		result["workspace name"] = e.workspaceArg
	}

	if _, err := strconv.ParseInt(e.flowArg, 10, 64); err == nil {
		result["flow id"] = e.flowArg
	} else {
		result["workflow handle"] = e.flowArg
	}

	return result
}

func (e FlowStartExecutor) newFlowParams(brokerParams *broker.Broker) *broker.WorkspaceFlowReadParams {
	var result = &broker.WorkspaceFlowReadParams{
		Broker:         *brokerParams,
		FlowReadParams: &workspace.FlowReadParams{},
	}

	if id, err := strconv.ParseInt(e.flowArg, 10, 64); err == nil {
		result.FlowId = &id
	}

	return result
}

func (e FlowStartExecutor) newRunResolveParams(brokerParams *broker.Broker,
	flowParams *workspace.FlowReadParams, startParams *run.StartParams) *broker.RunResolveParams {
	var result = &broker.RunResolveParams{
		Broker:         *brokerParams,
		FlowReadParams: flowParams,
		StartParams:    startParams,
	}

	if flowParams.FlowId == nil {
		result.WorkflowHandle = e.flowArg
	}

	return result
}

func (e FlowStartExecutor) newWorkspaceParams(brokerParams *broker.Broker, flowParams *workspace.FlowReadParams) *broker.WorkspaceFlowResolveParams {
	var result = &broker.WorkspaceFlowResolveParams{
		Broker:         *brokerParams,
		FlowReadParams: flowParams,
	}

	if id, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result.SetWorkspaceId(id)
	} else {
		result.WorkspaceName = e.workspaceArg
	}

	return result
}

func (e FlowStartExecutor) Start(workspaceArg, flowArg string) error {
	e.workspaceArg = workspaceArg
	e.flowArg = flowArg
	e.Cli.Exec(e.Ledger, e)
	return nil
}

func NewFlow(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var createOptions = flow.NewCreateOptions()
	var startOptions = flow.NewStartOptions()
	var createFactory = newFlowCreateExecutorFactory(ledger, wsCli, createOptions)
	var startFactory = newFlowStartExecutorFactory(ledger, wsCli, startOptions)
	var flowCmd = &cobra.Command{
		Use:     "flow",
		Aliases: []string{"fl"},
		Short:   "Manages flows under workspaces",
	}

	flowCmd.AddCommand(flow.NewCreate(ledger, createOptions, createFactory))
	flowCmd.AddCommand(flow.NewStart(ledger, startOptions, startFactory))
	return flowCmd
}

func NewFlowCreateExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *flow.CreateOptions) *FlowCreateExecutor {
	return &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		CreateOptions: options,

		accountParams: options.AccountParams(ledger),
		printerParams: options.PrinterParams(ledger),

		flowCreateTaskFactory:       broker.WorkspaceTasks.CreateFlowTaskFactory,
		flowResolveTaskFactory:      broker.WorkspaceTasks.ResolveFlowTaskFactory,
		workspaceResolveTaskFactory: broker.WorkspaceTasks.ResolveTaskFactory,
	}
}

func NewFlowStartExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *flow.StartOptions) *FlowStartExecutor {
	return &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		StartOptions: options,

		accountParams: options.AccountParams(ledger),
		printerParams: options.PrinterParams(ledger),

		flowPrepareTaskFactory:      broker.WorkspaceTasks.PrepareFlowTaskFactory,
		runResolveTaskFactory:       broker.RunTasks.ResolveRunTaskFactory,
		runStartTaskFactory:         broker.RunTasks.StartRunTaskFactory,
		workspaceResolveTaskFactory: broker.WorkspaceTasks.ResolveFlowTaskFactory,
	}
}

func newFlowCreateExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *flow.CreateOptions) flow.CreateExecutorFactory {
	return func(cmd *cobra.Command) flow.CreateExecutor {
		return NewFlowCreateExecutor(cmd.Context(), ledger, wsCli, options)
	}
}

func newFlowStartExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *flow.StartOptions) flow.StartExecutorFactory {
	return func(cmd *cobra.Command) flow.StartExecutor {
		return NewFlowStartExecutor(cmd.Context(), ledger, wsCli, options)
	}
}
