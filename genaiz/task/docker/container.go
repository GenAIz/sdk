package docker

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/lang/signalz"
	"genaiz.com/genaiz/lang/streamz"
	"genaiz.com/genaiz/task"
)

var (
	InputMount = ContainerMountPoint{
		EnvVar:    "SF_INPUT_PATH",
		MountPath: "/mnt/in",
		ReadOnly:  true,
	}

	OutputMount = ContainerMountPoint{
		EnvVar:    "SF_OUTPUT_PATH",
		MountPath: "/mnt/out",
		ReadOnly:  false,
	}

	LogMount = ContainerMountPoint{
		EnvVar:    "SF_LOG_PATH",
		MountPath: "/mnt/log",
		ReadOnly:  false,
	}

	VarMount = ContainerMountPoint{
		EnvVar:    "SF_VAR_PATH",
		MountPath: "/mnt/var",
		ReadOnly:  false,
	}
)

type ContainerMountBind struct {
	ContainerMountPoint
	HostPath string
}

type ContainerMountPoint struct {
	EnvVar    string
	MountPath string
	ReadOnly  bool
}

func (cm ContainerMountPoint) GetEnvKeyValuePair() string {
	return cm.EnvVar + "=" + cm.MountPath
}

func (cm ContainerMountPoint) MakeBind(hostPath string) ContainerMountBind {
	return ContainerMountBind{
		ContainerMountPoint: cm,
		HostPath:            hostPath,
	}
}

func (cm ContainerMountPoint) MakeMount(hostDir string) *mount.Mount {
	return &mount.Mount{
		Type:     mount.TypeBind,
		Source:   hostDir,
		Target:   cm.MountPath,
		ReadOnly: cm.ReadOnly,
		BindOptions: &mount.BindOptions{
			CreateMountpoint: true,
		},
	}
}

type ContainerParams struct {
	RunParams
	DockerImage string
	MountInput  string
	MountOutput string
	MountLog    string
	MountVar    string
	Name        string
	Prefix      string
	Force       bool
}

func (c *ContainerParams) GetName(summaries []container.Summary) (string, error) {
	if c.Name != "" {
		return c.Name, nil
	} else if c.Prefix != "" {
		return fmt.Sprintf("%s-%d", c.Prefix, len(summaries)), nil
	}

	return "", errors.New("could not provide container name")
}

func (c *ContainerParams) GetContainerMountBinds() []ContainerMountBind {
	var result []ContainerMountBind

	if c.MountInput != "" {
		result = append(result, InputMount.MakeBind(c.MountInput))
	}

	if c.MountLog != "" {
		result = append(result, LogMount.MakeBind(c.MountLog))
	}

	if c.MountOutput != "" {
		result = append(result, OutputMount.MakeBind(c.MountOutput))
	}

	if c.MountVar != "" {
		result = append(result, VarMount.MakeBind(c.MountVar))
	}

	return result
}

func (c *ContainerParams) MakeDisposableName() string {
	var result = ""

	if c.Name != "" {
		result = c.Name
	} else if c.Prefix != "" {
		if matched, _ := regexp.MatchString("-\\d+$", c.Prefix); matched {
			var i = strings.LastIndex(c.Prefix, "-")

			result = c.Prefix[:i]
		} else {
			result = c.Prefix
		}
	} else {
		result = "container"
	}

	return result + fmt.Sprintf("-d%d", time.Now().UnixMilli())
}

func (c *ContainerParams) fmtArgs() string {
	var stringParams = []string{
		c.fmtMountBindArg(&InputMount, c.MountInput),
		c.fmtMountBindArg(&OutputMount, c.MountOutput),
		c.fmtMountBindArg(&LogMount, c.MountLog),
		c.fmtMountBindArg(&VarMount, c.MountVar),
		c.fmtMountEnvArg(&InputMount, c.MountInput),
		c.fmtMountEnvArg(&OutputMount, c.MountOutput),
		c.fmtMountEnvArg(&LogMount, c.MountLog),
		c.fmtMountEnvArg(&VarMount, c.MountVar),
	}

	return strings.Join(stringz.AllNonEmpty(stringParams...), "   ")
}

func (c *ContainerParams) fmtMountBindArg(mount *ContainerMountPoint, path string) string {
	if path != "" {
		var local = filez.FromWorkDir(path)
		var readOnly = ""

		if mount.ReadOnly {
			readOnly = ",ro"
		}

		return fmt.Sprintf("--mount type=bind,src=%s,dst=%s%s \\\n", local, mount.MountPath, readOnly)
	}

	return ""
}

