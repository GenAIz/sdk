package wf

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
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestCreateExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewCreateOptions()
	var expectedHandle = "handle"
	var expectedName = "name-create"
	var expectedDesc = "desc-create"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDesc)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDesc), actual)
}

func TestCreateExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(),

		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionName.Key, "create-name")
	testViper.Set(testExecutor.optionHandle.Key, "create-pretend")
	testViper.Set(testExecutor.optionDescription.Key, "create-desc")
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestCreateExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(),

		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionHandle.Key, "create-proceed")
	testViper.Set(testExecutor.optionName.Key, "create-name")
	testViper.Set(testExecutor.optionDescription.Key, "create-desc")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
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

	testViper.Set(newOptionHandle("create").Key, "create-handle")
	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{expectedFolder})
	assert.NoError(t, testCreate.Execute())
	assert.True(t, createCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedFolder)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewCreateDisappearingWorkingDir(t *testing.T) {
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
	var testCreate = NewCreate(testLedger, testCli)
	var testFile, err = filez.CreateRecursiveTemp("/tmp/.create_test", "genaiz_wf_create*")
	var back string

	defer patch.Unpatch()
	panicz.PanicIfError(err)
	defer filez.RemoveSilently(filepath.Dir(testFile.Name()))
	back, err = os.Getwd()
	panicz.PanicIfError(err)
	panicz.PanicIfError(os.Chdir(filepath.Dir(testFile.Name())))
	defer func() { _ = os.Chdir(back) }()
	filez.RemoveSilently(filepath.Dir(testFile.Name()))
	testViper.Set(newOptionHandle("create").Key, "create-handle")
	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{})
	assert.NoError(t, testCreate.Execute())
	assert.True(t, createCompleted)
	assert.Empty(t, testOutput.String())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewCreateInvalidArgs(t *testing.T) {
	var createCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCreate = NewCreate(testLedger, testCli)

	testViper.Set(newOptionHandle("create").Key, "create-handle")
	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{"test.."})
	assert.Error(t, testCreate.Execute())
	assert.False(t, createCompleted)
	assert.Empty(t, testOutput.String())
}

func TestNewCreateInvalidWorkingDir(t *testing.T) {
	var createCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCreate = NewCreate(testLedger, testCli)
	var testFile, err = filez.CreateRecursiveTemp("/tmp", ".genaiz_wf_create*")

	panicz.PanicIfError(err)
	defer filez.RemoveSilently(testFile.Name())
	testViper.Set(newOptionHandle("create").Key, "create-handle")
	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{testFile.Name()})
	assert.Error(t, testCreate.Execute())
	assert.False(t, createCompleted)
	assert.Empty(t, testOutput.String())
}

func newWorkflowTaskCompleteStub(flag *bool) WorkflowTaskFactory {
	return func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams] {
		return &task.Task[broker.WorkflowParams]{
			OnPrepare: func(params *broker.WorkflowParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkflowParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newWorkflowTaskPretendStub(flag *bool) WorkflowTaskFactory {
	return func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams] {
		return &task.Task[broker.WorkflowParams]{
			OnPrepare: func(params *broker.WorkflowParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkflowParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
