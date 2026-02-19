package docker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

func TestBuildParams_GetFilters(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "testRepository",
		DockerVersion:    "version",
	}

	actual := testParams.GetFilters()
	assert.Equal(t, []string{fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion)},
		actual.Get("reference"))
}

func TestBuildParams_GetFiltersByRepo(t *testing.T) {
	var testField = "field"
	var expectedRepo = "repo"
	var testParams = &BuildParams{
		DockerRepository: expectedRepo,
	}

	actual := testParams.GetFiltersByRepo(testField)
	assert.Equal(t, []string{expectedRepo + "*"}, actual.Get(testField))
}

func TestBuildParams_GetFiltersByRepo_WithNamespace(t *testing.T) {
	var testField = "field"
	var expectedRepo = "repo"
	var testParams = &BuildParams{
		DockerRepository: "namespace/" + expectedRepo,
	}

	actual := testParams.GetFiltersByRepo(testField)
	assert.Equal(t, []string{expectedRepo + "*"}, actual.Get(testField))
}

func TestBuildParams_GetFiltersByVersion(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "testRepository",
		DockerVersion:    "version",
	}

	actual := testParams.GetFiltersByVersion()
	assert.Equal(t, []string{fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion)},
		actual.Get("reference"))
}

func TestBuildParams_GetFiltersByVersion_Latest(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "testRepository",
		DockerVersion:    "latest",
	}

	actual := testParams.GetFiltersByVersion()
	assert.Equal(t, []string{testParams.DockerRepository + ":*"},
		actual.Get("reference"))
}

func TestBuildParams_GetReference(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), testParams.GetReference())
}

func TestBuildParams_GetVersion(t *testing.T) {
	var testParams = &BuildParams{}
	var expectedVersion = "version"

	assert.Equal(t, "latest", testParams.GetVersion())
	testParams.DockerVersion = expectedVersion
	assert.Equal(t, expectedVersion, testParams.GetVersion())
}

func TestBuildParams_getErrorStream(t *testing.T) {
	var testParams = &BuildParams{}
	var testFile = filepath.Join(t.TempDir(), "testErr")

	assert.Equal(t, os.Stderr, testParams.getErrorStream())
	testParams.Streams = &BuildStreams{}
	assert.Equal(t, os.Stderr, testParams.getErrorStream())

	if fd, err := os.Create(testFile); err == nil {
		testParams.Streams.Err = fd
		assert.Equal(t, fd, testParams.getErrorStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBuildParams_getInputStream(t *testing.T) {
	var testParams = &BuildParams{}
	var testFile = filepath.Join(t.TempDir(), "testInput")

	assert.Equal(t, os.Stdin, testParams.getInputStream())
	testParams.Streams = &BuildStreams{}
	assert.Equal(t, os.Stdin, testParams.getInputStream())

	if fd, err := os.Create(testFile); err == nil {
		testParams.Streams.In = fd
		assert.Equal(t, fd, testParams.getInputStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBuildParams_getOutputStream(t *testing.T) {
	var testParams = &BuildParams{}
	var testFile = filepath.Join(t.TempDir(), "testOutput")

	assert.Equal(t, os.Stdout, testParams.getOutputStream())
	testParams.Streams = &BuildStreams{}
	assert.Equal(t, os.Stdout, testParams.getOutputStream())

	if fd, err := os.Create(testFile); err == nil {
		testParams.Streams.Out = fd
		assert.Equal(t, fd, testParams.getOutputStream())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBuildParams_toBuildArgs(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	actual := testParams.toBuildArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), actual[3])
	assert.Equal(t, ".", actual[4])
}

func TestBuildParams_toBuildArgs_WithContext(t *testing.T) {
	var expectedDir = t.TempDir()
	var testParams = &BuildParams{
		DockerContext:    expectedDir,
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	actual := testParams.toBuildArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), actual[3])
	assert.Equal(t, expectedDir, actual[4])
}

func TestBuildParams_toBuildArgs_WithDockerfile(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "Dockerfile")
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		Dockerfile:       expectedFile,
	}

	actual := testParams.toBuildArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), actual[3])
	assert.Equal(t, "-f", actual[4])
	assert.Equal(t, expectedFile, actual[5])
	assert.Equal(t, ".", actual[6])
}

func TestBuildParams_toBuildArgs_WithLabel(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		Label:            true,
	}

	actual := testParams.toBuildArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), actual[3])
	assert.Equal(t, "--label", actual[4])
	assert.Equal(t, ".", actual[5])
}

