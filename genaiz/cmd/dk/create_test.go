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

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestCreateExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewCreateOptions()
	var expectedHandle = "dkHandle"
	var expectedName = "name-create"
	var expectedOem = "dkOem"
	var expectedVersion = "1.0.0"
	var expectedDesc = "desc-create"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		linkHandle: expectedHandle,
	}

	testViper.Set(testOptions.optionDescription.Key, expectedDesc)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDesc), actual)
}

func TestCreateExecutor_Pretend(t *testing.T) {
	var calledParams broker.DataLinkParams
	var expectedDescription = "description"
	var expectedOem = "oem"
	var expectedVersion = "1.0.0"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(),

		linkHandle: "create-pretend",

		publishLinkTaskFactory: newPublishLinkTaskPretendCapture(&calledParams),
	}

	testViper.Set(testExecutor.optionDescription.Key, expectedDescription)
	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.Equal(t, expectedDescription, calledParams.Description)
	assert.Equal(t, testExecutor.linkHandle, calledParams.Handle)
	assert.Equal(t, testExecutor.linkHandle, calledParams.Name)
	assert.Equal(t, expectedOem, calledParams.Oem)
	assert.Equal(t, expectedVersion, calledParams.Version)
}

func TestCreateExecutor_Proceed(t *testing.T) {
	var calledParams broker.DataLinkParams
	var expectedDescription = "description"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "1.0.0"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(),

		linkHandle: "create-proceed",

		publishLinkTaskFactory: newPublishLinkTaskCompleteCapture(&calledParams),
	}

	testViper.Set(testExecutor.optionDescription.Key, expectedDescription)
	testViper.Set(testExecutor.optionName.Key, expectedName)
	testViper.Set(testExecutor.optionOem.Key, expectedOem)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedDescription, calledParams.Description)
	assert.Equal(t, testExecutor.linkHandle, calledParams.Handle)
	assert.Equal(t, expectedName, calledParams.Name)
	assert.Equal(t, expectedOem, calledParams.Oem)
	assert.Equal(t, expectedVersion, calledParams.Version)
}

func TestNewCreate(t *testing.T) {
	var createCompleted = false
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
	}
	var testCreate = NewCreate(testLedger, testCli)
	var expectedFolder = "test-folder"

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{"create-handle", expectedFolder})
	assert.NoError(t, testCreate.Execute())
	assert.True(t, createCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedFolder)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewCreate_DisappearingWorkingDir(t *testing.T) {
	if runtime.GOOS == "linux" {
		// This only happens on Linux, OSX prevents removing the current working dir with an EBUSY signal on remove
		var testDir = filepath.Join(t.TempDir(), ".dk_create_test")
		var patch = mock.Patches{T: t}.OsExit(func(int) {})
		var createCompleted = false
		var testOutput = new(bytes.Buffer)
		var testViper = viper.New()
		var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).
			WithViper(testViper).
			Build()
		var testCli = &Cli{
			BaseCli: cli.BaseCli{
				Dry: func(ledger *config.Ledger) bool {
					return true
				},
			},
		}
		var testFile *os.File
		var err error

		defer patch.Unpatch()

		if testFile, err = filez.CreateRecursiveTemp(testDir, "genaiz_dk_create*"); err == nil {
			var testCreate = NewCreate(testLedger, testCli)

			defer filez.CloseSilently(testFile)
			t.Chdir(testDir)

			if err = os.RemoveAll(testDir); err == nil {
				testCreate.PostRun = func(cmd *cobra.Command, args []string) {
					createCompleted = true
				}
				testCreate.SetArgs([]string{"create-handle"})

				assert.NoError(t, testCreate.Execute())
				assert.True(t, createCompleted)
				assert.Empty(t, testOutput.String())
				assert.True(t, patch.Called)
				assert.EqualValues(t, 1, patch.CalledWith)
				return
			}
		}

		assert.NoError(t, err)
	}
}

func newPublishLinkTaskCompleteCapture(capture *broker.DataLinkParams) PublishLinkTaskFactory {
	return func() *task.Task[broker.DataLinkParams] {
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
	return func() *task.Task[broker.DataLinkParams] {
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
