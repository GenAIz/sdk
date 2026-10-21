package docker

import (
	"context"
	"io"
	"os/exec"
	"sort"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"

	"genaiz.com/genaiz-lib/lang/ioz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	dockerFactory = &ClientFactory{
		get: getDockerClient,
	}

	forkFactory = &ForkFactory{
		get: getFork,
	}
)

// Client is a wrapper interface for the concrete object client.NewClientWithOpts under docker/api. This is useful for mocking calls to the client avoiding the need for a Docker installation when running tests
type Client interface {
	ContainerAttach(context.Context, string, client.ContainerAttachOptions) (client.ContainerAttachResult, error)

	ContainerCreate(context.Context, client.ContainerCreateOptions) (client.ContainerCreateResult, error)

	ContainerInspect(context.Context, string, client.ContainerInspectOptions) (client.ContainerInspectResult, error)

	ContainerKill(context.Context, string, client.ContainerKillOptions) (client.ContainerKillResult, error)

	ContainerList(context.Context, client.ContainerListOptions) (client.ContainerListResult, error)

	ContainerRemove(context.Context, string, client.ContainerRemoveOptions) (client.ContainerRemoveResult, error)

	ContainerStart(context.Context, string, client.ContainerStartOptions) (client.ContainerStartResult, error)

	ContainerStop(context.Context, string, client.ContainerStopOptions) (client.ContainerStopResult, error)

	ContainerWait(context.Context, string, client.ContainerWaitOptions) client.ContainerWaitResult

	ImageBuild(context.Context, io.Reader, client.ImageBuildOptions) (client.ImageBuildResult, error)

	ImageInspect(context.Context, string, ...client.ImageInspectOption) (client.ImageInspectResult, error)

	ImageList(context.Context, client.ImageListOptions) (client.ImageListResult, error)

	ImagePrune(context.Context, client.ImagePruneOptions) (client.ImagePruneResult, error)

	ImagePush(context.Context, string, client.ImagePushOptions) (client.ImagePushResponse, error)

	ImageTag(context.Context, client.ImageTagOptions) (client.ImageTagResult, error)

	RegistryLogin(context.Context, client.RegistryLoginOptions) (client.RegistryLoginResult, error)
}

type ClientFactory struct {
	dockerClient *client.Client

	get func(cf *ClientFactory) (Client, error)
}

func (cf *ClientFactory) Get() Client {
	var result, err = cf.get(cf)

	panicz.PanicIfError(err)
	return result
}

type clientTracking struct {
	shared.VarSpecTracking
	containers []container.Summary
	images     []image.Summary
}

type ClientState struct {
	clientTracking
	state *task.State
}

func (cs *ClientState) AddContainers(containers ...container.Summary) {
	cs.containers = append(cs.containers, containers...)
	cs.state.Internal = cs.clientTracking
}

func (cs *ClientState) AddImages(images ...image.Summary) {
	cs.images = append(cs.images, images...)
	cs.state.Internal = cs.clientTracking
}

func (cs *ClientState) DisplayString(summary *container.Summary) string {
	if len(summary.Names) > 0 {
		return summary.Names[0][1:]
	}

	return summary.ID[0:12]
}

func (cs *ClientState) GetContainers() []container.Summary {
	return cs.containers
}

func (cs *ClientState) GetContainersSize() int {
	return len(cs.containers)
}

func (cs *ClientState) GetImages() []image.Summary {
	return cs.images
}

func (cs *ClientState) GetImagesSize() int {
	return len(cs.images)
}

func (cs *ClientState) HasContainers() bool {
	return cs.GetContainersSize() > 0
}

func (cs *ClientState) HasImages() bool {
	return cs.GetImagesSize() > 0
}

func (cs *ClientState) Reset() {
	cs.state.Internal = nil
}

func (cs *ClientState) SelectLatestContainer() *container.Summary {
	if len(cs.containers) > 0 {
		var summaries = cs.containers

		sort.SliceStable(summaries, func(i, j int) bool {
			return summaries[i].Created > summaries[j].Created
		})
		return &summaries[0]
	}

	return nil
}

func NewClientState(state *task.State) *ClientState {
	var containers []container.Summary
	var images []image.Summary
	var varSpecs []shared.VarSpec

	if ct, ok := state.Internal.(clientTracking); ok {
		containers = ct.containers
		images = ct.images
	} else if st, ok := state.Internal.(shared.VarSpecTracking); ok {
		varSpecs = st.VarSpecs
	}

	return &ClientState{
		clientTracking: clientTracking{
			VarSpecTracking: shared.VarSpecTracking{
				VarSpecs: varSpecs,
			},
			containers: containers,
			images:     images,
		},
		state: state,
	}
}

func getDockerClient(cf *ClientFactory) (Client, error) {
	var err error

	if cf.dockerClient == nil {
		cf.dockerClient, err = client.New(client.FromEnv, client.WithAPIVersionFromEnv())
	}

	return cf.dockerClient, err
}

type ForkFactory struct {
	fork ioz.Fork

	get func(*ForkFactory, *exec.Cmd) ioz.Fork
}

func (ff *ForkFactory) Get(cmd *exec.Cmd) ioz.Fork {
	return ff.get(ff, cmd)
}

func getFork(ff *ForkFactory, cmd *exec.Cmd) ioz.Fork {
	if ff.fork == nil {
		ff.fork = ioz.NewFork(cmd)
	}

	return ff.fork
}
