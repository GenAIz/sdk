package dataset

import "genaiz.com/genaiz/task"

var (
	Errors = &dataSetErrors{}
)

type dataSetErrors struct {
	task.EagerErrors
	dataSetFileExists        task.Error
	dataSetFileUnknown       task.Error
	dataSetIdKnown           task.Error
	dataSetIdUnknown         task.Error
	dataSetHandleRequired    task.Error
	dataSetWorkspaceRequired task.Error
}

func (e *dataSetErrors) DataSetFileExists() error {
	return e.SetOrReturn(&e.dataSetFileExists,
		task.MakeErrorProvider("file exists and should not be overwritten"))
}

func (e *dataSetErrors) DataSetFileUnknown() error {
	return e.SetOrReturn(&e.dataSetFileUnknown,
		task.MakeErrorProvider("local file path is missing"))
}

func (e *dataSetErrors) DataSetHandleRequired() error {
	return e.SetOrReturn(&e.dataSetHandleRequired,
		task.MakeErrorProvider("data set handle is required"))
}

func (e *dataSetErrors) DataSetIdKnown() error {
	return e.SetOrReturn(&e.dataSetIdKnown,
		task.MakeErrorProvider("data set id known"))
}

func (e *dataSetErrors) DataSetIdUnknown() error {
	return e.SetOrReturn(&e.dataSetIdUnknown,
		task.MakeErrorProvider("data set can not be located"))
}

func (e *dataSetErrors) DataSetWorkspaceRequired() error {
	return e.SetOrReturn(&e.dataSetWorkspaceRequired,
		task.MakeErrorProvider("data set workspace is required"))
}
