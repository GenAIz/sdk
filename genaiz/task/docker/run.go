package docker

import (
	"genaiz.com/genaiz/task"
)

type RunParams struct {
	DockerImage string
	MountInput  string
	MountOutput string
}

func DebugTask() task.Task[RunParams] {
	return task.Task[RunParams]{
		Name:       "docker-debug",
		OnPrepare:  handleRunContext,
		OnComplete: handleDebugCompletion,
		OnPretend:  handleDebugPretend,
	}
}

func RunTask() task.Task[RunParams] {
	return task.Task[RunParams]{
		Name:       "docker-run",
		OnPrepare:  handleRunContext,
		OnComplete: handleRunCompletion,
		OnPretend:  handleRunPretend,
	}
}

func TestTask() task.Task[RunParams] {
	return task.Task[RunParams]{
		Name:       "docker-test",
		OnPrepare:  handleRunContext,
		OnComplete: handleTestCompletion,
		OnPretend:  handleTestPretend,
	}
}

func handleDebugCompletion(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Running a docker image in interactive mode with the entrypoint set to sh")
	// TODO
	return nil
}

func handleDebugPretend(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Pretending a docker image run in interactive mode with the entrypoint set to sh")
	// TODO
	return nil
}

func handleRunCompletion(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Running a docker image in detached mode with the specified params")
	// TODO
	state.Completed = true
	return nil
}

func handleRunContext(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Finding a docker image with the specified parameters")
	// TODO
	return nil
}

func handleRunPretend(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Pretending to run a docker image run in detached mode with the specified params")
	// TODO
	return nil
}

func handleTestCompletion(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Running a docker image in attached mode with the specified params")
	// TODO
	state.Completed = true
	return nil
}

func handleTestPretend(params *RunParams, state *task.State) error {
	state.Logger.Debugf("Pretending to run a docker image in attached mode with the specified params")
	// TODO
	return nil
}
