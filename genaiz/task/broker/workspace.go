package broker

import (
	"errors"
	"fmt"
	"slices"
	"strings"
	"time"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
)

var (
	ErrorWorkflowIdKnown         = task.NewError("workflow id is already known")
	ErrorWorkflowIdRequired      = task.NewError("workflow id is required")
	ErrorWorkflowOemRequired     = task.NewError("workflow solution oem is required")
	ErrorWorkflowHandleRequired  = task.NewError("workflow solution handle is required")
	ErrorWorkflowVersionRequired = task.NewError("workflow solution version is required")
	ErrorWorkspaceEmpty          = task.NewError("no workspace definition provided")
	ErrorWorkspaceFlowConflict   = task.NewError("workflow present under multiple workspace flows")
	ErrorWorkspaceFlowInvalid    = task.NewError("workspace flow definition is invalid")
	ErrorWorkspaceFlowRequired   = task.NewError("workspace flow can not be located")
	ErrorWorkspaceFlowUnresolved = task.NewError("workspace flow needs to be located")
	ErrorWorkspaceConflict       = task.NewError("workspace can be several possibilities")
	ErrorWorkspaceInvalidNco     = task.NewError("workspace creation date is invalid")
	ErrorWorkspaceInvalidOwner   = task.NewError("workspace ownership can not be established with the selected session")
	ErrorWorkspaceIdKnown        = task.NewError("workspace id is already known")
	ErrorWorkspaceIdRequired     = task.NewError("workspace id is required")
	ErrorWorkspaceNameRequired   = task.NewError("workspace name is required")
	ErrorWorkspaceNotFound       = task.NewError("workspace could not be found")
	ErrorWorkspaceVisibility     = task.NewError("workspace visibility is required")
)

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
	if lp.ReadyOnly {
		return WorkspaceFlowFlags.Active | WorkspaceFlowFlags.Ready,
			WorkspaceFlowFlags.Active | WorkspaceFlowFlags.Ready
	}

	return WorkspaceFlowFlags.Active, WorkspaceFlowFlags.Active
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

func NewWorkspaceCreateTask() *task.Task[WorkspaceCreateParams] {
	return &task.Task[WorkspaceCreateParams]{
		Name:       "workspace-create",
		OnPrepare:  handleWorkspaceCreateContext,
		OnComplete: handleWorkspaceCreateComplete,
		OnPretend:  handleWorkspaceCreatePretend,
	}
}

func NewWorkspaceFlowCreateTask() *task.Task[WorkspaceFlowCreateParams] {
	return &task.Task[WorkspaceFlowCreateParams]{
		Name:       "workspace-flow-create",
		OnPrepare:  handleWorkspaceFlowCreateContext,
		OnComplete: handleWorkspaceFlowCreateComplete,
		OnPretend:  handleWorkspaceFlowCreatePretend,
	}
}

func NewWorkspaceFlowListTask() *task.Task[WorkspaceFlowListParams] {
	return &task.Task[WorkspaceFlowListParams]{
		Name:       "workspace-flow-list",
		OnPrepare:  handleWorkspaceFlowListContext,
		OnComplete: handleWorkspaceFlowListComplete,
		OnPretend:  handleWorkspaceFlowListPretend,
	}
}

func NewWorkspaceFlowResolveTask() *task.Task[WorkspaceFlowResolveParams] {
	return &task.Task[WorkspaceFlowResolveParams]{
		Name:         "workspace-flow-resolve",
		OnPrepare:    handleWorkspaceFlowResolveContext,
		OnComplete:   handleWorkspaceFlowResolveComplete,
		OnIncomplete: handleWorkspaceFlowResolveIncomplete,
		OnPretend:    handleWorkspaceFlowResolvePretend,
	}
}

func NewWorkspaceFlowSolutionTask() *task.Task[WorkspaceFlowResolveParams] {
	return &task.Task[WorkspaceFlowResolveParams]{
		Name:         "workspace-flow-solution",
		OnPrepare:    handleWorkspaceFlowSolutionContext,
		OnComplete:   handleWorkspaceFlowSolutionComplete,
		OnIncomplete: handleWorkspaceFlowSolutionIncomplete,
		OnPretend:    handleWorkspaceFlowSolutionPretend,
	}
}

func NewWorkspaceNodeListTask() *task.Task[WorkspaceNodeListParams] {
	return &task.Task[WorkspaceNodeListParams]{
		Name:         "workspace-nodes-resolve",
		OnPrepare:    handleWorkspaceNodeListContext,
		OnComplete:   handleWorkspaceNodeListComplete,
		OnIncomplete: handleWorkspaceNodeListIncomplete,
		OnPretend:    handleWorkspaceNodeListPretend,
	}
}

