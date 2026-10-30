package docker

import (
	"context"
	"errors"
	"io"
	"os"
	"os/exec"
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/moby/moby/api/types/image"
	"github.com/moby/moby/client"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/ioz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	testLogger, testHook = test.NewNullLogger()
)

type stubDockerClient struct {
	containerAttachError    error
	containerAttachId       string
	containerAttachOptions  *client.ContainerAttachOptions
	containerAttachResult   client.ContainerAttachResult
	containerCreateOptions  *client.ContainerCreateOptions
	containerCreateResult   client.ContainerCreateResult
	containerCreateError    error
	containerInspectError   error
	containerInspectId      string
	containerInspectOptions *client.ContainerInspectOptions
	containerInspectResult  client.ContainerInspectResult
	containerKillError      error
	containerKillId         string
	containerKillOptions    *client.ContainerKillOptions
	containerListError      error
	containerListOptions    *client.ContainerListOptions
	containerListResult     client.ContainerListResult
	containerRemoveError    error
	containerRemoveId       string
	containerRemoveOptions  *client.ContainerRemoveOptions
	containerRemoveResult   client.ContainerRemoveResult
	containerStartError     error
	containerStartId        string
	containerStartOptions   *client.ContainerStartOptions
	containerStartResult    client.ContainerStartResult
	containerStopError      error
	containerStopId         []string
	containerStopOptions    *client.ContainerStopOptions
	containerStopResult     client.ContainerStopResult
	containerWaitOptions    *client.ContainerWaitOptions
	containerWaitResult     client.ContainerWaitResult
	containerWaitSignal     string
	imageBuildError         error
	imageBuildOptions       *client.ImageBuildOptions
	imageBuildReader        io.Reader
	imageBuildResult        client.ImageBuildResult
	imageInspectError       error
	imageInspectId          string
	imageInspectOptions     []client.ImageInspectOption
	imageInspectResult      client.ImageInspectResult
	imageListError          error
	imageListOptions        *client.ImageListOptions
	imageListResult         client.ImageListResult
	imagePruneError         error
	imagePruneOptions       *client.ImagePruneOptions
	imagePruneResult        client.ImagePruneResult
	imagePushError          error
	imagePushOptions        *client.ImagePushOptions
	imagePushPath           string
	imagePushResult         client.ImagePushResponse
	imageTagError           error
	imageTagOptions         *client.ImageTagOptions
	imageTagResult          client.ImageTagResult
}

func (s *stubDockerClient) ContainerAttach(ctx context.Context, id string, options client.ContainerAttachOptions) (client.ContainerAttachResult, error) {
	_ = ctx
	s.containerAttachOptions = &options
	s.containerAttachId = id
	return s.containerAttachResult, s.containerAttachError
}

func (s *stubDockerClient) ContainerCreate(ctx context.Context, options client.ContainerCreateOptions) (client.ContainerCreateResult, error) {
	_ = ctx
	s.containerCreateOptions = &options
	return s.containerCreateResult, s.containerCreateError
}

func (s *stubDockerClient) ContainerInspect(ctx context.Context, id string, options client.ContainerInspectOptions) (client.ContainerInspectResult, error) {
	_ = ctx
	s.containerInspectId = id
	s.containerInspectOptions = &options
	return s.containerInspectResult, s.containerInspectError
}

func (s *stubDockerClient) ContainerKill(ctx context.Context, id string, options client.ContainerKillOptions) (client.ContainerKillResult, error) {
	_ = ctx
	s.containerKillId = id
	s.containerKillOptions = &options
	return client.ContainerKillResult{}, s.containerKillError
}

func (s *stubDockerClient) ContainerList(ctx context.Context, options client.ContainerListOptions) (client.ContainerListResult, error) {
	_ = ctx
	s.containerListOptions = &options
	return s.containerListResult, s.containerListError
}

func (s *stubDockerClient) ContainerRemove(ctx context.Context, id string, options client.ContainerRemoveOptions) (client.ContainerRemoveResult, error) {
	_ = ctx
	s.containerRemoveId = id
	s.containerRemoveOptions = &options
	return s.containerRemoveResult, s.containerRemoveError
}

func (s *stubDockerClient) ContainerStart(ctx context.Context, id string, options client.ContainerStartOptions) (client.ContainerStartResult, error) {
	_ = ctx
	s.containerStartId = id
	s.containerStartOptions = &options
	return s.containerStartResult, s.containerStartError
}

func (s *stubDockerClient) ContainerStop(ctx context.Context, id string, options client.ContainerStopOptions) (client.ContainerStopResult, error) {
	_ = ctx
	s.containerStopId = append(s.containerStopId, id)
	s.containerStopOptions = &options
	return s.containerStopResult, s.containerStopError
}

func (s *stubDockerClient) ContainerWait(ctx context.Context, signal string, options client.ContainerWaitOptions) client.ContainerWaitResult {
	_ = ctx
	s.containerWaitSignal = signal
	s.containerWaitOptions = &options
	return s.containerWaitResult
}

func (s *stubDockerClient) ImageBuild(ctx context.Context, reader io.Reader, options client.ImageBuildOptions) (client.ImageBuildResult, error) {
	_ = ctx
	s.imageBuildReader = reader
	s.imageBuildOptions = &options
	return s.imageBuildResult, s.imageBuildError
}

func (s *stubDockerClient) ImageInspect(ctx context.Context, id string, options ...client.ImageInspectOption) (client.ImageInspectResult, error) {
	_ = ctx
	s.imageInspectId = id
	s.imageInspectOptions = options
	return s.imageInspectResult, s.imageInspectError
}

func (s *stubDockerClient) ImageList(ctx context.Context, options client.ImageListOptions) (client.ImageListResult, error) {
	_ = ctx
	s.imageListOptions = &options
	return s.imageListResult, s.imageListError
}

func (s *stubDockerClient) ImagePrune(ctx context.Context, options client.ImagePruneOptions) (client.ImagePruneResult, error) {
	_ = ctx
	s.imagePruneOptions = &options
	return s.imagePruneResult, s.imagePruneError
}

func (s *stubDockerClient) ImagePush(ctx context.Context, path string, options client.ImagePushOptions) (client.ImagePushResponse, error) {
	_ = ctx
	s.imagePushPath = path
	s.imagePushOptions = &options
	return s.imagePushResult, s.imagePushError
}

func (s *stubDockerClient) ImageTag(ctx context.Context, options client.ImageTagOptions) (client.ImageTagResult, error) {
	_ = ctx
	s.imageTagOptions = &options
	return s.imageTagResult, s.imageTagError
}

func (s *stubDockerClient) RegistryLogin(context.Context, client.RegistryLoginOptions) (client.RegistryLoginResult, error) {
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

func newTestState() *task.State {
	testHook.Reset()
	return &task.State{
		Logger: testLogger,
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
