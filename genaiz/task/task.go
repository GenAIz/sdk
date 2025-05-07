package task

import (
	"context"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Executor[P any] interface {
	Execute(params P, logger logrus.Logger) State
	Pretend(params P, logger logrus.Logger)
}

type Env struct {
	Context context.Context
}

type State struct {
	Completed bool
	Error     error
	Output    string
	Logger    *logrus.Logger
}

type Task[P any] struct {
	Name          string
	OnPrepare     func(params *P, state *State) error
	OnIncomplete  func(params *P, state *State) error
	OnComplete    func(params *P, state *State) error
	OnPretend     func(params *P, state *State) error
	StopOnFailure bool
}

func (t Task[P]) Execute(params *P, logger *logrus.Logger) *State {
	var result = &State{Completed: false, Logger: logger}

	logger.Debugf("Preparing task %s", t.Name)

	if err := t.OnPrepare(params, result); err != nil {
		if t.OnIncomplete == nil {
			logger.Errorf("Preparing task %s failed: %s", t.Name, err)
			result.Error = err
			result.Completed = true
		} else {
			if err := t.OnIncomplete(params, result); err != nil {
				logger.Errorf("Handling incomplete task %s failed: %s", t.Name, err)
				result.Error = err
			} else {
				result.Error = nil
			}
		}
	}

	if !result.Completed && t.OnComplete != nil {
		if err := t.OnComplete(params, result); err == nil {
			result.Completed = true
		} else {
			result.Completed = false
			result.Error = err
		}
	} else {
		result.Completed = true
	}

	logger.Debugf("Completed task %s", t.Name)
	return result
}

func (t Task[P]) Pretend(params *P, logger *logrus.Logger) {
	var result = &State{Completed: false, Logger: logger}

	if t.OnPretend == nil {
		logger.Warningf("No pretend for task %s, skipping", t.Name)
	} else {
		if err := t.OnPretend(params, result); err != nil {
			logger.Errorf("Pretending task %s failed with err: %s", t.Name, err)
			cobra.CheckErr(err)
		}
	}
}
