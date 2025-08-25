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
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

func TestStartExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = &StartOptions{
		RunOptions: &RunOptions{
			optionMountInput:  newOptionMountInput("_test", false),
			optionMountOutput: newOptionMountOutput("_test", false),
			optionMountLog:    newOptionMountLog("_test", nil),
			optionMountVar:    newOptionMountVar("_test", nil),
			optionRunImage:    newOptionCmdImage("_test"),
		},
		StopOptions: &StopOptions{
			optionContainerName:   newOptionContainerName("_test"),
			optionContainerPrefix: newOptionContainerPrefix("_test", testCli),
		},
		optionContainerReplace: newOptionContainerReplace(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: testOptions,
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
	var expectedContainerName = "TestContainerName"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionMountInput.Key, expectedMountInput)
	testViper.Set(testOptions.optionMountOutput.Key, expectedMountOutput)
	testViper.Set(testOptions.optionMountLog.Key, expectedMountLog)
	testViper.Set(testOptions.optionMountVar.Key, expectedMountVar)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testViper.Set(testOptions.optionContainerReplace.Key, true)
	testViper.Set(testOptions.optionContainerName.Key, expectedContainerName)
	testViper.Set(testOptions.optionContainerPrefix.Key, expectedRunPrefix)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerTag)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedMountInput)
		assert.Contains(t, actual, expectedMountOutput)
		assert.Contains(t, actual, expectedMountLog)
		assert.Contains(t, actual, expectedMountVar)
		assert.Contains(t, actual, expectedRunImage)
		assert.Contains(t, actual, expectedRunPrefix)
		assert.Contains(t, actual, expectedContainerName)
		assert.Regexp(t, regexp.MustCompile(testOptions.optionContainerReplace.Param+`:[\s\t]*true`), actual)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestStartExecutor_PretendNoDispose(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 0, calledDispose)
		assert.EqualValues(t, 1, calledStart)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_PretendNoPreserve(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
		stopTaskFactory:      newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, false)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
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

func TestStartExecutor_PretendReplace(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		containerTaskFactory: newContainerTaskPretendStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskPretendStub(&calledDispose),
		startTaskFactory:     newContainerTaskPretendStub(&calledStart),
		stopTaskFactory:      newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, true)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 2, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_ProceedNoDispose(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 0, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 0, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartExecutor_ProceedNoPreserve(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, false)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
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

func TestStartExecutor_ProceedReplace(t *testing.T) {
	var calledBuild bool
	var calledCreate, calledDispose, calledStart, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptionOutput = newOptionMountOutput("_test", false)
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StartExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StartOptions: &StartOptions{
			RunOptions: &RunOptions{
				optionMountInput:  newOptionMountInput("_test", false),
				optionMountOutput: testOptionOutput,
				optionMountLog:    newOptionMountLog("_test", testOptionOutput),
				optionMountVar:    newOptionMountVar("_test", testOptionOutput),
				optionRunImage:    newOptionCmdImage("_test"),
				optionRunPrefix:   newOptionContainerPrefix("test", testCli),
				rebuildImage:      true,
			},
			StopOptions: &StopOptions{
				optionContainerName:     newOptionContainerName("_test"),
				optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
				optionContainerPreserve: newOptionContainerPreserve(true),
			},
			optionContainerReplace: newOptionContainerReplace(),
		},

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		containerTaskFactory: newContainerTaskCompleteStub(&calledCreate),
		disposeTaskFactory:   newContainerTaskCompleteStub(&calledDispose),
		startTaskFactory:     newContainerTaskCompleteStub(&calledStart),
		stopTaskFactory:      newContainerTaskCompleteStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerReplace.Key, true)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledCreate)
		assert.EqualValues(t, 2, calledDispose)
		assert.EqualValues(t, 1, calledStart)
		assert.EqualValues(t, 1, calledStop)
	} else {
		assert.NoError(t, err)
	}
}

func TestStartOptions_allDefiners(t *testing.T) {
	var expectedOptionMountInput = newOptionMountInput("_test", false)
	var expectedOptionMountOutput = newOptionMountOutput("_test", false)
	var expectedOptionMountLog = newOptionMountLog("_test", expectedOptionMountOutput)
	var expectedOptionMountVar = newOptionMountVar("_test", expectedOptionMountOutput)
	var expectedOptionRunImage = newOptionCmdImage("_test")
	var expectedOptionContainerName = newOptionContainerName("_test")
	var expectedOptionContainerPrefix = newOptionContainerPrefix("_test", NewSfCli(nil, nil, nil))
	var expectedOptionContainerReplace = newOptionContainerReplace()
	var testOptions = &StartOptions{
		RunOptions: &RunOptions{
			optionMountInput:  expectedOptionMountInput,
			optionMountOutput: expectedOptionMountOutput,
			optionMountLog:    expectedOptionMountLog,
			optionMountVar:    expectedOptionMountVar,
			optionRunImage:    expectedOptionRunImage,
		},
		StopOptions: &StopOptions{
			optionContainerName:   expectedOptionContainerName,
			optionContainerPrefix: expectedOptionContainerPrefix,
		},
		optionContainerReplace: expectedOptionContainerReplace,
	}
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, expectedOptionMountInput)
	assert.Contains(t, definers, expectedOptionMountOutput)
	assert.Contains(t, definers, expectedOptionMountLog)
	assert.Contains(t, definers, expectedOptionMountVar)
	assert.Contains(t, definers, expectedOptionRunImage)
	assert.Contains(t, definers, expectedOptionContainerName)
	assert.Contains(t, definers, expectedOptionContainerPrefix)
	assert.Contains(t, definers, expectedOptionContainerReplace)
}

func TestNewStartOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewStartOptions(testCli)

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

func TestNewStart(t *testing.T) {
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
		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testStart = NewStart(testLedger, testCli)
	var testParam = newOptionMountInput("_test", false).Param
	var expectedTag = "dockerTag"
	var expectedFolder = "folder"
	var expectedWorkDir = "work"

	testStart.PostRun = func(cmd *cobra.Command, args []string) {
		startCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, expectedTag)
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
