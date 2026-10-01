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
)

var (
	errorWorkflowIdKnown               = task.NewError("workflow id is already known")
	errorWorkflowIdRequired            = task.NewError("workflow id is required")
	errorWorkflowOemRequired           = task.NewError("workflow solution oem is required")
	errorWorkflowHandleRequired        = task.NewError("workflow handle is required")
	errorWorkflowVersionRequired       = task.NewError("workflow solution version is required")
	errorWorkspaceEmpty                = task.NewError("no workspace definition provided")
	errorWorkspaceFlowConflict         = task.NewError("workflow present under multiple workspace flows")
	errorWorkspaceFlowInvalid          = task.NewError("workspace flow definition is invalid")
	errorWorkspaceFlowRequired         = task.NewError("workspace flow can not be located")
	errorWorkspaceFlowUnresolved       = task.NewError("workspace flow needs to be located")
	errorWorkspaceConflict             = task.NewError("workspace can be several possibilities")
	errorWorkspaceInvalidNco           = task.NewError("workspace creation date is invalid")
	errorWorkspaceInvalidOwner         = task.NewError("workspace ownership can not be established with the selected session")
	errorWorkspaceIdKnown              = task.NewError("workspace id is already known")
	errorWorkspaceIdRequired           = task.NewError("workspace id is required")
	errorWorkspaceNameRequired         = task.NewError("workspace name is required")
	errorWorkspaceNotFound             = task.NewError("workspace could not be found")
	errorWorkspaceNodeDsRequired       = task.NewError("workspace node datasource is required")
	errorWorkspaceNodeKnown            = task.NewError("workspace node is already known")
	errorWorkspaceNodeRequired         = task.NewError("workspace node can not be located")
	errorWorkspaceNodeSfHandleRequired = task.NewError("smart function handle is required")
	errorWorkspaceVisibility           = task.NewError("workspace visibility is required")
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
	FlowId         *int64
	NodeId         *int64
	FnHandle       string
	WorkspaceId    *int64
	WorkspaceName  string
	WorkflowHandle string
}

type WorkspaceNodeSourceParams struct {
	*DataSourceResolveParams
	NodeId *int64
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
		Name:         "workspace-node-list",
		OnPrepare:    handleWorkspaceNodeListContext,
		OnComplete:   handleWorkspaceNodeListComplete,
		OnIncomplete: handleWorkspaceNodeListIncomplete,
		OnPretend:    handleWorkspaceNodeListPretend,
	}
}

func NewWorkspaceNodeResolveTask() *task.Task[WorkspaceNodeResolveParams] {
	return &task.Task[WorkspaceNodeResolveParams]{
		Name:         "workspace-node-resolve",
		OnPrepare:    handleWorkspaceNodeResolveContext,
		OnComplete:   handleWorkspaceNodeResolveComplete,
		OnIncomplete: handleWorkspaceNodeResolveIncomplete,
		OnPretend:    handleWorkspaceNodeResolvePretend,
	}
}

func NewWorkspaceNodeSourceAddTask() *task.Task[WorkspaceNodeSourceParams] {
	return &task.Task[WorkspaceNodeSourceParams]{
		Name:       "workspace-node-source-add",
		OnPrepare:  handleWorkspaceNodeSourceContext,
		OnComplete: handleWorkspaceNodeSourceAddComplete,
		OnPretend:  handleWorkspaceNodeSourcePretend,
	}
}

