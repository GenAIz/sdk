package docker

import (
	"context"
	"errors"
	"io"
	"testing"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/build"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/network"
	"github.com/docker/docker/api/types/registry"
	"github.com/docker/docker/client"
	ocispec "github.com/opencontainers/image-spec/specs-go/v1"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
)

type stubDockerClient struct {
	containerAttachError    error
	containerAttachId       string
	containerAttachOptions  *container.AttachOptions
	containerAttachResponse *types.HijackedResponse
	containerCreate         *container.Summary
	containerCreateConfig   *container.Config
	containerCreateName     string
	containerCreatePlatform *ocispec.Platform
	containerCreateWarnings []string
	containerHostConfig     *container.HostConfig
	containerNetworkConfig  *network.NetworkingConfig
	containerCreateError    error
	containerInspect        *container.Summary
	containerInspectConfig  *container.Config
	containerInspectError   error
	containerInspectId      string
	containerList           []container.Summary
	containerListFilter     filters.Args
	containerListError      error
	containerKillError      error
	containerRemoveError    error
	containerRemoveId       string
	containerRemoveOptions  *container.RemoveOptions
	containerStartError     error
	containerStartId        string
	containerStartOptions   *container.StartOptions
	containerStopError      error
	containerStopId         []string
	containerStopOptions    *container.StopOptions
	containerWaitResponse   chan container.WaitResponse
	containerWaitError      chan error
	imageList               []image.Summary
	imageListError          error
	imageTagError           error
}

func (s *stubDockerClient) ContainerAttach(ctx context.Context, id string, options container.AttachOptions) (types.HijackedResponse, error) {
	s.containerAttachOptions = &options
	s.containerAttachId = id
	_ = ctx

	if s.containerAttachResponse != nil {
		return *s.containerAttachResponse, nil
	}

	return types.HijackedResponse{}, s.containerAttachError
}

func (s *stubDockerClient) ContainerCreate(ctx context.Context, config *container.Config, hostConfig *container.HostConfig, networkConfig *network.NetworkingConfig, platform *ocispec.Platform, name string) (container.CreateResponse, error) {
	s.containerCreateConfig = config
	s.containerHostConfig = hostConfig
	s.containerNetworkConfig = networkConfig
	s.containerCreatePlatform = platform
	s.containerCreateName = name
	_ = ctx

	if s.containerCreateError == nil && s.containerCreate != nil {
		return container.CreateResponse{
			ID:       s.containerCreate.ID,
			Warnings: s.containerCreateWarnings,
		}, nil
	}

	return container.CreateResponse{}, s.containerCreateError
}

func (s *stubDockerClient) ContainerInspect(ctx context.Context, id string) (container.InspectResponse, error) {
	s.containerInspectId = id
	_ = ctx

	if s.containerInspectError == nil {
		return container.InspectResponse{
			ContainerJSONBase: &container.ContainerJSONBase{ID: s.containerInspect.ID},
			Config:            s.containerInspectConfig,
		}, nil
	}

	return container.InspectResponse{}, s.containerInspectError
}

func (s *stubDockerClient) ContainerKill(context.Context, string, string) error {
	return s.containerKillError
}

func (s *stubDockerClient) ContainerList(ctx context.Context, options container.ListOptions) ([]container.Summary, error) {
	_ = ctx
	s.containerListFilter = options.Filters

	if s.containerListError == nil {
		return s.containerList, nil
	}

	return nil, s.containerListError
}

func (s *stubDockerClient) ContainerRemove(ctx context.Context, id string, options container.RemoveOptions) error {
	s.containerRemoveOptions = &options
	s.containerRemoveId = id
	_ = ctx
	return s.containerRemoveError
}

func (s *stubDockerClient) ContainerStart(ctx context.Context, id string, options container.StartOptions) error {
	s.containerStartId = id
	s.containerStartOptions = &options
	_ = ctx
	return s.containerStartError
}

func (s *stubDockerClient) ContainerStop(ctx context.Context, id string, options container.StopOptions) error {
	s.containerStopId = append(s.containerStopId, id)
	s.containerStopOptions = &options
	_ = ctx
	return s.containerStopError
}

func (s *stubDockerClient) ContainerWait(context.Context, string, container.WaitCondition) (<-chan container.WaitResponse, <-chan error) {
	return s.containerWaitResponse, s.containerWaitError
}

func (s *stubDockerClient) ImageBuild(context.Context, io.Reader, build.ImageBuildOptions) (build.ImageBuildResponse, error) {
	panic("implement me")
}

func (s *stubDockerClient) ImageInspect(context.Context, string, ...client.ImageInspectOption) (image.InspectResponse, error) {
	panic("implement me")
}

