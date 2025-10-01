package task

import (
	"context"

	"github.com/docker/docker/api/types/container"
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/lang"
)

// Executor qualifies methods for executing and pretending provided params P potentially handling a task.State
type Executor[P any] interface {
	Execute(*P, *State) *State

	Pretend(*P, *State) *State
}

// Env contains references to the environment Context handling the task execution
type Env struct {
	Context context.Context
}

// State tracks the result(s) and state values of Task(s)' execution
type State struct {
	Abort      bool                 // Abort indicates that any executing plan should exit without completing the rest of the tasks, usually done without an Error
	Completed  bool                 // Indicates whether a Task completed, this could be completed with error, incomplete tasks may invalidate Plan execution or not
	Containers *[]container.Summary // Containers contains the docker container summaries handled by tasks in a Plan
	Error      error                // If a Task completed with an error it will be in this field, it should be mutually exclusive with Output
	Internal   interface{}          // Internal is meant to allow tracking of structured task data
	Logger     *logrus.Logger       // A reference to the Logger used by the Task and its associated Plan if any
	Output     string               // If a Task completed successfully, it may have stored a value in this field, it should be mutually exclusive with Error
	Reports    []string             // reports can be used by a Task to print warnings or other user-bound information to STDOUT once all tasks are completed
}

// GetContainersSize returns the amount of containers held under Containers, or 0 if it was never initialized
func (s *State) GetContainersSize() int {
	if s.Containers == nil {
		return 0
	}

	return len(*s.Containers)
}

// HasContainers indicates whether the Containers have any member container.Summary
func (s *State) HasContainers() bool {
	return s.GetContainersSize() > 0
}

func (s *State) Report(message string) {
	s.Reports = append(s.Reports, message)
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
	return t.execute(params, &State{Logger: logger})
}

// Pretend invokes the OnPretend member of a task, after OnPrepare is called. The returned errors from OnPrepare are ignored and should be processed by OnPretend.
func (t Task[P]) Pretend(params *P, logger *logrus.Logger) *State {
	return t.pretend(params, &State{Logger: logger})
}

func (t Task[P]) execute(params *P, state *State) *State {
	var logger = state.Logger
	var err error

	logger.Debugf("Preparing task [%s]", t.Name)

	if err = t.OnPrepare(params, state); err != nil {
		state.Error = err

		if t.OnIncomplete == nil {
			logger.Errorf("Preparing task [%s] failed with error: %s", t.Name, err)
			state.Error = err
			state.Completed = true
		} else {
			if err = t.OnIncomplete(params, state); err != nil {
				logger.Errorf("Handling incomplete task [%s] failed with error: %s", t.Name, err)
				state.Error = err
			} else {
				state.Error = nil
			}
		}
	}

	if !state.Completed && t.OnComplete != nil {
		if err = t.OnComplete(params, state); err == nil {
			state.Completed = true
		} else {
			state.Completed = false
			state.Error = err
		}
	} else {
		state.Completed = true
	}

	logger.Debugf("Completed task [%s]", t.Name)
	return state
}

func (t Task[P]) pretend(params *P, state *State) *State {
	var logger = state.Logger

	state.Error = t.OnPrepare(params, state)

	if t.OnPretend == nil {
		logger.Warningf("No pretend for task [%s], skipping", t.Name)
	} else if err := t.OnPretend(params, state); err != nil {
		logger.Errorf("Pretending task [%s] failed with error: %s", t.Name, err)
		lang.HandleExit(err)
	} else {
		state.Error = nil
	}

	return state
}
