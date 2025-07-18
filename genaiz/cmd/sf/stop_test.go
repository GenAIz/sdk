package sf

import (
	"bytes"
	"io"
	"os"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
)

func TestStopExecutor_Display(t *testing.T) {
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
	var testOptions = &StopOptions{
		RunOptions: &RunOptions{
			optionRunImage: newOptionCmdImage("_test"),
		},
		optionContainerName:     newOptionContainerName("_test"),
		optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
		optionContainerPreserve: newOptionContainerPreserve(false),
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
	var expectedDockerTag = "TestDockerTag"
	var expectedDockerVersion = "TestDockerVersion"
	var expectedRunImage = "TestRunImage"
	var expectedRunPrefix = "TestRunPrefix"
	var expectedContainerName = "TestContainerName"

	testViper.Set(testCli.optionDockerContext.Key, expectedDockerContext)
	testViper.Set(testCli.optionDockerFile.Key, expectedDockerFile)
	testViper.Set(testCli.optionDockerTag.Key, expectedDockerTag)
	testViper.Set(testCli.optionDockerVersion.Key, expectedDockerVersion)
	testViper.Set(testOptions.optionRunImage.Key, expectedRunImage)
	testViper.Set(testOptions.optionContainerName.Key, expectedContainerName)
	testViper.Set(testOptions.optionContainerPrefix.Key, expectedRunPrefix)
	testViper.Set(testOptions.optionContainerPreserve.Key, true)
	testExecutor.Display()

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedDockerContext)
		assert.Contains(t, actual, expectedDockerFile)
		assert.Contains(t, actual, expectedDockerTag)
		assert.Contains(t, actual, expectedDockerVersion)
		assert.Contains(t, actual, expectedRunImage)
		assert.Contains(t, actual, expectedRunPrefix)
		assert.Contains(t, actual, expectedContainerName)
		assert.Regexp(t, regexp.MustCompile(testOptions.optionContainerPreserve.Param+`:[\s\t]*true`), actual)
	} else {
		assert.Fail(t, "output is empty")
	}
}

func TestStopExecutor_Pretend(t *testing.T) {
	var calledDispose, calledStop int
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: newOptionCmdImage("_test"),
			},
			optionContainerName:     newOptionContainerName("_test"),
			optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
			optionContainerPreserve: newOptionContainerPreserve(false),
		},

		disposeTaskFactory: newContainerTaskPretendStub(&calledDispose),
		stopTaskFactory:    newContainerTaskPretendStub(&calledStop),
	}

	testViper.Set(testExecutor.optionContainerPreserve.Key, true)

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
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
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: newOptionCmdImage("_test"),
			},
			optionContainerName:     newOptionContainerName("_test"),
			optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
			optionContainerPreserve: newOptionContainerPreserve(false),
		},

		disposeTaskFactory: newContainerTaskPretendStub(&calledDispose),
		stopTaskFactory:    newContainerTaskPretendStub(&calledStop),
	}

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
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
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{
		optionDockerFile:    newOptionDockerFile(),
		optionDockerContext: newOptionDockerContext(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testExecutor = &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		StopOptions: &StopOptions{
			RunOptions: &RunOptions{
				optionRunImage: newOptionCmdImage("_test"),
			},
			optionContainerName:     newOptionContainerName("_test"),
			optionContainerPrefix:   newOptionContainerPrefix("_test", testCli),
			optionContainerPreserve: newOptionContainerPreserve(false),
		},

		disposeTaskFactory: newContainerTaskCompleteStub(&calledDispose),
		stopTaskFactory:    newContainerTaskCompleteStub(&calledStop),
	}

	if fd, err := os.CreateTemp("/tmp", "genaizDockerfile"); err == nil {
		defer filez.RemoveSilently(fd.Name())
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
	var expectedOptionRunImage = newOptionCmdImage("_test")
	var expectedOptionContainerName = newOptionContainerName("_test")
	var expectedOptionContainerPrefix = newOptionContainerPrefix("_test", NewSfCli(nil, nil, nil))
	var expectedOptionContainerPreserve = newOptionContainerPreserve(false)
	var testOptions = &StopOptions{
		RunOptions: &RunOptions{
			optionRunImage: expectedOptionRunImage,
		},
		optionContainerName:     expectedOptionContainerName,
		optionContainerPrefix:   expectedOptionContainerPrefix,
		optionContainerPreserve: expectedOptionContainerPreserve,
	}
	var definers = testOptions.allDefiners()

	assert.Contains(t, definers, expectedOptionRunImage)
	assert.Contains(t, definers, expectedOptionContainerName)
	assert.Contains(t, definers, expectedOptionContainerPrefix)
	assert.Contains(t, definers, expectedOptionContainerPreserve)
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
	var testStop = NewStop(testLedger, testCli)
	var testCmdImageOption = newOptionCmdImage("Stop")
	var expectedImage = "dockerImage"

	testStop.PostRun = func(cmd *cobra.Command, args []string) {
		stopCompleted = true
	}

	testViper.Set(testCmdImageOption.Key, expectedImage)
	assert.NoError(t, testStop.Execute())
	assert.True(t, stopCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedImage)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func Test_newOptionContainerPrefix_DefaultWorkSpace(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionContainerPrefix("_test", testCli)
	var testWorkspace = &config.StringOption{
		Option: config.Option{
			Param: "workspace",
		},
	}
	var expectedTag = "tag"
	var expectedWorkspace = "workspace"

	testViper.Set(testCli.optionDockerTag.Key, expectedTag)
	testViper.Set(testWorkspace.Param, expectedWorkspace)
	testLedger.InitWorkspace(testWorkspace)
	assert.EqualValues(t, expectedWorkspace+"-"+expectedTag, testOption.DefaultGetter(testLedger))
}
