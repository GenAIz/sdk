package docker

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"os"
	"slices"
	"strings"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/mount"
	"github.com/moby/moby/client"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/signalz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/lang/streamz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker/adapt"
	"genaiz.com/genaiz/task/shared"
)

var (
	ContainerTasks struct {
		CreateTaskFactory  CreateTaskFactory  // CreateTaskFactory builds a task which can be used with an instance of ContainerParams to instruct the creation of a container with dockerd.
		DisposeTaskFactory DisposeTaskFactory // DisposeTaskFactory builds a task which can be used to automatically remove a container after it has been stopped.
		StartTaskFactory   StartTaskFactory   // StartTaskFactory builds a task which can be used with an instance of ContainerParams to start a previously created container.
		StopTaskFactory    StopTaskFactory    // StopTaskFactory builds a task which can be used with an instance of ContainerParams to stop a previously started container.
	}

	signalProvider = signalz.NewSignalChannel
)

type CreateTask = *task.Task[ContainerParams]
type CreateTaskFactory = func() CreateTask
type DisposeTask = *task.Task[ContainerParams]
type DisposeTaskFactory = func() DisposeTask
type StartTask = *task.Task[ContainerParams]
type StartTaskFactory = func() StartTask
type StopTask = *task.Task[ContainerParams]
type StopTaskFactory = func() StopTask

type ContainerParams struct {
	task.Env
	adapt.RunArguments
	adapt.ContainerArguments
	VarSpecs []shared.VarSpec
}

func handleContainerAttach(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if summary := dockerState.SelectLatestContainer(); summary != nil {
		var dockerClient = dockerFactory.Get()
		var wait, cancel = context.WithCancel(params.Context)
		var inspect client.ContainerInspectResult
		var options = client.ContainerInspectOptions{
			Size: true,
		}
		var err error
		defer cancel()

		state.Logger.Debugf("Selected container id [%s]", summary.ID)

		if inspect, err = dockerClient.ContainerInspect(wait, summary.ID, options); err == nil {
			var attachOptions = client.ContainerAttachOptions{Stream: true, Stdout: true, Stderr: true}
			var channelReader chan error
			var channelContainer chan int
			var response client.ContainerAttachResult

			if !inspect.Container.Config.Tty {
				var noCancels = context.WithoutCancel(wait)
				var channelShell = signalProvider()

				go signalz.ForwardTerminate(noCancels, channelShell, func(code string) {
					var killOptions = client.ContainerKillOptions{
						Signal: code,
					}

					if _, err = dockerClient.ContainerKill(noCancels, summary.ID, killOptions); err != nil {
						state.Logger.Errorf("Could not kill container: %s", err)
						state.Error = err
					}
				})
				defer signalz.StopCatch(channelShell)
			}

			if response, err = dockerClient.ContainerAttach(wait, summary.ID, attachOptions); err == nil {
				channelReader = streamz.NewHiJackedChannel(wait, streamz.NewHiJackedStreamerStd(response.Reader))
				channelContainer = makeContainerChannel(wait, params, summary.ID)
				defer response.Close()
			} else {
				return err
			}

			if err = handleContainerStart(params, state); err != nil {
				cancel()
				<-channelReader

				if params.Dispose {
					<-channelContainer
				}

				return err
			}

			if err = <-channelReader; err != nil {
				return err
			}

			if status := <-channelContainer; status != 0 {
				return fmt.Errorf("received exit status error %d from the container", status)
			}

			return nil
		}

		return err
	}

	return errors.New("no containers to attach")
}

