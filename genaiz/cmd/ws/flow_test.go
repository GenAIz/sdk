package ws

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/flow"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
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

		accountParams: testOptions.AccountParams(testLedger),
	}
	var expectedWorkflowId = "42"
	var expectedWorkspaceId = "37"

	// Case where the command was entered with specific ids
	assert.NoError(t, testExecutor.Create(expectedWorkspaceId, "", expectedWorkflowId))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+expectedWorkspaceId), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow id:[\s\t]*`+expectedWorkflowId), actual)
}

func TestFlowCreateExecutor_Create_NoIds(t *testing.T) {
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

		accountParams: testOptions.AccountParams(testLedger),
	}
	var expectedWorkflowHandle = "workflowHandle"
	var expectedWorkspaceName = "workspaceName"

	// Case where the command was entered with specific ids
	assert.NoError(t, testExecutor.Create(expectedWorkspaceName, "", expectedWorkflowHandle))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspaceName), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedWorkflowHandle), actual)
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

		solutionArg:  fmt.Sprintf("%s/%s:%s", expectedSolutionOem, expectedSolutionHandle, expectedSolutionVersion),
		workflowArg:  expectedWorkflowHandle,
		workspaceArg: expectedWorkspaceName,

		accountParams: testOptions.AccountParams(testLedger),
	}

	testViper.Set(schema.Genaiz.Workspace.Flow.Create.Name.Doc, expectedName)
	testViper.Set(schema.Genaiz.Workspace.Flow.Create.Description.Doc, expectedDesc)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspaceName), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedWorkflowHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`solution oem:[\s\t]*`+expectedSolutionOem), actual)
	assert.Regexp(t, regexp.MustCompile(`solution handle:[\s\t]*`+expectedSolutionHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`solution version:[\s\t]*`+expectedSolutionVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(`description:[\s\t]*`+expectedDesc), actual)
}

func TestFlowCreateExecutor_Pretend(t *testing.T) {
	var calledCreate, calledResolve, calledSolution bool
	var expectedWorkflowId = int64(37)
	var expectedWorkspaceId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workflowArg:  cast.ToString(expectedWorkflowId),
		workspaceArg: cast.ToString(expectedWorkspaceId),

		accountParams: testOptions.AccountParams(testLedger),

		flowCreateTaskFactory:       newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:      newFlowResolveTaskPretendStub(&calledSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskPretendStub(&calledResolve),
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
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workspaceArg: cast.ToString(expectedWorkspaceId),

		accountParams: testOptions.AccountParams(testLedger),

		flowCreateTaskFactory:       newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:      newFlowResolveTaskPretendStub(&calledSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskPretendStub(&calledResolve),
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
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workflowArg: cast.ToString(expectedWorkflowId),

		accountParams: testOptions.AccountParams(testLedger),

		flowCreateTaskFactory:       newFlowCreateTaskPretendStub(&calledCreate),
		flowResolveTaskFactory:      newFlowResolveTaskPretendStub(&calledSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskPretendStub(&calledResolve),
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
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		printerParams: &stubPrinterParametric{
			defaultPrinter: true,
			printer:        testPrinter,
		},

		accountParams: testOptions.AccountParams(testLedger),

		flowCreateTaskFactory:       newFlowCreateTaskCompleteCapture(&captureCreate),
		flowResolveTaskFactory:      newFlowResolveTaskCompleteCapture(&captureSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskCompleteCapture(&captureResolve),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureResolve)
	assert.NotEmpty(t, captureSolution)
}

func TestFlowCreateExecutor_Proceed_WithWorkflowId(t *testing.T) {
	var captureResolve broker.WorkspaceFlowResolveParams
	var captureSolution broker.WorkspaceFlowResolveParams
	var expectedFlowId = int64(37)
	var expectedFlow = &broker.WorkspaceFlow{
		Id:          expectedFlowId,
		WorkspaceId: int64(42),
		Name:        "expectedName",
		Description: "expectedDescription",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workflowArg: cast.ToString(expectedFlowId),

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		flowCreateTaskFactory:       newFlowCreateTaskCompleteStub(expectedFlow),
		flowResolveTaskFactory:      newFlowResolveTaskCompleteCapture(&captureSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskCompleteCapture(&captureResolve),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureResolve)
	assert.Empty(t, captureSolution)
	assert.Equal(t, expectedFlow, testPrinter.printOut)
}

func TestFlowCreateExecutor_Proceed_WithWorkspaceId(t *testing.T) {
	var captureResolve broker.WorkspaceFlowResolveParams
	var captureSolution broker.WorkspaceFlowResolveParams
	var expectedFlowId = int64(73)
	var expectedFlow = &broker.WorkspaceFlow{
		Id:          expectedFlowId,
		WorkspaceId: int64(42),
		Name:        "expectedName",
		Description: "expectedDescription",
	}
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testPrinter = &stubPrinter{}
	var testOptions = flow.NewCreateOptions()
	var testExecutor = &FlowCreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		workflowArg: cast.ToString(expectedFlowId),

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		flowCreateTaskFactory:       newFlowCreateTaskCompleteStub(expectedFlow),
		flowResolveTaskFactory:      newFlowResolveTaskCompleteCapture(&captureSolution),
		workspaceResolveTaskFactory: newFlowResolveTaskCompleteCapture(&captureResolve),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureResolve)
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

func TestFlowStartExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var expectedAccount = "account"
	var expectedName = "name"
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "37",
		flowArg:      "42",
	}

	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Account.Doc, expectedAccount)
	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Name.Doc, expectedName)
	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Printer.Doc, true)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+testExecutor.workspaceArg), actual)
	assert.Regexp(t, regexp.MustCompile(`flow id:[\s\t]*`+testExecutor.flowArg), actual)
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*true`), actual)
}

