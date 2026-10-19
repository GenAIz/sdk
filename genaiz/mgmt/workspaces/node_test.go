package workspaces

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestUserWorkspaceNode_MarshalJSON(t *testing.T) {
	var testNode = &UserWorkspaceNode{
		Id:              37,
		WorkspaceId:     int64(1337),
		WorkspaceFlowId: int64(73),
		WorkflowNodeId:  int64(42),
		SmartFunctionId: int64(69),
	}
	var bytes []byte
	var err error

	if bytes, err = testNode.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testNode.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"workspaceId\":%d", testNode.WorkspaceId))
		assert.Contains(t, actual, fmt.Sprintf("\"workspaceFlowId\":%d", testNode.WorkspaceFlowId))
		assert.Contains(t, actual, fmt.Sprintf("\"workflowNodeId\":%d", testNode.WorkflowNodeId))
		assert.Contains(t, actual, fmt.Sprintf("\"smartFunctionId\":%d", testNode.SmartFunctionId))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceNode_MarshalSlice(t *testing.T) {
	var testWorkspaceNode = &UserWorkspaceNode{
		Id:                   37,
		WorkspaceId:          42,
		WorkspaceFlowId:      1337,
		SmartFunctionId:      69,
		SmartFunctionOem:     "expectedOem",
		SmartFunctionHandle:  "expectedHandle",
		SmartFunctionVersion: "expectedVersion",
		WorkflowNodeId:       31337,
		WorkflowNodeHandle:   "expectedNodeHandle",
	}
	var values []string
	var err error

	if values, err = testWorkspaceNode.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testWorkspaceNode.Id))
		assert.Equal(t, values[1], testWorkspaceNode.SmartFunctionOem)
		assert.Equal(t, values[2], testWorkspaceNode.SmartFunctionHandle)
		assert.Equal(t, values[3], testWorkspaceNode.SmartFunctionVersion)
		assert.Equal(t, values[4], testWorkspaceNode.WorkflowNodeHandle)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceNode_Match(t *testing.T) {
	var testWorkspaceNode = &UserWorkspaceNode{
		Id:                   37,
		WorkspaceId:          42,
		WorkspaceFlowId:      1337,
		SmartFunctionId:      69,
		SmartFunctionOem:     "expectedOem",
		SmartFunctionHandle:  "expectedHandle",
		SmartFunctionVersion: "expectedVersion",
		WorkflowNodeId:       31337,
		WorkflowNodeHandle:   "expectedNodeHandle",
	}

	if actual := testWorkspaceNode.Match("3"); actual != nil {
		assert.Equal(t, cast.ToString(testWorkspaceNode.Id), actual.matched)
	} else {
		assert.Fail(t, "could not match node")
	}
}

func TestUserWorkspaceNode_Matched(t *testing.T) {
	var testNode = &UserWorkspaceNode{
		Id:                  int64(37),
		SmartFunctionHandle: "handle",
	}
	var expected = fmt.Sprintf("%d\t%s", testNode.Id, testNode.SmartFunctionHandle)

	assert.Equal(t, expected, testNode.Matched())
}

func TestUserWorkspaceNode_Matched_Custom(t *testing.T) {
	var testNode = &UserWorkspaceNode{
		Id:                  int64(37),
		SmartFunctionHandle: "handle",
		matched:             "custom",
	}
	var expected = fmt.Sprintf("%s\t%d", testNode.matched, testNode.Id)

	assert.Equal(t, expected, testNode.Matched())
}

func TestUserWorkspaceNode_Matched_Id(t *testing.T) {
	var testNode = &UserWorkspaceNode{
		Id:                  int64(37),
		SmartFunctionHandle: "handle",
		matched:             "1337",
	}
	var expected = fmt.Sprintf("%s\t%s", testNode.matched, testNode.SmartFunctionHandle)

	assert.Equal(t, expected, testNode.Matched())
}

func TestUserWorkspaceNodesProvider_Get(t *testing.T) {
	var capturedParams broker.WorkspaceNodeListParams
	var expectedParams = &broker.WorkspaceNodeListParams{
		WorkspaceName:  "workspace",
		WorkflowHandle: "workflow",
	}
	var expectedNodes = []broker.WorkspaceNode{
		{
			Id: 37,
		},
		{
			Id: 42,
			SmartFunction: &broker.Function{
				Oem:     "oem",
				Handle:  "function",
				Version: "version",
			},
		},
	}
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		filter: new("funct"),
		params: expectedParams,

		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteCapture(&capturedParams, expectedNodes),
	}

	actual, err := testProvider.Get()
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedNodes[1].Id, actual[0].Id)
	assert.NoError(t, err)
	assert.Equal(t, *expectedParams, capturedParams)
}

func TestUserWorkspaceNodesProvider_Get_Error(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteError(expectedError),
	}

	actual, err := testProvider.Get()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, expectedError)
}

func TestUserWorkspaceNodesProvider_Get_NoFilter(t *testing.T) {
	var capturedParams broker.WorkspaceNodeListParams
	var expectedParams = &broker.WorkspaceNodeListParams{
		WorkspaceName:  "workspace",
		WorkflowHandle: "workflow",
	}
	var expectedNodes = []broker.WorkspaceNode{
		{
			Id: int64(37),
		},
	}
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       expectedParams,
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteCapture(&capturedParams, expectedNodes),
	}

	actual, err := testProvider.Get()
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedNodes[0].Id, actual[0].Id)
	assert.NoError(t, err)
	assert.Equal(t, *expectedParams, capturedParams)
}

func TestUserWorkspaceNodesProvider_Get_NoResults(t *testing.T) {
	var capturedParams broker.WorkspaceNodeListParams
	var expectedParams = &broker.WorkspaceNodeListParams{
		WorkspaceName:  "workspace",
		WorkflowHandle: "workflow",
	}
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       expectedParams,
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteCapture(&capturedParams, []broker.WorkspaceNode{}),
	}

	actual, err := testProvider.Get()
	assert.Empty(t, actual)
	assert.NoError(t, err)
	assert.Equal(t, *expectedParams, capturedParams)
}

func TestNewUserWorkspaceNodesFactory(t *testing.T) {
	var testFactory = NewUserWorkspaceNodesFactory()

	assert.NotNil(t, testFactory.Provider())
}

func newWorkspaceNodeListTaskCompleteCapture(capture *broker.WorkspaceNodeListParams, seeded []broker.WorkspaceNode) broker.WorkspaceNodeListTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeListParams] {
		return &task.Task[broker.WorkspaceNodeListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newWorkspaceNodeResolveTaskCompleteCapture(capture *broker.WorkspaceNodeResolveParams) broker.WorkspaceNodeResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeResolveParams] {
		return &task.Task[broker.WorkspaceNodeResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetListTaskCompleteCapture(capture *broker.WorkspaceNodeSetListParams, seeded []broker.WorkspaceNodePort) broker.WorkspaceNodeSetListTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeSetListParams] {
		return &task.Task[broker.WorkspaceNodeSetListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeSetListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newWorkspaceNodeListTaskCompleteError(expected task.Error) broker.WorkspaceNodeListTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeListParams] {
		return &task.Task[broker.WorkspaceNodeListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return expected
			},
		}
	}
}

func newWorkspaceNodeResolveTaskCompleteError(expected task.Error) broker.WorkspaceNodeResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeResolveParams] {
		return &task.Task[broker.WorkspaceNodeResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return expected
			},
		}
	}
}