func (c *ContainerParams) fmtMountEnvArg(mount *ContainerMountPoint, path string) string {
	if path != "" {
		return fmt.Sprintf("--env %s=%s \\\n", mount.EnvVar, mount.MountPath)
	}

	return ""
}

func NewCreateTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:         "docker-create",
		OnPrepare:    handleContainerContext,
		OnIncomplete: handleContainerCreate,
		OnPretend:    handleContainerCreatePretend,
	}
}

func NewDisposeTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-rm",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerDisposal,
		OnPretend:  handleContainerDisposalPretend,
	}
}

func NewStartTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-start",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerStart,
		OnPretend:  handleContainerStartPretend,
	}
}

func NewStopTask() *task.Task[ContainerParams] {
	return &task.Task[ContainerParams]{
		Name:       "docker-stop",
		OnPrepare:  handleContainerContext,
		OnComplete: handleContainerStop,
		OnPretend:  handleContainerStopPretend,
	}
}

func handleContainerAttach(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if count := dockerState.GetContainersSize(); count > 0 {
		state.Logger.Debugf("Found [%d] summaries for attach", count)

		if summary := dockerState.SelectLatestContainer(); summary != nil {
			var dockerClient = dockerFactory.Get()
			var wait, cancel = context.WithCancel(params.Context)
			defer cancel()

			if inspect, err := dockerClient.ContainerInspect(wait, summary.ID); err == nil {
				var attachOptions = container.AttachOptions{Stream: true, Stdout: true, Stderr: true}
				var channelReader chan error
				var channelContainer chan int
				var response types.HijackedResponse

				if !inspect.Config.Tty {
					var noCancels = context.WithoutCancel(wait)
					var channelShell = signalz.NewSignalChannel()

					go signalz.ForwardTerminate(noCancels, channelShell, func(code string) {
						if err = dockerClient.ContainerKill(noCancels, summary.ID, code); err != nil {
							panic(err)
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
					var escapeError signalz.EscapeError

					if errors.As(err, &escapeError) {
						return nil
					}

					return err
				}

				if status := <-channelContainer; status != 0 {
					return fmt.Errorf("received exit status error %d from the container", status)
				}

				return nil
			} else {
				return err
			}
		}
	}

	return errors.New("no containers to attach")
}

func handleContainerContext(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if !dockerState.HasContainers() {
		var dockerClient = dockerFactory.Get()
		var listFilters = filters.NewArgs()
		var listOptions = container.ListOptions{
			All:     true,
			Filters: listFilters,
		}

		if params.Name != "" {
			state.Logger.Debugf("Finding a docker container for name [%s]", params.Name)
			listFilters.Add("name", params.Name)
		} else if params.Prefix != "" {
			state.Logger.Debugf("Finding all docker containers for prefix [%s]", params.Prefix)
			listFilters.Add("name", params.Prefix)
		} else {
			panic("name or prefix required for container commands")
		}

		if containers, err := dockerClient.ContainerList(params.Context, listOptions); err == nil {
			if len(containers) == 0 {
				return errors.New("container not found")
			} else {
				dockerState.AddContainers(containers...)
			}
		} else {
			return err
		}
	}

	return nil
}

func handleContainerCreate(params *ContainerParams, state *task.State) error {
	var dockerClient = dockerFactory.Get()
	var dockerState = NewClientState(state)
	var dockerImage = stringz.FirstNonEmpty(state.Output, params.DockerImage)
	var hostConfig = &container.HostConfig{AutoRemove: params.Dispose}
	var containerBinds = params.GetContainerMountBinds()
	var containerName string
	var err error

	if containerName, err = params.GetName(dockerState.GetContainers()); err == nil {
		var createConfig = &container.Config{
			Env: []string{
				InputMount.GetEnvKeyValuePair(),
				LogMount.GetEnvKeyValuePair(),
				OutputMount.GetEnvKeyValuePair(),
				VarMount.GetEnvKeyValuePair(),
			},
			Image: dockerImage,
			Tty:   params.Interactive,
		}
		var resp container.CreateResponse
		var mounts []mount.Mount

		state.Logger.Debugf("Creating a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

		if params.Dispose {
			state.Logger.Debugf("Auto-Remove of container after start is enabled")
		}

		if mounts = makeContainerMounts(containerBinds); len(mounts) > 0 {
			hostConfig.Mounts = mounts

			for _, bind := range hostConfig.Mounts {
				state.Logger.Debugf("Binding %s to %s", bind.Source, bind.Target)
			}
		}

		if resp, err = dockerClient.ContainerCreate(params.Context, createConfig, hostConfig,
			nil, nil, containerName); err == nil {
			for _, w := range resp.Warnings {
				state.Logger.Warningf("%s", w)
			}

			state.Output = resp.ID[0:12]
			state.Logger.Debugf("Created docker container Id [%s]", state.Output)
			return nil
		}
	}

	return err
}

func handleContainerCreatePretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var envInputMount = InputMount.GetEnvKeyValuePair()
	var envOutputMount = OutputMount.GetEnvKeyValuePair()
	var mountDefs = params.GetContainerMountBinds()
	var mountValues []string
	var containerName string
	var err error

	state.Logger.Debugf("Pretending to create a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

	if containerName, err = params.GetName(dockerState.GetContainers()); err == nil {
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
	}

	if err == nil {
		fmt.Printf("docker create -e %s -e %s %s --name %s %s", envInputMount, envOutputMount,
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
		var removeOptions = container.RemoveOptions{
			RemoveVolumes: true,
			Force:         params.Force,
		}

		for _, ct := range dockerState.containers {
			var name = dockerState.DisplayString(&ct)

			state.Logger.Debugf("Removing container [%s]", name)

			if err = dockerClient.ContainerRemove(params.Context, ct.ID, removeOptions); err == nil {
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

	if count := dockerState.GetContainersSize(); count > 0 {
		state.Logger.Debugf("Found [%d] summaries for start", count)

		if summary := dockerState.SelectLatestContainer(); summary != nil {
			var dockerClient = dockerFactory.Get()
			var name = dockerState.DisplayString(summary)

			state.Logger.Debugf("Starting docker container [%s]", name)

			if err := dockerClient.ContainerStart(params.Context, summary.ID, container.StartOptions{}); err != nil {
				return err
			}

			state.Reportf("Started container %s for image %s", name, params.DockerImage)
			state.Output = summary.ID
			return nil
		}
	}

	return errors.New("no containers selected")
}

func handleContainerStartPretend(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)

	if dockerState.HasContainers() {
		var summary = dockerState.SelectLatestContainer()

		if summary != nil {
			var attached, interactive string

			state.Logger.Debugf("Pretending starting docker container [%s]", summary.ID)

			if params.RunParams.Attached {
				attached = "--attach "
			}

			if params.RunParams.Interactive {
				interactive = "--interactive "
			}

			fmt.Printf("docker start %s%s%s\n", attached, interactive, summary.ID)
			state.Completed = true
		}

		return errors.New("could not select a container to start")
	}

	return nil
}

func handleContainerStop(params *ContainerParams, state *task.State) error {
	var dockerState = NewClientState(state)
	var count = dockerState.GetContainersSize()
	var err error

	if count > 0 {
		var dockerClient = dockerFactory.Get()
		var stopOptions = container.StopOptions{}

		if count > 1 {
			state.Logger.Debugf("[%d] containers selected for stop", count)
		}

		if !params.Force {
			stopOptions.Timeout = lang.Ref(-1)
		}

		for _, ct := range dockerState.GetContainers() {
			var name = dockerState.DisplayString(&ct)

			state.Logger.Debugf("Stopping container [%s]", name)

			if err = dockerClient.ContainerStop(params.Context, ct.ID, stopOptions); err == nil {
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
		for _, summary := range dockerState.GetContainers() {
			var timeoutStr = ""

			if !params.Force {
				timeoutStr = "-t -1"
			}

			fmt.Printf("docker stop %s %s\n", timeoutStr, summary.ID)
		}
	}

	return nil
}

func makeContainerChannel(ctx context.Context, params *ContainerParams, containerId string) chan int {
	var dockerClient = dockerFactory.Get()
	var channel, channelErr = dockerClient.ContainerWait(ctx, containerId, params.WaitCondition())
	var channelStatus = make(chan int)

	go func() {
		defer close(channelStatus)
		select {
		case <-params.Context.Done():
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
		}
	}()

	return channelStatus
}

func makeContainerMounts(definitions []ContainerMountBind) []mount.Mount {
	var result []mount.Mount

	for _, def := range definitions {
		_ = dirz.DoIfPathExist(def.HostPath, func() error {
			result = append(result, *def.MakeMount(def.HostPath))
			return nil
		})
	}

	return result
}