func NewWorkspaceListTask() *task.Task[WorkspaceListParams] {
	return &task.Task[WorkspaceListParams]{
		Name:       "workspace-list",
		OnPrepare:  handleWorkspaceListContext,
		OnComplete: handleWorkspaceListComplete,
		OnPretend:  handleWorkspaceListPretend,
	}
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

func handleWorkspaceCreateContext(params *WorkspaceCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.Workspace != nil {
			var nameSuffix string

			if params.Workspace.Visibility == "" {
				return ErrorWorkspaceVisibility
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

		return ErrorWorkspaceEmpty
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

	return ErrorWorkspaceEmpty
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

	return ErrorWorkspaceEmpty
}

func handleWorkspaceFlowCreateContext(params *WorkspaceFlowCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkspaceId == nil {
			return ErrorWorkspaceIdRequired
		}

		if params.WorkflowId == nil {
			return ErrorWorkflowIdRequired
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

	return ErrorWorkspaceFlowInvalid
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

	return ErrorWorkspaceFlowInvalid
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

	return ErrorWorkspaceIdRequired
}

func handleWorkspaceFlowListContext(params *WorkspaceFlowListParams, state *task.State) error {
	if state.Output == "" {
		if params.GetWorkspaceId() == nil {
			return ErrorWorkspaceIdRequired
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

	return ErrorWorkspaceIdRequired
}

func handleWorkspaceFlowResolveComplete(params *WorkspaceFlowResolveParams, state *task.State) error {
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

func handleWorkspaceFlowResolveContext(params *WorkspaceFlowResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.HasWorkspaceId() {
			state.Logger.Warnf("Workspace id is already known: [%d]", *params.WorkspaceId)
			return ErrorWorkspaceIdKnown
		}

		if params.WorkspaceName == "" {
			return ErrorWorkspaceNameRequired
		}
	}

	return nil
}

func handleWorkspaceFlowResolveIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, ErrorWorkspaceIdKnown) {
		state.Logger.Debugf("Resolved workspace id [%d]", *params.WorkspaceId)
		state.Completed = true
		state.Output = ""
		return nil
	}

	return state.Error
}

func handleWorkspaceFlowResolvePretend(params *WorkspaceFlowResolveParams, state *task.State) error {
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
			return ErrorWorkflowIdKnown
		}

		if params.SolutionOem == "" {
			return ErrorWorkflowOemRequired
		}

		if params.SolutionHandle == "" {
			return ErrorWorkflowHandleRequired
		}

		if params.SolutionVersion == "" {
			return ErrorWorkflowVersionRequired
		}
	}

	return nil
}

func handleWorkspaceFlowSolutionIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, ErrorWorkflowIdKnown) {
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
			var result []WorkspaceNode

			state.Logger.Debugf("Listing workspace flow nodes for flow id [%d]", params.WorkspaceFlowId)

			if nodes, err = brokerClient.ListWorkspaceNodes(*params.WorkspaceFlowId); err == nil {
				var sf *Function

				state.Logger.Debugf("Found %d nodes under workspace [%d] for flow id [%d]",
					len(nodes), *params.WorkspaceId, *params.WorkspaceFlowId)

				for _, node := range nodes {
					if sf, err = brokerClient.GetFunction(node.SmartFunctionId); err == nil {
						result = append(result, node.withFunction(sf))
					} else {
						result = append(result, node)
					}
				}

				state.Internal = result
				return nil
			}
		}

		return err
	}

	return ErrorWorkspaceFlowRequired
}

func handleWorkspaceNodeListContext(params *WorkspaceNodeListParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkflowHandle == "" && params.WorkspaceFlowId == nil {
			return ErrorWorkspaceFlowRequired
		}

		if params.WorkspaceFlowId == nil {
			return ErrorWorkspaceFlowUnresolved
		}
	}

	return nil
}

func handleWorkspaceNodeListIncomplete(params *WorkspaceNodeListParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, ErrorWorkspaceFlowUnresolved) {
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
					return ErrorWorkspaceNotFound
				} else if len(results) > 1 {
					return ErrorWorkspaceConflict
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
					return ErrorWorkspaceFlowConflict
				}

				return ErrorWorkspaceFlowRequired
			}
		}

		return err
	}

	return ErrorWorkspaceFlowRequired
}

func handleWorkspaceNodeListPretend(params *WorkspaceNodeListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		if errors.Is(state.Error, ErrorWorkspaceFlowUnresolved) {
			if params.WorkspaceId == nil {
				state.Logger.Debugf("Pretending to list workspaces to resolve a workspace by name")
				fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspacesUrl(), WorkspaceFlags.Active, WorkspaceFlags.Active)
			}

			state.Logger.Debugf("Pretending to list workspace flows for the workspace id")
			fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -G -d workspaceId=%s", "[workspaceId]")
			fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspaceFlowsUrl(), WorkspaceFlags.Active, WorkspaceFlags.Active)
		}

		state.Logger.Debugf("Pretending to list workspace flow nodes")
		fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("  -G -d workspaceFlowId=%s", "[workspaceFlowId]")
		fmt.Println(brokerClient.ListWorkspaceNodesUrl())
	}

	return err
}

func handleWorkspaceListContext(params *WorkspaceListParams, state *task.State) error {
	if state.Output == "" {
		var brokerClient Client
		var err error

		if params.FromDate != nil && time.Now().Before(*params.FromDate) {
			return ErrorWorkspaceInvalidNco
		}

		if params.OwnerOnly {
			if brokerClient, err = params.GetClient(); err == nil {
				if brokerClient.GetUserId() > 0 {
					return nil
				}

				return ErrorWorkspaceInvalidOwner
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