func TestBuildParams_toBuildArgs_WithNoCache(t *testing.T) {
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		NoCache:          true,
	}

	actual := testParams.toBuildArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testParams.DockerRepository, testParams.DockerVersion), actual[3])
	assert.Equal(t, "--no-cache", actual[4])
	assert.Equal(t, ".", actual[5])
}

func TestBuildParams_toPruneArgs(t *testing.T) {
	var testParams = &BuildParams{}

	actual := testParams.toPruneArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "buildx", actual[0])
	assert.Equal(t, "prune", actual[1])
	assert.Equal(t, "--force", actual[2])
	assert.Equal(t, "--filter", actual[3])
	assert.Equal(t, "until=12h", actual[4])
}

func TestBuildParams_toPruneArgs_NoCache(t *testing.T) {
	var testParams = &BuildParams{
		NoCache: true,
	}

	actual := testParams.toPruneArgs()
	assert.NotEmpty(t, actual)
	assert.Equal(t, 3, len(actual))
	assert.Equal(t, "buildx", actual[0])
	assert.Equal(t, "prune", actual[1])
	assert.Equal(t, "--force", actual[2])
}

func TestNewBuildTask_Fork(t *testing.T) {
	var patchedUid = mock.Patches{T: t}.OsGeteuid(1)
	var patchLookup = mock.Patches{T: t}.ExecLookPath(func(string) (string, error) {
		return "path", nil
	})

	defer patchLookup.Unpatch()
	defer patchedUid.Unpatch()
	actual := NewBuildTask()

	if actual != nil {
		assert.Equal(t, "docker-build-fork", actual.Name)
		assert.NotEmpty(t, actual.OnPrepare)
		assert.NotEmpty(t, actual.OnComplete)
		assert.NotEmpty(t, actual.OnIncomplete)
		assert.NotEmpty(t, actual.OnPretend)
	} else {
		assert.Fail(t, "no task built")
	}
}

func TestNewBuildTask_Legacy(t *testing.T) {
	var expectedError = errors.New("expected")
	var patchPrintln = mock.Patches{T: t}.FmtPrintln(func(...any) {})
	var patchLookup = mock.Patches{T: t}.ExecLookPath(func(string) (string, error) {
		return "", expectedError
	})

	defer patchLookup.Unpatch()
	defer patchPrintln.Unpatch()
	actual := NewBuildTask()

	if actual != nil {
		assert.Equal(t, "docker-build-legacy", actual.Name)
		assert.NotEmpty(t, actual.OnPrepare)
		assert.NotEmpty(t, actual.OnComplete)
		assert.NotEmpty(t, actual.OnIncomplete)
		assert.NotEmpty(t, actual.OnPretend)
		values := cast.ToStringSlice(patchPrintln.CalledWith)
		assert.Contains(t, values[0], "DEPRECATED")
	} else {
		assert.Fail(t, "no task built")
	}
}

func TestNewBuildTask_RootError(t *testing.T) {
	var patchedUid = mock.Patches{T: t}.OsGeteuid(0)
	var patchLookup = mock.Patches{T: t}.ExecLookPath(func(string) (string, error) {
		return "path", nil
	})

	defer patchLookup.Unpatch()
	defer patchedUid.Unpatch()
	assert.Panics(t, func() { NewBuildTask() })
}

func TestNewInspectTask(t *testing.T) {
	var testTask = NewInspectTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewListTask(t *testing.T) {
	var testTask = NewListTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnPretend)
}

