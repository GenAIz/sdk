package run

import (
	"errors"

	"genaiz.com/genaiz/task"
)

var (
	Errors = &runErrors{}
)

type runErrors struct {
	task.EagerErrors
	runDuplicated       task.Error
	runFlowUnknown      task.Error
	runIdKnown          task.Error
	runIdRequired       task.Error
	runSolutionUnknown  task.Error
	runUnprepared       task.Error
	runWorkflowUnknown  task.Error
	runWorkspaceUnknown task.Error
}

func (e *runErrors) IsRunDuplicated(err error) bool {
	return errors.Is(err, e.runDuplicated)
}

func (e *runErrors) RunDuplicated(id int64) error {
	var parent = e.SetOrReturn(&e.runDuplicated,
		task.MakeErrorProvider("workspace run is duplicated for flow"))

	return task.NewFormattedError(parent, "flow [%d] is already executing on this workspace", id)
}

func (e *runErrors) RunFlowUnknown() error {
	return e.SetOrReturn(&e.runFlowUnknown,
		task.MakeErrorProvider("run flow can not be located"))
}

func (e *runErrors) RunIdKnown() error {
	return e.SetOrReturn(&e.runIdKnown,
		task.MakeErrorProvider("run id is already known"))
}

func (e *runErrors) RunIdRequired() error {
	return e.SetOrReturn(&e.runIdRequired,
		task.MakeErrorProvider("run id is required"))
}

func (e *runErrors) RunSolutionUnknown() error {
	return e.SetOrReturn(&e.runSolutionUnknown,
		task.MakeErrorProvider("run solution is unknown"))
}

func (e *runErrors) RunUnprepared() error {
	return e.SetOrReturn(&e.runUnprepared,
		task.MakeErrorProvider("run workspace flow is unprepared"))
}

func (e *runErrors) RunWorkflowUnknown() error {
	return e.SetOrReturn(&e.runWorkflowUnknown,
		task.MakeErrorProvider("run solution workflow is unknown"))
}

func (e *runErrors) RunWorkspaceUnknown() error {
	return e.SetOrReturn(&e.runWorkspaceUnknown,
		task.MakeErrorProvider("run workspace can not be located"))
}
