package workspace

type FlowCreateParams struct {
	WorkspaceId *int64
	WorkflowId  *int64
	Name        string
	Description string
}

func (p FlowCreateParams) IsValid() bool {
	return p.WorkspaceId != nil && p.WorkflowId != nil
}

type FlowReadParams struct {
	WorkspaceId *int64
	FlowId      *int64
}
