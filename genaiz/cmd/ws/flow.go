package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/flow"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type workspaceFlowCreateTaskFactory func() *task.Task[broker.WorkspaceFlowCreateParams]
type workspaceFlowResolveTaskFactory func() *task.Task[broker.WorkspaceFlowResolveParams]
type workspaceFlowSolutionTaskFactory func() *task.Task[broker.WorkspaceFlowResolveParams]

type FlowCreateExecutor struct {
	BaseExecutor
	*flow.CreateOptions

	accountParams           config.AccountParametric
	printerParams           cli.PrinterParametric
	resolveParams           *broker.WorkspaceFlowResolveParams
	flowCreateTaskFactory   workspaceFlowCreateTaskFactory
	flowResolveTaskFactory  workspaceFlowResolveTaskFactory
	flowSolutionTaskFactory workspaceFlowSolutionTaskFactory
}

func (cfe *FlowCreateExecutor) Create(workspace, solution, workflow string) error {
	cfe.resolveParams = cfe.newResolveParams(workspace, solution, workflow)
	cfe.Cli.Exec(cfe.Ledger, cfe)
	return nil
}

func (cfe *FlowCreateExecutor) Display() {
	var optionMap = make(map[string]string)

	panicz.RequiresNotNil("resolveParams", cfe.resolveParams)

	if cfe.resolveParams.WorkspaceId == nil {
		optionMap["workspace name"] = cast.ToString(cfe.resolveParams.WorkspaceName)
	} else {
		optionMap["workspace id"] = cast.ToString(cfe.resolveParams.WorkspaceId)
	}

	if cfe.resolveParams.SolutionOem != "" {
		optionMap["solution oem"] = cfe.resolveParams.SolutionOem
		optionMap["solution handle"] = cfe.resolveParams.SolutionHandle
		optionMap["solution version"] = cfe.resolveParams.SolutionVersion
	}

	if cfe.resolveParams.WorkflowId == nil {
		optionMap["workflow handle"] = cfe.resolveParams.WorkflowHandle
	} else {
		optionMap["workflow id"] = cast.ToString(cfe.resolveParams.WorkflowId)
	}

	cfe.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&cfe.OptionAccount.Option,
		&cfe.OptionDescription.Option,
		&cfe.OptionJsonPrinter.Option,
		&cfe.OptionName.Option,
	)
}

func (cfe *FlowCreateExecutor) Pretend() {
	var plan = task.NewPlan("WorkspaceFlowCreate", cfe.Ledger.Logger)
	var workers []task.Worker

	if cfe.resolveParams.WorkspaceId == nil {
		workers = append(workers, task.NewPretender(cfe.resolveParams, cfe.flowResolveTaskFactory()))
	}

	if cfe.resolveParams.WorkflowId == nil {
		workers = append(workers, task.NewPretender(cfe.resolveParams, cfe.flowSolutionTaskFactory()))
	}

	workers = append(workers, task.NewPretender(cfe.resolveParams.WorkspaceFlowCreateParams, cfe.flowCreateTaskFactory()))
	plan.Sequence(workers...)
}

func (cfe *FlowCreateExecutor) Proceed() {
	var workers []task.Worker
	var plan *task.Plan

	if cfe.printerParams.IsDefault() {
		plan = task.NewPlan("WorkspaceFlowCreate", cfe.Ledger.Logger)
		plan.PrintReportsOnly = true
	} else {
		var printer = cfe.printerParams.Printer()

		plan = task.NewPlanBuilder(cfe.Ledger.Logger).
			WithReturn(cli.HandlePrint(printer)).
			WithFailures(cli.HandleError(printer)).
			Build()
	}

	if cfe.resolveParams.WorkspaceId == nil {
		workers = append(workers, task.NewWorker(cfe.resolveParams, cfe.flowResolveTaskFactory()))
	}

	if cfe.resolveParams.WorkflowId == nil {
		workers = append(workers, task.NewWorker(cfe.resolveParams, cfe.flowSolutionTaskFactory()))
	}

	workers = append(workers, task.NewWorker(cfe.resolveParams.WorkspaceFlowCreateParams, cfe.flowCreateTaskFactory()))
	plan.Sequence(workers...)
}

func (cfe *FlowCreateExecutor) newResolveParams(workspace, solution, workflow string) *broker.WorkspaceFlowResolveParams {
	var result = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			Broker:      cfe.accountParams.BrokerParams(),
			Name:        cfe.Ledger.GetString(cfe.OptionName),
			Description: cfe.Ledger.GetString(cfe.OptionDescription),
		},
		// kludge: temporarily hardcoded to avoid explosion of test cases
		RcEnabled: true,
	}

	if workspaceId, err := strconv.Atoi(workspace); err == nil {
		result.WorkspaceId = new(int64(workspaceId))
	} else {
		result.WorkspaceName = workspace
	}

	if _, err := strconv.Atoi(solution); err != nil {
		var oem, handle, ver = broker.ParseFqdnVersion(solution)

		result.SolutionOem = oem
		result.SolutionHandle = handle
		result.SolutionVersion = ver
	}

	if workflowId, err := strconv.Atoi(workflow); err == nil {
		result.WorkflowId = new(int64(workflowId))
	}

	return result
}

func NewFlow(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var createOptions = flow.NewCreateOptions()
	var createFactory = newFlowCreateExecutorFactory(ledger, wsCli, createOptions)
	var flowCmd = &cobra.Command{
		Use:     "flow",
		Aliases: []string{"fl"},
		Short:   "Manages flows under workspaces",
	}

	flowCmd.AddCommand(flow.NewCreate(ledger, createOptions, createFactory))
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

		accountParams: config.NewAccountParams(ledger, options.OptionAccount),
		printerParams: cli.NewPrinterParam(ledger, options.OptionJsonPrinter),

		flowCreateTaskFactory:   broker.NewWorkspaceFlowCreateTask,
		flowResolveTaskFactory:  broker.NewWorkspaceFlowResolveTask,
		flowSolutionTaskFactory: broker.NewWorkspaceFlowSolutionTask,
	}
}

func newFlowCreateExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *flow.CreateOptions) flow.CreateExecutorFactory {
	return func(cmd *cobra.Command) flow.CreateExecutor {
		return NewFlowCreateExecutor(cmd.Context(), ledger, wsCli, options)
	}
}
