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

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
)

func TestStopExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = &StopOptions{
		RunOptions: &RunOptions{
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Run.Image).
				BuildStringOption(),
		},
		optionContainerName: cli.Options.Docker.ContainerName().
			WithKeys(&schema.Genaiz.Function.Stop.Name).
			BuildStringOption(),
		optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Stop.Prefix).
			BuildStringOption(),
		optionContainerPreserve: cli.Options.Docker.Preserve().
			WithKeys(&schema.Genaiz.Function.Stop.Preserve).
			BuildBoolOption(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: testOptions,
	}
	var expectedDockerContext = "TestDockerContext"
	var expectedDockerFile = "TestDockerfile"
	var expectedDockerRepo = "TestDockerRepo"
	var expectedDockerVersion = "TestDockerVersion"
	var expectedRunImage = "TestRunImage"
	var expectedContainerPrefix = "TestContainerPrefix"
	var expectedContainerName = "TestContainerName"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerRepo.Key, expectedDockerRepo)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testViper.Set(testOptions.optionContainerName.Key, expectedContainerName)
	testViper.Set(testOptions.optionContainerPrefix.Key, expectedContainerPrefix)
	testViper.Set(testOptions.optionContainerPreserve.Key, true)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerRepo)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedRunImage)
		assert.Contains(t, actual, expectedContainerPrefix)
		assert.Contains(t, actual, expectedContainerName)
		assert.Regexp(t, regexp.MustCompile(testOptions.optionContainerPreserve.Param+`:[\s\t]*true`), actual)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestStopExecutor_Pretend(t *testing.T) {
	var calledDispose, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: cli.Options.Docker.Image().
					WithKeys(&schema.Genaiz.Function.Run.Image).
					BuildStringOption(),
			},
			optionContainerName: cli.Options.Docker.ContainerName().
				WithKeys(&schema.Genaiz.Function.Stop.Name).
				BuildStringOption(),
			optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Stop.Prefix).
				BuildStringOption(),
			optionContainerPreserve: cli.Options.Docker.Preserve().
				WithKeys(&schema.Genaiz.Function.Stop.Preserve).
				BuildBoolOption(),
		},

		disposeTaskFactory: newContainerTaskPretendStub(&calledDispose),
		stopTaskFactory:    newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.EqualValues(t, 1, calledStop)
		assert.EqualValues(t, 0, calledDispose)
	} else {
		assert.NoError(t, err)
	}
}

func TestStopExecutor_PretendDispose(t *testing.T) {
	var calledDispose, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: cli.Options.Docker.Image().
					WithKeys(&schema.Genaiz.Function.Run.Image).
					BuildStringOption(),
			},
			optionContainerName: cli.Options.Docker.ContainerName().
				WithKeys(&schema.Genaiz.Function.Stop.Name).
				BuildStringOption(),
			optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Stop.Prefix).
				BuildStringOption(),
			optionContainerPreserve: cli.Options.Docker.Preserve().
				WithKeys(&schema.Genaiz.Function.Stop.Preserve).
				BuildBoolOption(),
		},

		disposeTaskFactory: newContainerTaskPretendStub(&calledDispose),
		stopTaskFactory:    newContainerTaskPretendStub(&calledStop),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.EqualValues(t, 0, calledStop)
		assert.EqualValues(t, 1, calledDispose)
	} else {
		assert.NoError(t, err)
	}
}

func TestStopExecutor_Proceed(t *testing.T) {
	var calledDispose, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: cli.Options.Docker.Image().
					WithKeys(&schema.Genaiz.Function.Run.Image).
					BuildStringOption(),
			},
			optionContainerName: cli.Options.Docker.ContainerName().
				WithKeys(&schema.Genaiz.Function.Stop.Name).
				BuildStringOption(),
			optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Stop.Prefix).
				BuildStringOption(),
			optionContainerPreserve: cli.Options.Docker.Preserve().
				WithKeys(&schema.Genaiz.Function.Stop.Preserve).
				BuildBoolOption(),
		},

		disposeTaskFactory: newContainerTaskCompleteStub(&calledDispose),
		stopTaskFactory:    newContainerTaskCompleteStub(&calledStop),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.EqualValues(t, 0, calledStop)
		assert.EqualValues(t, 1, calledDispose)
	} else {
		assert.NoError(t, err)
	}
}

func TestStopExecutor_allDefiners(t *testing.T) {
	var testOptions = &StopOptions{
		RunOptions: &RunOptions{
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Run.Image).
				BuildStringOption(),
		},
		optionContainerName: cli.Options.Docker.ContainerName().
			WithKeys(&schema.Genaiz.Function.Stop.Name).
			BuildStringOption(),
		optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Stop.Prefix).
			BuildStringOption(),
		optionContainerPreserve: cli.Options.Docker.Preserve().
			WithKeys(&schema.Genaiz.Function.Stop.Preserve).
			BuildBoolOption(),
	}
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, testOptions.optionRunImage)
	assert.Contains(t, definers, testOptions.optionContainerName)
	assert.Contains(t, definers, testOptions.optionContainerPrefix)
	assert.Contains(t, definers, testOptions.optionContainerPreserve)
}

func TestNewStopOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStopOptions(testCli)

	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionContainerPreserve)
	assert.NotEmpty(t, testOptions.optionContainerPrefix)
	assert.NotEmpty(t, testOptions.optionContainerName)
}

func TestNewStop(t *testing.T) {
	var stopCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testStop = NewStop(testLedger, testCli)
	var testCmdImageOption = cli.Options.Docker.Image().
		WithKeys(&schema.Genaiz.Function.Stop.Image).
		BuildStringOption()
	var expectedImage = "dockerImage"

	testStop.PostRun = func(cmd *cobra.Command, args []string) {
		stopCompleted = true
	}

	testViper.Set(testCli.optionDockerRepo.Key, "namespace/repo")
	testViper.Set(testCmdImageOption.Key, expectedImage)
	assert.NoError(t, testStop.Execute())
	assert.True(t, stopCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}
