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
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

func TestStartExecutor_Display(t *testing.T) {
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
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),
	}
	var expectedDockerContext = "StartDockerContext"
	var expectedDockerFile = "StartDockerfile"
	var expectedDockerTag = "StartDockerTag"
	var expectedDockerVersion = "StartDockerVersion"
	var expectedEnvFile = "StartEnvFile"
	var expectedEnvVars = "StartEnvVars"
	var expectedMountInput = "StartMountInput"
	var expectedMountOutput = "StartMountOutput"
	var expectedMountLog = "StartMountLog"
	var expectedMountVar = "StartMountVar"
	var expectedRunImage = "StartRunImage"
	var expectedContainerPrefix = "StartContainerPrefix"
	var expectedContainerName = "StartContainerName"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testExecutor.optionMountInput.Key, expectedMountInput)
	testViper.Set(testExecutor.optionMountOutput.Key, expectedMountOutput)
	testViper.Set(testExecutor.optionMountLog.Key, expectedMountLog)
	testViper.Set(testExecutor.optionMountVar.Key, expectedMountVar)
	testViper.Set(testExecutor.optionRunImage.Key, expectedRunImage)
	testViper.Set(testExecutor.optionContainerReplace.Key, true)
	testViper.Set(testExecutor.optionContainerName.Key, expectedContainerName)
	testViper.Set(testExecutor.optionContainerPrefix.Key, expectedContainerPrefix)
	testViper.Set(testExecutor.optionEnvFile.Key, expectedEnvFile)
	testViper.Set(testExecutor.optionEnvVars.Key, expectedEnvVars)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerTag)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedEnvFile)
		assert.Contains(t, actual, expectedEnvVars)
		assert.Contains(t, actual, expectedMountInput)
		assert.Contains(t, actual, expectedMountOutput)
		assert.Contains(t, actual, expectedMountLog)
		assert.Contains(t, actual, expectedMountVar)
		assert.Contains(t, actual, expectedRunImage)
		assert.Contains(t, actual, expectedContainerPrefix)
		assert.Contains(t, actual, expectedContainerName)
		assert.Regexp(t, regexp.MustCompile(testExecutor.optionContainerReplace.Param+`:[\s\t]*true`), actual)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestStartExecutor_Pretend_EnvMapError(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart int
	var testFile = filepath.Join(t.TempDir(), ".env-start")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
	}
	var fd *os.File
	var err error

	testViper.Set(testExecutor.optionEnvFile.Key, testFile)

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testExecutor.Pretend()
			assert.False(t, calledBuild)
			assert.EqualValues(t, 0, calledCreate)
			assert.EqualValues(t, 0, calledDispose)
			assert.EqualValues(t, 0, calledStart)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestStartExecutor_Pretend_NoDispose(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 0, calledDispose)
		assert.EqualValues(t, 1, calledStart)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_Pretend_NoPreserve(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
		stopTaskFactory:      newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, false)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 1, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_Pretend_Replace(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
		stopTaskFactory:      newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, true)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 2, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_Proceed_EnvMapError(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart int
	var testFile = filepath.Join(t.TempDir(), ".env-start")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
	}
	var fd *os.File
	var err error

	testViper.Set(testExecutor.optionEnvFile.Key, testFile)

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("syntax_error")); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			testExecutor.Proceed()
			assert.False(t, calledBuild)
			assert.EqualValues(t, 0, calledCreate)
			assert.EqualValues(t, 0, calledDispose)
			assert.EqualValues(t, 0, calledStart)
			assert.True(t, patch.Called)
			assert.EqualValues(t, 1, patch.CalledWith)
			return
		}
	}

	assert.NoError(t, err)
}

func TestStartExecutor_Proceed_NoDispose(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 0, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 0, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_Proceed_NoPreserve(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, false)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.rebuildImage = true
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 1, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_Proceed_Replace(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: newStartTestOptions(),

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, true)

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 2, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartOptions_allDefiners(t *testing.T) {
	var testOptions = newStartTestOptions()
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, testOptions.optionEnvFile)
	assert.Contains(t, definers, testOptions.optionEnvVars)
	assert.Contains(t, definers, testOptions.optionMountInput)
	assert.Contains(t, definers, testOptions.optionMountOutput)
	assert.Contains(t, definers, testOptions.optionMountLog)
	assert.Contains(t, definers, testOptions.optionMountVar)
	assert.Contains(t, definers, testOptions.optionRunImage)
	assert.Contains(t, definers, testOptions.optionContainerPreserve)
	assert.Contains(t, definers, testOptions.optionContainerPrefix)
	assert.Contains(t, definers, testOptions.optionContainerName)
	assert.Contains(t, definers, testOptions.optionContainerReplace)
}

func TestNewStartOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStartOptions(testCli)

	assert.NotEmpty(t, testOptions.optionEnvFile)
	assert.NotEmpty(t, testOptions.optionEnvVars)
	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionContainerReplace)
	assert.NotEmpty(t, testOptions.optionContainerPreserve)
	assert.NotEmpty(t, testOptions.optionContainerPrefix)
	assert.NotEmpty(t, testOptions.optionContainerName)
	assert.False(t, testOptions.rebuildImage)
}

func TestNewStartOptions_DefaultRunMounts(t *testing.T) {
	var testInputDir = t.TempDir()
	var testOutputDir = t.TempDir()
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStartOptions(testCli)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	testViper.Set(schema.Genaiz.Function.Run.MountInput.Doc, testInputDir)
	testViper.Set(schema.Genaiz.Function.Run.MountOutput.Doc, testOutputDir)

	assert.Equal(t, testInputDir, testLedger.GetString(testOptions.optionMountInput))
	assert.Equal(t, testOutputDir, testLedger.GetString(testOptions.optionMountOutput))
	assert.Equal(t, testOutputDir, testLedger.GetString(testOptions.optionMountLog))
	assert.Equal(t, testOutputDir, testLedger.GetString(testOptions.optionMountVar))
}

func TestNewStart(t *testing.T) {
	var testDir = t.TempDir()
	var startCompleted = false
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
	var testStart = NewStart(testLedger, testCli)
	var testParam = cli.Options.Functions.MountInput().BuildStringOption().Param
	var expectedTag = "dockerTag"
	var expectedFolder = "folder"
	var expectedWorkDir = "work"

	testStart.PostRun = func(cmd *cobra.Command, args []string) {
		startCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, expectedTag)
	testViper.Set(schema.Genaiz.Function.Start.MountOutput.Doc, testDir)
	assert.NoError(t, testStart.PersistentFlags().Lookup(testParam).Value.Set(expectedFolder))
	testLedger.WorkDir = expectedWorkDir
	assert.NoError(t, testStart.Execute())
	assert.True(t, startCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.EqualValues(t, filepath.Join(expectedWorkDir, expectedFolder), testStart.PersistentFlags().Lookup(testParam).Value.String())
		assert.Contains(t, actual, expectedTag)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func newContainerTaskPretendStub(counter *int) func() *task.Task[docker.ContainerParams] {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *docker.ContainerParams, state *task.State) error {
				*counter++
				return nil
			},
		}
	}
}

func newContainerTaskCompleteStub(counter *int) func() *task.Task[docker.ContainerParams] {
	return func() *task.Task[docker.ContainerParams] {
		return &task.Task[docker.ContainerParams]{
			Name: "build_test",
			OnPrepare: func(params *docker.ContainerParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *docker.ContainerParams, state *task.State) error {
				*counter++
				return nil
			},
		}
	}
}

func newStartTestOptions() *StartOptions {
	return &StartOptions{
		RunOptions: &RunOptions{
			EnvOptions: EnvOptions{
				optionEnvFile: cli.Options.Docker.EnvFile().
					WithKeys(&schema.Genaiz.Function.Start.EnvFile).
					BuildStringOption(),
				optionEnvVars: cli.Options.Docker.EnvVar().
					WithKeys(&schema.Genaiz.Function.Start.EnvVars).
					BuildListOption(),
			},
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Start.MountInput).
				BuildStringOption(),
			optionMountOutput: cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Start.MountOutput).
				BuildStringOption(),
			optionMountLog: cli.Options.Functions.MountLog().
				WithKeys(&schema.Genaiz.Function.Start.MountLog).
				BuildStringOption(),
			optionMountVar: cli.Options.Functions.MountVar().
				WithKeys(&schema.Genaiz.Function.Start.MountVar).
				BuildStringOption(),
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Start.Image).
				BuildStringOption(),
		},
		StopOptions: &StopOptions{
			optionContainerName: cli.Options.Docker.ContainerName().
				WithKeys(&schema.Genaiz.Function.Start.Name).
				BuildStringOption(),
			optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Start.Prefix).
				BuildStringOption(),
			optionContainerPreserve: cli.Options.Docker.Preserve().
				WithKeys(&schema.Genaiz.Function.Start.Preserve).
				BuildBoolOption(),
		},
		optionContainerReplace: cli.Options.Docker.Replace().
			BuildBoolOption(),
	}
}
