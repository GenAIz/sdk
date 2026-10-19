package args

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/broker"
)

func TestNodeResolution_MakeOptionMap(t *testing.T) {
	var expectedWorkspaceId = "37"
	var expectedWorkflowId = "42"
	var expectedNodeId = "69"
	var testResolution = NewNodeResolution(expectedWorkspaceId, expectedWorkflowId, expectedNodeId)

	actual := testResolution.MakeOptionMap()
	assert.Equal(t, expectedWorkspaceId, actual["workspace id"])
	assert.Equal(t, expectedWorkflowId, actual["flow id"])
	assert.Equal(t, expectedNodeId, actual["node id"])
}

func TestNodeResolution_MakeOptionMap_FunctionHandle(t *testing.T) {
	var expectedWorkspaceId = "37"
	var expectedWorkflowId = "42"
	var expectedFunctionHandle = "function"
	var testResolution = NewNodeResolution(expectedWorkspaceId, expectedWorkflowId, expectedFunctionHandle)

	actual := testResolution.MakeOptionMap()
	assert.Equal(t, expectedWorkspaceId, actual["workspace id"])
	assert.Equal(t, expectedWorkflowId, actual["flow id"])
	assert.Equal(t, expectedFunctionHandle, actual["function handle"])
}

func TestNodeResolution_MakeOptionMap_WorkflowHandle(t *testing.T) {
	var expectedWorkspaceId = "37"
	var expectedWorkflowHandle = "workflow"
	var expectedNodeId = "69"
	var testResolution = NewNodeResolution(expectedWorkspaceId, expectedWorkflowHandle, expectedNodeId)

	actual := testResolution.MakeOptionMap()
	assert.Equal(t, expectedWorkspaceId, actual["workspace id"])
	assert.Equal(t, expectedWorkflowHandle, actual["workflow handle"])
	assert.Equal(t, expectedNodeId, actual["node id"])
}

func TestNodeResolution_MakeOptionMap_WorkspaceName(t *testing.T) {
	var expectedWorkspaceName = "workspace"
	var expectedWorkflowId = "42"
	var expectedNodeId = "69"
	var testResolution = NewNodeResolution(expectedWorkspaceName, expectedWorkflowId, expectedNodeId)

	actual := testResolution.MakeOptionMap()
	assert.Equal(t, expectedWorkspaceName, actual["workspace name"])
	assert.Equal(t, expectedWorkflowId, actual["flow id"])
	assert.Equal(t, expectedNodeId, actual["node id"])
}

func TestNodeResolution_ResolveNodeParams(t *testing.T) {
	var expectedWorkspaceId = int64(37)
	var expectedWorkflowId = int64(42)
	var expectedNodeId = int64(69)
	var testResolution = NewNodeResolution(cast.ToString(expectedWorkspaceId),
		cast.ToString(expectedWorkflowId), cast.ToString(expectedNodeId))
	var testBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), ".auth"),
	}

	actual := testResolution.ResolveNodeParams(testBroker)

	if actual.WorkspaceId == nil {
		assert.Fail(t, "expected workspaceId not nil")
	} else {
		assert.Equal(t, expectedWorkspaceId, *actual.WorkspaceId)
	}

	if actual.FlowId == nil {
		assert.Fail(t, "expected flowId not nil")
	} else {
		assert.Equal(t, expectedWorkflowId, *actual.FlowId)
	}

	if actual.NodeId == nil {
		assert.Fail(t, "expected nodeId not nil")
	} else {
		assert.Equal(t, expectedNodeId, *actual.NodeId)
	}
}

func TestNodeResolution_ResolveNodeParams_FunctionHandle(t *testing.T) {
	var expectedWorkspaceId = int64(37)
	var expectedWorkflowId = int64(42)
	var expectedFunctionHandle = "function"
	var testResolution = NewNodeResolution(cast.ToString(expectedWorkspaceId),
		cast.ToString(expectedWorkflowId), expectedFunctionHandle)
	var testBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), ".auth"),
	}

	actual := testResolution.ResolveNodeParams(testBroker)

	if actual.WorkspaceId == nil {
		assert.Fail(t, "expected workspaceId not nil")
	} else {
		assert.Equal(t, expectedWorkspaceId, *actual.WorkspaceId)
	}

	if actual.FlowId == nil {
		assert.Fail(t, "expected flowId not nil")
	} else {
		assert.Equal(t, expectedWorkflowId, *actual.FlowId)
	}

	assert.Nil(t, actual.NodeId)
	assert.Equal(t, expectedFunctionHandle, actual.FnHandle)
}

func TestNodeResolution_ResolveNodeParams_WorkflowHandle(t *testing.T) {
	var expectedWorkspaceId = int64(37)
	var expectedWorkflowHandle = "workflow"
	var expectedNodeId = int64(69)
	var testResolution = NewNodeResolution(cast.ToString(expectedWorkspaceId),
		expectedWorkflowHandle, cast.ToString(expectedNodeId))
	var testBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), ".auth"),
	}

	actual := testResolution.ResolveNodeParams(testBroker)

	if actual.WorkspaceId == nil {
		assert.Fail(t, "expected workspaceId not nil")
	} else {
		assert.Equal(t, expectedWorkspaceId, *actual.WorkspaceId)
	}

	assert.Nil(t, actual.FlowId)
	assert.Equal(t, expectedWorkflowHandle, actual.WorkflowHandle)

	if actual.NodeId == nil {
		assert.Fail(t, "expected nodeId not nil")
	} else {
		assert.Equal(t, expectedNodeId, *actual.NodeId)
	}
}

func TestNodeResolution_ResolveNodeParams_WorkspaceName(t *testing.T) {
	var expectedWorkspaceName = "workspace"
	var expectedWorkflowId = int64(42)
	var expectedNodeId = int64(69)
	var testResolution = NewNodeResolution(expectedWorkspaceName,
		cast.ToString(expectedWorkflowId), cast.ToString(expectedNodeId))
	var testBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), ".auth"),
	}

	actual := testResolution.ResolveNodeParams(testBroker)

	assert.Nil(t, actual.WorkspaceId)
	assert.Equal(t, expectedWorkspaceName, actual.WorkspaceName)

	if actual.FlowId == nil {
		assert.Fail(t, "expected flowId not nil")
	} else {
		assert.Equal(t, expectedWorkflowId, *actual.FlowId)
	}

	if actual.NodeId == nil {
		assert.Fail(t, "expected nodeId not nil")
	} else {
		assert.Equal(t, expectedNodeId, *actual.NodeId)
	}
}