func Test_handleBuildContext(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerVersion: "version",
	}
	var expectedSummary = image.Summary{ID: "expectedId"}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{expectedSummary},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleBuildContext(testParams, testState))
	assert.Equal(t, expectedSummary.ID, testState.Output)
}

func Test_handleBuildContext_latestError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerVersion: "latest",
	}
	var expectedSummary = image.Summary{ID: "expectedId"}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{expectedSummary},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleBuildContext(testParams, testState), ErrorLatestBuild)
	assert.Equal(t, expectedSummary.ID, testState.Output)
}

func Test_handleBuildContext_listError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		imageListError: expectedError,
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleBuildContext(&BuildParams{}, testState), expectedError)
}

func Test_handleBuildContext_noResults(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleBuildContext(&BuildParams{}, testState), ErrorNoBuild)
}

func Test_handleBuildForkCreate(t *testing.T) {
	var testFolder = t.TempDir()
	var testState = &task.State{Logger: logrus.New(), Output: "notEmpty"}

	var testParams = &BuildParams{
		Env: task.Env{
			Context: t.Context(),
		},
		Streams: &BuildStreams{},
	}
	var fErr, fOut *os.File
	var err error

	if fErr, err = os.Create(filepath.Join(testFolder, "errFile")); err == nil {
		if fOut, err = os.Create(filepath.Join(testFolder, "outFile")); err == nil {
			var testFork = &stubFork{}

			defer installFork(testFork)()
			testParams.Streams.Err = fErr
			testParams.Streams.Out = fOut
			assert.NoError(t, handleBuildForkCreate("_path", testParams, testState))
			assert.Empty(t, testState.Output)
			assert.Same(t, fErr, testFork.stdErr)
			assert.Same(t, fOut, testFork.stdOut)
			return
		}
	}

	assert.NoError(t, err)
}

func Test_handleBuildForkCreate_RunError(t *testing.T) {
	var expectedError = errors.New("expectedError")
	var testLogger = logrus.New()
	var testState = &task.State{Logger: testLogger}
	var testFork = &stubFork{runError: expectedError}
	var testParams = &BuildParams{
		Env: task.Env{
			Context: t.Context(),
		},
	}

	defer installFork(testFork)()
	testLogger.Level = logrus.DebugLevel
	assert.ErrorIs(t, handleBuildForkCreate("_path", testParams, testState), expectedError)
}

func Test_handleBuildForkCreate_WaitError(t *testing.T) {
	var expectedError = errors.New("expectedError")
	var testLogger = logrus.New()
	var testState = &task.State{Logger: testLogger}
	var testFork = &stubFork{waitError: expectedError}
	var testParams = &BuildParams{
		Env: task.Env{
			Context: t.Context(),
		},
	}

	defer installFork(testFork)()
	testLogger.Level = logrus.DebugLevel
	assert.ErrorIs(t, handleBuildForkCreate("_path", testParams, testState), expectedError)
	assert.NotNil(t, testFork.pipeErrFunc)
}

func Test_handleBuildForkPrune(t *testing.T) {
	var expectedShortId = "123456789ABC"
	var expectedSize = int64(1024 * 1024)
	var testLogger = logrus.New()
	var testState = &task.State{Logger: testLogger}
	var testParams = &BuildParams{
		Env: task.Env{
			Context: t.Context(),
		},
		Prune: true,
	}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{{
			ID:   "sha256:" + expectedShortId,
			Size: expectedSize,
		}},
	}
	var testFork = &stubFork{}

	defer installFork(testFork)()
	defer installDockerClient(stubClient)()
	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleBuildForkPrune("_path", testParams, testState))
	assert.NotEmpty(t, testState.Reports)
	assert.Contains(t, testState.Reports[0], expectedShortId)
	assert.Contains(t, testState.Reports[0], "1.0MB")
	assert.NotNil(t, testFork.pipeOutFunc)
}

