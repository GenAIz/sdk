package workspaces

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserWorkspaceNodesFactory interface {
	mgmt.Factory[broker.WorkspaceNodeListParams, []UserWorkspaceNode]
}

type UserWorkspaceNodesProvider = mgmt.Provider[[]UserWorkspaceNode]
type UserWorkspaceNodesProviderFactory = mgmt.ProviderFactory[broker.WorkspaceNodeListParams, []UserWorkspaceNode]

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

	matched string
}

func (un UserWorkspaceNode) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Id              int64 `json:"id"`
		WorkspaceId     int64 `json:"workspaceId"`
		WorkspaceFlowId int64 `json:"workspaceFlowId"`
		WorkflowNodeId  int64 `json:"workflowNodeId"`
		SmartFunctionId int64 `json:"smartFunctionId"`
	}{
		Id:              un.Id,
		WorkspaceId:     un.WorkspaceId,
		WorkspaceFlowId: un.WorkspaceFlowId,
		WorkflowNodeId:  un.WorkflowNodeId,
		SmartFunctionId: un.SmartFunctionId,
	})
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

func (un UserWorkspaceNode) Match(filter string) *UserWorkspaceNode {
	var idString = cast.ToString(un.Id)
	var matched string

	if strings.EqualFold(idString, filter) ||
		strings.HasPrefix(idString, filter) {
		matched = idString
	} else if strings.EqualFold(un.SmartFunctionHandle, filter) ||
		strings.HasPrefix(un.SmartFunctionHandle, filter) {
		matched = un.SmartFunctionHandle
	}

	if matched == "" {
		return nil
	}

	return &UserWorkspaceNode{
		Id:                   un.Id,
		WorkspaceId:          un.WorkspaceId,
		WorkspaceFlowId:      un.WorkspaceFlowId,
		SmartFunctionId:      un.SmartFunctionId,
		SmartFunctionOem:     un.SmartFunctionOem,
		SmartFunctionHandle:  un.SmartFunctionHandle,
		SmartFunctionVersion: un.SmartFunctionVersion,
		WorkflowNodeId:       un.WorkflowNodeId,
		WorkflowNodeHandle:   un.WorkflowNodeHandle,

		matched: matched,
	}
}

func (un UserWorkspaceNode) Matched() cobra.Completion {
	if un.matched == "" {
		return cobra.CompletionWithDesc(cast.ToString(un.Id), un.SmartFunctionHandle)
	}

	if _, err := strconv.Atoi(un.matched); err == nil {
		return cobra.CompletionWithDesc(un.matched, un.SmartFunctionHandle)
	}

	return cobra.CompletionWithDesc(un.matched, cast.ToString(un.Id))
}

type userWorkspaceNodesFactory struct {
	mgmt.ProviderFactory[broker.WorkspaceNodeListParams, []UserWorkspaceNode]
}

func (uf userWorkspaceNodesFactory) provider(factory mgmt.ProviderFactory[broker.WorkspaceNodeListParams, []UserWorkspaceNode]) mgmt.Provider[[]UserWorkspaceNode] {
	return &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: factory.Logger,
		},

		filter: factory.Filter,
		format: factory.Format,
		params: factory.Params,

		workspaceNodeListTaskFactory: broker.WorkspaceTasks.ListNodeTaskFactory,
	}
}

type userWorkspaceNodesProvider struct {
	task.Plan

	filter *string
	format timez.Formatter
	params *broker.WorkspaceNodeListParams

	workspaceNodeListTaskFactory broker.WorkspaceNodeListTaskFactory
}

func (up userWorkspaceNodesProvider) Get() ([]UserWorkspaceNode, task.Error) {
	var nodes []broker.WorkspaceNode
	var failure interface{}

	up.OnReturn = func(i interface{}) { nodes = i.([]broker.WorkspaceNode) }
	up.OnFailure = func(i interface{}) { failure = i }
	up.Sequence(task.NewWorker(up.params, up.workspaceNodeListTaskFactory()))

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

			if up.filter == nil {
				result = append(result, *userNode)
			} else if matched := userNode.Match(*up.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserWorkspaceNodesFactory() UserWorkspaceNodesFactory {
	var factory = &userWorkspaceNodesFactory{
		mgmt.ProviderFactory[broker.WorkspaceNodeListParams, []UserWorkspaceNode]{
			Format: mgmt.DefaultTimeFormat,
		},
	}

	return factory.WithProvider(factory.provider)
}
