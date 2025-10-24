package docker

import (
	"errors"
	"fmt"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
)

var (
	ErrorNoImage = errors.New("image not found")
)

type RunParams struct {
	task.Env
	Attached    bool
	Dispose     bool
	Interactive bool
}

func (r RunParams) WaitCondition() container.WaitCondition {
	if r.Dispose {
		return container.WaitConditionRemoved
	}

	return container.WaitConditionNextExit
}

func NewRunTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-run",
		OnPrepare:  handleRunContext,
		OnComplete: handleRunCompletion,
		OnPretend:  handleRunPretend,
	}
}

func NewTestTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-test",
		OnPrepare:  handleRunContext,
		OnComplete: handleTestCompletion,
		OnPretend:  handleTestPretend,
	}
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
	var listFilters = filters.NewArgs(
		filters.Arg("reference", params.DockerImage),
	)

	state.Logger.Debugf("Finding docker image [%s]", params.DockerImage)

	if summaries, err := dockerClient.ImageList(params.Context, image.ListOptions{Filters: listFilters}); err == nil {
		if len(summaries) == 0 {
			state.Logger.Errorf("Could not find an image for reference [%s]", params.DockerImage)
			return ErrorNoImage
		} else {
			var imageSummary = summaries[0]

			state.Logger.Debugf("Found image id [%s] with reference [%s]", imageSummary.ID[7:19], params.DockerImage)
			state.Output = imageSummary.ID
			return nil
		}
	} else {
		return err
	}
}

func handleRunPretend(params *ContainerParams, state *task.State) error {
	var dockerImage = stringz.FirstNonEmpty(state.Output, params.DockerImage)
	var optionParams = params.fmtArgs()
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
	var optionParams = params.fmtArgs()
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