func Test_handleBuildForkPrune_forkError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &BuildParams{
		Env: task.Env{
			Context: t.Context(),
		},
		Prune: true,
	}
	var testFork = &stubFork{
		runError: expectedError,
	}

	defer installFork(testFork)()
	// prune errors should not fail the build
	assert.NoError(t, handleBuildForkPrune("_path", testParams, testState))
}

func Test_handleBuildForkPrune_noPruning(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleBuildForkPrune("_path", &BuildParams{}, testState))
}

func Test_handleBuildForkPrune_stateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Error: expectedError}

	assert.ErrorIs(t, handleBuildForkPrune("_path", &BuildParams{}, testState), expectedError)
}

func Test_handleBuildLegacyCreate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "notEmpty",
	}
	var testParams = &BuildParams{
		DockerContext:    t.TempDir(),
		DockerRepository: "repository",
		Label:            true,
	}
	var testResponse = []string{
		"An illegal json string",
		"{}",
		"{\"Stream\": \"a stream\"}",
	}
	var stubClient = &stubDockerClient{
		imageBuildResponseReader: io.NopCloser(strings.NewReader(strings.Join(testResponse, "\n"))),
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleBuildLegacyCreate(testParams, testState))
	assert.Empty(t, testState.Output)
	assert.NotNil(t, stubClient.imageBuildOptions)
	assert.Equal(t, testParams.DockerRepository, stubClient.imageBuildOptions.Labels["sf"])
}

func Test_handleBuildLegacyCreate_buildError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &BuildParams{
		DockerContext: t.TempDir(),
	}
	var stubClient = &stubDockerClient{imageBuildError: expectedError}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleBuildLegacyCreate(testParams, testState), expectedError)
}

func Test_handleBuildPretend(t *testing.T) {
	var lines []string
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {
		lines = append(lines, fmt.Sprintf(format, a...))
	})
	var testState = &task.State{
		Error:  ErrorNoBuild,
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		DockerContext:    t.TempDir(),
	}

	defer patch.Unpatch()
	assert.NoError(t, handleBuildPretend(testParams, testState))
	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.DockerContext)
}

func Test_handleBuildPretend_alreadyBuilt(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	defer patch.Unpatch()
	assert.NoError(t, handleBuildPretend(testParams, testState))
	actual := cast.ToStringSlice(patch.CalledWith)
	assert.NotEmpty(t, actual)
	assert.Contains(t, actual, testParams.GetReference())
}

func Test_handleBuildPretend_dockerFile(t *testing.T) {
	var lines []string
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {
		lines = append(lines, fmt.Sprintf(format, a...))
	})
	var testState = &task.State{
		Error:  ErrorNoBuild,
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		DockerContext:    t.TempDir(),
		Dockerfile:       "expectedFile",
	}

	defer patch.Unpatch()
	assert.NoError(t, handleBuildPretend(testParams, testState))
	assert.Equal(t, 2, len(lines))
	assert.Contains(t, lines[0], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.Dockerfile)
	assert.Contains(t, lines[1], testParams.DockerContext)
}

func Test_handleBuildPretend_withPruning(t *testing.T) {
	var lines []string
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {
		lines = append(lines, fmt.Sprintf(format, a...))
	})
	var testState = &task.State{
		Error:  ErrorNoBuild,
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerRepository: "repository",
		DockerVersion:    "version",
		DockerContext:    t.TempDir(),
		Prune:            true,
	}

	defer patch.Unpatch()
	assert.NoError(t, handleBuildPretend(testParams, testState))
	assert.Equal(t, 3, len(lines))
	assert.Contains(t, lines[0], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.GetReference())
	assert.Contains(t, lines[1], testParams.DockerContext)
	assert.Contains(t, lines[2], testParams.DockerRepository)
}

func Test_handleBuildPrune(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &BuildParams{
		Prune: true,
	}
	var stubClient = &stubDockerClient{
		imageList:           []image.Summary{},
		imagePruneDeletions: []string{"deleted"},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleBuildPrune(testParams, testState))
}

func Test_handleBuildPrune_noPrune(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleBuildPrune(&BuildParams{}, testState))
}