func TestFlowStartExecutor_Display_WorkflowHandle(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var expectedAccount = "account"
	var expectedName = "name"
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "37",
		flowArg:      "myWorkflow",
	}

	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Account.Doc, expectedAccount)
	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Name.Doc, expectedName)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+testExecutor.workspaceArg), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+testExecutor.flowArg), actual)
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*\n`), actual)
}

func TestFlowStartExecutor_Display_WorkspaceName(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var expectedAccount = "account"
	var expectedName = "name"
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "myWorkspace",
		flowArg:      "myWorkflow",
	}

	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Account.Doc, expectedAccount)
	testViper.Set(schema.Genaiz.Workspace.Flow.Start.Name.Doc, expectedName)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+testExecutor.workspaceArg), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+testExecutor.flowArg), actual)
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*\n`), actual)
}

func TestFlowStartExecutor_Pretend(t *testing.T) {
	var prepareCalled, resolveCalled, startCalled, workspaceCalled bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "myWorkspace",
		flowArg:      "myFlow",

		accountParams: testOptions.AccountParams(testLedger),

		flowPrepareTaskFactory:      newFlowPreparePretendStub(&prepareCalled),
		runResolveTaskFactory:       newRunResolvePretendStub(&resolveCalled),
		runStartTaskFactory:         newRunStartPretendStub(&startCalled),
		workspaceResolveTaskFactory: newWorkspaceResolvePretendStub(&workspaceCalled),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, prepareCalled)
	assert.True(t, resolveCalled)
	assert.True(t, startCalled)
	assert.True(t, workspaceCalled)
}

func TestFlowStartExecutor_Pretend_WithFlowId(t *testing.T) {
	var prepareCalled, resolveCalled, startCalled, workspaceCalled bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "myWorkspace",
		flowArg:      "37",

		accountParams: testOptions.AccountParams(testLedger),

		flowPrepareTaskFactory:      newFlowPreparePretendStub(&prepareCalled),
		runResolveTaskFactory:       newRunResolvePretendStub(&resolveCalled),
		runStartTaskFactory:         newRunStartPretendStub(&startCalled),
		workspaceResolveTaskFactory: newWorkspaceResolvePretendStub(&workspaceCalled),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, prepareCalled)
	assert.False(t, resolveCalled)
	assert.True(t, startCalled)
	assert.False(t, workspaceCalled)
}