func handleContainerContext(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if !dockerState.HasContainers() {
		var dockerClient = dockerFactory.Get()
		var listFilters = make(client.Filters)
		var listOptions client.ContainerListOptions

		if params.Name != "" {
			state.Logger.Debugf("Finding a docker container for name [%s]", params.Name)
			listFilters.Add("name", params.Name)
		} else if params.Prefix != "" {
			state.Logger.Debugf("Finding all docker containers for prefix [%s]", params.Prefix)
			listFilters.Add("name", params.Prefix)
		} else {
			return Errors.ContainerNameRequired()
		}

		listOptions = client.ContainerListOptions{
			All:     true,
			Filters: listFilters,
		}

		if results, err := dockerClient.ContainerList(params.Context, listOptions); err == nil {
			if len(results.Items) == 0 {
				return Errors.ContainerNotFound()
			}

			dockerState.AddContainers(results.Items...)
		} else {
			return err
		}
	}

	return nil
}

func handleContainerCreate(params *ContainerParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var dockerState = NewClientState(state)
	var containerName = params.GetName(dockerState.GetContainers())
	var err error

	state.Logger.Debugf("Creating a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

	if err = dockerState.MergeSpecs(params.VarSpecs); err == nil {
		var envValues = makeEnvironmentValues(params.GetEnvironment(), dockerState.VarSpecs)

		if err = validateEnvironmentValues(dockerState.VarSpecs, envValues); err == nil {
			var dockerImage = stringz.FirstNonEmpty(state.Output, params.DockerImage)
			var containerBinds = params.GetMountBinds()
			var createConfig = &container.Config{
				Env:   envValues,
				Image: dockerImage,
				Tty:   params.Interactive,
			}
			var hostConfig = &container.HostConfig{AutoRemove: params.Dispose}
			var createOptions = client.ContainerCreateOptions{
				Config:     createConfig,
				HostConfig: hostConfig,
				Name:       containerName,
			}
			var result client.ContainerCreateResult
			var mounts []mount.Mount

			if uid, gid := os.Getuid(), os.Getgid(); uid >= 0 && gid >= 0 {
				state.Logger.Debugf("Runtime user selected: [%d:%d]", uid, gid)
				createConfig.User = fmt.Sprintf("%d:%d", uid, gid)
			} else {
				state.Logger.Warningf("Could not determine a runtime user [%d:%d]", uid, gid)
			}

			if params.Dispose {
				state.Logger.Debugf("Auto-Remove of container after start is enabled")
			}

			if mounts = makeContainerMounts(containerBinds); len(mounts) > 0 {
				hostConfig.Mounts = mounts

				for _, bind := range hostConfig.Mounts {
					state.Logger.Debugf("Binding %s to %s", bind.Source, bind.Target)
				}
			}

			if result, err = dockerClient.ContainerCreate(params.Context, createOptions); err == nil {
				for _, w := range result.Warnings {
					state.Logger.Warningf("%s", w)
				}

				state.Output = result.ID[0:12]
				state.Logger.Debugf("Created docker container Id [%s]", state.Output)
				return nil
			}
		}
	}

	return err
}

func handleContainerCreatePretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var containerName = params.GetName(dockerState.GetContainers())
	var envValues = makeEnvironmentValues(params.GetEnvironment(), params.VarSpecs)
	var err error

	state.Logger.Debugf("Pretending to create a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

	if err = validateEnvironmentValues(params.VarSpecs, envValues); err == nil {
		var mountDefs = params.GetMountBinds()
		var mountValues []string
		var mounts []mount.Mount

		if mounts = makeContainerMounts(mountDefs); len(mounts) > 0 {
			for _, bind := range mounts {
				var readonly = ""

				if bind.ReadOnly {
					readonly = "ro=true"
				}

				mountValues = append(mountValues, fmt.Sprintf("--mount src=%s,dst=%s,type=%s%s",
					bind.Source, bind.Target, bind.Type, readonly))
			}
		}

		for i, envVar := range envValues {
			envValues[i] = "--env " + envVar
		}

		fmt.Printf("docker create %s %s --name %s %s", strings.Join(envValues, " "),
			strings.Join(mountValues, " "), containerName, params.DockerImage)
	}

	return err
}

func handleContainerDisposal(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var count = dockerState.GetContainersSize()
	var err error

	if count > 0 {
		var dockerClient = dockerFactory.Get()
		// Removed volumes and links as we'll just re-create them if needed on Create
		var removeOptions = client.ContainerRemoveOptions{
			RemoveVolumes: true,
			Force:         params.Force,
		}

		for _, ct := range dockerState.containers {
			var name = dockerState.DisplayString(&ct)

			state.Logger.Debugf("Removing container [%s]", name)

			if _, err = dockerClient.ContainerRemove(params.Context, ct.ID, removeOptions); err == nil {
				state.Reportf("Disposed of container %s with status %s", name, ct.Status)
			} else {
				state.Logger.Debugf("Could not remove container [%s] with error [%s]", name, err)
				break
			}
		}

		dockerState.Reset()
	} else {
		state.Logger.Debugf("No containers selected for disposal")
	}

	return err
}

func handleContainerDisposalPretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var count = dockerState.GetContainersSize()

	if count > 0 {
		var forceString = ""

		state.Logger.Debugf("Pretending removing docker container(s)")

		if params.Force {
			forceString = "-f"
		}

		for _, ct := range dockerState.GetContainers() {
			fmt.Printf("docker rm -l -v %s %s", forceString, ct.ID)
		}
	}

	return nil
}

