package docker

import (
	"errors"

	"genaiz.com/genaiz/task"
)

type BuildParams struct {
	Dockerfile    string
	DockerContext string
	DockerTag     string
	DockerVersion string
}

func BuildTask() task.Task[BuildParams] {
	return task.Task[BuildParams]{
		Name:         "docker-build",
		OnPrepare:    handleBuildContext,
		OnIncomplete: handleBuildCreate,
		OnPretend:    handleBuildPretend,
	}
}

func handleBuildContext(params *BuildParams, state *task.State) error {
	state.Logger.Debugf("Finding a docker image with the specified parameters")
	// TODO
	return errors.New("build not found")
}

func handleBuildCreate(params *BuildParams, state *task.State) error {
	state.Logger.Debugf("Building a docker image with the specified parameters")
	// TODO
	return nil
}

func handleBuildPretend(params *BuildParams, state *task.State) error {
	state.Logger.Debugf("Pretending a docker image build with the specified parameters")
	// TODO
	return nil
}