func Test_handleBuildPrune_pruneError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &BuildParams{
		Prune: true,
	}
	var stubClient = &stubDockerClient{
		imageList:       []image.Summary{},
		imagePruneError: expectedError,
	}

	defer installDockerClient(stubClient)()
	// pruning can not fail the build even if there's an error logged
	assert.NoError(t, handleBuildPrune(testParams, testState))
}

func Test_handleBuildPrune_stateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleBuildPrune(&BuildParams{}, testState), expectedError)
}

func Test_handleBuildReport_stateError(t *testing.T) {
	var testState = &task.State{Error: ErrorLatestBuild}

	assert.ErrorIs(t, handleBuildReport(&BuildParams{}, testState), ErrorLatestBuild)
}

func Test_handleListContext(t *testing.T) {
	var expectedId = "expectedId"
	var expectedContainer = "expectedContainer"
	var testState = &task.State{}
	var testParams = &BuildParams{
		DockerRepository: "repository",
	}
	var stubClient = &stubDockerClient{
		containerList: []container.Summary{{
			ID: expectedContainer,
		}},
		imageList: []image.Summary{{
			ID: expectedId,
		}},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleListContext(testParams, testState))
	assert.NotNil(t, stubClient.imageListOptions)
	assert.True(t, stubClient.imageListOptions.Filters.Contains("reference"))
	assert.NotNil(t, testState.Internal)
	tracking, ok := testState.Internal.(clientTracking)
	assert.True(t, ok)
	assert.Equal(t, expectedId, tracking.images[0].ID)
	assert.Equal(t, expectedContainer, tracking.containers[0].ID)
}

func Test_handleListContext_containerListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedId = "expectedId"
	var testState = &task.State{}
	var testParams = &BuildParams{
		DockerRepository: "repository",
	}
	var stubClient = &stubDockerClient{
		containerListError: expectedError,
		imageList: []image.Summary{{
			ID: expectedId,
		}},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleListContext(testParams, testState), expectedError)
	assert.NotNil(t, stubClient.imageListOptions)
	assert.True(t, stubClient.imageListOptions.Filters.Contains("reference"))
	assert.NotNil(t, testState.Internal)
	tracking, ok := testState.Internal.(clientTracking)
	assert.True(t, ok)
	assert.Equal(t, expectedId, tracking.images[0].ID)
}

func Test_handleListContext_imageListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var stubClient = &stubDockerClient{
		imageListError: expectedError,
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleListContext(&BuildParams{}, &task.State{}), expectedError)
}

func Test_handleListContext_noBuildError(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var stubClient = &stubDockerClient{}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleListContext(&BuildParams{}, testState), ErrorNoBuild)
}

func Test_handleListComplete(t *testing.T) {
	var expectedImageId = "123456789ABC"
	var expectedAlternateId = "PQRSTUVWXYZ0"
	var expectedImageVersion = "imageVersion"
	var expectedContainerId = "DEFGHIJKLMNO"
	var expectedContainerName = "expectedName"
	var imageTimestamp = time.Now().Add(time.Duration(-1) * time.Hour)
	var containerTimestamp = time.Now().Add(time.Duration(-24) * time.Hour)
	var testState = &task.State{
		Internal: clientTracking{
			containers: []container.Summary{
				{
					ID:      hashPrefix + expectedContainerId,
					Created: containerTimestamp.Unix(),
					ImageID: hashPrefix + expectedImageId,
				},
				{
					ID:      hashPrefix + "untestedContainerId",
					ImageID: hashPrefix + expectedAlternateId,
					Names:   []string{"/" + expectedContainerName},
				},
			},
			images: []image.Summary{
				{
					ID:       hashPrefix + expectedImageId,
					Created:  imageTimestamp.Unix(),
					RepoTags: []string{"tag:" + expectedImageVersion},
				},
				{
					ID:      hashPrefix + expectedAlternateId,
					Created: imageTimestamp.Unix(),
				},
			},
		},
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerRepository: "repository",
	}

	assert.NoError(t, handleListComplete(testParams, testState))
	assert.Contains(t, testState.Output, expectedAlternateId)
	assert.Contains(t, testState.Output, expectedImageId)
	assert.Contains(t, testState.Output, expectedImageVersion)
	assert.Contains(t, testState.Output, expectedContainerId)
	assert.Contains(t, testState.Output, expectedContainerName)
}

