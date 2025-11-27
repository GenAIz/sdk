package task

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang"
)

// Worker denotes a function acting on a task.State returning a pointer to a task.State, which may or may not be the original
type Worker func(*State) *State

// Planner qualifies methods for scheduling Worker(s)
type Planner interface {
	Sequence(...Worker)

	Single(Worker, func(interface{}))
}

// Plan is a group of success, failure and logging facilities to apply in particular manner
type Plan struct {
	Logger            *logrus.Logger    // Logger to use for all task.Task managed by this Plan
	OnSuccess         func(interface{}) // OnSuccess is a function called with the content of State.Output
	OnFailure         func(interface{}) // OnFailure is a function called with the content of State.Error
	ContinueOnFailure bool              // ContinueOnFailure will keep calling sequence workers even when a failure is reported
	PrintReportsOnly  bool              // PrintReportsOnly relies on State.Report for displaying the result of a Plan
}

// Sequence will execute all workers provided up until a worker's returned State signals an error or until the sequence is over. The state returned by each Worker is always passed to the following Worker in the list.
//
//   - If the last Worker returns no error and OnSuccess is set, it will be called with the output
//   - If the last Worker returned an error and OnFailure is set, it will be called with the error, and the Sequence will exit the program
//   - If the last Worker returned an error with no OnFailure set, the Sequence exits the program with the error
func (p Plan) Sequence(workers ...Worker) {
	var result = &State{Completed: false, Logger: p.Logger}

	for _, work := range workers {
		if result = work(result); (result.Error != nil && !p.ContinueOnFailure) || result.Abort {
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
		if p.PrintReportsOnly {
			p.OnSuccess(result.Reports)
		} else {
			p.OnSuccess(result.Output)
		}
	}
}

// Single will execute a single Worker provided.
//
//   - If the Worker returns an error and OnFailure is set, it will be called with the error and Single will call the exitHandler
//   - If the Worker is completed and OnSuccess is set, it will be called with the output
func (p Plan) Single(worker Worker, exitHandler func(interface{})) {
	var result = &State{Completed: false, Logger: p.Logger}

	if result = worker(result); result.Error != nil {
		if p.OnFailure != nil {
			p.OnFailure(result.Error)
		}

		exitHandler(result.Error)
	}

	if p.OnSuccess != nil {
		if p.PrintReportsOnly {
			p.OnSuccess(result.Reports)
		} else {
			p.OnSuccess(result.Output)
		}
	}
}

// NewPlan returns a standard simple plan with name, using the provided logger with a failure logger
func NewPlan(planName string, logger *logrus.Logger) *Plan {
	return newPlan(logger,
		func(msg interface{}) {
			logger.Errorf("%s failed with error: %s", planName, msg)
		}, successWriter)
}

// NewPretender encapsulates the type of params handled by the task to make it easier to sequence task pretends requiring different parameter types
func NewPretender[P any](params *P, task *Task[P]) Worker {
	return func(state *State) *State {
		state.Completed = false
		return task.pretend(params, state)
	}
}

// NewWorker encapsulates the type of params handled by the task to make it easier to sequence tasks executions requiring different parameter types
func NewWorker[P any](params *P, task *Task[P]) Worker {
	return func(state *State) *State {
		state.Completed = false
		return task.execute(params, state)
	}
}

// Attempt will execute a single Worker provided but will simply return without handling program exit
func Attempt[P any](plan Planner, params *P, task *Task[P]) {
	plan.Single(NewWorker(params, task), func(msg interface{}) {})
}

// Conditional is a shorthand way to express the execution of 2 task.Task(s) with the same params P depending on the veracity of the provided condition
func Conditional[P any](plan Planner, condition bool, params *P, ifTask, elseTask func() *Task[P]) {
	var worker Worker

	if condition {
		worker = NewWorker(params, ifTask())
	} else {
		worker = NewWorker(params, elseTask())
	}

	plan.Single(worker, lang.HandleExit)
}

// HandleFlag is a utility for creating handling functions which simply set the value of a flag on call
func HandleFlag(flag *bool, value bool) func(interface{}) {
	return func(i interface{}) {
		*flag = value
	}
}

// HandleString is a utility for creating handling function which simply set the value of a string by casting the result passed as a string
func HandleString(str *string) func(interface{}) {
	return func(i interface{}) {
		*str = cast.ToString(i)
	}
}

// Single is a shorthand way to express the execution of a task.Task with the provided par
func Single[P any](plan Planner, params *P, task *Task[P]) {
	plan.Single(NewWorker(params, task), lang.HandleExit)
}

func newPlan(logger *logrus.Logger, onFailure, onSuccess func(interface{})) *Plan {
	return &Plan{
		Logger:    logger,
		OnFailure: onFailure,
		OnSuccess: onSuccess,
	}
}

func successString(msg string) {
	if msg != "" {
		_, _ = fmt.Printf("%s\n", msg)
	}
}

func successWriter(i interface{}) {
	if messages, ok := i.([]string); ok {
		for _, msg := range messages {
			successString(msg)
		}
	} else {
		successString(cast.ToString(i))
	}
}
