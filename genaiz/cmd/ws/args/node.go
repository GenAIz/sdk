package args

import (
	"strconv"

	"genaiz.com/genaiz/task/broker"
)

type NodeResolution interface {
	MakeOptionMap() map[string]string

	ResolveNodeParams(*broker.Broker) *broker.WorkspaceNodeResolveParams
}

type nodeResolution struct {
	node      string
	flow      string
	workspace string
}

func (na nodeResolution) MakeOptionMap() map[string]string {
	var result = map[string]string{}

	if _, err := strconv.ParseInt(na.workspace, 10, 64); err == nil {
		result["workspace id"] = na.workspace
	} else {
		result["workspace name"] = na.workspace
	}

	if _, err := strconv.ParseInt(na.flow, 10, 64); err == nil {
		result["flow id"] = na.flow
	} else {
		result["workflow handle"] = na.flow
	}

	if _, err := strconv.ParseInt(na.node, 10, 64); err == nil {
		result["node id"] = na.node
	} else {
		result["function handle"] = na.node
	}

	return result
}

func (na nodeResolution) ResolveNodeParams(brokerParams *broker.Broker) *broker.WorkspaceNodeResolveParams {
	var result = &broker.WorkspaceNodeResolveParams{
		Broker: *brokerParams,
	}

	if wsId, err := strconv.ParseInt(na.workspace, 10, 64); err == nil {
		result.WorkspaceId = new(wsId)
	} else {
		result.WorkspaceName = na.workspace
	}

	if flId, err := strconv.ParseInt(na.flow, 10, 64); err == nil {
		result.FlowId = new(flId)
	} else {
		result.WorkflowHandle = na.flow
	}

	if ndId, err := strconv.ParseInt(na.node, 10, 64); err == nil {
		result.NodeId = new(ndId)
	} else {
		result.FnHandle = na.node
	}

	return result
}

func NewNodeResolution(workspace, workflow, node string) NodeResolution {
	return &nodeResolution{
		workspace: workspace,
		flow:      workflow,
		node:      node,
	}
}
