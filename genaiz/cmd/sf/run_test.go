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
	"genaiz.com/genaiz-lib/mock"
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
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),
	}
	var expectedDockerContext = "RunDockerContext"
	var expectedDockerFile = "RunDockerfile"
	var expectedDockerTag = "RunDockerTag"
	var expectedDockerVersion = "RunDockerVersion"
	var expectedEnvFile = "RunEnvFile"
	var expectedEnvVars = "RunEnvVars"
	var expectedMountInput = "RunMountInput"
	var expectedMountOutput = "RunMountOutput"
	var expectedMountLog = "RunMountLog"
	var expectedMountVar = "RunMountVar"
	var expectedRunImage = "RunRunImage"
	var expectedRunPrefix = "RunRunPrefix"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testExecutor.optionEnvFile.Key, expectedEnvFile)
	testViper.Set(testExecutor.optionEnvVars.Key, expectedEnvVars)
	testViper.Set(testExecutor.optionMountInput.Key, expectedMountInput)
	testViper.Set(testExecutor.optionMountOutput.Key, expectedMountOutput)
	testViper.Set(testExecutor.optionMountLog.Key, expectedMountLog)
	testViper.Set(testExecutor.optionMountVar.Key, expectedMountVar)
	testViper.Set(testExecutor.optionRunImage.Key, expectedRunImage)
	testViper.Set(testExecutor.optionRunPrefix.Key, expectedRunPrefix)
	testExecutor.Display()

	if out := testOutput.String(); out != "" {
		assert.Contains(t, out, expectedDockerContext)
		assert.Contains(t, out, expectedDockerFile)
		assert.Contains(t, out, expectedDockerTag)
		assert.Contains(t, out, expectedDockerVersion)
		assert.Contains(t, out, expectedEnvFile)
		assert.Contains(t, out, expectedEnvVars)
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

func TestRunExecutor_Pretend_EnvMapError(t *testing.T) {
	var calledBuild, calledRun bool
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testExecutor.Pretend()
			assert.False(t, calledBuild)
			assert.False(t, calledRun)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestRunExecutor_Pretend_RebuildImage(t *testing.T) {
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
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
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
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
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
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		runTaskFactory:   newRunTaskCompleteStub(&calledRun),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledRun)
	} else {
		assert.NoError(t, err)
	}
}

func TestRunExecutor_Proceed_EnvMapError(t *testing.T) {
	var calledBuild, calledRun bool
	var testFile = filepath.Join(t.TempDir(), ".env-test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &RunExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		RunOptions: newRunTestOptions(),

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		runTaskFactory:   newRunTaskPretendStub(&calledRun),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testExecutor.Proceed()
			assert.False(t, calledBuild)
			assert.False(t, calledRun)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestRunOptions_allDefiners(t *testing.T) {
	var testOptions = newRunTestOptions()
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, testOptions.optionEnvFile)
	assert.Contains(t, definers, testOptions.optionEnvVars)
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

	assert.NotEmpty(t, testOptions.optionEnvFile)
	assert.NotEmpty(t, testOptions.optionEnvVars)
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

func newRunTestOptions() *RunOptions {
	return &RunOptions{
		EnvOptions: EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Genaiz.Function.Run.EnvFile).
				BuildStringOption(),
			optionEnvVars: cli.Options.Docker.EnvVar().
				WithKeys(&schema.Genaiz.Function.Run.EnvVars).
				BuildListOption(),
		},
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
}
