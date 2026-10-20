package broker

import (
	"errors"
	"fmt"
	"math/rand/v2"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/workspace"
)

var (
	WorkspaceTasks struct {
		CreateTaskFactory  WorkspaceCreateTaskFactory  // CreateTaskFactory builds a task which can be used with an instance of WorkspaceCreateParams to create a Workspace on a Broker account.
		ResolveTaskFactory WorkspaceResolveTaskFactory // ResolveTaskFactory builds a task which can be used with an instance of WorkspaceFlowResolveParams to find a workspace to modify by name.
		ListTaskFactory    WorkspaceListTaskFactory    // ListTaskFactory builds a task which can be used with an instance of WorkspaceListParams to list workspaces on a Broker account.

		CreateFlowTaskFactory  WorkspaceFlowCreateTaskFactory  // CreateFlowTaskFactory builds a task which can be used with an instance of WorkspaceFlowCreateParams to create a Workspace Flow on a Broker account.
		ListFlowTaskFactory    WorkspaceFlowListTaskFactory    // ListFlowTaskFactory builds a task which can be used with an instance of WorkspaceFlowListParams to list the flows available under a Workspace.
		ResolveFlowTaskFactory WorkspaceFlowResolveTaskFactory // ResolveFlowTaskFactory builds a task which can be used with an instance of WorkspaceFlowResolveParams to find a workflow in a solution by handle.

		ListNodeTaskFactory    WorkspaceNodeListTaskFactory    // ListNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeListParams to list the nodes under a workspace flow.
		ResolveNodeTaskFactory WorkspaceNodeResolveTaskFactory // ResolveNodeTaskFactory build a task which can be used with an instance of WorkspaceNodeResolveParams to find the Node Id of a workspace node.

		ListSetNodeTaskFactory    WorkspaceNodeSetListTaskFactory    // ListSetNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeSetListParams to list the data sets under a workspace node.
		ResolveSetNodeTaskFactory WorkspaceNodeSetResolveTaskFactory // ResolveSetNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeSetResolveParams to find an existing dataset on a workspace node by port handle.
		UpdateSetNodeTaskFactory  WorkspaceNodeSetUpdateTaskFactory  // UpdateSetNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeSetResolveParams to update a node with a dataSet id and a port handle

		AddSourceNodeTaskFactory    WorkspaceNodeSourceAddTaskFactory    // AddSourceNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeSourceParams to add a data source to a workspace node.
		RemoveSourceNodeTaskFactory WorkspaceNodeSourceRemoveTaskFactory // RemoveSourceNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeSourceParams to remove a data source from a workspace node.

		AddStoreNodeTaskFactory    WorkspaceNodeStoreAddTaskFactory    // AddStoreNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeStoreParams to add a data store to a workspace node.
		RemoveStoreNodeTaskFactory WorkspaceNodeStoreRemoveTaskFactory // RemoveStoreNodeTaskFactory builds a task which can be used with an instance of WorkspaceNodeStoreParams to remove a data store from a workspace node.
	}
)

type WorkspaceCreateTask = *task.Task[WorkspaceCreateParams]
type WorkspaceCreateTaskFactory = func() WorkspaceCreateTask
type WorkspaceListTask = *task.Task[WorkspaceListParams]
type WorkspaceListTaskFactory = func() WorkspaceListTask
type WorkspaceResolveTask = *task.Task[WorkspaceFlowResolveParams]
type WorkspaceResolveTaskFactory = func() WorkspaceResolveTask

type WorkspaceFlowCreateTask = *task.Task[WorkspaceFlowCreateParams]
type WorkspaceFlowCreateTaskFactory = func() WorkspaceFlowCreateTask
type WorkspaceFlowListTask = *task.Task[WorkspaceFlowListParams]
type WorkspaceFlowListTaskFactory = func() WorkspaceFlowListTask
type WorkspaceFlowResolveTask = *task.Task[WorkspaceFlowResolveParams]
type WorkspaceFlowResolveTaskFactory = func() WorkspaceFlowResolveTask

type WorkspaceNodeListTask = *task.Task[WorkspaceNodeListParams]
type WorkspaceNodeListTaskFactory = func() WorkspaceNodeListTask
type WorkspaceNodeResolveTask = *task.Task[WorkspaceNodeResolveParams]
type WorkspaceNodeResolveTaskFactory = func() WorkspaceNodeResolveTask

type WorkspaceNodeSetListTask = *task.Task[WorkspaceNodeSetListParams]
type WorkspaceNodeSetListTaskFactory = func() WorkspaceNodeSetListTask
type WorkspaceNodeSetResolveTask = *task.Task[WorkspaceNodeSetResolveParams]
type WorkspaceNodeSetResolveTaskFactory = func() WorkspaceNodeSetResolveTask
type WorkspaceNodeSetUpdateTask = *task.Task[WorkspaceNodeSetResolveParams]
type WorkspaceNodeSetUpdateTaskFactory = func() WorkspaceNodeSetUpdateTask

type WorkspaceNodeSourceAddTask = *task.Task[WorkspaceNodeSourceParams]
type WorkspaceNodeSourceAddTaskFactory = func() WorkspaceNodeSourceAddTask
type WorkspaceNodeSourceRemoveTask = *task.Task[WorkspaceNodeSourceParams]
type WorkspaceNodeSourceRemoveTaskFactory = func() WorkspaceNodeSourceRemoveTask

type WorkspaceNodeStoreAddTask = *task.Task[WorkspaceNodeStoreParams]
type WorkspaceNodeStoreAddTaskFactory = func() WorkspaceNodeStoreAddTask
type WorkspaceNodeStoreRemoveTask = *task.Task[WorkspaceNodeStoreParams]
type WorkspaceNodeStoreRemoveTaskFactory = func() WorkspaceNodeStoreRemoveTask

type WorkspaceCreateParams struct {
	Broker
	Workspace *Workspace
}

type WorkspaceFlowCreateParams struct {
	*Broker
	WorkspaceId *int64
	WorkflowId  *int64
	Name        string
	Description string
}

func (cp WorkspaceFlowCreateParams) IsValid() bool {
	return cp.WorkspaceId != nil && cp.WorkflowId != nil
}

