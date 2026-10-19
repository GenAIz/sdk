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
	"genaiz.com/genaiz/cmd/ws/node"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/workspaces"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserWorkspaceNodesFactory struct {
	mgmt.ProviderFactory[broker.WorkspaceNodeListParams, []workspaces.UserWorkspaceNode]
}

type stubUserWorkspaceNodesProvider struct {
	nodes    []workspaces.UserWorkspaceNode
	getError task.Error
}

func (s stubUserWorkspaceNodesProvider) Get() ([]workspaces.UserWorkspaceNode, task.Error) {
	return s.nodes, s.getError
}

func TestNodeListExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = node.NewListOptions()
	var expectedAccount = "expectedAccount"
	var testExecutor = &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
	}

	testViper.Set(testOptions.OptionAccount.Key, expectedAccount)
	testViper.Set(testOptions.OptionJsonPrinter.Key, cast.ToString(true))
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.OptionAccount.Param+`:[\s\t]*`+expectedAccount), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.OptionJsonPrinter.Param+`:[\s\t]*true`), actual)
}

func TestNodeListExecutor_List(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = node.NewListOptions()
	var expectedWorkspace = "theWorkspace"
	var expectedFlow = "theFlow"
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		ListOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
	}

	testViper.Set(testOptions.OptionJsonPrinter.Key, cast.ToString(true))
	assert.NoError(t, testExecutor.List(expectedWorkspace, expectedFlow))
	actual := testOutput.String()
	assert.NotContains(t, actual, testOptions.OptionAccount.Param)
	assert.Regexp(t, regexp.MustCompile(`workspace:[\s\t]*`+expectedWorkspace), actual)
	assert.Regexp(t, regexp.MustCompile(`flow:[\s\t]*`+expectedFlow), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.OptionJsonPrinter.Param+`:[\s\t]*true`), actual)
}

func TestNodeListExecutor_Pretend(t *testing.T) {
	var calledList bool
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = node.NewListOptions()
	var testExecutor = &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams:                config.NewAccountParams(testLedger, testOptions.OptionAccount),
		workspaceNodeListTaskFactory: newNodeListTaskPretendStub(&calledList),
	}

	testExecutor.Pretend()
	assert.True(t, calledList)
}

func TestNodeListExecutor_Proceed(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPrinter = &stubPrinter{}
	var testProvider = &stubUserWorkspaceNodesProvider{
		nodes: []workspaces.UserWorkspaceNode{
			{
				Id:                 37,
				WorkflowNodeHandle: "nodeHandle",
			},
		},
	}
	var testFactory = &stubUserWorkspaceNodesFactory{}
	var testOptions = node.NewListOptions()
	var testExecutor = &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
		printerParams: &stubPrinterParametric{printer: testPrinter},

		workspaceNodesFactory: testFactory.WithProvider(
			func(workspaces.UserWorkspaceNodesProviderFactory) workspaces.UserWorkspaceNodesProvider {
				return testProvider
			}),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printOut.([]workspaces.UserWorkspaceNode); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testProvider.nodes[0].Id, actual[0].Id)
		assert.Equal(t, testProvider.nodes[0].WorkflowNodeHandle, actual[0].WorkflowNodeHandle)
	} else {
		assert.Fail(t, "did not receive the workspace nodes")
	}
}

func TestNodeListExecutor_Proceed_Error(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPrinter = &stubPrinter{}
	var testProvider = &stubUserWorkspaceNodesProvider{
		getError: task.NewError("expected"),
	}
	var testFactory = &stubUserWorkspaceNodesFactory{}
	var testOptions = node.NewListOptions()
	var testExecutor = &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		workspaceArg: "1",
		flowArg:      "37",

		accountParams: config.NewAccountParams(testLedger, testOptions.OptionAccount),
		printerParams: &stubPrinterParametric{printer: testPrinter},

		workspaceNodesFactory: testFactory.WithProvider(
			func(workspaces.UserWorkspaceNodesProviderFactory) workspaces.UserWorkspaceNodesProvider {
				return testProvider
			}),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printError.(task.Error); ok {
		assert.ErrorIs(t, actual, testProvider.getError)
	} else {
		assert.Fail(t, "did not receive the error")
	}
}

func TestNewNodeListExecutor(t *testing.T) {
	var expectedWorkspace = "workspace name"
	var expectedFlowId = "1337"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = node.NewListOptions()
	var testCmd = &cobra.Command{}
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = newNodeListExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	assert.NoError(t, testExecutor.List(expectedWorkspace, expectedFlowId))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace:[\s\t]*`+expectedWorkspace), actual)
	assert.Regexp(t, regexp.MustCompile(`flow:[\s\t]*`+expectedFlowId), actual)
}

func newNodeListTaskPretendStub(flag *bool) broker.WorkspaceNodeListTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeListParams] {
		return &task.Task[broker.WorkspaceNodeListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newWorkspaceNodeResolveProceedCapture(capture *broker.WorkspaceNodeResolveParams) broker.WorkspaceNodeResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeResolveParams] {
		return &task.Task[broker.WorkspaceNodeResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
