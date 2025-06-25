package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

func TestBuildExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var expectedDockerContext = "context"
	var testDockerContext = newOptionDockerContext()
	var expectedDockerFile = "file"
	var testDockerFile = newOptionDockerFile()
	var expectedDockerTag = "tag"
	var testDockerTag = newOptionDockerTag()
	var expectedDockerVersion = "version"
	var testDockerVersion = newOptionDockerVersion()
	var testExecutor = &BuildExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli: &Cli{
				optionDockerContext: testDockerContext,
				optionDockerFile:    testDockerFile,
				optionDockerTag:     testDockerTag,
				optionDockerVersion: testDockerVersion,
			},
		},
	}

	testLedger.Register(&cobra.Command{}, testDockerContext, testDockerFile, testDockerTag, testDockerVersion)
	testLedger.InitValue(testDockerContext, expectedDockerContext)
	testLedger.InitValue(testDockerFile, expectedDockerFile)
	testLedger.InitValue(testDockerTag, expectedDockerTag)
	testLedger.InitValue(testDockerVersion, expectedDockerVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testDockerContext.Param+`:[\s\t]*`+expectedDockerContext), actual)
	assert.Regexp(t, regexp.MustCompile(testDockerFile.Param+`:[\s\t]*`+expectedDockerFile), actual)
	assert.Regexp(t, regexp.MustCompile(testDockerTag.Param+`:[\s\t]*`+expectedDockerTag), actual)
	assert.Regexp(t, regexp.MustCompile(testDockerVersion.Param+`:[\s\t]*`+expectedDockerVersion), actual)
}

func TestBuildExecutor_Pretend(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledBuild = false
	var testExecutor = &BuildExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testExecutor.Pretend()
		assert.True(t, calledBuild)
	} else {
		assert.NoError(t, err)
	}
}

func TestBuildExecutor_Proceed(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledBuild = false
	var testExecutor = &BuildExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewBuild(t *testing.T) {
	var buildCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		Dry: func(ledger *config.Ledger) bool {
			return true
		},
		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testBuild = NewBuild(testLedger, testCli)

	testBuild.PostRun = func(cmd *cobra.Command, args []string) {
		buildCompleted = true
	}
	assert.NoError(t, testBuild.Execute())
	assert.True(t, buildCompleted)
	assert.NotEmpty(t, testOutput.String())
}

func Test_makeBuildParamsCwdContext(t *testing.T) {
	var expectedFolder = "/tmp/.genaiz"
	var cwd, _ = os.Getwd()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &BaseExecutor{
		Cli: &Cli{
			optionDockerContext: newOptionDockerContext(),
			optionDockerFile:    newOptionDockerFile(),
			optionDockerTag:     newOptionDockerTag(),
			optionDockerVersion: newOptionDockerVersion(),
		},
		Ledger: testLedger,
	}

	if fd, err := filez.CreateRecursive(expectedFolder, "Dockerfile"); err == nil {
		_ = os.Chdir(expectedFolder)
		defer func() {
			_ = os.Chdir(cwd)
			_ = os.Remove(fd.Name())
		}()

		testViper.Set(testExecutor.Cli.optionDockerContext.Key, expectedFolder)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, filepath.Join(expectedFolder, "Dockerfile"))
		testParam := makeBuildParams(testExecutor)
		assert.EqualValues(t, ".", testParam.DockerContext)
		assert.Empty(t, testParam.Dockerfile)
	}
}

func Test_makeBuildParamsCwdModule(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expectedContext = "/tmp/.genaiz"
	var expectedModule = "ModuleA"
	var expectedFile = "Dockerfile2"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &BaseExecutor{
		Cli: &Cli{
			optionDockerContext: newOptionDockerContext(),
			optionDockerFile:    newOptionDockerFile(),
			optionDockerTag:     newOptionDockerTag(),
			optionDockerVersion: newOptionDockerVersion(),
		},
		Ledger: testLedger,
	}

	if fd, err := filez.CreateRecursive(filepath.Join(expectedContext, expectedModule), expectedFile); err == nil {
		_ = os.Chdir(expectedContext)
		defer func() {
			_ = os.Chdir(cwd)
			_ = os.Remove(fd.Name())
		}()

		testViper.Set(testExecutor.Cli.optionDockerContext.Key, expectedContext)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fd.Name())
		testParam := makeBuildParams(testExecutor)
		assert.EqualValues(t, ".", testParam.DockerContext)
		assert.EqualValues(t, filepath.Join(expectedModule, expectedFile), testParam.Dockerfile)
	}
}

func Test_makeBuildParamsCwdNotStandard(t *testing.T) {
	var expectedFolder = "/tmp/.genaiz"
	var expectedFile = "Dockerfile3"
	var cwd, _ = os.Getwd()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &BaseExecutor{
		Cli: &Cli{
			optionDockerContext: newOptionDockerContext(),
			optionDockerFile:    newOptionDockerFile(),
			optionDockerTag:     newOptionDockerTag(),
			optionDockerVersion: newOptionDockerVersion(),
		},
		Ledger: testLedger,
	}

	if fd, err := filez.CreateRecursive(expectedFolder, expectedFile); err == nil {
		_ = os.Chdir(expectedFolder)
		defer func() {
			_ = os.Chdir(cwd)
			_ = os.Remove(fd.Name())
		}()

		testViper.Set(testExecutor.Cli.optionDockerContext.Key, expectedFolder)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, filepath.Join(expectedFolder, expectedFile))
		testParam := makeBuildParams(testExecutor)
		assert.EqualValues(t, ".", testParam.DockerContext)
		assert.EqualValues(t, expectedFile, testParam.Dockerfile)
	}
}

func Test_makeBuildParamsExternalContext(t *testing.T) {
	var expectedContext = "/tmp/.genaiz"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &BaseExecutor{
		Cli: &Cli{
			optionDockerContext: newOptionDockerContext(),
			optionDockerFile:    newOptionDockerFile(),
			optionDockerTag:     newOptionDockerTag(),
			optionDockerVersion: newOptionDockerVersion(),
		},
		Ledger: testLedger,
	}

	if fd, err := filez.CreateRecursive(expectedContext, "Dockerfile2"); err == nil {
		defer filez.RemoveSilently(expectedContext)

		testViper.Set(testExecutor.Cli.optionDockerContext.Key, expectedContext)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fd.Name())
		testParam := makeBuildParams(testExecutor)
		assert.EqualValues(t, expectedContext, testParam.DockerContext)
		assert.EqualValues(t, fd.Name(), testParam.Dockerfile)
	}
}

func newBuildTaskPretendStub(flag *bool) BuildTaskFactory {
	return func() *task.Task[docker.BuildParams] {
		return &task.Task[docker.BuildParams]{
			OnPrepare: func(params *docker.BuildParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *docker.BuildParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newBuildTaskCompleteStub(flag *bool) BuildTaskFactory {
	return func() *task.Task[docker.BuildParams] {
		return &task.Task[docker.BuildParams]{
			Name: "build_test",
			OnPrepare: func(params *docker.BuildParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *docker.BuildParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