type WorkspaceFlowListParams struct {
	*WorkspaceFlowResolveParams
	ReadyOnly bool
}

func (lp WorkspaceFlowListParams) GetMaskFlags() (int, int) {
	return getWorkspaceFlowListFlags(lp.ReadyOnly)
}

func (lp WorkspaceFlowListParams) GetWorkspaceId() *int64 {
	if lp.WorkspaceFlowResolveParams == nil || lp.WorkspaceFlowCreateParams == nil {
		return nil
	}

	return lp.WorkspaceFlowResolveParams.WorkspaceId
}

func (lp WorkspaceFlowListParams) IsValid() bool {
	if lp.WorkspaceFlowResolveParams == nil || lp.WorkspaceFlowCreateParams == nil {
		return false
	}

	return lp.WorkspaceId != nil
}

type WorkspaceFlowResolveParams struct {
	*WorkspaceFlowCreateParams
	WorkspaceName   string
	SolutionOem     string
	SolutionHandle  string
	SolutionVersion string
	WorkflowHandle  string
	RcEnabled       bool
}

func (rp WorkspaceFlowResolveParams) GetMaskFlags() (int, int) {
	return getWorkspaceListFlags(rp.RcEnabled)
}

func (rp WorkspaceFlowResolveParams) HasWorkflowId() bool {
	return rp.WorkspaceFlowCreateParams != nil && rp.WorkflowId != nil
}

func (rp WorkspaceFlowResolveParams) HasWorkspaceId() bool {
	return rp.WorkspaceFlowCreateParams != nil && rp.WorkspaceId != nil
}

type WorkspaceListParams struct {
	Broker
	FromDate  *time.Time
	OwnerOnly bool
	RcEnabled bool
}

func (wlp WorkspaceListParams) GetMaskFlags() (int, int) {
	return getWorkspaceListFlags(wlp.RcEnabled)
}

type WorkspaceNodeListParams struct {
	*Broker
	WorkspaceName   string
	WorkspaceId     *int64
	WorkspaceFlowId *int64
	WorkflowHandle  string
}

type WorkspaceNodeResolveParams struct {
	Broker
	workspace.NodeResolveParams
}

type WorkspaceNodeSetListParams struct {
	Broker
	*workspace.NodeResolveParams
	Sealed bool
}

func (p WorkspaceNodeSetListParams) getNodeId() *int64 {
	if p.NodeResolveParams == nil {
		return nil
	}

	return p.NodeResolveParams.NodeId
}

type WorkspaceNodeSetResolveParams struct {
	Broker
	workspace.NodeSetParams
	*workspace.NodeResolveParams
}

func (p WorkspaceNodeSetResolveParams) getNodeId() *int64 {
	if p.NodeResolveParams == nil {
		return nil
	}

	return p.NodeResolveParams.NodeId
}

type WorkspaceNodeSourceParams struct {
	*DataSourceResolveParams
	NodeId *int64
}

type WorkspaceNodeStoreParams struct {
	*DataStoreResolveParams
	NodeId *int64
}

func getWorkspaceFlowListFlags(readOnly bool) (int, int) {
	if readOnly {
		return WorkspaceFlowFlags.Active | WorkspaceFlowFlags.Ready,
			WorkspaceFlowFlags.Active | WorkspaceFlowFlags.Ready
	}

	return WorkspaceFlowFlags.Active, WorkspaceFlowFlags.Active
}

func getWorkspaceListFlags(rcEnabled bool) (int, int) {
	if rcEnabled {
		// see the orchestrator API
		return WorkspaceFlags.Active | WorkspaceFlags.RcEnabled,
			WorkspaceFlags.Active | WorkspaceFlags.RcEnabled
	}

	return WorkspaceFlags.Active | WorkspaceFlags.RcEnabled,
		WorkspaceFlags.Active
}

func handleWorkspaceListContext(params *WorkspaceListParams, state *task.State) error {
	if state.Output == "" {
		var brokerClient Client
		var err error

		if params.FromDate != nil && time.Now().Before(*params.FromDate) {
			return workspace.Errors.WorkspaceInvalidNco()
		}

		if params.OwnerOnly {
			if brokerClient, err = params.GetClient(); err == nil {
				if brokerClient.GetUserId() > 0 {
					return nil
				}

				return workspace.Errors.WorkspaceInvalidOwner()
			}
		}

		return err
	}

	return nil
}

func handleWorkspaceListComplete(params *WorkspaceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var mask, flag = params.GetMaskFlags()
		var workspaces []Workspace

		if workspaces, err = brokerClient.ListWorkspaces(mask, flag); err == nil {
			var results = workspaces

			if params.OwnerOnly {
				var ownerId = brokerClient.GetUserId()
				var filtered []Workspace

				state.Logger.Debugf("Filtering workspaces by owner [%d]", ownerId)

				for _, ws := range results {
					if ownerId == ws.OwnerUserId {
						filtered = append(filtered, ws)
					}
				}

				results = filtered
			}

			if params.FromDate != nil {
				var filtered []Workspace

				state.Logger.Debugf("Filtering workspaces after date [%s]", params.FromDate.Format(time.DateOnly))

				for _, ws := range results {
					if time.UnixMilli(ws.Created).After(*params.FromDate) {
						filtered = append(filtered, ws)
					}
				}

				results = filtered
			}

			state.Output = cast.ToString(len(results))
			state.Internal = results
			return nil
		}

		return err
	}

	return err
}

func handleWorkspaceListPretend(params *WorkspaceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var mask, flags = params.GetMaskFlags()
		var loggingSuffix string

		if params.FromDate != nil {
			loggingSuffix = fmt.Sprintf(" after date [%s]", params.FromDate.Format(time.DateOnly))
		}

		state.Logger.Debugf("Pretending to list workspaces%s", loggingSuffix)
		fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspacesUrl(), mask, flags)
		return nil
	}

	return err
}

