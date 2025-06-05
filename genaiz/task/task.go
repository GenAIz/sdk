package task

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/lang"
)

// Executor qualifies methods for executing and pretending provided params P potentially handling a task.State
type Executor[P any] interface {
	Execute(params *P, logger *logrus.Logger) *State

	Pretend(params *P, logger *logrus.Logger)
}

// Env contains references to the environment Context handling the task execution
type Env struct {
	Context context.Context
}

// State tracks the result(s) and state values of Task(s)' execution
type State struct {
	Completed  bool                 // Indicates whether a Task completed, this could be completed with error, incomplete tasks should invalidate any Plan associated with their execution
	Containers *[]container.Summary // Containers contains the docker container summaries handled by tasks in a Plan
	Error      error                // If a Task completed with an error it will be in this field, it should be mutually exclusive with Output
	Output     string               // If a Task completed successfully, it may have stored a value in this field, it should be mutually exclusive with Error
	Logger     *logrus.Logger       // A reference to the Logger used by the Task and its associated Plan if any
}

// GetContainersSize returns the amount of containers held under Containers, or 0 if it was never initialized
func (s State) GetContainersSize() int {
	if s.Containers == nil {
		return 0
	}

	return len(*s.Containers)
}

// HasContainers indicates whether the Containers have any member container.Summary
func (s State) HasContainers() bool {
	return s.GetContainersSize() > 0
}

// Task is a representation of something to accomplish, referred to by Name. Tasks can either be Completed, Not Completed or Pretended. In all cases there needs to be a Preparation method specified to dictate the course of execution.
type Task[P any] struct {
	Name         string
	OnPrepare    func(params *P, state *State) error
	OnIncomplete func(params *P, state *State) error
	OnComplete   func(params *P, state *State) error
	OnPretend    func(params *P, state *State) error
}

// Execute will execute the task described provided it has a OnPrepare function set. This method will panic if no preparation can be called
//
//   - If the task returns an Error OnPrepare, Execute will trigger OnIncomplete if it set, otherwise the task will complete with an error
//   - If the task returns an Error OnPrepare and an Error OnIncomplete, the task completes with the error returned on OnIncomplete
//   - If preparation with or without incomplete is not completed and OnComplete is set, Execute will invoke it and return the result of this call in the returned State
//   - The task is always deemed completed unless OnComplete returns an error, in which case the task is considered incomplete
func (t Task[P]) Execute(params *P, logger *logrus.Logger) *State {
	var result = &State{Completed: false, Logger: logger}
	var err error

	logger.Debugf("Preparing task [%s]", t.Name)

	if err = t.OnPrepare(params, result); err != nil {
		if t.OnIncomplete == nil {
			logger.Errorf("Preparing task [%s] failed with error: %s", t.Name, err)
			result.Error = err
			result.Completed = true
		} else {
			if err = t.OnIncomplete(params, result); err != nil {
				logger.Errorf("Handling incomplete task [%s] failed with error: %s", t.Name, err)
				result.Error = err
			} else {
				result.Error = nil
			}
		}
	}

	if !result.Completed && t.OnComplete != nil {
		if err = t.OnComplete(params, result); err == nil {
			result.Completed = true
		} else {
			result.Completed = false
			result.Error = err
		}
	} else {
		result.Completed = true
	}

	logger.Debugf("Completed task [%s]", t.Name)
	return result
}

// Pretend invokes the OnPretend member of a task, after OnPrepare is called.
//
//   - If OnPretend is not set, the task simply does not support pretending and the execution returns with no errors
//   - If OnPrepare fails with an error Pretending is considered failed, no OnIncomplete handlers are called, as those may not be pretending anything
func (t Task[P]) Pretend(params *P, logger *logrus.Logger) {
	var result = &State{Completed: false, Logger: logger}

	if err := t.OnPrepare(params, result); err != nil {
		lang.HandleExit(err)
		return
	}

	if t.OnPretend == nil {
		logger.Warningf("No pretend for task [%s], skipping", t.Name)
	} else if err := t.OnPretend(params, result); err != nil {
		logger.Errorf("Pretending task [%s] failed with error: %s", t.Name, err)
		lang.HandleExit(err)
	}
}
