package wf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"runtime"
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
	var testOptions = NewCreateOptions(NewWfCli(nil, nil, nil))
	var expectedHandle = "wfHandle"
	var expectedName = "name-create"
	var expectedDesc = "desc-create"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workflowArg: expectedHandle,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDesc)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDesc), actual)
}

func TestCreateExecutor_Pretend(t *testing.T) {
	var calledParams broker.WorkflowParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(NewWfCli(nil, nil, nil)),

		workflowArg: "create-pretend",

		workflowTaskFactory: newWorkflowTaskPretendCapture(&calledParams),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.Equal(t, testExecutor.workflowArg, calledParams.Description)
	assert.Equal(t, testExecutor.workflowArg, calledParams.Handle)
	assert.Equal(t, testExecutor.workflowArg, calledParams.Name)
}

func TestCreateExecutor_Pretend_InvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions:       NewCreateOptions(NewWfCli(nil, nil, nil)),
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "invalid")
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestCreateExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(NewWfCli(nil, nil, nil)),

		workflowArg: "create-proceed",

		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionName.Key, "create-name")
	testViper.Set(testExecutor.optionDescription.Key, "create-desc")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestCreateExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions:       NewCreateOptions(NewWfCli(nil, nil, nil)),
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "invalid")
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestCreateExecutor_Proceed_InvalidHandle(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions:       NewCreateOptions(NewWfCli(nil, nil, nil)),
		workflowArg:         "not--valid",
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestCreateExecutor_Proceed_WithNameDesc(t *testing.T) {
	var calledParams broker.WorkflowParams
	var expectedDesc = "create-desc"
	var expectedHandle = "create-proceed"
	var expectedName = "create-name"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: NewCreateOptions(NewWfCli(nil, nil, nil)),

		workflowArg: expectedHandle,

		workflowTaskFactory: newWorkflowTaskCompleteCapture(&calledParams),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testViper.Set(testExecutor.optionDescription.Key, expectedDesc)
	testViper.Set(testExecutor.optionName.Key, expectedName)
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.Equal(t, expectedDesc, calledParams.Description)
	assert.Equal(t, expectedHandle, calledParams.Handle)
	assert.Equal(t, expectedName, calledParams.Name)
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
		var testDir = filepath.Join(t.TempDir(), ".wf_create_test")
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
		var testFile, err = filez.CreateRecursiveTemp(testDir, "genaiz_wf_create*")

		defer patch.Unpatch()
		defer filez.CloseSilently(testFile)
		panicz.PanicIfError(err)
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
		} else {
			assert.Fail(t, err.Error())
		}
	}
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

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{"create-handle", "test.."})
	assert.Error(t, testCreate.Execute())
	assert.False(t, createCompleted)
	assert.Empty(t, testOutput.String())
}

func TestNewCreateInvalidWorkingDir(t *testing.T) {
	var testDir = t.TempDir()
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
	var testFile, err = filez.CreateRecursiveTemp(testDir, ".genaiz_wf_create*")

	defer filez.CloseSilently(testFile)
	panicz.PanicIfError(err)

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{"create-handle", testFile.Name()})
	assert.Error(t, testCreate.Execute())
	assert.False(t, createCompleted)
	assert.Empty(t, testOutput.String())
}

func newWorkflowTaskCompleteCapture(capture *broker.WorkflowParams) WorkflowTaskFactory {
	return func(writer broker.WorkflowWriter) *task.Task[broker.WorkflowParams] {
		return &task.Task[broker.WorkflowParams]{
			OnPrepare: func(params *broker.WorkflowParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkflowParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkflowTaskPretendCapture(capture *broker.WorkflowParams) WorkflowTaskFactory {
	return func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams] {
		return &task.Task[broker.WorkflowParams]{
			OnPrepare: func(params *broker.WorkflowParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkflowParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
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