func (s *stubDockerClient) ImageList(context.Context, image.ListOptions) ([]image.Summary, error) {
	if s.imageListError == nil {
		return s.imageList, nil
	}

	return nil, s.imageListError
}

func (s *stubDockerClient) ImagesPrune(context.Context, filters.Args) (image.PruneReport, error) {
	panic("implement me")
}

func (s *stubDockerClient) ImagePush(context.Context, string, image.PushOptions) (io.ReadCloser, error) {
	panic("implement me")
}

func (s *stubDockerClient) ImageTag(context.Context, string, string) error {
	return s.imageTagError
}

func (s *stubDockerClient) RegistryLogin(context.Context, registry.AuthConfig) (registry.AuthenticateOKBody, error) {
	panic("implement me")
}

func installDockerClient(cl Client) func() {
	var resetFactory = dockerFactory.get

	dockerFactory.get = func(cf *ClientFactory) (Client, error) {
		return cl, nil
	}

	return func() {
		dockerFactory.get = resetFactory
	}
}

func TestClientFactory_Get(t *testing.T) {
	dockerFactory.get = func(cf *ClientFactory) (Client, error) {
		return nil, nil
	}
	assert.Empty(t, dockerFactory.Get())
}

func TestClientFactory_GetPanic(t *testing.T) {
	dockerFactory.get = func(cf *ClientFactory) (Client, error) {
		return nil, errors.New("panic")
	}
	assert.Panics(t, func() { dockerFactory.Get() })
}

func TestNewClientState(t *testing.T) {
	var expectedTaskState = &task.State{}
	var testState = NewClientState(expectedTaskState)

	assert.Empty(t, testState.containers)
	assert.Empty(t, testState.images)
	assert.Same(t, expectedTaskState, testState.state)
	assert.Equal(t, 0, testState.GetContainersSize())
	assert.False(t, testState.HasContainers())
	assert.Equal(t, 0, testState.GetImagesSize())
	assert.False(t, testState.HasImages())
	assert.Empty(t, testState.SelectLatestContainer())
}

func TestClientState_DisplayString(t *testing.T) {
	var testState = &ClientState{}
	var testSummary = &container.Summary{}
	var expectedName = "name"
	var expectedId = "123456789ABCDE"

	testSummary.ID = expectedId
	assert.Equal(t, expectedId[0:12], testState.DisplayString(testSummary))
	testSummary.Names = []string{"/" + expectedName}
	assert.Equal(t, expectedName, testState.DisplayString(testSummary))
}

func TestClientState_GetContainers(t *testing.T) {
	var testState = NewClientState(&task.State{})
	var expectedId = "id"
	var expectedCreated = int64(37)

	testState.AddContainers(container.Summary{
		ID:      expectedId,
		Created: expectedCreated,
	})
	assert.True(t, testState.HasContainers())
	assert.Equal(t, 1, testState.GetContainersSize())
	assert.Equal(t, expectedId, testState.GetContainers()[0].ID)
	assert.Equal(t, expectedCreated, testState.GetContainers()[0].Created)
}

func TestClientState_GetImages(t *testing.T) {
	var testState = NewClientState(&task.State{})
	var expectedId = "id"
	var expectedCreated = int64(37)

	testState.AddImages(image.Summary{
		ID:      expectedId,
		Created: expectedCreated,
	})
	assert.True(t, testState.HasImages())
	assert.Equal(t, 1, testState.GetImagesSize())
	assert.Equal(t, expectedId, testState.GetImages()[0].ID)
	assert.Equal(t, expectedCreated, testState.GetImages()[0].Created)
}

func TestClientState_Reset(t *testing.T) {
	var notExpected = "notExpected"
	var expectedTaskState = &task.State{Internal: notExpected}
	var testState = NewClientState(expectedTaskState)

	testState.Reset()
	assert.Empty(t, expectedTaskState.Internal)
}

func TestClientState_SelectLatestContainer(t *testing.T) {
	var expectedId = "latest"
	var expectedTaskState = &task.State{Internal: clientTracking{
		containers: []container.Summary{
			{
				ID:      expectedId,
				Created: int64(42),
			},
			{
				ID:      "notExpected",
				Created: int64(37),
			},
		},
	}}
	var testState = NewClientState(expectedTaskState)

	if actual := testState.SelectLatestContainer(); actual != nil {
		assert.Equal(t, expectedId, actual.ID)
	} else {
		assert.NotEmpty(t, actual)
	}
}

func Test_getClientFactory(t *testing.T) {
	var result, err = getDockerClient(dockerFactory)

	assert.NotEmpty(t, result)
	assert.NoError(t, err)
}
