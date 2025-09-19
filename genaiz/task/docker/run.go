package docker

import (
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/version"
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

func NewDebugTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-debug",
		OnPrepare:  handleRunContext,
		OnComplete: handleDebugCompletion,
		OnPretend:  handleDebugPretend,
	}
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

func handleDebugCompletion(params *ContainerParams, state *task.State) error {
	return fmt.Errorf("not implemented in version [%s]", version.GetVersion())
}

func handleDebugPretend(params *ContainerParams, state *task.State) error {
	var optionParams = fmtParams(params)
	var dispose string

	state.Logger.Debugf("Pretending debugging docker image [%s]", params.DockerImage)

	if params.Dispose {
		dispose = "--rm"
	}

	fmt.Printf("docker run %s -it --entrypoint sh %s%s\n", dispose, optionParams, params.DockerImage)
	state.Completed = true
	return nil
}

func handleRunCompletion(params *ContainerParams, state *task.State) error {
	var err error

	params.Name = makeDisposableName(params)

	if err = handleContainerCreate(params, state); err == nil {
		state.Containers = &[]container.Summary{
			{
				ID:      state.Output,
				Created: time.Now().UnixMilli(),
			},
		}

		err = handleContainerStart(params, state)
	}

	state.Completed = true
	return err
}

func handleRunContext(params *ContainerParams, state *task.State) error {
	var listFilters = filters.NewArgs(
		filters.Arg("reference", params.DockerImage),
	)

	state.Logger.Debugf("Finding docker image [%s]", params.DockerImage)

	if summaries, err := dockerClient.ImageList(params.Context, image.ListOptions{Filters: listFilters}); err == nil {
		if len(summaries) == 0 {
			state.Logger.Errorf("Could not find an image for reference [%s]", params.DockerImage)
			return errors.New("image not found")
		} else {
			var imageSummary = summaries[0]

			state.Logger.Debugf("Found image id [%s] with reference [%s]", imageSummary.ID[:12], params.DockerImage)
			state.Output = fmt.Sprintf("%s", imageSummary.ID[:12])
			return nil
		}
	} else {
		return err
	}
}

func handleRunPretend(params *ContainerParams, state *task.State) error {
	var optionParams = fmtParams(params)
	var detach, dispose string

	state.Logger.Debugf("Pretending running docker image [%s]", state.Output)

	if params.Dispose {
		dispose = "--rm "
	}

	if !params.RunParams.Attached {
		detach = "-d "
	}

	fmt.Printf("docker run %s%s%s%s\n", dispose, detach, optionParams, state.Output)
	state.Completed = true
	state.Output = ""
	return nil
}

func handleTestCompletion(params *ContainerParams, state *task.State) error {
	var err error

	params.Name = makeDisposableName(params)

	if err = handleContainerCreate(params, state); err == nil {
		state.Containers = &[]container.Summary{
			{
				ID:      state.Output,
				Created: time.Now().UnixMilli(),
			},
		}

		err = handleContainerAttach(params, state)
	}

	state.Completed = true
	state.Output = ""
	return err
}

func handleTestPretend(params *ContainerParams, state *task.State) error {
	var optionParams = fmtParams(params)
	var dispose string

	state.Logger.Debugf("Pretending testing docker image [%s]", params.DockerImage)

	if params.Dispose {
		dispose = "--rm"
	}

	fmt.Printf("docker run %s %s%s\n", dispose, optionParams, params.DockerImage)
	state.Completed = true
	return nil
}

func makeDisposableName(params *ContainerParams) string {
	var result = ""

	if params.Name != "" {
		result = params.Name
	} else if params.Prefix != "" {
		if matched, err := regexp.MatchString("-\\d+$", params.Prefix); matched && err == nil {
			var i = strings.LastIndex(params.Prefix, "-")

			result = params.Prefix[:i]
		} else {
			result = params.Prefix
		}
	}

	return result + fmt.Sprintf("-d%d", time.Now().UnixMilli())
}
