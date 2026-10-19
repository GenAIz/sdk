package workspace

import (
	"genaiz.com/genaiz/task"
)

var (
	Errors = &workspaceErrors{}
)

type workspaceErrors struct {
	task.EagerErrors
	workflowHandleRequired        task.Error
	workflowIdKnown               task.Error
	workflowIdRequired            task.Error
	workflowOemRequired           task.Error
	workflowVersionRequired       task.Error
	workspaceConflict             task.Error
	workspaceDefInvalid           task.Error
	workspaceFlowConflict         task.Error
	workspaceFlowDefInvalid       task.Error
	workspaceFlowRequired         task.Error
	workspaceFlowUnresolved       task.Error
	workspaceIdKnown              task.Error
	workspaceIdRequired           task.Error
	workspaceInvalidNco           task.Error
	workspaceInvalidOwner         task.Error
	workspaceNameRequired         task.Error
	workspaceNodeFnHandleRequired task.Error
	workspaceNodeFnPortNotFound   task.Error
	workspaceNodeFnPortRequired   task.Error
	workspaceNodeKnown            task.Error
	workspaceNodeRequired         task.Error
	workspaceNodeSetKnown         task.Error
	workspaceNodeSetRequired      task.Error
	workspaceNodeSourceRequired   task.Error
	workspaceNodeStoreRequired    task.Error
	workspaceNotFound             task.Error
	workspaceVisibilityRequired   task.Error
}

func (e *workspaceErrors) WorkflowHandleRequired() error {
	return e.SetOrReturn(&e.workflowHandleRequired,
		task.MakeErrorProvider("workflow handle is required"))
}

func (e *workspaceErrors) WorkflowIdKnown() error {
	return e.SetOrReturn(&e.workflowIdKnown,
		task.MakeErrorProvider("workflow id is already known"))
}

func (e *workspaceErrors) WorkflowIdRequired() error {
	return e.SetOrReturn(&e.workflowIdRequired,
		task.MakeErrorProvider("workflow id is required"))
}

func (e *workspaceErrors) WorkflowOemRequired() error {
	return e.SetOrReturn(&e.workflowOemRequired,
		task.MakeErrorProvider("workflow solution oem is required"))
}

func (e *workspaceErrors) WorkflowVersionRequired() error {
	return e.SetOrReturn(&e.workflowVersionRequired,
		task.MakeErrorProvider("workflow solution version is required"))
}

func (e *workspaceErrors) WorkspaceConflict() error {
	return e.SetOrReturn(&e.workspaceConflict,
		task.MakeErrorProvider("workspace can be several possibilities"))
}

func (e *workspaceErrors) WorkspaceDefInvalid() error {
	return e.SetOrReturn(&e.workspaceDefInvalid,
		task.MakeErrorProvider("no workspace definition provided"))
}

func (e *workspaceErrors) WorkspaceFlowConflict() error {
	return e.SetOrReturn(&e.workspaceFlowConflict,
		task.MakeErrorProvider("workflow present under multiple workspace flows"))
}

func (e *workspaceErrors) WorkspaceFlowDefInvalid() error {
	return e.SetOrReturn(&e.workspaceFlowDefInvalid,
		task.MakeErrorProvider("workspace flow definition is invalid"))
}

func (e *workspaceErrors) WorkspaceFlowRequired() error {
	return e.SetOrReturn(&e.workspaceFlowRequired,
		task.MakeErrorProvider("workspace flow can not be located"))
}

func (e *workspaceErrors) WorkspaceFlowUnresolved() error {
	return e.SetOrReturn(&e.workspaceFlowUnresolved,
		task.MakeErrorProvider("workspace flow needs to be located"))
}

func (e *workspaceErrors) WorkspaceIdKnown() error {
	return e.SetOrReturn(&e.workspaceIdKnown,
		task.MakeErrorProvider("workspace id is already known"))
}

func (e *workspaceErrors) WorkspaceIdRequired() error {
	return e.SetOrReturn(&e.workspaceIdRequired,
		task.MakeErrorProvider("workspace id is required"))
}

func (e *workspaceErrors) WorkspaceInvalidNco() error {
	return e.SetOrReturn(&e.workspaceInvalidNco,
		task.MakeErrorProvider("workspace creation date is invalid"))
}

func (e *workspaceErrors) WorkspaceInvalidOwner() error {
	return e.SetOrReturn(&e.workspaceInvalidOwner,
		task.MakeErrorProvider("workspace ownership can not be established with the selected session"))
}

func (e *workspaceErrors) WorkspaceNameRequired() error {
	return e.SetOrReturn(&e.workspaceNameRequired,
		task.MakeErrorProvider("workspace name is required"))
}

func (e *workspaceErrors) WorkspaceNodeFnHandleRequired() error {
	return e.SetOrReturn(&e.workspaceNodeFnHandleRequired,
		task.MakeErrorProvider("workspace node function handle is required"))
}

func (e *workspaceErrors) WorkspaceNodeFnPortNotFound() error {
	return e.SetOrReturn(&e.workspaceNodeFnPortNotFound,
		task.MakeErrorProvider("workspace node function port can not be located"))
}

func (e *workspaceErrors) WorkspaceNodeFnPortRequired() error {
	return e.SetOrReturn(&e.workspaceNodeFnPortRequired,
		task.MakeErrorProvider("workspace node function port is required"))
}

func (e *workspaceErrors) WorkspaceNodeKnown() error {
	return e.SetOrReturn(&e.workspaceNodeKnown,
		task.MakeErrorProvider("workspace node is already known"))
}

func (e *workspaceErrors) WorkspaceNodeRequired() error {
	return e.SetOrReturn(&e.workspaceNodeRequired,
		task.MakeErrorProvider("workspace node can not be located"))
}

func (e *workspaceErrors) WorkspaceNodeSetKnown() error {
	return e.SetOrReturn(&e.workspaceNodeSetKnown,
		task.MakeErrorProvider("workspace node data set is known"))
}

func (e *workspaceErrors) WorkspaceNodeSetRequired() error {
	return e.SetOrReturn(&e.workspaceNodeSetRequired,
		task.MakeErrorProvider("workspace node data set is required"))
}

func (e *workspaceErrors) WorkspaceNodeSourceRequired() error {
	return e.SetOrReturn(&e.workspaceNodeSourceRequired,
		task.MakeErrorProvider("workspace node data source is required"))
}

func (e *workspaceErrors) WorkspaceNodeStoreRequired() error {
	return e.SetOrReturn(&e.workspaceNodeStoreRequired,
		task.MakeErrorProvider("workspace node data store is required"))
}

func (e *workspaceErrors) WorkspaceNotFound() error {
	return e.SetOrReturn(&e.workspaceNotFound,
		task.MakeErrorProvider("workspace could not be found"))
}

func (e *workspaceErrors) WorkspaceVisibilityRequired() error {
	return e.SetOrReturn(&e.workspaceVisibilityRequired,
		task.MakeErrorProvider("workspace visibility is required"))
}