func TestFlowStartExecutor_Pretend_WithWorkspaceId(t *testing.T) {
	var prepareCalled, resolveCalled, startCalled, workspaceCalled bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "37",
		flowArg:      "myFlow",

		accountParams: testOptions.AccountParams(testLedger),

		flowPrepareTaskFactory:      newFlowPreparePretendStub(&prepareCalled),
		runResolveTaskFactory:       newRunResolvePretendStub(&resolveCalled),
		runStartTaskFactory:         newRunStartPretendStub(&startCalled),
		workspaceResolveTaskFactory: newWorkspaceResolvePretendStub(&workspaceCalled),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, prepareCalled)
	assert.True(t, resolveCalled)
	assert.True(t, startCalled)
	assert.False(t, workspaceCalled)
}

func TestFlowStartExecutor_Proceed(t *testing.T) {
	var capturedRead broker.WorkspaceFlowReadParams
	var capturedStart broker.RunStartParams
	var capturedResolve broker.RunResolveParams
	var capturedWorkspace broker.WorkspaceFlowResolveParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "myWorkspace",
		flowArg:      "myFlow",

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			defaultPrinter: true,
			printer:        testPrinter,
		},

		flowPrepareTaskFactory:      newFlowPrepareCompleteCapture(&capturedRead),
		runResolveTaskFactory:       newRunResolveCompleteCapture(&capturedResolve),
		runStartTaskFactory:         newRunStartCompleteCapture(&capturedStart),
		workspaceResolveTaskFactory: newWorkspaceResolveCompleteCapture(&capturedWorkspace),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, capturedRead)
	assert.Nil(t, capturedRead.WorkspaceId)
	assert.Nil(t, capturedRead.FlowId)
	assert.NotEmpty(t, capturedResolve)
	assert.Equal(t, testExecutor.flowArg, capturedResolve.WorkflowHandle)
	assert.Nil(t, capturedResolve.WorkspaceId)
	assert.Nil(t, capturedResolve.FlowId)
	assert.Nil(t, capturedResolve.RunId)
	assert.NotEmpty(t, capturedStart)
	assert.NotEmpty(t, capturedWorkspace)
	assert.Equal(t, testExecutor.workspaceArg, capturedWorkspace.WorkspaceName)
}

func TestFlowStartExecutor_Proceed_WithFlowId(t *testing.T) {
	var capturedRead broker.WorkspaceFlowReadParams
	var capturedStart broker.RunStartParams
	var capturedResolve broker.RunResolveParams
	var capturedWorkspace broker.WorkspaceFlowResolveParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "myWorkspace",
		flowArg:      "37",

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			defaultPrinter: true,
			printer:        testPrinter,
		},

		flowPrepareTaskFactory:      newFlowPrepareCompleteCapture(&capturedRead),
		runResolveTaskFactory:       newRunResolveCompleteCapture(&capturedResolve),
		runStartTaskFactory:         newRunStartCompleteCapture(&capturedStart),
		workspaceResolveTaskFactory: newWorkspaceResolveCompleteCapture(&capturedWorkspace),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, capturedRead)
	assert.Nil(t, capturedRead.WorkspaceId)
	assert.Equal(t, int64(37), *capturedRead.FlowId)
	assert.Empty(t, capturedResolve)
	assert.NotEmpty(t, capturedStart)
	assert.Empty(t, capturedWorkspace)
}

