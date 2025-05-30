package task

import (
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/lang"
)

// Worker denotes a function acting on a task.State returning a pointer to a task.State, which may or may not be the original
type Worker func(*State) *State

// Planner qualifies methods for scheduling Worker(s)
type Planner interface {
	Sequence(...Worker)

	Single(Worker)
}

// Plan is a group of success, failure and logging facilities to apply in particular manner
type Plan struct {
	Logger    *logrus.Logger    // Logger to use for all task.Task managed by this Plan
	OnSuccess func(interface{}) // OnSuccess is a function called with the content of State.Output
	OnFailure func(interface{}) // OnFailure is a function called with the content of State.Error
}

// Sequence will execute all workers provided up until a worker's returned State signals an error or until the sequence is over. The state returned by each Worker is always passed to the following Worker in the list.
//
//   - If the last Worker returns no error and OnSuccess is set, it will be called with the output
//   - If the last Worker returned an error and OnFailure is set, it will be called with the error, and the Sequence will exit the program
//   - If the last Worker returned an error with no OnFailure set, the Sequence exits the program with the error
func (p Plan) Sequence(workers ...Worker) {
	var result = &State{Completed: false, Logger: p.Logger}

	for _, work := range workers {
		if result = work(result); result.Error != nil {
			break
		}
	}

	if result.Error != nil {
		if p.OnFailure != nil {
			p.OnFailure(result.Error)
		} else {
			p.Logger.Errorf("Failure with error: %s", result.Error)
		}

		lang.HandleExit(result.Error)
	} else if p.OnSuccess != nil {
		p.OnSuccess(result.Output)
	}
}

// Single will execute a single Worker provided.
//
//   - If the Worker returns an error and OnFailure is set, it will be called with the error and Single will exit the program.
//   - If the Worker is completed and OnSuccess is set, it will be called with the output
func (p Plan) Single(worker Worker) {
	var result = &State{Completed: false, Logger: p.Logger}

	if result = worker(result); result.Error != nil {
		if p.OnFailure != nil {
			p.OnFailure(result.Error)
		}

		lang.HandleExit(result.Error)
	}

	if p.OnSuccess != nil {
		p.OnSuccess(result.Output)
	}
}

// NewWorker encapsulates the type of params handled by the task to make it easier to sequence tasks requiring different parameter types
func NewWorker[P any](params *P, task *Task[P]) Worker {
	return func(state *State) *State {
		return task.Execute(params, state.Logger)
	}
}

// Conditional is a shorthand way to express the execution of 2 task.Task(s) with the same params P depending on the veracity of the provided condition
func Conditional[P any](plan Planner, condition bool, params *P, ifTask, elseTask func() *Task[P]) {
	var worker Worker

	if condition {
		worker = NewWorker(params, ifTask())
	} else {
		worker = NewWorker(params, elseTask())
	}

	plan.Single(worker)
}

// Single is a shorthand way to express the execution of a task.Task with the provided par
func Single[P any](plan Planner, params *P, task *Task[P]) {
	plan.Single(NewWorker(params, task))
}
