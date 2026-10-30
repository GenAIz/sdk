package docker

import (
	"fmt"
	"time"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/client"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker/adapt"
)

var (
	RunTasks struct {
		StandardTaskFactory RunTaskFactory
		TestTaskFactory     RunTaskFactory
	}
)

type RunTask = *task.Task[ContainerParams]
type RunTaskFactory = func() *task.Task[ContainerParams]

type RunParams struct {
	task.Env
	adapt.RunArguments
}

func handleRunCompletion(params *ContainerParams, state *task.State) error {
	var err error

	params.Name = params.MakeDisposableName()

	if err = handleContainerCreate(params, state); err == nil {
		var dockerState = NewClientState(state)

		dockerState.AddContainers(container.Summary{
			ID:      state.Output,
			Created: time.Now().UnixMilli(),
		})

		if err = handleContainerStart(params, state); err == nil {
			state.Reportf("Ran container %s with image %s successfully", state.Output, params.DockerImage)
		}
	}

	state.Completed = true
	return err
}

func handleRunContext(params *ContainerParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var listOptions = client.ImageListOptions{
		Filters: make(client.Filters).
			Add("reference", params.DockerImage),
	}
	var result client.ImageListResult
	var err error

	state.Logger.Debugf("Finding docker image [%s]", params.DockerImage)

	if result, err = dockerClient.ImageList(params.Context, listOptions); err == nil {
		if len(result.Items) > 0 {
			var imageSummary = result.Items[0]

			state.Logger.Debugf("Found image id [%s] with reference [%s]", imageSummary.ID[7:19], params.DockerImage)
			state.Output = imageSummary.ID
			return nil
		}

		state.Logger.Errorf("Could not find an image for reference [%s]", params.DockerImage)
		return Errors.RunImageNotFound()
	}

	return err
}

func handleRunPretend(params *ContainerParams, state *task.State) error {
	var dockerImage = stringz.FirstNonEmpty(state.Output, params.DockerImage)
	var optionParams = params.MountArguments()
	var detach, dispose string

	state.Logger.Debugf("Pretending running docker image [%s]", dockerImage)

	if params.Dispose {
		dispose = "--rm "
	}

	if !params.Attached {
		detach = "-d "
	}

	fmt.Printf("docker run %s%s%s%s\n", dispose, detach, optionParams, dockerImage)
	state.Completed = true
	state.Output = ""
	return nil
}

func handleTestCompletion(params *ContainerParams, state *task.State) error {
	var err error

	params.Name = params.MakeDisposableName()

	if err = handleContainerCreate(params, state); err == nil {
		var dockerState = NewClientState(state)

		dockerState.AddContainers(container.Summary{
			ID:      state.Output,
			Created: time.Now().UnixMilli(),
		})
		err = handleContainerAttach(params, state)
	}

	state.Completed = true
	state.Output = ""
	return err
}

func handleTestPretend(params *ContainerParams, state *task.State) error {
	var optionParams = params.MountArguments()
	var dockerImage = stringz.FirstNonEmpty(state.Output, params.DockerImage)
	var dispose string

	state.Logger.Debugf("Pretending testing docker image [%s]", dockerImage)

	if params.Dispose {
		dispose = "--rm "
	}

	fmt.Printf("docker run %s%s%s\n", dispose, optionParams, dockerImage)
	state.Completed = true
	state.Output = ""
	return nil
}

func init() {
	RunTasks.StandardTaskFactory = newRunTask
	RunTasks.TestTaskFactory = newTestTask
}

func newRunTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-run",
		OnPrepare:  handleRunContext,
		OnComplete: handleRunCompletion,
		OnPretend:  handleRunPretend,
	}
}

func newTestTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-test",
		OnPrepare:  handleRunContext,
		OnComplete: handleTestCompletion,
		OnPretend:  handleTestPretend,
	}
}