func handleContainerStart(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if summary := dockerState.SelectLatestContainer(); summary != nil {
		var dockerClient = dockerFactory.Get()
		var name = dockerState.DisplayString(summary)
		var startOptions = client.ContainerStartOptions{}

		state.Logger.Debugf("Starting docker container [%s]", name)

		if _, err := dockerClient.ContainerStart(params.Context, summary.ID, startOptions); err != nil {
			return err
		}

		state.Reportf("Started container %s for image %s", name, params.DockerImage)
		state.Output = summary.ID
		return nil
	}

	return errors.New("no containers selected")
}

func handleContainerStartPretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if summary := dockerState.SelectLatestContainer(); summary != nil {
		var attached, interactive string

		state.Logger.Debugf("Pretending starting docker container [%s]", summary.ID)

		if params.Attached {
			attached = "--attach "
		}

		if params.Interactive {
			interactive = "--interactive "
		}

		fmt.Printf("docker start %s%s%s\n", attached, interactive, summary.ID)
		state.Completed = true
		return nil
	}

	return errors.New("could not select a container to start")
}

func handleContainerStop(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var count = dockerState.GetContainersSize()
	var err error

	if count > 0 {
		var dockerClient = dockerFactory.Get()
		var stopOptions = client.ContainerStopOptions{}

		if count > 1 {
			state.Logger.Debugf("[%d] containers selected for stop", count)
		}

		if !params.Force {
			stopOptions.Timeout = new(-1)
		}

		for _, ct := range dockerState.GetContainers() {
			var name = dockerState.DisplayString(&ct)

			state.Logger.Debugf("Stopping container [%s]", name)

			if _, err = dockerClient.ContainerStop(params.Context, ct.ID, stopOptions); err == nil {
				state.Reportf("Stopped container %s with status %s", name, ct.Status)
			} else {
				state.Logger.Debugf("Could not stop container [%s] with error [%s]", ct.ID, err)
				break
			}
		}
	}

	return err
}

func handleContainerStopPretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if dockerState.HasContainers() {
		var timeoutStr = ""

		if !params.Force {
			timeoutStr = "-t -1 "
		}

		for _, summary := range dockerState.GetContainers() {
			fmt.Printf("docker stop %s%s\n", timeoutStr, summary.ID)
		}
	}

	return nil
}

func init() {
	ContainerTasks.CreateTaskFactory = newCreateTask
	ContainerTasks.DisposeTaskFactory = newDisposeTask
	ContainerTasks.StartTaskFactory = newStartTask
	ContainerTasks.StopTaskFactory = newStopTask
}

