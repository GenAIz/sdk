package dk

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestPublishExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testFolder = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		WithUserPath(testFolder).
		Build()
	var expectedFolder = filepath.Join(testFolder, ".config", "genaiz")
	var expectedHandle = "expectedHandle"
	var expectedOem = "expectedOem"
	var expectedVersion = "expectedVersion"
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: NewPublishOptions(),
	}

	testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testExecutor.optionUserDefined.Key, "True")
	testViper.Set(testExecutor.optionHandle.Key, expectedHandle)
	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`folder:[\s\t]*`+expectedFolder), actual)
}

func TestPublishExecutor_Pretend(t *testing.T) {
	var calledParams broker.DataLinkParams
	var expectedOem = "oem"
	var expectedVersion = "1.0.0"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: NewPublishOptions(),

		linkArgument: "handle",

		publishLinkTaskFactory: newPublishLinkTaskPretendCapture(&calledParams),
		dataLinksWriterFactory: newDataLinksWriterFactory(nil),
	}

	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.Equal(t, testExecutor.linkArgument, calledParams.Handle)
	assert.Equal(t, expectedOem, calledParams.Oem)
	assert.Equal(t, expectedVersion, calledParams.Version)
}

func TestPublishExecutor_Pretend_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: NewPublishOptions(),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPublishExecutor_Proceed(t *testing.T) {
	var calledParams broker.DataLinkParams
	var expectedOem = "oem"
	var expectedVersion = "1.0.0"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: NewPublishOptions(),

		linkArgument: "handle",

		publishLinkTaskFactory: newPublishLinkTaskCompleteCapture(&calledParams),
		dataLinksWriterFactory: newDataLinksWriterFactory(nil),
	}

	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, testExecutor.linkArgument, calledParams.Handle)
	assert.Equal(t, expectedOem, calledParams.Oem)
	assert.Equal(t, expectedVersion, calledParams.Version)
}

func TestPublishExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: NewPublishOptions(),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "notValid")
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewPublish(t *testing.T) {
	var publishCompleted = false
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testPublish = NewPublish(testLedger, &Cli{
		BaseCli: cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
			return true
		}},
	})

	testPublish.PostRun = func(cmd *cobra.Command, args []string) {
		publishCompleted = true
	}
	testPublish.SetArgs([]string{"handle"})
	assert.NoError(t, testPublish.Execute())
	assert.True(t, publishCompleted)
}

func TestNewPublish_ConfigFolderOverride(t *testing.T) {
	var configFolder = t.TempDir()
	var publishCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testPublish = NewPublish(testLedger, &Cli{
		BaseCli: cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
			return true
		}},
	})

	testPublish.PostRun = func(cmd *cobra.Command, args []string) {
		publishCompleted = true
	}
	testPublish.SetArgs([]string{"handle", configFolder})
	assert.NoError(t, testPublish.Execute())
	assert.True(t, publishCompleted)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`folder:[\s\t]*`+configFolder), actual)
}

func TestNewPublish_WorkDirInvalid(t *testing.T) {
	if runtime.GOOS == "linux" {
		var configFolder = filepath.Join(t.TempDir(), "remove-able")
		var testViper = viper.New()
		var testLedger = config.NewBuilder().WithViper(testViper).Build()
		var testPublish = NewPublish(testLedger, &Cli{
			BaseCli: cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
				return true
			}},
		})
		var err error

		if err = os.MkdirAll(configFolder, 0750); err == nil {
			var patch = mock.Patches{T: t}.OsExit(func(int) {})

			defer patch.Unpatch()
			t.Chdir(configFolder)
			testPublish.PreRun = func(cmd *cobra.Command, args []string) {
				err = os.Remove(configFolder)
			}
			testPublish.SetArgs([]string{"handle"})
			assert.NoError(t, testPublish.Execute())
			assert.NotEmpty(t, patch.CalledWith)
			assert.EqualValues(t, 1, patch.CalledWith)
		}
	}
}

func newPublishLinkTaskCompleteCapture(capture *broker.DataLinkParams) PublishLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newPublishLinkTaskPretendCapture(capture *broker.DataLinkParams) PublishLinkTaskFactory {
	return func(writer broker.DataLinkWriter) *task.Task[broker.DataLinkParams] {
		return &task.Task[broker.DataLinkParams]{
			OnPrepare: func(params *broker.DataLinkParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataLinkParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