func handleWorkspaceCreateContext(params *WorkspaceCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.Workspace != nil {
			var nameSuffix string

			if params.Workspace.Visibility == "" {
				return workspace.Errors.WorkspaceVisibilityRequired()
			}

			if params.Workspace.Name == "" {
				state.Logger.Warn("Nameless workspaces are allowed, but not recommended")
			} else {
				nameSuffix = fmt.Sprintf(" with name [%s]", params.Workspace.Name)
			}

			if params.Workspace.IsRcEnabled() {
				state.Logger.Debugf("Creating a development workspace%s", nameSuffix)
			} else {
				state.Logger.Debugf("Creating a production workspace%s", nameSuffix)
			}

			return nil
		}

		return workspace.Errors.WorkspaceDefInvalid()
	}

	return nil
}

func handleWorkspaceCreateComplete(params *WorkspaceCreateParams, state *task.State) error {
	if params.Workspace != nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var created *Workspace

			state.Logger.Debugf("Workspace created on host [%s]", brokerClient.GetHostAddr())

			if created, err = brokerClient.CreateWorkspace(params.Workspace); err == nil {
				var formatter = timez.NewTodayFormatter()

				state.Logger.Debugf("Workspace id [%d] created on [%s]", created.Id, formatter.FormatMillis(created.Created))
				state.Reportf("Created workspace id [%d]", created.Id)
				state.Output = cast.ToString(created.Id)
				state.Internal = created
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceDefInvalid()
}

func handleWorkspaceCreatePretend(params *WorkspaceCreateParams, state *task.State) error {
	if params.Workspace != nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending creating workspace [%s]", params.Workspace.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d name=%s\\\n", params.Workspace.Name)
			fmt.Printf("  -d description=%s\\\n", params.Workspace.Description)
			fmt.Printf("  -d visibility=%s\\\n", params.Workspace.Visibility)
			fmt.Printf("  -d rcEnabled=%s\\\n", cast.ToString(params.Workspace.RcEnabled))
			fmt.Printf("%s\n", brokerClient.CreateWorkspaceUrl())
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceDefInvalid()
}

func handleWorkspaceFlowCreateContext(params *WorkspaceFlowCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkspaceId == nil {
			return workspace.Errors.WorkspaceIdRequired()
		}

		if params.WorkflowId == nil {
			return workspace.Errors.WorkflowIdRequired()
		}
	}

	return nil
}

func handleWorkspaceFlowCreateComplete(params *WorkspaceFlowCreateParams, state *task.State) error {
	if params.IsValid() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var wsId = *params.WorkspaceId
			var wfId = *params.WorkflowId
			var name = params.Name
			var desc = params.Description
			var flow *WorkspaceFlow

			if params.Name == "" {
				state.Logger.Debugf("Creating nameless workspace flow for workspace [%d], workflow [%d]", wsId, wfId)
			} else {
				state.Logger.Debugf("Creating workspace flow [%s] under workspace [%d] for workflow [%d]", name, wsId, wfId)
			}

			if flow, err = brokerClient.CreateWorkspaceFlow(wsId, wfId, name, desc); err == nil {
				state.Logger.Debugf("Created workspace flow [%d]", flow.Id)
				state.Reportf("Workspace flow [%d:%s] created", flow.Id, flow.Name)
				state.Internal = flow
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceFlowDefInvalid()
}

func handleWorkspaceFlowCreatePretend(params *WorkspaceFlowCreateParams, state *task.State) error {
	if params.IsValid() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending creating workspace flow [%s]", params.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d workspaceId=%d\\\n", *params.WorkspaceId)
			fmt.Printf("  -d workflowId=%d\\\n", *params.WorkflowId)
			fmt.Printf("  -d name=%s\\\n", params.Name)
			fmt.Printf("  -d description=%s\\\n", cast.ToString(params.Description))
			fmt.Printf("%s\n", brokerClient.CreateWorkspaceFlowUrl())
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceFlowDefInvalid()
}

func handleWorkspaceFlowListComplete(params *WorkspaceFlowListParams, state *task.State) error {
	if params.IsValid() {
		var workspaceId = *params.GetWorkspaceId()
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var mask, flags = params.GetMaskFlags()
			var flows []WorkspaceFlow

			if params.ReadyOnly {
				state.Logger.Debugf("Finding ready workspace flows for workspace [%d]", params.WorkspaceId)
			} else {
				state.Logger.Debugf("Finding active workspace flows for workspace [%d]", params.WorkspaceId)
			}

			if flows, err = brokerClient.ListWorkspaceFlows(workspaceId, mask, flags); err == nil {
				state.Logger.Debugf("Found %d workspace flows under workspace id [%d]", len(flows), workspaceId)
				state.Internal = flows
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceIdRequired()
}

func handleWorkspaceFlowListContext(params *WorkspaceFlowListParams, state *task.State) error {
	if state.Output == "" {
		if params.GetWorkspaceId() == nil {
			return workspace.Errors.WorkspaceIdRequired()
		}

		return nil
	}

	return nil
}

func handleWorkspaceFlowListPretend(params *WorkspaceFlowListParams, state *task.State) error {
	if params.IsValid() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var mask, flags = params.GetMaskFlags()

			state.Logger.Debugf("Pretending to resolve workspace flows with a list request")
			fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d workspaceId=%d", *params.WorkspaceId)
			fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspaceFlowsUrl(), mask, flags)
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceIdRequired()
}

func handleWorkspaceFlowSolutionComplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if !params.HasWorkflowId() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var solution *Solution

			if solution, err = brokerClient.FindSolution(params.SolutionOem, params.SolutionHandle, params.SolutionVersion); err == nil {
				state.Logger.Debugf("Found solution id [%d]", *solution.Id)

				for _, wf := range solution.Workflows {
					if wf.Handle == params.WorkflowHandle {
						params.WorkflowId = wf.Id
						return nil
					}
				}

				return ErrorWorkflowNotFound
			}
		}

		return err
	}

	return nil
}

func handleWorkspaceFlowSolutionContext(params *WorkspaceFlowResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.HasWorkflowId() {
			state.Logger.Warnf("Workflow id is already known: [%d]", *params.WorkflowId)
			return workspace.Errors.WorkflowIdKnown()
		}

		if params.SolutionOem == "" {
			return workspace.Errors.WorkflowOemRequired()
		}

		if params.SolutionHandle == "" {
			return workspace.Errors.WorkflowHandleRequired()
		}

		if params.SolutionVersion == "" {
			return workspace.Errors.WorkflowVersionRequired()
		}
	}

	return nil
}

func handleWorkspaceFlowSolutionIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, workspace.Errors.WorkflowIdKnown()) {
		state.Logger.Debugf("Resolved workflow id [%d]", *params.WorkflowId)
		state.Completed = true
		state.Output = ""
		return nil
	}

	return state.Error
}

func handleWorkspaceFlowSolutionPretend(params *WorkspaceFlowResolveParams, state *task.State) error {
	if params.WorkflowId == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to resolve solution by oem, handle and version")
			fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("%s?oem=%s&handle=%s&version=%s\n", brokerClient.FindSolutionUrl(),
				params.SolutionOem, params.SolutionHandle, params.SolutionVersion)
		}

		return err
	}

	return nil
}

func handleWorkspaceNodeListComplete(params *WorkspaceNodeListParams, state *task.State) error {
	if params.WorkspaceFlowId != nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var nodes []WorkspaceNode

			state.Logger.Debugf("Listing workspace flow nodes for flow id [%d]", params.WorkspaceFlowId)

			if nodes, err = brokerClient.ListWorkspaceNodes(*params.WorkspaceFlowId); err == nil {
				state.Logger.Debugf("Found %d nodes under workspace flow id [%d]",
					len(nodes), *params.WorkspaceFlowId)
				state.Internal = nodes
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceFlowRequired()
}

func handleWorkspaceNodeListContext(params *WorkspaceNodeListParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkflowHandle == "" && params.WorkspaceFlowId == nil {
			return workspace.Errors.WorkspaceFlowRequired()
		}

		if params.WorkspaceFlowId == nil {
			return workspace.Errors.WorkspaceFlowUnresolved()
		}
	}

	return nil
}