func Test_handleListComplete_noContainers(t *testing.T) {
	var expectedImageId = "123456789ABC"
	var timestamp = time.Now().Add(time.Duration(-1) * time.Hour)
	var testState = &task.State{
		Internal: clientTracking{
			images: []image.Summary{
				{
					ID:       hashPrefix + expectedImageId,
					Created:  timestamp.Unix(),
					RepoTags: []string{"tag"},
				},
			},
		},
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{
		DockerRepository: "repository",
	}

	assert.NoError(t, handleListComplete(testParams, testState))
	assert.Contains(t, testState.Output, "No containers")
	assert.Contains(t, testState.Output, expectedImageId)
}

func Test_handleListComplete_noImages(t *testing.T) {
	var testState = &task.State{}
	var testParams = &BuildParams{
		DockerRepository: "repository",
	}

	assert.NoError(t, handleListComplete(testParams, testState))
	assert.Contains(t, testState.Output, testParams.DockerRepository)
}

func Test_handleInspectComplete(t *testing.T) {
	var expectedDigest = hashPrefix + "123456789ABC"
	var expectedImage = &image.Summary{
		ID:          "expectedId",
		RepoDigests: []string{expectedDigest},
	}
	var testParams = &BuildParams{
		DockerVersion: "expectedVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedImage.ID,
	}
	var stubClient = &stubDockerClient{
		imageInspect: expectedImage,
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleInspectComplete(testParams, testState))
	assert.Equal(t, stubClient.imageInspectId, expectedImage.ID)
	actual, ok := testState.Internal.(*shared.Identity)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedImage.ID, actual.Id)
	assert.Equal(t, testParams.DockerVersion, actual.Version)
	assert.Equal(t, expectedDigest, actual.Hash)
}

func Test_handleInspectComplete_inspectError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &BuildParams{}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "id",
	}
	var stubClient = &stubDockerClient{
		imageInspectError: expectedError,
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleInspectComplete(testParams, testState), expectedError)
	assert.Equal(t, stubClient.imageInspectId, testState.Output)
}

func Test_handleInspectComplete_noBuild(t *testing.T) {
	assert.ErrorIs(t, handleInspectComplete(&BuildParams{}, &task.State{}), ErrorNoBuild)
}

func Test_handleInspectComplete_invalidRepoDigest(t *testing.T) {
	var expectedImage = &image.Summary{
		ID: "expectedId",
		RepoDigests: []string{
			"invalidDigest",
		},
	}
	var testParams = &BuildParams{
		DockerVersion: "expectedVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedImage.ID,
	}
	var stubClient = &stubDockerClient{
		imageInspect: expectedImage,
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleInspectComplete(testParams, testState))
	assert.Equal(t, stubClient.imageInspectId, expectedImage.ID)
	actual, ok := testState.Internal.(*shared.Identity)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedImage.ID, actual.Id)
	assert.Equal(t, testParams.DockerVersion, actual.Version)
	assert.Empty(t, actual.Hash)
}

func Test_handleInspectComplete_noRepoDigest(t *testing.T) {
	var expectedImage = &image.Summary{
		ID: "expectedId",
	}
	var testParams = &BuildParams{
		DockerVersion: "expectedVersion",
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedImage.ID,
	}
	var stubClient = &stubDockerClient{
		imageInspect: expectedImage,
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleInspectComplete(testParams, testState))
	assert.Equal(t, stubClient.imageInspectId, expectedImage.ID)
	actual, ok := testState.Internal.(*shared.Identity)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedImage.ID, actual.Id)
	assert.Equal(t, testParams.DockerVersion, actual.Version)
	assert.Empty(t, actual.Hash)
}

