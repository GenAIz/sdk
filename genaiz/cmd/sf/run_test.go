package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

func TestRunExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = &RunOptions{
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Run.MountOutput).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Run.MountLog).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Run.MountVar).
			BuildStringOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Run.Image).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Run.Prefix).
			BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: testOptions,
	}
	var expectedDockerContext = "TestDockerContext"
	var expectedDockerFile = "TestDockerfile"
	var expectedDockerTag = "TestDockerTag"
	var expectedDockerVersion = "TestDockerVersion"
	var expectedMountInput = "TestMountInput"
	var expectedMountOutput = "TestMountOutput"
	var expectedMountLog = "TestMountLog"
	var expectedMountVar = "TestMountVar"
	var expectedRunImage = "TestRunImage"
	var expectedRunPrefix = "TestRunPrefix"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionMountInput.Key, expectedMountInput)
	testViper.Set(testOptions.optionMountOutput.Key, expectedMountOutput)
	testViper.Set(testOptions.optionMountLog.Key, expectedMountLog)
	testViper.Set(testOptions.optionMountVar.Key, expectedMountVar)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testViper.Set(testOptions.optionRunPrefix.Key, expectedRunPrefix)
	testExecutor.Display()

	if out := testOutput.String(); out != "" {
		assert.Contains(t, out, expectedDockerContext)
		assert.Contains(t, out, expectedDockerFile)
		assert.Contains(t, out, expectedDockerTag)
		assert.Contains(t, out, expectedDockerVersion)
		assert.Contains(t, out, expectedMountInput)
		assert.Contains(t, out, expectedMountOutput)
		assert.Contains(t, out, expectedMountLog)
		assert.Contains(t, out, expectedMountVar)
		assert.Contains(t, out, expectedRunImage)
		assert.Contains(t, out, expectedRunPrefix)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestRunExecutor_PretendRebuildImage(t *testing.T) {
	var calledBuild, calledRun bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Run.MountInput).
				BuildStringOption(),
			optionMountOutput: cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Run.MountOutput).
				BuildStringOption(),
			optionMountLog: cli.Options.Functions.MountLog().
				WithKeys(&schema.Genaiz.Function.Run.MountLog).
				BuildStringOption(),
			optionMountVar: cli.Options.Functions.MountVar().
				WithKeys(&schema.Genaiz.Function.Run.MountVar).
				BuildStringOption(),
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Run.Image).
				BuildStringOption(),
			optionRunPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Run.Prefix).
				BuildStringOption(),
			rebuildImage: true,
		},

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.True(t, calledRun)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Proceed(t *testing.T) {
	var calledBuild, calledRun bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Run.MountInput).
				BuildStringOption(),
			optionMountOutput: cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Run.MountOutput).
				BuildStringOption(),
			optionMountLog: cli.Options.Functions.MountLog().
				WithKeys(&schema.Genaiz.Function.Run.MountLog).
				BuildStringOption(),
			optionMountVar: cli.Options.Functions.MountVar().
				WithKeys(&schema.Genaiz.Function.Run.MountVar).
				BuildStringOption(),
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Run.Image).
				BuildStringOption(),
			optionRunPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Run.Prefix).
				BuildStringOption(),
			rebuildImage: true,
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		runTaskFactory:   newRunTaskCompleteStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledRun)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunOptions_allDefiners(t *testing.T) {
	var testOptions = &RunOptions{
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Run.MountOutput).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Run.MountLog).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Run.MountVar).
			BuildStringOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Run.Image).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Run.Prefix).
			BuildStringOption(),
	}
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, testOptions.optionMountInput)
	assert.Contains(t, definers, testOptions.optionMountOutput)
	assert.Contains(t, definers, testOptions.optionMountLog)
	assert.Contains(t, definers, testOptions.optionMountVar)
	assert.Contains(t, definers, testOptions.optionRunImage)
	assert.Contains(t, definers, testOptions.optionRunPrefix)
}

func TestNewRunOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewRunOptions(testCli)

	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
	assert.False(t, testOptions.rebuildImage)
}

func TestNewRun(t *testing.T) {
	var runCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testRun = NewRun(testLedger, testCli)
	var expectedTag = "dockerTag"

	testRun.PostRun = func(cmd *cobra.Command, args []string) {
		runCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, expectedTag)
	assert.NoError(t, testRun.Execute())
	assert.True(t, runCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedTag)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func newRunTaskPretendStub(flag *bool) RunTaskFactory {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *docker.ContainerParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newRunTaskCompleteStub(flag *bool) RunTaskFactory {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			Name: "run_test",
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *docker.ContainerParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
