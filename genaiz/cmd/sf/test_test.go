package sf

import (
	"bytes"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/filez"
)

func TestTestExecutor_Display(t *testing.T) {
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
	var testMountOption = newOptionMountOutput("_test", false)
	var testOptions = &RunOptions{
		optionMountInput:  newOptionMountInput("_test", false),
		optionMountOutput: testMountOption,
		optionMountLog:    newOptionMountLog("_test", testMountOption),
		optionMountVar:    newOptionMountVar("_test", testMountOption),
		optionRunImage:    newOptionCmdImage("_test"),
		optionRunPrefix:   newOptionContainerPrefix("_test", testCli),
	}
	var testExecutor = &TestExecutor{
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

func TestTestExecutor_Pretend(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testMountOptions = newOptionMountOutput("_test", false)
	var testExecutor = &TestExecutor{
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
		testTaskFactory:  newContainerTaskPretendStub(&calledTest),
	}

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
	} else {
		assert.NoError(t, err)
	}
}

func TestTestExecutor_Proceed(t *testing.T) {
	var calledBuild bool
	var calledTest int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testMountOptions = newOptionMountOutput("_test", false)
	var testExecutor = &TestExecutor{
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
		testTaskFactory:  newContainerTaskCompleteStub(&calledTest),
	}

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.EqualValues(t, 1, calledTest)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewTestOptions(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testOptions = NewTestOptions(testCli)

	assert.NotEmpty(t, testOptions.optionMountInput)
	assert.NotEmpty(t, testOptions.optionMountOutput)
	assert.NotEmpty(t, testOptions.optionMountLog)
	assert.NotEmpty(t, testOptions.optionMountVar)
	assert.NotEmpty(t, testOptions.optionRunImage)
	assert.NotEmpty(t, testOptions.optionRunPrefix)
}

func TestNewTest(t *testing.T) {
	var testCompleted = false
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
	var testTest = NewTest(testLedger, testCli)
	var testCmdImageOption = newOptionCmdImage("Test")
	var expectedImage = "dockerImage"

	testTest.PostRun = func(cmd *cobra.Command, args []string) {
		testCompleted = true
	}

	testViper.Set(testCmdImageOption.Key, expectedImage)
	assert.NoError(t, testTest.Execute())
	assert.True(t, testCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}