func NewWorkspaceNodeSourceRemoveTask() *task.Task[WorkspaceNodeSourceParams] {
	return &task.Task[WorkspaceNodeSourceParams]{
		Name:       "workspace-node-source-remove",
		OnPrepare:  handleWorkspaceNodeSourceContext,
		OnComplete: handleWorkspaceNodeSourceRemoveComplete,
		OnPretend:  handleWorkspaceNodeSourcePretend,
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
			return errorWorkspaceInvalidNco
		}

		if params.OwnerOnly {
			if brokerClient, err = params.GetClient(); err == nil {
				if brokerClient.GetUserId() > 0 {
					return nil
				}

				return errorWorkspaceInvalidOwner
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
				return errorWorkspaceVisibility
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

		return errorWorkspaceEmpty
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

	return errorWorkspaceEmpty
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

	return errorWorkspaceEmpty
}

func handleWorkspaceFlowCreateContext(params *WorkspaceFlowCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkspaceId == nil {
			return errorWorkspaceIdRequired
		}

		if params.WorkflowId == nil {
			return errorWorkflowIdRequired
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

	return errorWorkspaceFlowInvalid
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

	return errorWorkspaceFlowInvalid
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

	return errorWorkspaceIdRequired
}

func handleWorkspaceFlowListContext(params *WorkspaceFlowListParams, state *task.State) error {
	if state.Output == "" {
		if params.GetWorkspaceId() == nil {
			return errorWorkspaceIdRequired
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

	return errorWorkspaceIdRequired
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
			return errorWorkspaceIdKnown
		}

		if params.WorkspaceName == "" {
			return errorWorkspaceNameRequired
		}
	}

	return nil
}

func handleWorkspaceFlowResolveIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, errorWorkspaceIdKnown) {
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
			return errorWorkflowIdKnown
		}

		if params.SolutionOem == "" {
			return errorWorkflowOemRequired
		}

		if params.SolutionHandle == "" {
			return errorWorkflowHandleRequired
		}

		if params.SolutionVersion == "" {
			return errorWorkflowVersionRequired
		}
	}

	return nil
}

func handleWorkspaceFlowSolutionIncomplete(params *WorkspaceFlowResolveParams, state *task.State) error {
	if errors.Is(state.Error, errorWorkflowIdKnown) {
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
				state.Logger.Debugf("Found %d nodes under workspace [%d] for flow id [%d]",
					len(nodes), *params.WorkspaceId, *params.WorkspaceFlowId)
				state.Internal = nodes
				return nil
			}
		}

		return err
	}

	return errorWorkspaceFlowRequired
}

func handleWorkspaceNodeListContext(params *WorkspaceNodeListParams, state *task.State) error {
	if state.Output == "" {
		if params.WorkflowHandle == "" && params.WorkspaceFlowId == nil {
			return errorWorkspaceFlowRequired
		}

		if params.WorkspaceFlowId == nil {
			return errorWorkspaceFlowUnresolved
		}
	}

	return nil
}

func handleWorkspaceNodeListIncomplete(params *WorkspaceNodeListParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorWorkspaceFlowUnresolved) {
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
					return errorWorkspaceNotFound
				} else if len(results) > 1 {
					return errorWorkspaceConflict
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
					return errorWorkspaceFlowConflict
				}

				return errorWorkspaceFlowRequired
			}
		}

		return err
	}

	return errorWorkspaceFlowRequired
}

func handleWorkspaceNodeListPretend(params *WorkspaceNodeListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		if errors.Is(state.Error, errorWorkspaceFlowUnresolved) {
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
					return errorWorkspaceNotFound
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
				return errorWorkspaceFlowRequired
			}

			state.Output = ""
			return nil
		}

		return err
	}

	return errorWorkspaceNodeSfHandleRequired
}

func handleWorkspaceNodeResolveContext(params *WorkspaceNodeResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.NodeId != nil {
			return errorWorkspaceNodeKnown
		}

		if params.FnHandle == "" {
			return errorWorkspaceNodeSfHandleRequired
		}

		if params.FlowId == nil && params.WorkflowHandle == "" {
			return errorWorkflowHandleRequired
		}

		if params.WorkspaceId == nil && params.WorkspaceName == "" {
			return errorWorkspaceNameRequired
		}

		state.Output = params.FnHandle
	}

	return nil
}

func handleWorkspaceNodeResolveIncomplete(params *WorkspaceNodeResolveParams, state *task.State) error {
	if errors.Is(state.Error, errorWorkspaceNodeKnown) {
		state.Logger.Debugf("Node id provided [%d], skipping resolution", *params.NodeId)
		state.Completed = true
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

	if errors.Is(state.Error, errorWorkspaceNodeKnown) {
		state.Logger.Debugf("Node id provided [%d], skipping resolution", *params.NodeId)
		return nil
	}

	return state.Error
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

	return errorWorkspaceNodeDsRequired
}

func handleWorkspaceNodeSourceContext(params *WorkspaceNodeSourceParams, state *task.State) error {
	if state.Output == "" {
		if params.NodeId == nil {
			return errorWorkspaceNodeRequired
		}

		if params.DataSourceId == nil {
			return errorWorkspaceNodeDsRequired
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

	return errorWorkspaceNodeDsRequired
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

	return errorWorkspaceNodeDsRequired
}
