package sn

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
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
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
	var testOptions = &CreateOptions{
		PublishOptions: PublishOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.Solution.Create.ConfigType).
				BuildStringOption(),
			optionDescription: cli.Options.Solutions.Description().
				WithKeys(&schema.Genaiz.Solution.Create.Description).
				BuildStringOption(),
			optionHandle: cli.Options.Solutions.Handle().
				WithKeys(&schema.Genaiz.Solution.Create.Handle).
				BuildStringOption(),
			optionName: cli.Options.Solutions.Name().
				WithKeys(&schema.Genaiz.Solution.Create.Name).
				BuildStringOption(),
			optionOem: cli.Options.Solutions.Oem().
				WithKeys(&schema.Genaiz.Solution.Create.Oem).
				BuildStringOption(),
			optionVersion: cli.Options.Solutions.Version().
				WithKeys(&schema.Genaiz.Solution.Create.Version).
				BuildStringOption(),
		},
		optionWorkflowDesc: cli.Options.Solutions.WorkflowDesc().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Description).
			BuildStringOption(),
		optionWorkflowHandle: cli.Options.Solutions.WorkflowHandle().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Handle).
			BuildStringOption(),
		optionWorkflowName: cli.Options.Solutions.WorkflowName().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Name).
			BuildStringOption(),
	}
	var expectedDescription = "description"
	var expectedFolder = "folder"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var expectedWorkflowDesc = "workflowDesc"
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
	testViper.Set(testOptions.optionWorkflowDesc.Key, expectedWorkflowDesc)
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
	assert.Regexp(t, regexp.MustCompile(testOptions.optionWorkflowDesc.Param+`:[\s\t]*`+expectedWorkflowDesc), actual)
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

func TestNewCreate_DisappearingWorkingDir(t *testing.T) {
	if runtime.GOOS == "linux" {
		// This only happens on Linux, OSX prevents removing the current working dir with an EBUSY signal on remove
		var testDir = filepath.Join(t.TempDir(), ".sn_create_test")
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
		var testFile, err = filez.CreateRecursiveTemp(testDir, "genaiz_sn_create*")

		defer patch.Unpatch()
		defer filez.CloseSilently(testFile)
		panicz.PanicIfError(err)
		t.Chdir(testDir)

		if err = os.RemoveAll(testDir); err == nil {
			testViper.Set(schema.Genaiz.Solution.Create.Handle.Doc, "create-handle")
			testCreate.PostRun = func(cmd *cobra.Command, args []string) {
				createCompleted = true
			}
			testCreate.SetArgs([]string{})
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
