package broker

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/run"
	"genaiz.com/genaiz/task/broker/workspace"
)

var (
	RunTasks struct {
		ResolveRunTaskFactory RunResolveTaskFactory // ResolveRunTaskFactory builds a task, which can be used with an instance of RunResolveParams to find if a specified Flow already has a run ready to fire.
		StartRunTaskFactory   RunStartTaskFactory   // StartRunTaskFactory builds a task, which can be used with an instance of RunStartParams to create a run from a Workspace Flow.
	}
)

type RunResolveTask = *task.Task[RunResolveParams]
type RunResolveTaskFactory = func() RunResolveTask
type RunStartTask = *task.Task[RunStartParams]
type RunStartTaskFactory = func() RunStartTask

type RunFilter struct {
	FlowId         *int64
	WorkflowHandle string
}

func (f RunFilter) FirstStartable(runs []WorkspaceRun) *WorkspaceRun {
	if i := slices.IndexFunc(runs, func(workspaceRun WorkspaceRun) bool {
		if workspaceRun.IsStartable() {
			if workspaceRun.Workflow != nil && f.WorkflowHandle != "" {
				return strings.EqualFold(workspaceRun.Workflow.Handle, f.WorkflowHandle)
			}

			return f.FlowId != nil && *f.FlowId == workspaceRun.WorkspaceFlowId
		}

		return false
	}); i >= 0 {
		return &runs[i]
	}

	return nil
}

func (f RunFilter) FirstStarted(runs []WorkspaceRun) *WorkspaceRun {
	if i := slices.IndexFunc(runs, func(workspaceRun WorkspaceRun) bool {
		if workspaceRun.IsStarted() {
			if workspaceRun.Workflow != nil && f.WorkflowHandle != "" {
				return strings.EqualFold(workspaceRun.Workflow.Handle, f.WorkflowHandle)
			}

			return f.FlowId != nil && *f.FlowId == workspaceRun.WorkspaceFlowId
		}

		return false
	}); i >= 0 {
		return &runs[i]
	}

	return nil
}

type RunResolveParams struct {
	Broker
	*workspace.FlowReadParams
	*run.StartParams
	WorkflowHandle string
}

func (p *RunResolveParams) GetFlowId() *int64 {
	if p.FlowReadParams == nil {
		return nil
	}

	return p.FlowReadParams.FlowId
}

func (p *RunResolveParams) GetRunFlags() (int, int) {
	return WorkspaceRunFlags.Active, WorkspaceRunFlags.Active
}

func (p *RunResolveParams) GetRunId() *int64 {
	if p.StartParams == nil {
		return nil
	}

	return p.StartParams.RunId
}

func (p *RunResolveParams) GetWorkspaceId() *int64 {
	if p.FlowReadParams == nil {
		return nil
	}

	return p.FlowReadParams.WorkspaceId
}

func (p *RunResolveParams) SetFlowId(flowId *int64) {
	if p.FlowReadParams == nil {
		p.FlowReadParams = &workspace.FlowReadParams{}
	}

	p.FlowReadParams.FlowId = flowId
}

func (p *RunResolveParams) SetRunId(runId *int64) {
	if p.StartParams == nil {
		p.StartParams = &run.StartParams{}
	}

	p.StartParams.RunId = runId
}

type RunStartParams struct {
	Broker
	*workspace.FlowReadParams
	*run.StartParams
	Name string
}

func (p *RunStartParams) GetFlowId() *int64 {
	if p.FlowReadParams == nil {
		return nil
	}

	return p.FlowReadParams.FlowId
}

func (p *RunStartParams) GetRunId() *int64 {
	if p.StartParams == nil {
		return nil
	}

	return p.StartParams.RunId
}

func handleRunResolveComplete(params *RunResolveParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var workspaceId = *params.GetWorkspaceId()
			var mask, flags = params.GetRunFlags()
			var runs []WorkspaceRun

			state.Logger.Debugf("Resolving runs for workspace [%d]", workspaceId)

			if runs, err = brokerClient.ListWorkspaceRuns(workspaceId, mask, flags); err == nil {
				var runFilter = &RunFilter{
					FlowId:         params.GetFlowId(),
					WorkflowHandle: params.WorkflowHandle,
				}

				if startable := runFilter.FirstStartable(runs); startable != nil {
					state.Logger.Debugf("Found existing run id [%d], with status [%s]", *startable.Id, startable.Status)
					state.Reportf("Selected run [%d]", *startable.Id)
					params.SetFlowId(&startable.WorkspaceFlowId)
					params.SetRunId(startable.Id)
					state.Output = ""
					return nil
				}

				if started := runFilter.FirstStarted(runs); started != nil {
					state.Logger.Debugf("Found run id [%d] currently tracked, exiting", *started.Id)
					return run.Errors.RunDuplicated(started.WorkspaceFlowId)
				}

				if params.GetFlowId() == nil {
					// We need to create a new Run with the resolved flow
					var flows []WorkspaceFlow

					if flows, err = brokerClient.ListWorkspaceFlows(workspaceId, WorkspaceFlowFlags.Active, WorkspaceFlowFlags.Active); err == nil {
						// List them all, because we are going to update their status before create/start
						if i := slices.IndexFunc(flows, func(flow WorkspaceFlow) bool {
							if flow.Solution != nil {
								if _, err = flow.Solution.FindWorkflowByHandle(params.WorkflowHandle); err == nil {
									return true
								}
							}

							return false
						}); i >= 0 {
							state.Logger.Debugf("Found flow id [%d] for handle [%s]", flows[i].Id, params.WorkflowHandle)
							params.SetFlowId(&flows[i].Id)
							return nil
						}

						return run.Errors.RunFlowUnknown()
					}
				}
			}
		}

		return err
	}

	return run.Errors.RunFlowUnknown()
}