func makeContainerChannel(ctx context.Context, params *ContainerParams, containerId string) chan int {
	var dockerClient = dockerFactory.Get()
	var waitOptions = client.ContainerWaitOptions{
		Condition: params.WaitCondition(),
	}
	var waitResult = dockerClient.ContainerWait(ctx, containerId, waitOptions)
	var channel, channelErr = waitResult.Result, waitResult.Error
	var channelStatus = make(chan int)

	go func() {
		defer close(channelStatus)
		select {
		case <-ctx.Done():
			return
		case result := <-channel:
			if result.Error != nil {
				channelStatus <- 125
			} else {
				channelStatus <- int(result.StatusCode)
			}
		case err := <-channelErr:
			if errors.Is(err, context.Canceled) {
				return
			}
			channelStatus <- 125
		default:
		}
	}()

	return channelStatus
}

func makeContainerMounts(definitions []adapt.ContainerMountBind) []mount.Mount {
	var result []mount.Mount

	for _, def := range definitions {
		_ = dirz.DoIfPathExist(def.HostPath, func() error {
			result = append(result, def.MakeMount(def.HostPath))
			return nil
		})
	}

	return result
}

func makeEnvironmentSpecs(propSpecs []shared.VarSpec) []string {
	var result []string

	result = append(result, adapt.EnvKeys...)

	for _, spec := range propSpecs {
		result = append(result, spec.GetKey())
	}

	return result
}

func makeEnvironmentValues(envMap map[string]string, specs []shared.VarSpec) []string {
	var allowedSpecs = makeEnvironmentSpecs(specs)
	var merged = make(map[string]string)
	var valueFunc = func(key string) string {
		var value string
		var ok bool

		if value, ok = merged[key]; !ok {
			// Forward the variable to the container, if it's not resolved
			return "$" + key
		}

		return value
	}
	var mountPoints = []adapt.ContainerMountPoint{
		adapt.Mounts.Input,
		adapt.Mounts.Output,
		adapt.Mounts.Log,
		adapt.Mounts.Var,
	}
	var result []string

	maps.Copy(merged, adapt.EnvFileMap)
	maps.Copy(merged, envMap)

	for _, mountPoint := range mountPoints {
		if _, ok := merged[mountPoint.EnvVar]; !ok {
			k, v := mountPoint.GetEnvKeyValue()
			merged[k] = v
		}
	}

	for k, v := range merged {
		if slices.Contains(allowedSpecs, k) {
			result = append(result, fmt.Sprintf("%s=%s", k, os.Expand(v, valueFunc)))
		}
	}

	for _, spec := range specs {
		if !slices.ContainsFunc(result, func(s string) bool {
			return strings.HasPrefix(s, spec.GetKey())
		}) {
			if spec.GetDefaultValue() != "" {
				result = append(result, fmt.Sprintf("%s=%s", spec.GetKey(), spec.GetDefaultValue()))
			}
		}
	}

	return result
}

func newCreateTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:         "docker-create",
		OnPrepare:    handleContainerContext,
		OnIncomplete: handleContainerCreate,
		OnPretend:    handleContainerCreatePretend,
	}
}

func newDisposeTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-rm",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerDisposal,
		OnPretend:  handleContainerDisposalPretend,
	}
}

func newStartTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-start",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerStart,
		OnPretend:  handleContainerStartPretend,
	}
}

func newStopTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-stop",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerStop,
		OnPretend:  handleContainerStopPretend,
	}
}

func validateEnvironmentValues(varSpecs []shared.VarSpec, envValues []string) error {
	for _, pair := range envValues {
		var keyValue = strings.SplitN(pair, "=", 2)

		if len(keyValue) == 2 {
			if i := slices.IndexFunc(varSpecs, func(spec shared.VarSpec) bool {
				return spec.GetKey() == keyValue[0]
			}); i >= 0 {
				if err := varSpecs[i].Validate(keyValue[1]); err != nil {
					return fmt.Errorf("env key [%s] is not valid: %s", keyValue[0], err)
				}
			}
		}
	}

	for _, spec := range varSpecs {
		if !slices.ContainsFunc(envValues, func(s string) bool {
			return strings.HasPrefix(s, spec.GetKey())
		}) {
			return fmt.Errorf("property [%s] is not specified and is mandatory", spec.GetKey())
		}
	}

	return nil
}
