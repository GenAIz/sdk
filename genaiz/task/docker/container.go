package docker

import "genaiz.com/genaiz/task"

type ContainerParams struct {
	RunParams
	Name  string
	Force bool
}

func CreateTask() task.Task[ContainerParams] {
	return task.Task[ContainerParams]{
		Name:         "docker-create",
		OnPrepare:    handleContainerContext,
		OnIncomplete: handleContainerCreate,
		OnPretend:    handleContainerCreatePretend,
	}
}

func DisposeTask() task.Task[ContainerParams] {
	return task.Task[ContainerParams]{
		Name:       "docker-rm",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerDisposal,
		OnPretend:  handleContainerDisposalPretend,
	}
}

func StartTask() task.Task[ContainerParams] {
	return task.Task[ContainerParams]{
		Name:       "docker-start",
		OnPrepare:  handleContainerImageContext,
		OnComplete: handleContainerStart,
		OnPretend:  handleContainerStartPretend,
	}
}

func StopTask() task.Task[ContainerParams] {
	return task.Task[ContainerParams]{
		Name:       "docker-stop",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerStop,
		OnPretend:  handleContainerStopPretend,
	}
}

func handleContainerContext(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Finding a docker container with the specified name")
	// TODO
	return nil
}

func handleContainerCreate(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Creating a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerCreatePretend(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Pretending to create a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerDisposal(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Disposing a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerDisposalPretend(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Pretending of disposing of a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerImageContext(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Finding a docker image and a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerStart(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Starting a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerStartPretend(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Pretending to start a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerStop(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Stopping a docker container with the specified parameters")
	// TODO
	return nil
}

func handleContainerStopPretend(params *ContainerParams, state *task.State) error {
	state.Logger.Debugf("Pretending to stop a docker container with the specified parameters")
	// TODO
	return nil
}