func handleRunResolveContext(params *RunResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.GetRunId() != nil {
			return run.Errors.RunIdKnown()
		}

		if params.GetWorkspaceId() == nil {
			// Always needed, because we need to make sure we're not creating supplementary undone runs
			return run.Errors.RunWorkspaceUnknown()
		}

		if params.GetFlowId() == nil && params.WorkflowHandle == "" {
			return run.Errors.RunFlowUnknown()
		}

		if params.FlowId != nil {
			state.Output = cast.ToString(*params.FlowId)
		} else {
			state.Output = params.WorkflowHandle
		}
	}

	return nil
}

func handleRunResolveIncomplete(params *RunResolveParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, run.Errors.RunIdKnown()) {
		state.Logger.Debugf("Resolved run id [%d]", *params.GetRunId())
		state.Output = ""
		return nil
	}

	return state.Error
}

func handleRunResolvePretend(params *RunResolveParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var workspaceId = *params.GetWorkspaceId()
		var flowId = params.GetFlowId()

		state.Logger.Debugf("Pretending to list workspace runs for workspace [%d]", workspaceId)
		fmt.Printf("genaiz raw get %s \\\n", brokerClient.ListWorkspaceRunsUrl())
		fmt.Printf("  -f workspaceId=%d\\\n", workspaceId)
		fmt.Println("  -r .runs")

		if flowId == nil {
			state.Logger.Debugf("Pretending to list workspace flows for workspace [%d]", workspaceId)
			fmt.Printf("genaiz raw get %s \\\n", brokerClient.ListWorkspaceFlowsUrl())
			fmt.Printf("  -f workspaceId=%d\\\n", workspaceId)
			fmt.Printf("  -r .workspaceFlows")
			return nil
		}
	}

	return err
}

func handleRunStartComplete(params *RunStartParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var runId = *params.GetRunId()
			var result *WorkspaceRun

			if result, err = brokerClient.StartWorkspaceRun(runId); err == nil {
				state.Reportf("Started run id [%d]", result.Id)
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return run.Errors.RunUnprepared()
}

func handleRunStartContext(params *RunStartParams, state *task.State) error {
	if state.Output == "" {
		if params.GetFlowId() == nil {
			return run.Errors.RunFlowUnknown()
		}

		if flow, ok := state.Internal.(WorkspaceFlow); !ok {
			return run.Errors.RunUnprepared()
		} else {
			if params.Name == "" {
				if flow.Solution == nil {
					return run.Errors.RunSolutionUnknown()
				}

				if i := slices.IndexFunc(flow.Solution.Workflows, func(workflow Workflow) bool {
					return workflow.Id != nil && *workflow.Id == flow.WorkflowId
				}); i >= 0 {
					state.Output = fmt.Sprintf("%s-%d", flow.Solution.Workflows[i].Handle, time.Now().UnixMilli())
					return nil
				}

				if state.Output == "" {
					return run.Errors.RunWorkflowUnknown()
				}
			}

			state.Output = params.Name
		}

		if params.GetRunId() == nil {
			return run.Errors.RunIdRequired()
		}

		return nil
	}

	return nil
}

func handleRunStartIncomplete(params *RunStartParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, run.Errors.RunIdRequired()) {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var flowId = *params.GetFlowId()
			var result *WorkspaceRun

			state.Logger.Debugf("Creating a new run for flow [%d], named [%s]", flowId, state.Output)

			if result, err = brokerClient.CreateWorkspaceRun(flowId, state.Output, true); err == nil {
				state.Reportf("Created run id [%d] for flow [%d]", result.Id, flowId)
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return state.Error
}

func handleRunStartPretend(params *RunStartParams, state *task.State) error {
	if state.Error == nil ||
		errors.Is(state.Error, run.Errors.RunIdRequired()) {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var runId = params.GetRunId()

			if runId == nil {
				var flowId = params.GetFlowId()

				if flowId != nil {
					state.Logger.Debugf("Pretending to create a workspace run for flow [%d]", *flowId)
					fmt.Printf("genaiz raw post %s \\\n", brokerClient.CreateWorkspaceRunUrl())
					fmt.Printf("  -f workspaceFlowId=%d\\\n", *flowId)
					fmt.Printf("  -f name=%s\\\n", state.Output)
					fmt.Printf("  -f start=true\\\n")
					fmt.Printf("  -r .run")
					return nil
				}
			} else {
				state.Logger.Debugf("Pretending to start workspace run id [%d]", *runId)
				fmt.Printf("genaiz raw get %s \\\n", brokerClient.StartWorkspaceRunUrl())
				fmt.Printf("  -f id=%d\\\n", *runId)
				fmt.Printf("  -r .run")
				return nil
			}

			return run.Errors.RunUnprepared()
		}

		return err
	}

	return state.Error
}

func init() {
	RunTasks.ResolveRunTaskFactory = newRunResolveTask
	RunTasks.StartRunTaskFactory = newRunStartTask
}

func newRunResolveTask() RunResolveTask {
	return &task.Task[RunResolveParams]{
		Name:         "run-resolve",
		OnPrepare:    handleRunResolveContext,
		OnComplete:   handleRunResolveComplete,
		OnIncomplete: handleRunResolveIncomplete,
		OnPretend:    handleRunResolvePretend,
	}
}

func newRunStartTask() RunStartTask {
	return &task.Task[RunStartParams]{
		Name:         "run-start",
		OnPrepare:    handleRunStartContext,
		OnComplete:   handleRunStartComplete,
		OnIncomplete: handleRunStartIncomplete,
		OnPretend:    handleRunStartPretend,
	}
}