func TestFlowStartExecutor_Proceed_WithWorkspaceId(t *testing.T) {
	var capturedRead broker.WorkspaceFlowReadParams
	var capturedStart broker.RunStartParams
	var capturedResolve broker.RunResolveParams
	var capturedWorkspace broker.WorkspaceFlowResolveParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = flow.NewStartOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &FlowStartExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		StartOptions: testOptions,

		workspaceArg: "37",
		flowArg:      "myFlow",

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			defaultPrinter: false,
			printer:        testPrinter,
		},

		flowPrepareTaskFactory:      newFlowPrepareCompleteCapture(&capturedRead),
		runResolveTaskFactory:       newRunResolveCompleteCapture(&capturedResolve),
		runStartTaskFactory:         newRunStartCompleteCapture(&capturedStart),
		workspaceResolveTaskFactory: newWorkspaceResolveCompleteCapture(&capturedWorkspace),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, capturedRead)
	assert.Equal(t, int64(37), *capturedRead.WorkspaceId)
	assert.Nil(t, capturedRead.FlowId)
	assert.NotEmpty(t, capturedResolve)
	assert.Equal(t, testExecutor.flowArg, capturedResolve.WorkflowHandle)
	assert.Equal(t, int64(37), *capturedRead.WorkspaceId)
	assert.Nil(t, capturedResolve.FlowId)
	assert.Nil(t, capturedResolve.RunId)
	assert.NotEmpty(t, capturedStart)
	assert.Empty(t, capturedWorkspace)
}

func TestNewFlowStartExecutor(t *testing.T) {
	var expectedName = "workspace name"
	var expectedWorkflowId = "1337"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = flow.NewStartOptions()
	var testCmd = &cobra.Command{}
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = newFlowStartExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	assert.NoError(t, testExecutor.Start(expectedName, expectedWorkflowId))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedName), actual)
}

func newFlowCreateTaskPretendStub(flag *bool) broker.WorkspaceFlowCreateTaskFactory {
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

func newFlowResolveTaskPretendStub(flag *bool) broker.WorkspaceFlowResolveTaskFactory {
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

func newFlowCreateTaskCompleteCapture(capture *broker.WorkspaceFlowCreateParams) broker.WorkspaceFlowCreateTaskFactory {
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

func newFlowCreateTaskCompleteStub(expected *broker.WorkspaceFlow) broker.WorkspaceFlowCreateTaskFactory {
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

func newFlowPrepareCompleteCapture(capture *broker.WorkspaceFlowReadParams) broker.WorkspaceFlowPrepareTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowReadParams] {
		return &task.Task[broker.WorkspaceFlowReadParams]{
			OnPrepare: func(params *broker.WorkspaceFlowReadParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowReadParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newFlowPreparePretendStub(flag *bool) broker.WorkspaceFlowPrepareTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowReadParams] {
		return &task.Task[broker.WorkspaceFlowReadParams]{
			OnPrepare: func(params *broker.WorkspaceFlowReadParams, state *task.State) error {
				*flag = true
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowReadParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newFlowResolveTaskCompleteCapture(capture *broker.WorkspaceFlowResolveParams) broker.WorkspaceFlowResolveTaskFactory {
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

func newRunResolveCompleteCapture(capture *broker.RunResolveParams) broker.RunResolveTaskFactory {
	return func() *task.Task[broker.RunResolveParams] {
		return &task.Task[broker.RunResolveParams]{
			OnPrepare: func(params *broker.RunResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.RunResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newRunResolvePretendStub(flag *bool) broker.RunResolveTaskFactory {
	return func() *task.Task[broker.RunResolveParams] {
		return &task.Task[broker.RunResolveParams]{
			OnPrepare: func(params *broker.RunResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
			OnComplete: func(params *broker.RunResolveParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newRunStartCompleteCapture(capture *broker.RunStartParams) broker.RunStartTaskFactory {
	return func() *task.Task[broker.RunStartParams] {
		return &task.Task[broker.RunStartParams]{
			OnPrepare: func(params *broker.RunStartParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.RunStartParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newRunStartPretendStub(flag *bool) broker.RunStartTaskFactory {
	return func() *task.Task[broker.RunStartParams] {
		return &task.Task[broker.RunStartParams]{
			OnPrepare: func(params *broker.RunStartParams, state *task.State) error {
				*flag = true
				return nil
			},
			OnComplete: func(params *broker.RunStartParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newWorkspaceResolveCompleteCapture(captured *broker.WorkspaceFlowResolveParams) broker.WorkspaceResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
}

func newWorkspaceResolvePretendStub(flag *bool) broker.WorkspaceResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
		}
	}
}
