package ws

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/run"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/runs"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserRunsFactory struct {
	mgmt.ProviderFactory[broker.RunListParams, []runs.UserRun]
}

type stubUserRunsProvider struct {
	runs     []runs.UserRun
	getError task.Error
}

func (s stubUserRunsProvider) Get() ([]runs.UserRun, task.Error) {
	return s.runs, s.getError
}

func TestRunListExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = run.NewListOptions()
	var expectedAccount = "expectedAccount"
	var expectedWorkspaceId = cast.ToString(int64(37))
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions:  testOptions,
		workspaceArg: expectedWorkspaceId,

		accountParams: testOptions.AccountParams(testLedger),
	}

	testViper.Set(schema.Genaiz.Workspace.Run.List.Account.Doc, expectedAccount)
	testViper.Set(schema.Genaiz.Workspace.Run.List.DateMonthly.Doc, true)
	testViper.Set(schema.Genaiz.Workspace.Run.List.DateWeekly.Doc, true)
	testViper.Set(schema.Genaiz.Workspace.Run.List.DateToday.Doc, false)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`account:[\s\t]*`+expectedAccount), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+expectedWorkspaceId), actual)
	assert.Regexp(t, regexp.MustCompile(`monthly:[\s\t]*true`), actual)
	assert.Regexp(t, regexp.MustCompile(`weekly:[\s\t]*true`), actual)
	assert.Regexp(t, regexp.MustCompile(`today:[\s\t]*false`), actual)
}

func TestRunListExecutor_List(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = run.NewListOptions()
	var testCli = &Cli{
		Dry: func(ledger *config.Ledger) bool {
			return true
		},
	}
	var expectedWorkspaceName = "expectedWsName"
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
	}

	testViper.Set(schema.Genaiz.Workspace.Run.List.DateMonthly.Doc, false)
	testViper.Set(schema.Genaiz.Workspace.Run.List.DateWeekly.Doc, false)
	testViper.Set(schema.Genaiz.Workspace.Run.List.DateToday.Doc, true)
	testViper.Set(schema.Genaiz.Workspace.Run.List.Printer.Doc, true)
	assert.NoError(t, testExecutor.List(expectedWorkspaceName))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspaceName), actual)
	assert.Regexp(t, regexp.MustCompile(`monthly:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`weekly:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`today:[\s\t]*true`), actual)
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*true`), actual)
}

func TestRunListExecutor_Pretend(t *testing.T) {
	var calledList, calledResolve bool
	var expectedWorkspaceId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithLogger(getTestLogger()).
		Build()
	var testOptions = run.NewListOptions()
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		workspaceArg: cast.ToString(expectedWorkspaceId),

		accountParams: testOptions.AccountParams(testLedger),

		runListTaskFactory:          newRunListTaskPretendStub(&calledList),
		workspaceResolveTaskFactory: newFlowResolveTaskPretendStub(&calledResolve),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, calledList)
	assert.False(t, calledResolve)
}

func TestRunListExecutor_Pretend_WorkspaceId(t *testing.T) {
	var calledList, calledResolve bool
	var expectedWorkspace = "workspace"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithLogger(getTestLogger()).
		Build()
	var testOptions = run.NewListOptions()
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		workspaceArg: expectedWorkspace,

		accountParams: testOptions.AccountParams(testLedger),

		runListTaskFactory:          newRunListTaskPretendStub(&calledList),
		workspaceResolveTaskFactory: newFlowResolveTaskPretendStub(&calledResolve),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, calledList)
	assert.True(t, calledResolve)
}

func TestRunListExecutor_Proceed(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFactory = &stubUserRunsFactory{}
	var testPrinter = &stubPrinter{}
	var testProvider = &stubUserRunsProvider{
		runs: []runs.UserRun{
			{
				Id: int64(37),
			},
		},
	}
	var testOptions = run.NewListOptions()
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{printer: testPrinter},

		workspaceRunsFactory: testFactory.WithProvider(
			func(p runs.UserRunsProviderFactory) runs.UserRunsProvider {
				return testProvider
			},
		),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printOut.([]runs.UserRun); ok {
		assert.Equal(t, testProvider.runs, actual)
	} else {
		assert.Fail(t, "expected a slice of runs.UserRun")
	}
}

func TestRunListExecutor_Proceed_Error(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFactory = &stubUserRunsFactory{}
	var testPrinter = &stubPrinter{}
	var testProvider = &stubUserRunsProvider{
		getError: task.NewError("expected"),
	}
	var testOptions = run.NewListOptions()
	var testExecutor = &RunListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{printer: testPrinter},

		workspaceRunsFactory: testFactory.WithProvider(
			func(p runs.UserRunsProviderFactory) runs.UserRunsProvider {
				return testProvider
			},
		),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printError.(task.Error); ok {
		assert.Equal(t, testProvider.getError, actual)
	} else {
		assert.Fail(t, "expected a task error")
	}
}

func TestNewRunListExecutor(t *testing.T) {
	var expectedName = "workspace name"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = run.NewListOptions()
	var testCmd = &cobra.Command{}
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = newRunListExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	assert.NoError(t, testExecutor.List(expectedName))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedName), actual)
}

func newRunListTaskPretendStub(flag *bool) broker.RunListTaskFactory {
	return func() *task.Task[broker.RunListParams] {
		return &task.Task[broker.RunListParams]{
			OnPrepare: func(params *broker.RunListParams, state *task.State) error {
				*flag = true
				return nil
			},
			OnComplete: func(params *broker.RunListParams, state *task.State) error {
				return nil
			},
		}
	}
}