func handleWorkspaceNodeListIncomplete(params *WorkspaceNodeListParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, workspace.Errors.WorkspaceFlowUnresolved()) {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Workspace flow is unknown")

		if brokerClient, err = params.GetClient(); err == nil {
			var flows []WorkspaceFlow
			var result []WorkspaceFlow

			if params.WorkspaceId == nil {
				var workspaces []Workspace
				var results []Workspace

				state.Logger.Debugf("Locating workspace [%s]", params.WorkspaceName)

				if workspaces, err = brokerClient.ListWorkspaces(WorkspaceFlags.Active, WorkspaceFlags.Active); err == nil {
					for _, ws := range workspaces {
						if strings.EqualFold(params.WorkspaceName, ws.Name) {
							results = append(results, ws)
						}
					}
				} else {
					return err
				}

				if len(results) == 0 {
					return workspace.Errors.WorkspaceNotFound()
				} else if len(results) > 1 {
					return workspace.Errors.WorkspaceConflict()
				}

				params.WorkspaceId = new(results[0].Id)
			}

			state.Logger.Debugf("Listing workspace flows for workspace [%d]", *params.WorkspaceId)

			if flows, err = brokerClient.ListWorkspaceFlows(*params.WorkspaceId, WorkspaceFlags.Active, WorkspaceFlags.Active); err == nil {
				for _, fl := range flows {
					if fl.Solution != nil {
						var wf *Workflow

						if wf, err = fl.Solution.FindWorkflowByHandle(params.WorkflowHandle); err == nil {
							if wf.Id != nil && *wf.Id == fl.WorkflowId {
								result = append(result, fl)
							}
						} else {
							break
						}
					}
				}
			}

			if err == nil {
				if len(result) == 1 {
					state.Logger.Debugf("Located workspace flow [%d]", result[0].Id)
					params.WorkspaceFlowId = new(result[0].Id)
					state.Completed = false
					return nil
				} else if len(result) > 1 {
					return workspace.Errors.WorkspaceFlowConflict()
				}

				return workspace.Errors.WorkspaceFlowRequired()
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceFlowRequired()
}

func handleWorkspaceNodeListPretend(params *WorkspaceNodeListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		if errors.Is(state.Error, workspace.Errors.WorkspaceFlowUnresolved()) {
			if params.WorkspaceId == nil {
				state.Logger.Debugf("Pretending to list workspaces to resolve a workspace by name")
				fmt.Printf("curl -X GET \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspacesUrl(), WorkspaceFlags.Active, WorkspaceFlags.Active)
			}

			state.Logger.Debugf("Pretending to list workspace flows for the workspace id")
			fmt.Printf("curl -X GET \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d workspaceId=%s", "[workspaceId]")
			fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspaceFlowsUrl(), WorkspaceFlags.Active, WorkspaceFlags.Active)
		}

		state.Logger.Debugf("Pretending to list workspace flow nodes")
		fmt.Printf("curl -X GET  \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("  -G -d workspaceFlowId=%s", "[workspaceFlowId]")
		fmt.Println(brokerClient.ListWorkspaceNodesUrl())
	}

	return err
}

func handleWorkspaceNodeResolveComplete(params *WorkspaceNodeResolveParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Resolving node from node name [%s]", params.FnHandle)

		if brokerClient, err = params.GetClient(); err == nil {
			var nodes []WorkspaceNode

			if params.FlowId == nil {
				var flows []WorkspaceFlow
				var mask, flags int

				if params.WorkspaceId == nil {
					var workspaces []Workspace

					state.Logger.Debugf("Finding workspace for name [%s]", params.WorkspaceName)
					mask, flags = getWorkspaceListFlags(true)

					if workspaces, err = brokerClient.ListWorkspaces(mask, flags); err != nil {
						return err
					}

					for _, ws := range workspaces {
						if ws.Name == params.WorkspaceName {
							params.WorkspaceId = new(ws.Id)
							break
						}
					}
				}

				if params.WorkspaceId != nil {
					mask, flags = getWorkspaceFlowListFlags(true)

					state.Logger.Debugf("Finding workspace flow for workflow handle [%s]", params.WorkflowHandle)

					if flows, err = brokerClient.ListWorkspaceFlows(*params.WorkspaceId, mask, flags); err != nil {
						return err
					}

					for _, fl := range flows {
						if wf, _ := fl.Solution.FindWorkflowByHandle(params.WorkflowHandle); wf != nil {
							params.FlowId = new(fl.Id)
							break
						}
					}
				} else {
					return workspace.Errors.WorkspaceNotFound()
				}
			}

			if params.FlowId != nil {
				state.Logger.Debugf("Finding workspace flow node for smart function handle [%s]", params.FnHandle)

				if nodes, err = brokerClient.ListWorkspaceNodes(*params.FlowId); err != nil {
					return err
				}

				for _, node := range nodes {
					if node.SmartFunction != nil && strings.EqualFold(node.SmartFunction.Handle, params.FnHandle) {
						params.NodeId = new(node.Id)
						break
					}
				}
			} else {
				return workspace.Errors.WorkspaceFlowRequired()
			}

			state.Output = ""
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeFnHandleRequired()
}

func handleWorkspaceNodeResolveContext(params *WorkspaceNodeResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.NodeId != nil {
			return workspace.Errors.WorkspaceNodeKnown()
		}

		if params.FnHandle == "" {
			return workspace.Errors.WorkspaceNodeFnHandleRequired()
		}

		if params.FlowId == nil && params.WorkflowHandle == "" {
			return workspace.Errors.WorkflowHandleRequired()
		}

		if params.WorkspaceId == nil && params.WorkspaceName == "" {
			return workspace.Errors.WorkspaceNameRequired()
		}

		state.Output = params.FnHandle
	}

	return nil
}

func handleWorkspaceNodeResolveIncomplete(params *WorkspaceNodeResolveParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, workspace.Errors.WorkspaceNodeKnown()) {
		state.Logger.Debugf("Node id provided [%d], skipping resolution", *params.NodeId)
		return nil
	}

	return state.Error
}

func handleWorkspaceNodeResolvePretend(params *WorkspaceNodeResolveParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var flowId = params.FlowId
			var mask, flags int

			if params.WorkspaceId == nil {
				mask, flags = getWorkspaceListFlags(true)
				state.Logger.Debugf("Pretending to find a workspace with name: [%s]", params.WorkspaceName)
				fmt.Printf("curl -X GET \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("  -G -d mask=%d\\\n", mask)
				fmt.Printf("  -d flags=%d\\\n", flags)
				fmt.Printf("%s\n", brokerClient.ListWorkspacesUrl())
			}

			if flowId == nil {
				mask, flags = getWorkspaceFlowListFlags(true)
				flowId = new(rand.Int64())
				state.Logger.Debugf("Pretending to find a workspace flow with workflow handle: [%s]", params.WorkflowHandle)
				fmt.Printf("curl -X GET \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("  -G -d mask=%d\\\n", mask)
				fmt.Printf("  -d flags=%d\\\n", flags)
				fmt.Printf("%s\n", brokerClient.ListWorkspaceFlowsUrl())
			}

			state.Logger.Debugf("Pretending to find a workspace flow node with smart function handle: [%s]", params.FnHandle)
			fmt.Printf("curl -X GET \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=%d\\\n", *flowId)
			fmt.Printf("%s\n", brokerClient.ListWorkspaceNodesUrl())
			return nil
		}

		return err
	}

	if errors.Is(state.Error, workspace.Errors.WorkspaceNodeKnown()) {
		state.Logger.Debugf("Node id provided [%d], skipping resolution", *params.NodeId)
		return nil
	}

	return state.Error
}

func handleWorkspaceNodeSetListComplete(params *WorkspaceNodeSetListParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Listing workspace node sets and ports for node [%s]", state.Output)

		if brokerClient, err = params.GetClient(); err == nil {
			var node *WorkspaceNode

			if node, err = brokerClient.InspectNode(*params.getNodeId()); err == nil {
				state.Logger.Debugf("Retrieved [%d] ports from node [%d]", len(node.InputPorts), node.Id)
				state.Internal = node.InputPorts
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeRequired()
}

func handleWorkspaceNodeSetListContext(params *WorkspaceNodeSetListParams, state *task.State) error {
	if state.Output == "" {
		var nodeId = params.getNodeId()

		if nodeId == nil {
			return workspace.Errors.WorkspaceNodeRequired()
		}

		state.Output = cast.ToString(*nodeId)
	}

	return nil
}

func handleWorkspaceNodeSetListPretend(params *WorkspaceNodeSetListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		state.Logger.Debugf("Pretending to list data sets for node [%s]", state.Output)
		fmt.Printf("genaiz raw get %s \\\n", brokerClient.InspectNodeUrl())
		fmt.Printf("  -f id=%d\\\n", *params.getNodeId())
		fmt.Printf("  -r .inputPorts")
		return nil
	}

	return err
}

func handleWorkspaceNodeSetResolveComplete(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	if state.Output != "" {
		var nodeId = *params.getNodeId()
		var portHandle = params.PortHandle
		var brokerClient Client
		var err error

		state.Logger.Debugf("Resolving port handle [%s] for node id [%d]", portHandle, nodeId)

		if brokerClient, err = params.GetClient(); err == nil {
			var node *WorkspaceNode

			if node, err = brokerClient.InspectNode(nodeId); err == nil {
				if port := node.FindInputPort(portHandle); port != nil {
					params.DataSetId = port.DataSetId

					if port.DataSet != nil {
						state.Internal = *port.DataSet
					}

					state.Output = ""
					return nil
				}

				return workspace.Errors.WorkspaceNodeFnPortNotFound()
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeFnPortRequired()
}

func handleWorkspaceNodeSetResolveContext(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.DataSetId != nil {
			return workspace.Errors.WorkspaceNodeSetKnown()
		}

		if params.getNodeId() == nil {
			return workspace.Errors.WorkspaceNodeRequired()
		}

		if params.PortHandle == "" {
			return workspace.Errors.WorkspaceNodeFnPortRequired()
		}

		state.Output = params.PortHandle
	}

	return nil
}

func handleWorkspaceNodeSetResolveIncomplete(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, workspace.Errors.WorkspaceNodeSetKnown()) {
		state.Logger.Debugf("DataSet id provided [%d], skipping resolution", *params.DataSetId)
		return nil
	}

	return state.Error
}

func handleWorkspaceNodeSetResolvePretend(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var nodeId = *params.getNodeId()

			state.Logger.Debugf("Pretending to resolve dataset on port [%s] on node [%d]", params.PortHandle, nodeId)
			fmt.Printf("genaiz raw get %s \\\n", brokerClient.InspectNodeUrl())
			fmt.Printf("  -f id=%d\\\n", nodeId)
			fmt.Printf("  -r .inputPorts[port == '%s']\n", params.PortHandle)
			return nil
		}

		return err
	}

	if errors.Is(state.Error, workspace.Errors.WorkspaceNodeSetKnown()) {
		state.Logger.Debugf("DataSet id provided [%d], skipping creation", *params.DataSetId)
		return nil
	}

	return state.Error
}

func handleWorkspaceNodeSetUpdateComplete(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var nodeId = *params.getNodeId()
			var dataSetId = *params.DataSetId
			var portHandle = params.PortHandle
			var updated *WorkspaceNode

			state.Logger.Debugf("Assigning data set [%s] to port [%s]", state.Output, portHandle)

			if updated, err = brokerClient.UpdateNodeInput(nodeId, dataSetId, portHandle); err == nil {
				state.Logger.Debugf("Updated node [%d]", nodeId)
				state.Reportf("Updated workspace node [%d] input port [%s] data set [%d]", updated.Id, portHandle, dataSetId)
				state.Internal = *updated
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeSetRequired()
}

func handleWorkspaceNodeSetUpdateContext(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	if state.Output == "" {

		if params.DataSetId == nil {
			return workspace.Errors.WorkspaceNodeSetRequired()
		}

		if params.getNodeId() == nil {
			return workspace.Errors.WorkspaceNodeRequired()
		}

		state.Output = cast.ToString(*params.DataSetId)
	}

	return nil
}

func handleWorkspaceNodeSetUpdatePretend(params *WorkspaceNodeSetResolveParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var nodeId = *params.getNodeId()
		var dataSetId = *params.DataSetId
		var portHandle = params.PortHandle

		state.Logger.Debugf("Pretending to assign data set [%d] to port [%s]", dataSetId, portHandle)
		fmt.Printf("genaiz raw get %s \\\n", brokerClient.UpdateNodeInputUrl())
		fmt.Printf("  -f id=%d\\\n", nodeId)
		fmt.Printf("  -f handle=%s\\\n", portHandle)
		fmt.Printf("  -f dataSetId=%d\\\n", dataSetId)
		fmt.Println("  -r .")
		return nil
	}

	return err
}

func handleWorkspaceNodeSourceAddComplete(params *WorkspaceNodeSourceParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Updating node [%d] adding data source [%d]", *params.NodeId, *params.DataSourceId)

		if brokerClient, err = params.GetClient(); err == nil {
			var currentNode *WorkspaceNode

			if currentNode, err = brokerClient.GetNode(*params.NodeId); err == nil {
				if !slices.Contains(currentNode.DataSourceIds, *params.DataSourceId) {
					var updated = currentNode.AddDataSource(*params.DataSourceId)

					if updated, err = brokerClient.UpdateNode(updated); err != nil {
						return err
					}

					state.Logger.Debugf("Updated node [%d]", *params.NodeId)
					state.Reportf("Added data source [%d] from node [%d]", *params.DataSourceId, *params.NodeId)
					state.Internal = *updated
					return nil
				}

				state.Logger.Warnf("Data source [%d] is already configured for node [%d]", *params.DataSourceId, *params.NodeId)
				state.Internal = *currentNode
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeSourceRequired()
}

func handleWorkspaceNodeSourceContext(params *WorkspaceNodeSourceParams, state *task.State) error {
	if state.Output == "" {
		if params.NodeId == nil {
			return workspace.Errors.WorkspaceNodeRequired()
		}

		if params.DataSourceId == nil {
			return workspace.Errors.WorkspaceNodeSourceRequired()
		}

		state.Output = cast.ToString(*params.DataSourceId)
	}

	return nil
}

func handleWorkspaceNodeSourcePretend(params *WorkspaceNodeSourceParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to get workspace node id: [%d]", *params.NodeId)
			fmt.Printf("curl -X GET \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=%d \\\n", *params.NodeId)
			fmt.Printf("%s\n", brokerClient.GetNodeUrl())
			state.Logger.Debugf("Pretending to update workspace node id: [%d]", *params.NodeId)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=%d \\\n", *params.NodeId)
			fmt.Println("   -d dataSourceIds=$SOURCE_IDS \\")
			fmt.Printf("%s\n", brokerClient.UpdateNodeUrl())
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeSourceRequired()
}

func handleWorkspaceNodeSourceRemoveComplete(params *WorkspaceNodeSourceParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Updating node [%d] removing data source [%d]", *params.NodeId, *params.DataSourceId)

		if brokerClient, err = params.GetClient(); err == nil {
			var currentNode *WorkspaceNode

			if currentNode, err = brokerClient.GetNode(*params.NodeId); err == nil {
				if slices.Contains(currentNode.DataSourceIds, *params.DataSourceId) {
					var updated = currentNode.RemoveDataSource(*params.DataSourceId)

					if updated, err = brokerClient.UpdateNode(updated); err != nil {
						return err
					}

					state.Logger.Debugf("Updated node [%d]", *params.NodeId)
					state.Reportf("Removed data source [%d] from node [%d]", *params.DataSourceId, *params.NodeId)
					state.Internal = *updated
					return nil
				}

				state.Logger.Warnf("Data source [%d] is not configured for node [%d]", *params.DataSourceId, *params.NodeId)
				state.Internal = *currentNode
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeSourceRequired()
}

func handleWorkspaceNodeStoreAddComplete(params *WorkspaceNodeStoreParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Updating node [%d] adding data store [%d]", *params.NodeId, *params.DataStoreId)

		if brokerClient, err = params.GetClient(); err == nil {
			var currentNode *WorkspaceNode

			if currentNode, err = brokerClient.GetNode(*params.NodeId); err == nil {
				if !slices.Contains(currentNode.DataStoreIds, *params.DataStoreId) {
					var updated = currentNode.AddDataStore(*params.DataStoreId)

					if updated, err = brokerClient.UpdateNode(updated); err != nil {
						return err
					}

					state.Logger.Debugf("Updated node [%d]", *params.NodeId)
					state.Reportf("Added data store [%d] from node [%d]", *params.DataStoreId, *params.NodeId)
					state.Internal = *updated
					return nil
				}

				state.Logger.Warnf("Data store [%d] is already configured for node [%d]", *params.DataStoreId, *params.NodeId)
				state.Internal = *currentNode
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeStoreRequired()
}

func handleWorkspaceNodeStoreContext(params *WorkspaceNodeStoreParams, state *task.State) error {
	if state.Output == "" {
		if params.NodeId == nil {
			return workspace.Errors.WorkspaceNodeRequired()
		}

		if params.DataStoreId == nil {
			return workspace.Errors.WorkspaceNodeStoreRequired()
		}

		state.Output = cast.ToString(*params.DataStoreId)
	}

	return nil
}

func handleWorkspaceNodeStorePretend(params *WorkspaceNodeStoreParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to get workspace node id: [%d]", *params.NodeId)
			fmt.Printf("curl -X GET \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=%d \\\n", *params.NodeId)
			fmt.Printf("%s\n", brokerClient.GetNodeUrl())
			state.Logger.Debugf("Pretending to update workspace node id: [%d]", *params.NodeId)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d id=%d \\\n", *params.NodeId)
			fmt.Println("   -d dataSourceIds=$STORE_IDS \\")
			fmt.Printf("%s\n", brokerClient.UpdateNodeUrl())
			return nil
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeStoreRequired()
}

func handleWorkspaceNodeStoreRemoveComplete(params *WorkspaceNodeStoreParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Updating node [%d] removing data store [%d]", *params.NodeId, *params.DataStoreId)

		if brokerClient, err = params.GetClient(); err == nil {
			var currentNode *WorkspaceNode

			if currentNode, err = brokerClient.GetNode(*params.NodeId); err == nil {
				if slices.Contains(currentNode.DataStoreIds, *params.DataStoreId) {
					var updated = currentNode.RemoveDataStore(*params.DataStoreId)

					if updated, err = brokerClient.UpdateNode(updated); err != nil {
						return err
					}

					state.Logger.Debugf("Updated node [%d]", *params.NodeId)
					state.Reportf("Removed data store [%d] from node [%d]", *params.DataStoreId, *params.NodeId)
					state.Internal = *updated
					return nil
				}

				state.Logger.Warnf("Data store [%d] is not configured for node [%d]", *params.DataStoreId, *params.NodeId)
				state.Internal = *currentNode
				return nil
			}
		}

		return err
	}

	return workspace.Errors.WorkspaceNodeStoreRequired()
}

func handleWorkspaceResolveComplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if !params.HasWorkspaceId() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var workspaces []Workspace

			if params.RcEnabled {
				state.Logger.Debugf("Finding release candidate enabled workspaces for name [%s]", params.WorkspaceName)
			} else {
				state.Logger.Debugf("Finding workspaces for name [%s]", params.WorkspaceName)
			}

			if workspaces, err = brokerClient.ListWorkspaces(params.GetMaskFlags()); err == nil {
				if i := slices.IndexFunc(workspaces, func(workspace Workspace) bool {
					return strings.EqualFold(workspace.Name, params.WorkspaceName)
				}); i >= 0 {
					state.Logger.Debugf("Found workspace id [%d]", workspaces[i].Id)
					params.WorkspaceId = &workspaces[i].Id
					return nil
				}
			}
		}

		return err
	}

	return nil
}

func handleWorkspaceResolveContext(params *WorkspaceFlowResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.HasWorkspaceId() {
			state.Logger.Warnf("Workspace id is already known: [%d]", *params.WorkspaceId)
			return workspace.Errors.WorkspaceIdKnown()
		}

		if params.WorkspaceName == "" {
			return workspace.Errors.WorkspaceNameRequired()
		}
	}

	return nil
}

func handleWorkspaceResolveIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, workspace.Errors.WorkspaceIdKnown()) {
		state.Logger.Debugf("Resolved workspace id [%d]", *params.WorkspaceId)
		state.Completed = true
		state.Output = ""
		return nil
	}

	return state.Error
}

