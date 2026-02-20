package docker

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
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

	"genaiz.com/genaiz-lib/lang/ioz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubDockerClient struct {
	containerAttachError     error
	containerAttachId        string
	containerAttachOptions   *container.AttachOptions
	containerAttachResponse  *types.HijackedResponse
	containerCreate          *container.Summary
	containerCreateConfig    *container.Config
	containerCreateName      string
	containerCreatePlatform  *ocispec.Platform
	containerCreateWarnings  []string
	containerHostConfig      *container.HostConfig
	containerNetworkConfig   *network.NetworkingConfig
	containerCreateError     error
	containerInspect         *container.Summary
	containerInspectConfig   *container.Config
	containerInspectError    error
	containerInspectId       string
	containerList            []container.Summary
	containerListFilter      filters.Args
	containerListError       error
	containerKillError       error
	containerRemoveError     error
	containerRemoveId        string
	containerRemoveOptions   *container.RemoveOptions
	containerStartError      error
	containerStartId         string
	containerStartOptions    *container.StartOptions
	containerStopError       error
	containerStopId          []string
	containerStopOptions     *container.StopOptions
	containerWaitResponse    chan container.WaitResponse
	containerWaitError       chan error
	imageBuildOptions        *build.ImageBuildOptions
	imageBuildReader         io.Reader
	imageBuildResponseReader io.ReadCloser
	imageBuildError          error
	imageInspect             *image.Summary
	imageInspectError        error
	imageInspectId           string
	imageInspectOptions      []client.ImageInspectOption
	imageList                []image.Summary
	imageListError           error
	imageListOptions         *image.ListOptions
	imagePruneArgs           *filters.Args
	imagePruneDeletions      []string
	imagePruneError          error
	imagePushError           error
	imagePushOptions         image.PushOptions
	imagePushPath            string
	imagePushReader          io.ReadCloser
	imageTagError            error
	imageTagId               string
	imageTagPath             string
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

func (s *stubDockerClient) ImageBuild(ctx context.Context, reader io.Reader, options build.ImageBuildOptions) (build.ImageBuildResponse, error) {
	s.imageBuildReader = reader
	s.imageBuildOptions = &options
	_ = ctx
	return build.ImageBuildResponse{
		Body: s.imageBuildResponseReader,
	}, s.imageBuildError
}

func (s *stubDockerClient) ImageInspect(ctx context.Context, id string, options ...client.ImageInspectOption) (image.InspectResponse, error) {
	var response image.InspectResponse

	_ = ctx
	s.imageInspectId = id
	s.imageInspectOptions = options

	if s.imageInspect != nil {
		response = image.InspectResponse{
			ID:          s.imageInspect.ID,
			RepoDigests: s.imageInspect.RepoDigests,
		}
	}

	return response, s.imageInspectError
}

func (s *stubDockerClient) ImageList(ctx context.Context, options image.ListOptions) ([]image.Summary, error) {
	_ = ctx
	s.imageListOptions = &options

	if s.imageListError == nil {
		return s.imageList, nil
	}

	return nil, s.imageListError
}

func (s *stubDockerClient) ImagesPrune(ctx context.Context, args filters.Args) (image.PruneReport, error) {
	var responses []image.DeleteResponse

	for _, deletion := range s.imagePruneDeletions {
		responses = append(responses, image.DeleteResponse{Deleted: deletion})
	}

	s.imagePruneArgs = &args
	_ = ctx
	return image.PruneReport{
		ImagesDeleted: responses,
	}, s.imagePruneError
}

func (s *stubDockerClient) ImagePush(ctx context.Context, imagePushPath string, imagePushOptions image.PushOptions) (io.ReadCloser, error) {
	_ = ctx
	s.imagePushPath = imagePushPath
	s.imagePushOptions = imagePushOptions
	return s.imagePushReader, s.imagePushError
}

func (s *stubDockerClient) ImageTag(ctx context.Context, imageTagId string, imageTagPath string) error {
	_ = ctx
	s.imageTagId = imageTagId
	s.imageTagPath = imageTagPath
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

type stubFork struct {
	pipeErrFunc func(string)
	pipeOutFunc func(string)
	runContext  context.Context
	runError    error
	stdErr      *os.File
	stdIn       *os.File
	stdOut      *os.File
	waitError   error
}

func (sf *stubFork) GetWaitError() error {
	return sf.waitError
}

func (sf *stubFork) Run(runContext context.Context) error {
	sf.runContext = runContext
	return sf.runError
}

func (sf *stubFork) WithPipeErr(pipeErrFunc func(string)) ioz.Fork {
	sf.pipeErrFunc = pipeErrFunc
	return sf
}

func (sf *stubFork) WithPipeOut(pipeOutFunc func(string)) ioz.Fork {
	sf.pipeOutFunc = pipeOutFunc
	return sf
}

func (sf *stubFork) WithStdErr(stdErr *os.File) ioz.Fork {
	sf.stdErr = stdErr
	return sf
}

func (sf *stubFork) WithStdIn(stdIn *os.File) ioz.Fork {
	sf.stdIn = stdIn
	return sf
}

func (sf *stubFork) WithStdOut(stdOut *os.File) ioz.Fork {
	sf.stdOut = stdOut
	return sf
}

func installFork(fork ioz.Fork) func() {
	var resetFactory = forkFactory.get

	forkFactory.get = func(ff *ForkFactory, cmd *exec.Cmd) ioz.Fork {
		return fork
	}

	return func() {
		forkFactory.get = resetFactory
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

func TestNewClientState_VarSpec(t *testing.T) {
	var expectedKey = "testKey"
	var expectedTaskState = &task.State{
		Internal: shared.VarSpecTracking{
			VarSpecs: []shared.VarSpec{
				broker.PropSpec{
					Key: expectedKey,
				},
			},
		},
	}
	var testState = NewClientState(expectedTaskState)

	assert.Empty(t, testState.containers)
	assert.Empty(t, testState.images)
	assert.Same(t, expectedTaskState, testState.state)
	assert.NotEmpty(t, testState.VarSpecs)
	assert.Equal(t, expectedKey, testState.VarSpecs[0].GetKey())
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

func Test_getFork(t *testing.T) {
	var testForkFactory = &ForkFactory{}
	var cmd = &exec.Cmd{}
	var testFork = getFork(testForkFactory, cmd)

	assert.Same(t, testFork, getFork(testForkFactory, cmd))
}
