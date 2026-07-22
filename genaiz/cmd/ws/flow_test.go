package ws

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/flow"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestFlowCreateExecutor_Create(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli: &Cli{
				BaseCli: cli.BaseCli{
					Dry: func(*config.Ledger) bool { return true },
				},
			},
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
	}
	var expectedWorkflowId = "42"
	var expectedWorkspaceId = "37"

	// Case where the command was entered with specific ids
	assert.NoError(t, testExecutor.Create(expectedWorkspaceId, "", expectedWorkflowId))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+expectedWorkspaceId), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow id:[\s\t]*`+expectedWorkflowId), actual)
}

func TestFlowCreateExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = flow.NewCreateOptions()
	var expectedSolutionOem = "expectedSnOem"
	var expectedSolutionHandle = "expectedSnHandle"
	var expectedSolutionVersion = "expectedSnVersion"
	var expectedWorkflowHandle = "expectedWfHandle"
	var expectedWorkspaceName = "expectedWsName"
	var expectedName = "expectedName"
	var expectedDesc = "expectedDesc"
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
		resolveParams: &broker.WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{},
			WorkspaceName:             expectedWorkspaceName,
			WorkflowHandle:            expectedWorkflowHandle,
			SolutionOem:               expectedSolutionOem,
			SolutionHandle:            expectedSolutionHandle,
			SolutionVersion:           expectedSolutionVersion,
		},
	}

	testViper.Set(testOptions.OptionName.Key, expectedName)
	testViper.Set(testOptions.OptionDescription.Key, expectedDesc)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspaceName), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedWorkflowHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`solution oem:[\s\t]*`+expectedSolutionOem), actual)
	assert.Regexp(t, regexp.MustCompile(`solution handle:[\s\t]*`+expectedSolutionHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`solution version:[\s\t]*`+expectedSolutionVersion), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.OptionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.OptionDescription.Param+`:[\s\t]*`+expectedDesc), actual)
}

func TestFlowCreateExecutor_Pretend(t *testing.T) {
	var calledCreate, calledResolve, calledSolution bool
	var expectedWorkflowId = int64(37)
	var expectedWorkspaceId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		flowCreateTaskFactory:   newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:  newFlowResolveTaskPretendStub(&calledResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskPretendStub(&calledSolution),
	}

	testExecutor.resolveParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			WorkflowId:  &expectedWorkflowId,
			WorkspaceId: &expectedWorkspaceId,
		},
	}
	testExecutor.Pretend()
	assert.True(t, calledCreate)
	assert.False(t, calledResolve)
	assert.False(t, calledSolution)
}

func TestFlowCreateExecutor_Pretend_NoWorkflowId(t *testing.T) {
	var calledCreate, calledResolve, calledSolution bool
	var expectedWorkspaceId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		flowCreateTaskFactory:   newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:  newFlowResolveTaskPretendStub(&calledResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskPretendStub(&calledSolution),
	}

	testExecutor.resolveParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			WorkspaceId: &expectedWorkspaceId,
		},
	}
	testExecutor.Pretend()
	assert.True(t, calledCreate)
	assert.False(t, calledResolve)
	assert.True(t, calledSolution)
}

func TestFlowCreateExecutor_Pretend_NoWorkspaceId(t *testing.T) {
	var calledCreate, calledResolve, calledSolution bool
	var expectedWorkflowId = int64(37)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		flowCreateTaskFactory:   newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:  newFlowResolveTaskPretendStub(&calledResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskPretendStub(&calledSolution),
	}

	testExecutor.resolveParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			WorkflowId: &expectedWorkflowId,
		},
	}
	testExecutor.Pretend()
	assert.True(t, calledCreate)
	assert.True(t, calledResolve)
	assert.False(t, calledSolution)
}

func TestFlowCreateExecutor_Proceed(t *testing.T) {
	var captureCreate broker.WorkspaceFlowCreateParams
	var captureResolve broker.WorkspaceFlowResolveParams
	var captureSolution broker.WorkspaceFlowResolveParams
	var expectedParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{},
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		printerParams: &stubPrinterParametric{
			defaultPrinter: true,
			printer:        testPrinter,
		},

		flowCreateTaskFactory:   newFlowCreateTaskCompleteCapture(&captureCreate),
		flowResolveTaskFactory:  newFlowResolveTaskCompleteCapture(&captureResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskCompleteCapture(&captureSolution),
	}

	testLedger.InitLogging()
	testExecutor.resolveParams = expectedParams
	testExecutor.Proceed()
	assert.Equal(t, expectedParams.WorkspaceFlowCreateParams, &captureCreate)
	assert.Equal(t, expectedParams, &captureResolve)
	assert.Equal(t, expectedParams, &captureSolution)
}

func TestFlowCreateExecutor_Proceed_WithWorkflowId(t *testing.T) {
	var captureResolve broker.WorkspaceFlowResolveParams
	var captureSolution broker.WorkspaceFlowResolveParams
	var expectedParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			WorkflowId: new(int64(37)),
		},
	}
	var expectedFlow = &broker.WorkspaceFlow{
		WorkspaceId: int64(42),
		WorkflowId:  *expectedParams.WorkflowId,
		Name:        "expectedName",
		Description: "expectedDescription",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		flowCreateTaskFactory:   newFlowCreateTaskCompleteStub(expectedFlow),
		flowResolveTaskFactory:  newFlowResolveTaskCompleteCapture(&captureResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskCompleteCapture(&captureSolution),
	}

	testLedger.InitLogging()
	testExecutor.resolveParams = expectedParams
	testExecutor.Proceed()
	assert.Equal(t, expectedParams, &captureResolve)
	assert.Empty(t, captureSolution)
	assert.Equal(t, expectedFlow, testPrinter.printOut)
}

func TestFlowCreateExecutor_Proceed_WithWorkspaceId(t *testing.T) {
	var captureResolve broker.WorkspaceFlowResolveParams
	var captureSolution broker.WorkspaceFlowResolveParams
	var expectedParams = &broker.WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
			WorkflowId: new(int64(37)),
		},
	}
	var expectedFlow = &broker.WorkspaceFlow{
		WorkspaceId: int64(42),
		WorkflowId:  *expectedParams.WorkflowId,
		Name:        "expectedName",
		Description: "expectedDescription",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: flow.NewCreateOptions(),

		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		flowCreateTaskFactory:   newFlowCreateTaskCompleteStub(expectedFlow),
		flowResolveTaskFactory:  newFlowResolveTaskCompleteCapture(&captureResolve),
		flowSolutionTaskFactory: newFlowSolutionTaskCompleteCapture(&captureSolution),
	}

	testLedger.InitLogging()
	testExecutor.resolveParams = expectedParams
	testExecutor.Proceed()
	assert.Equal(t, expectedParams, &captureResolve)
	assert.Empty(t, captureSolution)
	assert.Equal(t, expectedFlow, testPrinter.printOut)
}

func TestNewFlowCreateExecutor(t *testing.T) {
	var expectedName = "workspace name"
	var expectedWorkflowId = "1337"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = flow.NewCreateOptions()
	var testCmd = &cobra.Command{}
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = newFlowCreateExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	assert.NoError(t, testExecutor.Create(expectedName, "", expectedWorkflowId))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedName), actual)
}

func newFlowCreateTaskPretendStub(flag *bool) workspaceFlowCreateTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowCreateParams] {
		return &task.Task[broker.WorkspaceFlowCreateParams]{
			OnPrepare: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newFlowResolveTaskPretendStub(flag *bool) workspaceFlowResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newFlowSolutionTaskPretendStub(called *bool) workspaceFlowSolutionTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*called = true
				return nil
			},
		}
	}
}

func newFlowCreateTaskCompleteCapture(capture *broker.WorkspaceFlowCreateParams) workspaceFlowCreateTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowCreateParams] {
		return &task.Task[broker.WorkspaceFlowCreateParams]{
			OnPrepare: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newFlowCreateTaskCompleteStub(expected *broker.WorkspaceFlow) workspaceFlowCreateTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowCreateParams] {
		return &task.Task[broker.WorkspaceFlowCreateParams]{
			OnPrepare: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowCreateParams, state *task.State) error {
				state.Internal = expected
				return nil
			},
		}
	}
}

func newFlowResolveTaskCompleteCapture(capture *broker.WorkspaceFlowResolveParams) workspaceFlowResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newFlowSolutionTaskCompleteCapture(capture *broker.WorkspaceFlowResolveParams) workspaceFlowSolutionTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