func handleWorkspaceResolvePretend(params *WorkspaceFlowResolveParams, state *task.State) error {
	if !params.HasWorkspaceId() {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var mask, flags = params.GetMaskFlags()

			state.Logger.Debugf("Pretending to resolve workspace by name with a list request")
			fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspacesUrl(), mask, flags)
			return nil
		}

		return err
	}

	return nil
}

func init() {
	WorkspaceTasks.CreateTaskFactory = newWorkspaceCreateTask
	WorkspaceTasks.ListTaskFactory = newWorkspaceListTask
	WorkspaceTasks.ResolveTaskFactory = newWorkspaceResolveTask
	WorkspaceTasks.CreateFlowTaskFactory = newWorkspaceFlowCreateTask
	WorkspaceTasks.ListFlowTaskFactory = newWorkspaceFlowListTask
	WorkspaceTasks.ResolveFlowTaskFactory = newWorkspaceFlowResolveTask
	WorkspaceTasks.ListNodeTaskFactory = newWorkspaceNodeListTask
	WorkspaceTasks.ResolveNodeTaskFactory = newWorkspaceNodeResolveTask
	WorkspaceTasks.ListSetNodeTaskFactory = newWorkspaceNodeSetListTask
	WorkspaceTasks.ResolveSetNodeTaskFactory = newWorkspaceNodeSetResolveTask
	WorkspaceTasks.UpdateSetNodeTaskFactory = newWorkspaceNodeSetUpdateTask
	WorkspaceTasks.AddSourceNodeTaskFactory = newWorkspaceNodeSourceAddTask
	WorkspaceTasks.RemoveSourceNodeTaskFactory = newWorkspaceNodeSourceRemoveTask
	WorkspaceTasks.AddStoreNodeTaskFactory = newWorkspaceNodeStoreAddTask
	WorkspaceTasks.RemoveStoreNodeTaskFactory = newWorkspaceNodeStoreRemoveTask
}

