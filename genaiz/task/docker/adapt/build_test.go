package adapt

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestBuildArguments_CliArguments(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, ".", actual[4])
}

func TestBuildArguments_CliArguments_WithContext(t *testing.T) {
	var expectedDir = t.TempDir()
	var testArgs = &BuildArguments{
		DockerContext:    expectedDir,
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, expectedDir, actual[4])
}

func TestBuildArguments_CliArguments_WithDockerfile(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "Dockerfile")
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
		Dockerfile:       expectedFile,
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, "-f", actual[4])
	assert.Equal(t, expectedFile, actual[5])
	assert.Equal(t, ".", actual[6])
}

func TestBuildArguments_CliArguments_WithPlatform(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), "Dockerfile")
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
		Dockerfile:       expectedFile,
		Platform:         "testPlatform",
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, "-f", actual[4])
	assert.Equal(t, expectedFile, actual[5])
	assert.Equal(t, "--platform", actual[6])
	assert.Equal(t, testArgs.Platform, actual[7])
	assert.Equal(t, ".", actual[8])
}

func TestBuildArguments_CliArguments_WithLabel(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
		Label:            true,
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, "--label", actual[4])
	assert.Equal(t, ".", actual[5])
}

func TestBuildArguments_CliArguments_WithNoCache(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
		NoCache:          true,
	}

	actual := testArgs.CliArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "build", actual[0])
	assert.Equal(t, "--pull", actual[1])
	assert.Equal(t, "-t", actual[2])
	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), actual[3])
	assert.Equal(t, "--no-cache", actual[4])
	assert.Equal(t, ".", actual[5])
}

func TestBuildArguments_PruneArguments(t *testing.T) {
	var testArgs = &BuildArguments{}

	actual := testArgs.PruneArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, "buildx", actual[0])
	assert.Equal(t, "prune", actual[1])
	assert.Equal(t, "--force", actual[2])
	assert.Equal(t, "--filter", actual[3])
	assert.Equal(t, "until=12h", actual[4])
}

func TestBuildArguments_PruneArguments_NoCache(t *testing.T) {
	var testArgs = &BuildArguments{
		NoCache: true,
	}

	actual := testArgs.PruneArguments()
	assert.NotEmpty(t, actual)
	assert.Equal(t, 3, len(actual))
	assert.Equal(t, "buildx", actual[0])
	assert.Equal(t, "prune", actual[1])
	assert.Equal(t, "--force", actual[2])
}

func TestBuildArguments_GetFilters(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "testRepository",
		DockerVersion:    "version",
	}
	var expectedFilterKey = fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion)

	actual := testArgs.GetFilters()
	assert.True(t, actual["reference"][expectedFilterKey])
}

func TestBuildArguments_GetFiltersByRepo(t *testing.T) {
	var testField = "field"
	var expectedRepo = "repo"
	var testArgs = &BuildArguments{
		DockerRepository: expectedRepo,
	}
	var expectedFilterKey = fmt.Sprintf("%s*", testArgs.DockerRepository)

	actual := testArgs.GetFiltersByRepo(testField)
	assert.True(t, actual[testField][expectedFilterKey])
}

func TestBuildArguments_GetFiltersByRepo_WithNamespace(t *testing.T) {
	var testField = "field"
	var expectedRepo = "repo"
	var testArgs = &BuildArguments{
		DockerRepository: "namespace/" + expectedRepo,
	}
	var expectedFilterKey = fmt.Sprintf("%s*", expectedRepo)

	actual := testArgs.GetFiltersByRepo(testField)
	assert.True(t, actual[testField][expectedFilterKey])
}

func TestBuildArguments_GetFiltersByVersion(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "testRepository",
		DockerVersion:    "version",
	}
	var expectedFilterKey = fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion)

	actual := testArgs.GetFiltersByVersion()
	assert.True(t, actual["reference"][expectedFilterKey])
}

func TestBuildArguments_GetFiltersByVersion_Latest(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "testRepository",
		DockerVersion:    "latest",
	}
	var expectedFilterKey = fmt.Sprintf("%s:*", testArgs.DockerRepository)

	actual := testArgs.GetFiltersByVersion()
	assert.True(t, actual["reference"][expectedFilterKey])
}

func TestBuildArguments_GetReference(t *testing.T) {
	var testArgs = &BuildArguments{
		DockerRepository: "repository",
		DockerVersion:    "version",
	}

	assert.Equal(t, fmt.Sprintf("%s:%s", testArgs.DockerRepository, testArgs.DockerVersion), testArgs.GetReference())
}

func TestBuildArguments_GetPlatform(t *testing.T) {
	var expectedOS = "linux"
	var expectedArch = "amd64"
	var testArgs = &BuildArguments{
		Platform: fmt.Sprintf("%s/%s", expectedOS, expectedArch),
	}

	if actual := testArgs.GetPlatforms(); len(actual) == 1 {
		assert.Equal(t, expectedOS, actual[0].OS)
		assert.Equal(t, expectedArch, actual[0].Architecture)
	} else {
		assert.Fail(t, "expected a slice of 1 ocispec.Platform")
	}
}

func TestBuildArguments_GetPlatform_Empty(t *testing.T) {
	var testArgs = &BuildArguments{}

	assert.Empty(t, testArgs.GetPlatforms())
}

func TestBuildArguments_GetPlatform_SingleString(t *testing.T) {
	var testArgs = &BuildArguments{
		Platform: "weirdOs",
	}

	if actual := testArgs.GetPlatforms(); len(actual) == 1 {
		assert.Equal(t, testArgs.Platform, actual[0].OS)
	} else {
		assert.Fail(t, "expected a slice of 1 ocispec.Platform")
	}
}

func TestBuildArguments_GetVersion(t *testing.T) {
	var testArgs = &BuildArguments{}
	var expectedVersion = "version"

	assert.Equal(t, "latest", testArgs.GetVersion())
	testArgs.DockerVersion = expectedVersion
	assert.Equal(t, expectedVersion, testArgs.GetVersion())
}
