package task

import (
	"errors"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

type Planner[P any] interface {
	Single(params *P, task Task[P])
}

type Sequencer interface {
	Sequence(onSuccess func(out string), execs ...func(state *State)) error
}

type Plan[P any] struct {
	Logger    *logrus.Logger
	OnSuccess func(string)
	OnError   func(error)
}

func Execution[P any](params *P, task Task[P]) func(*State) error {
	return func(state *State) error {
		var st = task.Execute(params, state.Logger)

		state.Output = st.Output
		state.Error = st.Error
		state.Completed = st.Completed

		if state.Error != nil {
			return state.Error
		} else {
			return nil
		}
	}
}

func (p Plan[P]) Sequence(execs ...func(state *State) error) {
	var result = &State{Completed: false, Logger: p.Logger}

	for _, exec := range execs {
		if err := exec(result); err != nil {
			break
		}
	}

	if result.Completed {
		if p.OnSuccess != nil {
			p.OnSuccess(result.Output)
		}
	} else if p.OnError != nil {
		p.OnError(result.Error)
	} else {
		p.Logger.Errorf("Failure with error: %s", result.Error)
	}

	cobra.CheckErr(result.Error)
}

func (p Plan[P]) Single(params *P, task Task[P]) {
	if state := task.Execute(params, p.Logger); state.Error != nil {
		p.OnError(state.Error)
		cobra.CheckErr(state.Error)
	} else if state.Completed {
		p.OnSuccess(state.Output)
	} else {
		cobra.CheckErr(errors.New("incomplete task failed without error"))
	}
}