func newWorkspaceCreateTask() WorkspaceCreateTask {
	return &task.Task[WorkspaceCreateParams]{
		Name:       "workspace-create",
		OnPrepare:  handleWorkspaceCreateContext,
		OnComplete: handleWorkspaceCreateComplete,
		OnPretend:  handleWorkspaceCreatePretend,
	}
}

func newWorkspaceFlowCreateTask() WorkspaceFlowCreateTask {
	return &task.Task[WorkspaceFlowCreateParams]{
		Name:       "workspace-flow-create",
		OnPrepare:  handleWorkspaceFlowCreateContext,
		OnComplete: handleWorkspaceFlowCreateComplete,
		OnPretend:  handleWorkspaceFlowCreatePretend,
	}
}

func newWorkspaceFlowListTask() WorkspaceFlowListTask {
	return &task.Task[WorkspaceFlowListParams]{
		Name:       "workspace-flow-list",
		OnPrepare:  handleWorkspaceFlowListContext,
		OnComplete: handleWorkspaceFlowListComplete,
		OnPretend:  handleWorkspaceFlowListPretend,
	}
}

func newWorkspaceFlowResolveTask() WorkspaceFlowResolveTask {
	return &task.Task[WorkspaceFlowResolveParams]{
		Name:         "workspace-flow-solution",
		OnPrepare:    handleWorkspaceFlowSolutionContext,
		OnComplete:   handleWorkspaceFlowSolutionComplete,
		OnIncomplete: handleWorkspaceFlowSolutionIncomplete,
		OnPretend:    handleWorkspaceFlowSolutionPretend,
	}
}

