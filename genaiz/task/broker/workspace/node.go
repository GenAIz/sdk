package workspace

type NodeSetParams struct {
	DataSetId  *int64
	PortHandle string
}

type NodeResolveParams struct {
	FlowId         *int64
	NodeId         *int64
	FnHandle       string
	WorkspaceId    *int64
	WorkspaceName  string
	WorkflowHandle string
}
