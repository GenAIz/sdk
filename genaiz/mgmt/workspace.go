package mgmt

import (
	"encoding/json"
	"slices"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserWorkspacesFacade Facade[[]UserWorkspace, broker.WorkspaceListParams]
type UserWorkspaceFlowsFacade Facade[[]UserWorkspaceFlow, broker.WorkspaceFlowListParams]
type UserWorkspaceNodesFacade Facade[[]UserWorkspaceNode, broker.WorkspaceNodeListParams]
type WorkspaceListTaskFactory func() *task.Task[broker.WorkspaceListParams]
type WorkspaceFlowListTaskFactory func() *task.Task[broker.WorkspaceFlowListParams]
type WorkspaceFlowResolveTaskFactory func() *task.Task[broker.WorkspaceFlowResolveParams]
type WorkspaceNodeListTaskFactory func() *task.Task[broker.WorkspaceNodeListParams]

type UserWorkspace struct {
	Id          int64  `cli:"Id"`
	Name        string `cli:"Name"`
	Description string
	Created     int64 `cli:"Created"`
	Modified    int64
	OwnerAppId  int
	OwnerUserId int
	Visibility  string
	RcEnabled   bool `cli:"RC?"`
	Active      bool `cli:"Active"`
	Flags       *int

	matched string
}

func (uw UserWorkspace) MarshalJSON() ([]byte, error) {
	var created, modified string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	}

	if uw.Modified > 0 {
		modified = createdFormatter.FormatMillis(uw.Modified)
	}

	return json.Marshal(&struct {
		Id          int64  `json:"id,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		OwnerAppId  int    `json:"ownerAppId,omitempty"`
		OwnerUserId int    `json:"ownerUserId"`
		Created     string `json:"created,omitempty"`
		Modified    string `json:"modified,omitempty"`
		Visibility  string `json:"visibility"`
		Flags       *int   `json:"flags,omitempty"`
	}{
		Id:          uw.Id,
		Name:        uw.Name,
		Description: uw.Description,
		OwnerAppId:  uw.OwnerAppId,
		OwnerUserId: uw.OwnerUserId,
		Created:     created,
		Modified:    modified,
		Visibility:  uw.Visibility,
		Flags:       uw.Flags,
	})
}

func (uw UserWorkspace) MarshalSlice() ([]string, error) {
	var created string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(uw.Id),
		uw.Name,
		created,
		stringz.YesOrNo(uw.RcEnabled),
		stringz.YesOrNo(uw.Active),
	}, nil
}

func (uw UserWorkspace) Match(filter string) *UserWorkspace {
	var idString = cast.ToString(uw.Id)
	var matched string

	if strings.EqualFold(idString, filter) ||
		strings.HasPrefix(idString, filter) {
		matched = idString
	} else if strings.EqualFold(uw.Name, filter) ||
		strings.HasPrefix(uw.Name, filter) {
		matched = uw.Name
	}

	if matched == "" {
		return nil
	}

	return &UserWorkspace{
		Id:          uw.Id,
		Name:        uw.Name,
		Description: uw.Description,
		Created:     uw.Created,
		Modified:    uw.Modified,
		OwnerAppId:  uw.OwnerAppId,
		OwnerUserId: uw.OwnerUserId,
		Visibility:  uw.Visibility,
		RcEnabled:   uw.RcEnabled,
		Active:      uw.Active,
		Flags:       uw.Flags,
		matched:     matched,
	}
}

func (uw UserWorkspace) Matched() cobra.Completion {
	if uw.matched == "" {
		return cobra.CompletionWithDesc(cast.ToString(uw.Id), uw.Name)
	}

	if _, err := strconv.Atoi(uw.matched); err == nil {
		return cobra.CompletionWithDesc(uw.matched, uw.Name)
	}

	return cobra.CompletionWithDesc(uw.matched, cast.ToString(uw.Id))
}

type userWorkspacesFacade struct {
	baseLoggingFacade
	params *broker.WorkspaceListParams
}

func (uwf *userWorkspacesFacade) Filtering(filter string) Provider[[]UserWorkspace] {
	return &userWorkspacesProvider{
		Plan: task.Plan{
			Logger: uwf.logger,
		},
		filter:                   filter,
		params:                   uwf.params,
		workspaceListTaskFactory: broker.NewWorkspaceListTask,
	}
}

func (uwf *userWorkspacesFacade) Provider() Provider[[]UserWorkspace] {
	return uwf.Filtering("")
}

func (uwf *userWorkspacesFacade) WithLogger(logger *logrus.Logger) Facade[[]UserWorkspace, broker.WorkspaceListParams] {
	uwf.logger = logger
	return uwf
}

func (uwf *userWorkspacesFacade) WithParams(params *broker.WorkspaceListParams) Facade[[]UserWorkspace, broker.WorkspaceListParams] {
	uwf.params = params
	return uwf
}

type userWorkspacesProvider struct {
	task.Plan
	filter                   string
	params                   *broker.WorkspaceListParams
	workspaceListTaskFactory WorkspaceListTaskFactory
}

func (uwp userWorkspacesProvider) Get() ([]UserWorkspace, task.Error) {
	var workspaces []broker.Workspace
	var failure interface{}

	uwp.OnReturn = func(i interface{}) { workspaces = i.([]broker.Workspace) }
	uwp.OnFailure = func(i interface{}) { failure = i }
	uwp.Sequence(task.NewWorker(uwp.params, uwp.workspaceListTaskFactory()))

	if failure == nil {
		var result = make([]UserWorkspace, 0)

		for _, ws := range workspaces {
			var userWorkspace = &UserWorkspace{
				Id:          ws.Id,
				Name:        ws.Name,
				Description: ws.Description,
				OwnerAppId:  ws.OwnerAppId,
				OwnerUserId: ws.OwnerUserId,
				Visibility:  ws.Visibility,
				RcEnabled:   ws.IsRcEnabled(),
				Created:     ws.Created,
				Modified:    ws.Modified,
				Flags:       ws.Flags,
				Active:      ws.IsActive(),
			}

			if uwp.filter == "" {
				result = append(result, *userWorkspace)
			} else if matched := userWorkspace.Match(uwp.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

type UserWorkspaceFlow struct {
	Id              int64  `cli:"Id"`
	Name            string `cli:"Name"`
	Description     string
	SolutionId      int64
	SolutionOem     string `cli:"OEM"`
	SolutionHandle  string `cli:"Solution"`
	SolutionVersion string `cli:"Version"`
	WorkflowHandle  string `cli:"Workflow"`
	WorkflowId      int64
	Created         int64 `cli:"Created"`
	Active          bool  `cli:"Active"`
	Ready           bool  `cli:"Ready?"`
	Flags           *int

	matched string
}

func (uf UserWorkspaceFlow) MarshalJSON() ([]byte, error) {
	var created string

	if uf.Created > 0 {
		created = createdFormatter.FormatMillis(uf.Created)
	}

	return json.Marshal(&struct {
		Id              int64  `json:"id"`
		Name            string `json:"name,omitempty"`
		Description     string `json:"description,omitempty"`
		SolutionId      int64  `json:"solutionId"`
		SolutionOem     string `json:"solutionOem"`
		SolutionHandle  string `json:"solutionHandle"`
		SolutionVersion string `json:"solutionVersion"`
		WorkflowId      int64  `json:"workflowId"`
		WorkflowHandle  string `json:"workflowHandle"`
		Created         string `json:"created,omitempty"`
		Flags           *int   `json:"flags,omitempty"`
	}{
		Id:              uf.Id,
		Name:            uf.Name,
		Description:     uf.Description,
		SolutionId:      uf.SolutionId,
		SolutionOem:     uf.SolutionOem,
		SolutionHandle:  uf.SolutionHandle,
		SolutionVersion: uf.SolutionVersion,
		WorkflowId:      uf.WorkflowId,
		WorkflowHandle:  uf.WorkflowHandle,
		Created:         created,
		Flags:           uf.Flags,
	})
}

func (uf UserWorkspaceFlow) MarshalSlice() ([]string, error) {
	var created string

	if uf.Created > 0 {
		created = createdFormatter.FormatMillis(uf.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(uf.Id),
		uf.Name,
		uf.SolutionOem,
		uf.SolutionHandle,
		uf.SolutionVersion,
		uf.WorkflowHandle,
		created,
		stringz.YesOrNo(uf.Active),
		stringz.YesOrNo(uf.Ready),
	}, nil
}

func (uf UserWorkspaceFlow) Match(filter string) *UserWorkspaceFlow {
	var idString = cast.ToString(uf.Id)
	var matched string

	if strings.EqualFold(idString, filter) ||
		strings.HasPrefix(idString, filter) {
		matched = idString
	} else if strings.EqualFold(uf.WorkflowHandle, filter) ||
		strings.HasPrefix(uf.WorkflowHandle, filter) {
		matched = uf.WorkflowHandle
	}

	if matched == "" {
		return nil
	}

	return &UserWorkspaceFlow{
		Id:             uf.Id,
		Name:           uf.Name,
		Description:    uf.Description,
		SolutionId:     uf.SolutionId,
		SolutionOem:    uf.SolutionOem,
		SolutionHandle: uf.SolutionHandle,
		WorkflowId:     uf.WorkflowId,
		WorkflowHandle: uf.WorkflowHandle,
		Created:        uf.Created,
		Active:         uf.Active,
		Ready:          uf.Ready,
		Flags:          uf.Flags,
		matched:        matched,
	}
}

func (uf UserWorkspaceFlow) Matched() cobra.Completion {
	if uf.matched == "" {
		return cobra.CompletionWithDesc(cast.ToString(uf.Id), uf.WorkflowHandle)
	}

	if _, err := strconv.Atoi(uf.matched); err == nil {
		return cobra.CompletionWithDesc(uf.matched, uf.WorkflowHandle)
	}

	return cobra.CompletionWithDesc(uf.matched, cast.ToString(uf.Id))
}

type userWorkspaceFlowsFacade struct {
	baseLoggingFacade
	params *broker.WorkspaceFlowListParams
}

func (uwf userWorkspaceFlowsFacade) Filtering(filter string) Provider[[]UserWorkspaceFlow] {
	return &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: uwf.logger,
		},
		filter: filter,
		params: uwf.params,

		workspaceResolveTaskFactory:  broker.NewWorkspaceFlowResolveTask,
		workspaceFlowListTaskFactory: broker.NewWorkspaceFlowListTask,
	}
}

func (uwf userWorkspaceFlowsFacade) Provider() Provider[[]UserWorkspaceFlow] {
	return uwf.Filtering("")
}

func (uwf userWorkspaceFlowsFacade) WithLogger(logger *logrus.Logger) Facade[[]UserWorkspaceFlow, broker.WorkspaceFlowListParams] {
	uwf.logger = logger
	return uwf
}

func (uwf userWorkspaceFlowsFacade) WithParams(params *broker.WorkspaceFlowListParams) Facade[[]UserWorkspaceFlow, broker.WorkspaceFlowListParams] {
	uwf.params = params
	return uwf
}

type userWorkspaceFlowsProvider struct {
	task.Plan
	filter string
	params *broker.WorkspaceFlowListParams

	workspaceResolveTaskFactory  WorkspaceFlowResolveTaskFactory
	workspaceFlowListTaskFactory WorkspaceFlowListTaskFactory
}

func (uwp userWorkspaceFlowsProvider) Get() ([]UserWorkspaceFlow, task.Error) {
	var flows []broker.WorkspaceFlow
	var workers []task.Worker
	var failure interface{}

	uwp.OnReturn = func(i interface{}) { flows = i.([]broker.WorkspaceFlow) }
	uwp.OnFailure = func(i interface{}) { failure = i }

	if uwp.params.GetWorkspaceId() == nil {
		workers = append(workers, task.NewWorker(uwp.params.WorkspaceFlowResolveParams, uwp.workspaceResolveTaskFactory()))
	}

	workers = append(workers, task.NewWorker(uwp.params, uwp.workspaceFlowListTaskFactory()))
	uwp.Sequence(workers...)

	if failure == nil {
		var result = make([]UserWorkspaceFlow, 0)

		for _, fl := range flows {
			var userFlow = &UserWorkspaceFlow{
				Id:          fl.Id,
				Name:        fl.Name,
				Description: fl.Description,
				SolutionId:  fl.SolutionId,
				WorkflowId:  fl.WorkflowId,
				Created:     fl.Created,
				Active:      fl.IsActive(),
				Ready:       fl.IsReady(),
				Flags:       fl.Flags,
			}

			if fl.Solution != nil {
				userFlow.SolutionOem = fl.Solution.Oem
				userFlow.SolutionHandle = fl.Solution.Handle
				userFlow.SolutionVersion = fl.Solution.GetVersion()

				if i := slices.IndexFunc(fl.Solution.Workflows, func(workflow broker.Workflow) bool {
					return workflow.Id != nil && *workflow.Id == fl.WorkflowId
				}); i >= 0 {
					userFlow.WorkflowHandle = fl.Solution.Workflows[i].Handle
				}
			}

			if uwp.filter == "" {
				result = append(result, *userFlow)
			} else if matched := userFlow.Match(uwp.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

type UserWorkspaceNode struct {
	Id                   int64 `cli:"Id"`
	WorkspaceId          int64
	WorkspaceFlowId      int64
	SmartFunctionId      int64
	SmartFunctionOem     string `cli:"OEM"`
	SmartFunctionHandle  string `cli:"Smart Function"`
	SmartFunctionVersion string `cli:"Version"`
	WorkflowNodeId       int64
	WorkflowNodeHandle   string `cli:"Workflow Node"`
}

func (un UserWorkspaceNode) MarshalSlice() ([]string, error) {
	var nodeHandle = stringz.FirstNonEmpty(un.WorkflowNodeHandle, cast.ToString(un.WorkflowNodeId))
	var sfOem = stringz.FirstNonEmpty(un.SmartFunctionOem, "???")
	var sfHandle = stringz.FirstNonEmpty(un.SmartFunctionHandle, "???")
	var sfVersion = stringz.FirstNonEmpty(un.SmartFunctionVersion, "???")

	return []string{
		cast.ToString(un.Id),
		sfOem,
		sfHandle,
		sfVersion,
		nodeHandle,
	}, nil
}

type userWorkspaceNodesFacade struct {
	baseLoggingFacade
	params *broker.WorkspaceNodeListParams
}

func (uwn userWorkspaceNodesFacade) Filtering(filter string) Provider[[]UserWorkspaceNode] {
	return &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: uwn.logger,
		},
		filter:                       filter,
		params:                       uwn.params,
		workspaceNodeListTaskFactory: broker.NewWorkspaceNodeListTask,
	}
}

func (uwn userWorkspaceNodesFacade) Provider() Provider[[]UserWorkspaceNode] {
	return uwn.Filtering("")
}

func (uwn userWorkspaceNodesFacade) WithLogger(logger *logrus.Logger) Facade[[]UserWorkspaceNode, broker.WorkspaceNodeListParams] {
	uwn.logger = logger
	return uwn
}

func (uwn userWorkspaceNodesFacade) WithParams(params *broker.WorkspaceNodeListParams) Facade[[]UserWorkspaceNode, broker.WorkspaceNodeListParams] {
	uwn.params = params
	return uwn
}

type userWorkspaceNodesProvider struct {
	task.Plan
	filter string
	params *broker.WorkspaceNodeListParams

	workspaceNodeListTaskFactory WorkspaceNodeListTaskFactory
}

func (uwn userWorkspaceNodesProvider) Get() ([]UserWorkspaceNode, task.Error) {
	var nodes []broker.WorkspaceNode
	var failure interface{}

	uwn.OnReturn = func(i interface{}) { nodes = i.([]broker.WorkspaceNode) }
	uwn.OnFailure = func(i interface{}) { failure = i }
	uwn.Sequence(task.NewWorker(uwn.params, uwn.workspaceNodeListTaskFactory()))

	if failure == nil {
		var result = make([]UserWorkspaceNode, 0)

		for _, nd := range nodes {
			var userNode = &UserWorkspaceNode{
				Id:              nd.Id,
				WorkspaceId:     nd.WorkspaceId,
				WorkspaceFlowId: nd.WorkspaceFlowId,
				WorkflowNodeId:  nd.WorkflowNodeId,
				SmartFunctionId: nd.SmartFunctionId,
			}

			if nd.SmartFunction != nil {
				userNode.SmartFunctionOem = nd.SmartFunction.Oem
				userNode.SmartFunctionHandle = nd.SmartFunction.Handle
				userNode.SmartFunctionVersion = nd.SmartFunction.GetFullVersion()
			}

			result = append(result, *userNode)
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserWorkspacesFacade() UserWorkspacesFacade {
	return &userWorkspacesFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}

func NewUserWorkspaceFlowsFacade() UserWorkspaceFlowsFacade {
	return &userWorkspaceFlowsFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}

func NewUserWorkspaceNodesFacade() UserWorkspaceNodesFacade {
	return &userWorkspaceNodesFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