func newWorkspaceListTask() WorkspaceListTask {
	return &task.Task[WorkspaceListParams]{
		Name:       "workspace-list",
		OnPrepare:  handleWorkspaceListContext,
		OnComplete: handleWorkspaceListComplete,
		OnPretend:  handleWorkspaceListPretend,
	}
}

func newWorkspaceNodeListTask() WorkspaceNodeListTask {
	return &task.Task[WorkspaceNodeListParams]{
		Name:         "workspace-node-list",
		OnPrepare:    handleWorkspaceNodeListContext,
		OnComplete:   handleWorkspaceNodeListComplete,
		OnIncomplete: handleWorkspaceNodeListIncomplete,
		OnPretend:    handleWorkspaceNodeListPretend,
	}
}

func newWorkspaceNodeResolveTask() WorkspaceNodeResolveTask {
	return &task.Task[WorkspaceNodeResolveParams]{
		Name:         "workspace-node-resolve",
		OnPrepare:    handleWorkspaceNodeResolveContext,
		OnComplete:   handleWorkspaceNodeResolveComplete,
		OnIncomplete: handleWorkspaceNodeResolveIncomplete,
		OnPretend:    handleWorkspaceNodeResolvePretend,
	}
}

func newWorkspaceNodeSetListTask() WorkspaceNodeSetListTask {
	return &task.Task[WorkspaceNodeSetListParams]{
		Name:       "workspace-node-set-list",
		OnPrepare:  handleWorkspaceNodeSetListContext,
		OnComplete: handleWorkspaceNodeSetListComplete,
		OnPretend:  handleWorkspaceNodeSetListPretend,
	}
}

func newWorkspaceNodeSetResolveTask() WorkspaceNodeSetResolveTask {
	return &task.Task[WorkspaceNodeSetResolveParams]{
		Name:         "workspace-node-set-resolve",
		OnPrepare:    handleWorkspaceNodeSetResolveContext,
		OnComplete:   handleWorkspaceNodeSetResolveComplete,
		OnIncomplete: handleWorkspaceNodeSetResolveIncomplete,
		OnPretend:    handleWorkspaceNodeSetResolvePretend,
	}
}

func newWorkspaceNodeSetUpdateTask() WorkspaceNodeSetUpdateTask {
	return &task.Task[WorkspaceNodeSetResolveParams]{
		Name:       "workspace-node-set-update",
		OnPrepare:  handleWorkspaceNodeSetUpdateContext,
		OnComplete: handleWorkspaceNodeSetUpdateComplete,
		OnPretend:  handleWorkspaceNodeSetUpdatePretend,
	}
}

func newWorkspaceNodeSourceAddTask() WorkspaceNodeSourceAddTask {
	return &task.Task[WorkspaceNodeSourceParams]{
		Name:       "workspace-node-source-add",
		OnPrepare:  handleWorkspaceNodeSourceContext,
		OnComplete: handleWorkspaceNodeSourceAddComplete,
		OnPretend:  handleWorkspaceNodeSourcePretend,
	}
}

func newWorkspaceNodeSourceRemoveTask() WorkspaceNodeSourceRemoveTask {
	return &task.Task[WorkspaceNodeSourceParams]{
		Name:       "workspace-node-source-remove",
		OnPrepare:  handleWorkspaceNodeSourceContext,
		OnComplete: handleWorkspaceNodeSourceRemoveComplete,
		OnPretend:  handleWorkspaceNodeSourcePretend,
	}
}

func newWorkspaceNodeStoreAddTask() WorkspaceNodeStoreAddTask {
	return &task.Task[WorkspaceNodeStoreParams]{
		Name:       "workspace-node-store-add",
		OnPrepare:  handleWorkspaceNodeStoreContext,
		OnComplete: handleWorkspaceNodeStoreAddComplete,
		OnPretend:  handleWorkspaceNodeStorePretend,
	}
}

func newWorkspaceNodeStoreRemoveTask() WorkspaceNodeStoreRemoveTask {
	return &task.Task[WorkspaceNodeStoreParams]{
		Name:       "workspace-node-store-remove",
		OnPrepare:  handleWorkspaceNodeStoreContext,
		OnComplete: handleWorkspaceNodeStoreRemoveComplete,
		OnPretend:  handleWorkspaceNodeStorePretend,
	}
}

func newWorkspaceResolveTask() WorkspaceResolveTask {
	return &task.Task[WorkspaceFlowResolveParams]{
		Name:         "workspace-resolve",
		OnPrepare:    handleWorkspaceResolveContext,
		OnComplete:   handleWorkspaceResolveComplete,
		OnIncomplete: handleWorkspaceResolveIncomplete,
		OnPretend:    handleWorkspaceResolvePretend,
	}
}
