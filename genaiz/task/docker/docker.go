package docker

import (
	"context"
	"io"
	"sort"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	dockerFactory = &ClientFactory{
		get: getDockerClient,
	}
)

// Client is a wrapper interface for the concrete object client.NewClientWithOpts under docker/api. This is useful for mocking calls to the client avoiding the need for a Docker installation when running tests
type Client interface {
	ContainerAttach(context.Context, string, container.AttachOptions) (types.HijackedResponse, error)

	ContainerCreate(context.Context, *container.Config, *container.HostConfig, *network.NetworkingConfig, *ocispec.Platform, string) (container.CreateResponse, error)

	ContainerInspect(context.Context, string) (container.InspectResponse, error)

	ContainerKill(context.Context, string, string) error

	ContainerList(context.Context, container.ListOptions) ([]container.Summary, error)

	ContainerRemove(context.Context, string, container.RemoveOptions) error

	ContainerStart(context.Context, string, container.StartOptions) error

	ContainerStop(context.Context, string, container.StopOptions) error

	ContainerWait(context.Context, string, container.WaitCondition) (<-chan container.WaitResponse, <-chan error)

	ImageBuild(context.Context, io.Reader, build.ImageBuildOptions) (build.ImageBuildResponse, error)

	ImageInspect(context.Context, string, ...client.ImageInspectOption) (image.InspectResponse, error)

	ImageList(context.Context, image.ListOptions) ([]image.Summary, error)

	ImagesPrune(context.Context, filters.Args) (image.PruneReport, error)

	ImagePush(context.Context, string, image.PushOptions) (io.ReadCloser, error)

	ImageTag(context.Context, string, string) error

	RegistryLogin(context.Context, registry.AuthConfig) (registry.AuthenticateOKBody, error)
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
	} else {
		return summary.ID[0:12]
	}
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
		cf.dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	}

	return cf.dockerClient, err
}
