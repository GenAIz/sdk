package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/run"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt/runs"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type RunListExecutor struct {
	BaseExecutor
	*run.ListOptions

	workspaceArg string

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	runListTaskFactory          broker.RunListTaskFactory
	workspaceResolveTaskFactory broker.WorkspaceResolveTaskFactory
	workspaceRunsFactory        runs.UserRunsFactory
}

func (e *RunListExecutor) Display() {
	var optionMap = e.makeOptionMap()

	e.Ledger.DisplayOptionsWithMap(&optionMap, e.Options()...)
}

func (e *RunListExecutor) List(workspaceArg string) error {
	e.workspaceArg = workspaceArg
	e.Cli.Exec(e.Ledger, e)
	return nil
}

func (e *RunListExecutor) Pretend() {
	var brokerParams = e.accountParams.BrokerParams()
	var readParams = e.newReadParams()
	var listParams = e.ListParams(e.Ledger, brokerParams, readParams)
	var plan = task.NewPlan("RunList", e.Ledger.Logger)
	var workers []task.Worker

	if readParams.WorkspaceId == nil {
		var workspaceParams = &broker.WorkspaceFlowResolveParams{
			FlowReadParams: readParams,
			WorkspaceName:  e.workspaceArg,
		}

		workers = append(workers, task.NewPretender(workspaceParams, e.workspaceResolveTaskFactory()))
	}

	workers = append(workers, task.NewPretender(listParams, e.runListTaskFactory()))
	plan.Sequence(workers...)
}

func (e *RunListExecutor) Proceed() {
	var brokerParams = e.accountParams.BrokerParams()
	var readParams = e.newReadParams()
	var listParams = e.ListParams(e.Ledger, brokerParams, readParams)
	var printer = e.printerParams.Printer()
	var workspaceRuns, err = e.workspaceRunsFactory.
		WithLogger(e.Ledger.Logger).
		WithParams(listParams).
		Provider().
		Get()

	if err == nil {
		lang.HandleExit(printer.Print(workspaceRuns))
		return
	}

	lang.HandleExit(printer.Error(err))
}

func (e *RunListExecutor) makeOptionMap() map[string]string {
	var result = map[string]string{}

	if _, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result["workspace id"] = e.workspaceArg
	} else {
		result["workspace name"] = e.workspaceArg
	}

	return result
}

func (e *RunListExecutor) newReadParams() *workspace.FlowReadParams {
	var result = &workspace.FlowReadParams{}

	if id, err := strconv.ParseInt(e.workspaceArg, 10, 64); err == nil {
		result.WorkspaceId = new(id)
	}

	return result
}

func NewRun(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var listOptions = run.NewListOptions()
	var listFactory = newRunListExecutorFactory(ledger, wsCli, listOptions)
	var runCmd = &cobra.Command{
		Use:     "run",
		Aliases: []string{"rn"},
		Short:   "Manages runs under workspaces",
	}

	runCmd.AddCommand(run.NewList(ledger, listOptions, listFactory))
	return runCmd
}

func NewRunListExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *run.ListOptions) *RunListExecutor {
	return &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		ListOptions: options,

		accountParams: options.AccountParams(ledger),
		printerParams: options.PrinterParams(ledger),

		runListTaskFactory:          broker.RunTasks.ListRunTaskFactory,
		workspaceResolveTaskFactory: broker.WorkspaceTasks.ResolveTaskFactory,
		workspaceRunsFactory:        runs.NewUserRunsFactory(),
	}
}

func newRunListExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *run.ListOptions) run.ListExecutorFactory {
	return func(cmd *cobra.Command) run.ListExecutor {
		return NewRunListExecutor(cmd.Context(), ledger, wsCli, options)
	}
}
