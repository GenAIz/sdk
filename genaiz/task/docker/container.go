package docker

import (
	"context"
	"errors"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/mount"

	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/lang/ioz"
	"genaiz.com/genaiz/lang/signalz"
	"genaiz.com/genaiz/lang/stringz"
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

type HasNaming interface {
	GetName(i int) string
}

type HasMounting interface {
	GetEnvKeyValuePair() string

	MakeBindMount() (mount.Mount, error)
}

type ContainerMountBind struct {
	ContainerMountPoint
	HostPath string
}

type ContainerMountPoint struct {
	EnvVar    string
	MountPath string
	ReadOnly  bool
}

func (cm *ContainerMountPoint) GetEnvKeyValuePair() string {
	return cm.EnvVar + "=" + cm.MountPath
}

func (cm *ContainerMountPoint) MakeBindMount(hostDir string) (*mount.Mount, error) {
	if hostDir != "" {
		if _, err := os.Stat(hostDir); err != nil {
			return nil, err
		}

		return &mount.Mount{
			Type:     mount.TypeBind,
			Source:   hostDir,
			Target:   cm.MountPath,
			ReadOnly: cm.ReadOnly,
			BindOptions: &mount.BindOptions{
				CreateMountpoint: true,
			},
		}, nil
	}

	return nil, nil
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

func (c *ContainerParams) GetName(summaries *[]container.Summary) string {
	var i int

	if summaries != nil {
		i = len(*summaries)
	}

	if c.Name != "" {
		return c.Name
	} else if c.Prefix != "" {
		return c.Prefix + "-" + string(rune(i))
	}

	panic("could not provide container name")
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

func fmtMountParams(params *ContainerParams) string {
	var stringParams = []string{
		fmtMountBindParam(&InputMount, params.MountInput),
		fmtMountBindParam(&OutputMount, params.MountOutput),
		fmtMountBindParam(&LogMount, params.MountLog),
		fmtMountBindParam(&VarMount, params.MountVar),
		fmtMountEnvParam(&InputMount, params.MountInput),
		fmtMountEnvParam(&OutputMount, params.MountOutput),
		fmtMountEnvParam(&LogMount, params.MountLog),
		fmtMountEnvParam(&VarMount, params.MountVar),
	}

	return strings.Join(stringz.AllNonEmpty(stringParams...), "   ")
}

func fmtMountBindParam(mount *ContainerMountPoint, path string) string {
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

func fmtMountEnvParam(mount *ContainerMountPoint, path string) string {
	if path != "" {
		return fmt.Sprintf("--env %s=%s \\\n", mount.EnvVar, mount.MountPath)
	}

	return ""
}

func handleContainerAttach(params *ContainerParams, state *task.State) error {
	var summary = selectContainerLatest(state.Containers)

	if summary != nil {
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
				channelReader = ioz.NewHiJackedChannel(wait, ioz.NewHiJackedStreamerStd(response.Reader))
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

	return errors.New("no containers to attach")
}

func handleContainerContext(params *ContainerParams, state *task.State) error {
	if !state.HasContainers() {
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
				state.Containers = &containers
				return nil
			}
		} else {
			return err
		}
	}

	return nil
}

func handleContainerCreate(params *ContainerParams, state *task.State) error {
	var containerName = params.GetName(state.Containers)
	var createConfig = &container.Config{
		Env: []string{
			InputMount.GetEnvKeyValuePair(),
			LogMount.GetEnvKeyValuePair(),
			OutputMount.GetEnvKeyValuePair(),
			VarMount.GetEnvKeyValuePair(),
		},
		Image: params.DockerImage,
		Tty:   params.Interactive,
	}
	var hostConfig = &container.HostConfig{
		AutoRemove: params.Dispose,
	}
	var containerBinds = makeContainerMountBinds(params)

	state.Logger.Debugf("Creating a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

	for _, bind := range containerBinds {
		state.Logger.Debugf("Binding %s to %s", bind.HostPath, bind.MountPath)
	}

	if mounts, err := makeContainerMounts(containerBinds); err == nil && mounts != nil && len(*mounts) > 0 {
		hostConfig.Mounts = *mounts
	} else if err != nil {
		return err
	}

	if resp, err := dockerClient.ContainerCreate(params.Context, createConfig, hostConfig,
		nil, nil, containerName); err == nil {
		for _, w := range resp.Warnings {
			state.Logger.Warningf("%s", w)
		}

		state.Output = resp.ID[:12]
	} else {
		return err
	}

	return nil
}

func handleContainerCreatePretend(params *ContainerParams, state *task.State) error {
	var containerName = params.GetName(state.Containers)
	var envInputMount = InputMount.GetEnvKeyValuePair()
	var envOutputMount = OutputMount.GetEnvKeyValuePair()
	var mountDefs = makeContainerMountBinds(params)
	var mountValues []string

	state.Logger.Debugf("Pretending to create a docker container with name [%s] for image [%s]", containerName, params.DockerImage)

	if mountBinds, err := makeContainerMounts(mountDefs); err == nil && mountBinds != nil {
		for _, bind := range *mountBinds {
			var readonly = ""

			if bind.ReadOnly {
				readonly = "ro=true"
			}

			mountValues = append(mountValues, fmt.Sprintf("--mount src=%s,dst=%s,type=%s%s",
				bind.Source, bind.Target, bind.Type, readonly))
		}
	}

	fmt.Printf("docker create -e %s -e %s %s --name %s %s", envInputMount, envOutputMount,
		strings.Join(mountValues, " "), containerName, params.DockerImage)
	return nil
}

func handleContainerDisposal(params *ContainerParams, state *task.State) error {
	var count = state.GetContainersSize()
	var result error

	if count > 0 {
		// Removed volumes and links as we'll just re-create them if needed on Create
		var removeOptions = container.RemoveOptions{
			RemoveVolumes: true,
			RemoveLinks:   true,
			Force:         params.Force,
		}

		if count > 1 {
			state.Logger.Debugf("%d containers selected for disposal", count)
		}

		for _, ct := range *state.Containers {
			state.Logger.Debugf("Removing container [%s]", ct.ID)

			if err := dockerClient.ContainerRemove(params.Context, ct.ID, removeOptions); err != nil {
				state.Logger.Debugf("Could not remove container [%s] with error [%s]", ct.ID, err)
				result = err
			}
		}

		state.Containers = nil
	} else {
		state.Logger.Debugf("No containers selected for disposal")
	}

	return result
}

func handleContainerDisposalPretend(params *ContainerParams, state *task.State) error {
	var count = state.GetContainersSize()

	if count > 0 {
		var forceString = ""

		state.Logger.Debugf("Pretending removing docker container(s)")

		if params.Force {
			forceString = "-f"
		}

		for _, ct := range *state.Containers {
			fmt.Printf("docker rm -l -v %s %s", forceString, ct.ID)
		}
	}

	return nil
}

func handleContainerStart(params *ContainerParams, state *task.State) error {
	var summary = selectContainerLatest(state.Containers)

	if summary != nil {
		state.Logger.Debugf("Starting docker container [%s]", summary.ID)

		if err := dockerClient.ContainerStart(params.Context, summary.ID, container.StartOptions{}); err != nil {
			return err
		}

		state.Output = summary.ID[:12]
		return nil
	}

	return errors.New("no containers selected")
}

func handleContainerStartPretend(params *ContainerParams, state *task.State) error {
	if state.HasContainers() {
		var summary = selectContainerLatest(state.Containers)
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

	return nil
}

func handleContainerStop(params *ContainerParams, state *task.State) error {
	var count = state.GetContainersSize()
	var result error

	if count > 0 {
		var stopOptions = container.StopOptions{}

		if count > 1 {
			state.Logger.Debugf("[%d] containers selected for stop", count)
		}

		if !params.Force {
			*stopOptions.Timeout = -1
		}

		for _, ct := range *state.Containers {
			state.Logger.Debugf("Stopping container [%s]", ct.ID)

			if err := dockerClient.ContainerStop(params.Context, ct.ID, stopOptions); err != nil {
				state.Logger.Debugf("Could not stop container [%s] with error [%s]", ct.ID, err)
				result = err
			}
		}
	}

	return result
}

func handleContainerStopPretend(params *ContainerParams, state *task.State) error {
	if state.HasContainers() {
		for _, summary := range *state.Containers {
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

func makeContainerMounts(definitions []*ContainerMountBind) (*[]mount.Mount, error) {
	var result []mount.Mount
	var err error

	for _, def := range definitions {
		if err = filez.DoIfPathExist(def.HostPath, func() error {
			if bind, errB := def.MakeBindMount(def.HostPath); bind != nil && errB == nil {
				result = append(result, *bind)
			} else if errB != nil {
				return errB
			}

			return nil
		}); err != nil {
			break
		}
	}

	return &result, err
}

func makeContainerMountBinds(params *ContainerParams) []*ContainerMountBind {
	var result []*ContainerMountBind

	if params.MountInput != "" {
		result = append(result, &ContainerMountBind{
			ContainerMountPoint: InputMount,
			HostPath:            params.MountInput,
		})
	}

	if params.MountLog != "" {
		result = append(result, &ContainerMountBind{
			ContainerMountPoint: LogMount,
			HostPath:            params.MountLog,
		})
	}

	if params.MountOutput != "" {
		result = append(result, &ContainerMountBind{
			ContainerMountPoint: OutputMount,
			HostPath:            params.MountOutput,
		})
	}

	if params.MountVar != "" {
		result = append(result, &ContainerMountBind{
			ContainerMountPoint: VarMount,
			HostPath:            params.MountVar,
		})
	}

	return result
}

func selectContainerLatest(summaries *[]container.Summary) *container.Summary {
	if summaries != nil && len(*summaries) > 0 {
		sort.SliceStable(*summaries, func(i, j int) bool {
			return (*summaries)[i].Created > (*summaries)[j].Created
		})
		return &(*summaries)[0]
	}

	return nil
}
