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
)

func TestDebugExecutor_Display(t *testing.T) {
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
	var testMountOption = newOptionMountOutput("_test", false)
	var testOptions = &RunOptions{
		optionMountInput:  newOptionMountInput("_test", false),
		optionMountOutput: testMountOption,
		optionMountLog:    newOptionMountLog("_test", testMountOption),
		optionMountVar:    newOptionMountVar("_test", testMountOption),
		optionRunImage:    newOptionCmdImage("_test"),
		optionRunPrefix:   newOptionContainerPrefix("_test", testCli),
	}
	var testExecutor = &DebugExecutor{
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
	var expectedRunImage = "TestRunImage"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerTag)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedRunImage)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestDebugExecutor_Pretend(t *testing.T) {
	var calledBuild bool
	var calledDebug int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testMountOptions = newOptionMountOutput("_test", false)
	var testExecutor = &DebugExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
			optionMountInput:  newOptionMountInput("_test", false),
			optionMountOutput: testMountOptions,
			optionMountLog:    newOptionMountLog("_test", testMountOptions),
			optionMountVar:    newOptionMountVar("_test", testMountOptions),
			optionRunImage:    newOptionCmdImage("_test"),
			optionRunPrefix:   newOptionContainerPrefix("_test", testCli),
		},

		buildTaskFactory: newBuildTaskPretendStub(&calledBuild),
		debugTaskFactory: newContainerTaskPretendStub(&calledDebug),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledDebug)
	} else {
		assert.NoError(t, err)
	}
}

func TestDebugExecutor_Proceed(t *testing.T) {
	var calledBuild bool
	var calledDebug int
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testMountOptions = newOptionMountOutput("_test", false)
	var testExecutor = &DebugExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		RunOptions: &RunOptions{
			optionMountInput:  newOptionMountInput("_test", false),
			optionMountOutput: testMountOptions,
			optionMountLog:    newOptionMountLog("_test", testMountOptions),
			optionMountVar:    newOptionMountVar("_test", testMountOptions),
			optionRunImage:    newOptionCmdImage("_test"),
			optionRunPrefix:   newOptionContainerPrefix("_test", testCli),
			rebuildImage:      true,
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		debugTaskFactory: newContainerTaskCompleteStub(&calledDebug),
	}

	if fd, err := os.Create(filepath.Join(testDir, "genaizDockerfile")); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledDebug)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewDebugOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewDebugOptions(testCli)

	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
}

func TestNewDebug(t *testing.T) {
	var debugCompleted = false
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
	var testDebug = NewDebug(testLedger, testCli)
	var testCmdImageOption = newOptionCmdImage("Debug")
	var expectedImage = "dockerImage"

	testDebug.PostRun = func(cmd *cobra.Command, args []string) {
		debugCompleted = true
	}

	testViper.Set(testCli.optionDockerTag.Key, "tag/tag")
	testViper.Set(testCmdImageOption.Key, expectedImage)
	assert.NoError(t, testDebug.Execute())
	assert.True(t, debugCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}