func Test_handleInspectContext(t *testing.T) {
	var expectedId = "expectedId"
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{DockerVersion: "version"}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{
			{
				ID: expectedId,
			},
		},
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleInspectContext(testParams, testState))
	assert.Equal(t, expectedId, testState.Output)
}

func Test_handleInspectContext_buildError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		imageListError: expectedError,
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleInspectContext(&BuildParams{}, testState), expectedError)
}

func Test_handleInspectContext_latestBuildError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var stubClient = &stubDockerClient{
		imageListError: ErrorLatestBuild,
	}

	defer installDockerClient(stubClient)()
	assert.NoError(t, handleInspectContext(&BuildParams{}, testState))
}

func Test_handleInspectContext_knownOutput(t *testing.T) {
	var testState = &task.State{Output: "known"}

	assert.NoError(t, handleInspectContext(&BuildParams{}, testState))
}

func Test_handleInspectContext_noBuild(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &BuildParams{DockerVersion: "version"}
	var stubClient = &stubDockerClient{
		imageList: []image.Summary{
			{},
		},
	}

	defer installDockerClient(stubClient)()
	assert.ErrorIs(t, handleInspectContext(testParams, testState), ErrorNoBuild)
}

func Test_handleInspectPretend(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &BuildParams{
		DockerVersion: "version",
	}

	defer patch.Unpatch()
	assert.NoError(t, handleInspectPretend(testParams, testState))
	assert.True(t, patch.Called)
	assert.Contains(t, patch.CalledWith, testState.Output)
	actual, ok := testState.Internal.(*shared.Identity)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	assert.Equal(t, testState.Output, actual.Hash)
	assert.Equal(t, testParams.DockerVersion, actual.Version)
}

func Test_handleInspectPretend_latestBuild(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testState = &task.State{
		Error:  ErrorLatestBuild,
		Logger: logrus.New(),
		Output: "output",
	}

	defer patch.Unpatch()
	assert.NoError(t, handleInspectPretend(&BuildParams{}, testState))
	assert.True(t, patch.Called)
	assert.Contains(t, patch.CalledWith, testState.Output)
	actual, ok := testState.Internal.(*shared.Identity)
	assert.True(t, ok)
	assert.NotNil(t, actual)
	assert.Equal(t, testState.Output, actual.Hash)
	assert.Equal(t, "latest", actual.Version)
}

func Test_handleInspectPretend_stateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  expectedError,
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleInspectPretend(&BuildParams{}, testState), expectedError)
}

func Test_stateDebug(t *testing.T) {
	var testLogger = logrus.New()
	var testTask = &task.State{Logger: testLogger}
	var testWriter = stateDebug(testTask)
	var testHook = test.NewLocal(testLogger)
	var expectedValue = "value"

	testLogger.Level = logrus.DebugLevel
	testWriter(expectedValue)
	assert.NotEmpty(t, testHook.Entries)
	assert.Equal(t, expectedValue, testHook.Entries[0].Message)
}

func Test_stateError(t *testing.T) {
	var testLogger = logrus.New()
	var testTask = &task.State{Logger: testLogger}
	var testWriter = stateError(testTask)
	var testHook = test.NewLocal(testLogger)
	var expectedError = "this is an error"
	var expectedLine = "a line"
	var expectedDebug = "a debug comment"

	testLogger.Level = logrus.DebugLevel
	testWriter(fmt.Sprintf("error: %s", expectedError))
	testWriter("")
	testWriter(fmt.Sprintf("%s", expectedLine))
	testWriter("#")     // an empty comment line
	testWriter("#4323") // a weird status line
	testWriter(fmt.Sprintf("#4324 %s", expectedDebug))
	assert.NotEmpty(t, testHook.Entries)
	assert.Equal(t, 3, len(testHook.Entries))
	assert.Equal(t, expectedError, testHook.Entries[0].Message)
	assert.Equal(t, logrus.ErrorLevel, testHook.Entries[0].Level)
	assert.Equal(t, expectedLine, testHook.Entries[1].Message)
	assert.Equal(t, expectedDebug, testHook.Entries[2].Message)
}
