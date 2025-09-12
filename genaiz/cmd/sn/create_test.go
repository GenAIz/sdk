package sn

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

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
	var testDefaultOption = &config.StringOption{}
	var testOptions = &CreateOptions{
		PublishOptions: PublishOptions{
			optionConfigType:  newOptionConfigType("test"),
			optionDescription: newOptionDescription(testDefaultOption, "test"),
			optionHandle:      newOptionHandle("test"),
			optionName:        newOptionName(testDefaultOption, "test"),
			optionOem:         newOptionOem("test"),
			optionVersion:     newOptionVersion("test"),
		},
		optionWorkflowHandle: newOptionWorkflowHandle(),
		optionWorkflowName:   newOptionWorkflowName(testDefaultOption),
	}
	var expectedDescription = "description"
	var expectedFolder = "folder"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var expectedWorkflowHandle = "workflowHandle"
	var expectedWorkflowName = "workflowName"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: expectedFolder,
		},
		CreateOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testViper.Set(testOptions.optionWorkflowHandle.Key, expectedWorkflowHandle)
	testViper.Set(testOptions.optionWorkflowName.Key, expectedWorkflowName)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(`folder:[\s\t]*`+expectedFolder), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionWorkflowHandle.Param+`:[\s\t]*`+expectedWorkflowHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionWorkflowName.Param+`:[\s\t]*`+expectedWorkflowName), actual)
}

func TestCreateExecutor_Pretend(t *testing.T) {
	var calledSolution, calledWorkflow bool
	var testCli = NewSnCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var expectedHandle = "solution-handle"
	var expectedOem = "solution-oem"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(),

		solutionTaskFactory: newSolutionCreateTaskPretendStub(&calledSolution),
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.CreateOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testExecutor.CreateOptions.optionOem.Key, expectedOem)
	testLedger.Register(&cobra.Command{}, testExecutor.CreateOptions.allDefiners()...)
	testLedger.InitDefaults()
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, calledSolution)
	assert.True(t, calledWorkflow)
}

func TestCreateExecutor_Proceed(t *testing.T) {
	var calledSolution, calledWorkflow bool
	var testCli = NewSnCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var expectedHandle = "solution-handle"
	var expectedOem = "solution-oem"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		CreateOptions: NewCreateOptions(),

		solutionTaskFactory: newSolutionCreateTaskCompleteStub(&calledSolution),
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.CreateOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testExecutor.CreateOptions.optionOem.Key, expectedOem)
	testLedger.Register(&cobra.Command{}, testExecutor.CreateOptions.allDefiners()...)
	testLedger.InitDefaults()
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.True(t, calledSolution)
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
	var expectedSolution = "solution-test"

	testCreate.PostRun = func(cmd *cobra.Command, args []string) {
		createCompleted = true
	}
	testCreate.SetArgs([]string{expectedSolution})
	assert.NoError(t, testCreate.Execute())
	assert.True(t, createCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedSolution)
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
	var testFile, err = filez.CreateRecursiveTemp("/tmp/.sn_create_test", "genaiz_sn_create*")
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

func newSolutionCreateTaskCompleteStub(flag *bool) SolutionCreateTaskFactory {
	return func(broker.SolutionWriter) *task.Task[broker.SolutionParams] {
		return &task.Task[broker.SolutionParams]{
			OnPrepare: func(params *broker.SolutionParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.SolutionParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newSolutionCreateTaskPretendStub(flag *bool) SolutionCreateTaskFactory {
	return func(broker.SolutionWriter) *task.Task[broker.SolutionParams] {
		return &task.Task[broker.SolutionParams]{
			OnPrepare: func(params *broker.SolutionParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.SolutionParams, state *task.State) error {
				*flag = true
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
