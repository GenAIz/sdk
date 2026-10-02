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
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestDataStoreAutoBridge_bridgeArguments(t *testing.T) {
	var stubBridge = &stubAutoBridge{}
	var testCmd = &cobra.Command{}
	var testBridge = &DataStoreAutoBridge{
		dataStores: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{"workspace", "flow", "node", "ds"}, "toComplete")
	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Empty(t, stubBridge.toComplete)
}

func TestDataStoreAutoBridge_bridgeArguments_dataStores(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataStoreAutoBridge{
		dataStores: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{"workspace", "flow", "node"}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataStoreAutoBridge_bridgeArguments_workspaces(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataStoreAutoBridge{
		workspaces: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataStoreAutoBridge_bridgeArguments_workspaceFlows(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedWorkspace = "workspace"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataStoreAutoBridge{
		workspaceFlows: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{expectedWorkspace}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubBridge.workspace)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataStoreAutoBridge_bridgeArguments_workspaceNodes(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataStoreAutoBridge{
		workspaceNodes: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{expectedWorkspace, expectedFlow}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubBridge.workspace)
	assert.Equal(t, expectedFlow, stubBridge.workspaceFlow)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataStoreExecutor_Add(t *testing.T) {
	var expectedWorkspace = 37
	var expectedFlow = 42
	var expectedNode = 69
	var expectedDs = 73
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewDataStoreAddOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,
			accountParams:       config.NewAccountParams(testLedger, testOptions.optionAccount),
		},
	}

	assert.NoError(t, testExecutor.Add(cast.ToString(expectedWorkspace),
		cast.ToString(expectedFlow), cast.ToString(expectedNode), cast.ToString(expectedDs)))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`add data store:[\s\t]*`+cast.ToString(expectedDs)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+cast.ToString(expectedWorkspace)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace flow id:[\s\t]*`+cast.ToString(expectedFlow)), actual)
	assert.Regexp(t, regexp.MustCompile(`node id:[\s\t]*`+cast.ToString(expectedNode)), actual)
}

func TestDataStoreExecutor_PretendAdd(t *testing.T) {
	var calledNodeResolved, calledStoreResolved, calledStoreAdd bool
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataStoreAddOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskPretendStub(&calledNodeResolved),
		},
		addDataStore: "myDs",

		dataStoreResolveTaskFactory: newDataStoreResolveTaskPretendStub(&calledStoreResolved),
		nodeStoreAddTaskFactory:     newWorkspaceNodeStoreTaskPretendStub(&calledStoreAdd),
	}

	testExecutor.Pretend()
	assert.True(t, calledNodeResolved)
	assert.True(t, calledStoreResolved)
	assert.True(t, calledStoreAdd)
}

func TestDataStoreExecutor_PretendRemove(t *testing.T) {
	var calledNodeResolved, calledStoreResolved, calledStoreRm bool
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataStoreRmOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskPretendStub(&calledNodeResolved),
		},
		rmDataStore: "myDs",

		dataStoreResolveTaskFactory: newDataStoreResolveTaskPretendStub(&calledStoreResolved),
		nodeStoreRemoveTaskFactory:  newWorkspaceNodeStoreTaskPretendStub(&calledStoreRm),
	}

	testExecutor.Pretend()
	assert.True(t, calledNodeResolved)
	assert.True(t, calledStoreResolved)
	assert.True(t, calledStoreRm)
}

func TestDataStoreExecutor_ProceedAdd(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataStoreParams broker.DataStoreResolveParams
	var captureNodeStoreParams broker.WorkspaceNodeStoreParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataStoreAddOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{printer: testPrinter},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		addDataStore: "myDs",
		workspace:    cast.ToString(expectedWorkspaceId),
		flow:         cast.ToString(expectedFlowId),

		dataStoreResolveTaskFactory: newDataStoreResolveTaskProceedCapture(&captureDataStoreParams),
		nodeStoreAddTaskFactory:     newWorkspaceNodeStoreTaskProceedCapture(&captureNodeStoreParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.addDataStore, captureDataStoreParams.DataStoreName)
	assert.Equal(t, testExecutor.addDataStore, captureNodeStoreParams.DataStoreName)
}

func TestDataStoreExecutor_ProceedAdd_DefaultPrinter(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataStoreParams broker.DataStoreResolveParams
	var captureNodeStoreParams broker.WorkspaceNodeStoreParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{defaultPrinter: true},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		addDataStore: "myDs",
		workspace:    cast.ToString(expectedWorkspaceId),
		flow:         cast.ToString(expectedFlowId),

		dataStoreResolveTaskFactory: newDataStoreResolveTaskProceedCapture(&captureDataStoreParams),
		nodeStoreAddTaskFactory:     newWorkspaceNodeStoreTaskProceedCapture(&captureNodeStoreParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.addDataStore, captureDataStoreParams.DataStoreName)
	assert.Equal(t, testExecutor.addDataStore, captureNodeStoreParams.DataStoreName)
}

func TestDataStoreExecutor_ProceedRemove(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataStoreParams broker.DataStoreResolveParams
	var captureNodeStoreParams broker.WorkspaceNodeStoreParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataStoreAddOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{printer: testPrinter},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		rmDataStore: "myDs",
		workspace:   cast.ToString(expectedWorkspaceId),
		flow:        cast.ToString(expectedFlowId),

		dataStoreResolveTaskFactory: newDataStoreResolveTaskProceedCapture(&captureDataStoreParams),
		nodeStoreRemoveTaskFactory:  newWorkspaceNodeStoreTaskProceedCapture(&captureNodeStoreParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.rmDataStore, captureDataStoreParams.DataStoreName)
	assert.Equal(t, testExecutor.rmDataStore, captureNodeStoreParams.DataStoreName)
}

func TestDataStoreExecutor_ProceedRemove_DefaultPrinter(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataStoreParams broker.DataStoreResolveParams
	var captureNodeStoreParams broker.WorkspaceNodeStoreParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var expectedDataStoreId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{defaultPrinter: true},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		rmDataStore: cast.ToString(expectedDataStoreId),
		workspace:   cast.ToString(expectedWorkspaceId),
		flow:        cast.ToString(expectedFlowId),

		dataStoreResolveTaskFactory: newDataStoreResolveTaskProceedCapture(&captureDataStoreParams),
		nodeStoreRemoveTaskFactory:  newWorkspaceNodeStoreTaskProceedCapture(&captureNodeStoreParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, expectedDataStoreId, *captureDataStoreParams.DataStoreId)
	assert.Equal(t, expectedDataStoreId, *captureDataStoreParams.DataStoreId)
}

func TestDataStoreExecutor_Remove(t *testing.T) {
	var expectedWorkspace = "myWorkspace"
	var expectedFlow = "myFlow"
	var expectedNode = "myNode"
	var expectedDs = "myDs"
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewDataStoreRmOptions()
	var testExecutor = &DataStoreExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Cli:    testCli,
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,
			accountParams:       config.NewAccountParams(testLedger, testOptions.optionAccount),
		},
	}

	assert.NoError(t, testExecutor.Remove(cast.ToString(expectedWorkspace),
		cast.ToString(expectedFlow), cast.ToString(expectedNode), cast.ToString(expectedDs)))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`rm data store:[\s\t]*`+cast.ToString(expectedDs)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+cast.ToString(expectedWorkspace)), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+cast.ToString(expectedFlow)), actual)
	assert.Regexp(t, regexp.MustCompile(`smart function handle:[\s\t]*`+cast.ToString(expectedNode)), actual)
}

func TestNewDataStoreExecutor_addFactory(t *testing.T) {
	var testOptions = NewDataStoreAddOptions()
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()

	assert.NotEmpty(t, newStoreAddFactory(testLedger, nil, testOptions)(&cobra.Command{}))
}

func TestNewDataStoreExecutor_rmFactory(t *testing.T) {
	var testOptions = NewDataStoreRmOptions()
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()

	assert.NotEmpty(t, newStoreRemoveFactory(testLedger, nil, testOptions)(&cobra.Command{}))
}

func newDataStoreResolveTaskPretendStub(flag *bool) DataStoreResolveTaskFactory {
	return func() *task.Task[broker.DataStoreResolveParams] {
		return &task.Task[broker.DataStoreResolveParams]{
			OnPrepare: func(params *broker.DataStoreResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataStoreResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newDataStoreResolveTaskProceedCapture(capture *broker.DataStoreResolveParams) DataStoreResolveTaskFactory {
	return func() *task.Task[broker.DataStoreResolveParams] {
		return &task.Task[broker.DataStoreResolveParams]{
			OnPrepare: func(params *broker.DataStoreResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataStoreResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeStoreTaskPretendStub(flag *bool) func() *task.Task[broker.WorkspaceNodeStoreParams] {
	return func() *task.Task[broker.WorkspaceNodeStoreParams] {
		return &task.Task[broker.WorkspaceNodeStoreParams]{
			OnPrepare: func(params *broker.WorkspaceNodeStoreParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeStoreParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newWorkspaceNodeStoreTaskProceedCapture(capture *broker.WorkspaceNodeStoreParams) func() *task.Task[broker.WorkspaceNodeStoreParams] {
	return func() *task.Task[broker.WorkspaceNodeStoreParams] {
		return &task.Task[broker.WorkspaceNodeStoreParams]{
			OnPrepare: func(params *broker.WorkspaceNodeStoreParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeStoreParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
