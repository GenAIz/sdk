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
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubAutoBridge struct {
	directive     cobra.ShellCompDirective
	readyOnly     *config.BoolOption
	results       []cobra.Completion
	toComplete    string
	workspace     string
	workspaceFlow string
}

func (s *stubAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	s.toComplete = toComplete
	return s.results, s.directive
}

func (s *stubAutoBridge) WithReadyOnly(option *config.BoolOption) auto.WorkspaceBridge {
	s.readyOnly = option
	return s
}

func (s *stubAutoBridge) WithWorkspace(workspace string) auto.WorkspaceBridge {
	s.workspace = workspace
	return s
}

func (s *stubAutoBridge) WithWorkspaceFlow(workspace, flow string) auto.WorkspaceFlowBridge {
	s.workspace = workspace
	s.workspaceFlow = flow
	return s
}

func TestDataSourceAutoBridge_bridgeArguments(t *testing.T) {
	var stubBridge = &stubAutoBridge{}
	var testCmd = &cobra.Command{}
	var testBridge = &DataSourceAutoBridge{
		dataSources: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{"workspace", "flow", "node", "ds"}, "toComplete")
	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	assert.Empty(t, stubBridge.toComplete)
}

func TestDataSourceAutoBridge_bridgeArguments_dataSources(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataSourceAutoBridge{
		dataSources: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{"workspace", "flow", "node"}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataSourceAutoBridge_bridgeArguments_workspaces(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataSourceAutoBridge{
		workspaces: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataSourceAutoBridge_bridgeArguments_workspaceFlows(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedWorkspace = "workspace"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataSourceAutoBridge{
		workspaceFlows: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{expectedWorkspace}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubBridge.workspace)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataSourceAutoBridge_bridgeArguments_workspaceNodes(t *testing.T) {
	var expectedToComplete = "toComplete"
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var stubBridge = &stubAutoBridge{
		directive: expectedDirective,
	}
	var testCmd = &cobra.Command{}
	var testBridge = &DataSourceAutoBridge{
		workspaceNodes: stubBridge,
	}

	actual, directive := testBridge.bridgeArguments(testCmd, []string{expectedWorkspace, expectedFlow}, expectedToComplete)
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubBridge.workspace)
	assert.Equal(t, expectedFlow, stubBridge.workspaceFlow)
	assert.Equal(t, expectedToComplete, stubBridge.toComplete)
}

func TestDataSourceExecutor_Add(t *testing.T) {
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
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataSourceExecutor{
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
	assert.Regexp(t, regexp.MustCompile(`add data source:[\s\t]*`+cast.ToString(expectedDs)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace id:[\s\t]*`+cast.ToString(expectedWorkspace)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace flow id:[\s\t]*`+cast.ToString(expectedFlow)), actual)
	assert.Regexp(t, regexp.MustCompile(`node id:[\s\t]*`+cast.ToString(expectedNode)), actual)
}

func TestDataSourceExecutor_PretendAdd(t *testing.T) {
	var calledNodeResolved, calledSourceResolved, calledSourceAdd bool
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskPretendStub(&calledNodeResolved),
		},
		addDataSource: "myDs",

		dataSourceResolveTaskFactory: newDataSourceResolveTaskPretendStub(&calledSourceResolved),
		nodeSourceAddTaskFactory:     newWorkspaceNodeSourceTaskPretendStub(&calledSourceAdd),
	}

	testExecutor.Pretend()
	assert.True(t, calledNodeResolved)
	assert.True(t, calledSourceResolved)
	assert.True(t, calledSourceAdd)
}

func TestDataSourceExecutor_PretendRemove(t *testing.T) {
	var calledNodeResolved, calledSourceResolved, calledSourceRm bool
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceRmOptions()
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskPretendStub(&calledNodeResolved),
		},
		rmDataSource: "myDs",

		dataSourceResolveTaskFactory: newDataSourceResolveTaskPretendStub(&calledSourceResolved),
		nodeSourceRemoveTaskFactory:  newWorkspaceNodeSourceTaskPretendStub(&calledSourceRm),
	}

	testExecutor.Pretend()
	assert.True(t, calledNodeResolved)
	assert.True(t, calledSourceResolved)
	assert.True(t, calledSourceRm)
}

func TestDataSourceExecutor_ProceedAdd(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataSourceParams broker.DataSourceResolveParams
	var captureNodeSourceParams broker.WorkspaceNodeSourceParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{printer: testPrinter},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		addDataSource: "myDs",
		workspace:     cast.ToString(expectedWorkspaceId),
		flow:          cast.ToString(expectedFlowId),

		dataSourceResolveTaskFactory: newDataSourceResolveTaskProceedCapture(&captureDataSourceParams),
		nodeSourceAddTaskFactory:     newWorkspaceNodeSourceTaskProceedCapture(&captureNodeSourceParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.addDataSource, captureDataSourceParams.DataSourceName)
	assert.Equal(t, testExecutor.addDataSource, captureNodeSourceParams.DataSourceName)
}

func TestDataSourceExecutor_ProceedAdd_DefaultPrinter(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataSourceParams broker.DataSourceResolveParams
	var captureNodeSourceParams broker.WorkspaceNodeSourceParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{defaultPrinter: true},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		addDataSource: "myDs",
		workspace:     cast.ToString(expectedWorkspaceId),
		flow:          cast.ToString(expectedFlowId),

		dataSourceResolveTaskFactory: newDataSourceResolveTaskProceedCapture(&captureDataSourceParams),
		nodeSourceAddTaskFactory:     newWorkspaceNodeSourceTaskProceedCapture(&captureNodeSourceParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.addDataSource, captureDataSourceParams.DataSourceName)
	assert.Equal(t, testExecutor.addDataSource, captureNodeSourceParams.DataSourceName)
}

func TestDataSourceExecutor_ProceedRemove(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataSourceParams broker.DataSourceResolveParams
	var captureNodeSourceParams broker.WorkspaceNodeSourceParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{printer: testPrinter},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		rmDataSource: "myDs",
		workspace:    cast.ToString(expectedWorkspaceId),
		flow:         cast.ToString(expectedFlowId),

		dataSourceResolveTaskFactory: newDataSourceResolveTaskProceedCapture(&captureDataSourceParams),
		nodeSourceRemoveTaskFactory:  newWorkspaceNodeSourceTaskProceedCapture(&captureNodeSourceParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, testExecutor.rmDataSource, captureDataSourceParams.DataSourceName)
	assert.Equal(t, testExecutor.rmDataSource, captureNodeSourceParams.DataSourceName)
}

func TestDataSourceExecutor_ProceedRemove_DefaultPrinter(t *testing.T) {
	var capturedNodeResolveParams broker.WorkspaceNodeResolveParams
	var captureDataSourceParams broker.DataSourceResolveParams
	var captureNodeSourceParams broker.WorkspaceNodeSourceParams
	var expectedWorkspaceId = int64(37)
	var expectedFlowId = int64(73)
	var expectedDataSourceId = int64(42)
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewDataSourceAddOptions()
	var testExecutor = &DataSourceExecutor{
		DataInstanceExecutor: DataInstanceExecutor{
			BaseExecutor: BaseExecutor{
				Ledger: testLedger,
			},
			DataInstanceOptions: testOptions,

			accountParams:          config.NewAccountParams(testLedger, testOptions.optionAccount),
			printerParams:          &stubPrinterParametric{defaultPrinter: true},
			nodeResolveTaskFactory: newWorkspaceNodeResolveTaskProceedCapture(&capturedNodeResolveParams),
		},
		rmDataSource: cast.ToString(expectedDataSourceId),
		workspace:    cast.ToString(expectedWorkspaceId),
		flow:         cast.ToString(expectedFlowId),

		dataSourceResolveTaskFactory: newDataSourceResolveTaskProceedCapture(&captureDataSourceParams),
		nodeSourceRemoveTaskFactory:  newWorkspaceNodeSourceTaskProceedCapture(&captureNodeSourceParams),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.Equal(t, expectedWorkspaceId, *capturedNodeResolveParams.WorkspaceId)
	assert.Equal(t, expectedFlowId, *capturedNodeResolveParams.FlowId)
	assert.Equal(t, expectedDataSourceId, *captureDataSourceParams.DataSourceId)
	assert.Equal(t, expectedDataSourceId, *captureDataSourceParams.DataSourceId)
}

func TestDataSourceExecutor_Remove(t *testing.T) {
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
	var testOptions = NewDataSourceRmOptions()
	var testExecutor = &DataSourceExecutor{
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
	assert.Regexp(t, regexp.MustCompile(`rm data source:[\s\t]*`+cast.ToString(expectedDs)), actual)
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+cast.ToString(expectedWorkspace)), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+cast.ToString(expectedFlow)), actual)
	assert.Regexp(t, regexp.MustCompile(`smart function handle:[\s\t]*`+cast.ToString(expectedNode)), actual)
}

func TestNewDataSourceExecutor_addFactory(t *testing.T) {
	var testOptions = NewDataSourceAddOptions()
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()

	assert.NotEmpty(t, newSourceAddFactory(testLedger, nil, testOptions)(&cobra.Command{}))
}

func TestNewDataSourceExecutor_rmFactory(t *testing.T) {
	var testOptions = NewDataSourceRmOptions()
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()

	assert.NotEmpty(t, newSourceRemoveFactory(testLedger, nil, testOptions)(&cobra.Command{}))
}

func newDataSourceResolveTaskPretendStub(flag *bool) DataSourceResolveTaskFactory {
	return func() *task.Task[broker.DataSourceResolveParams] {
		return &task.Task[broker.DataSourceResolveParams]{
			OnPrepare: func(params *broker.DataSourceResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataSourceResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newDataSourceResolveTaskProceedCapture(capture *broker.DataSourceResolveParams) DataSourceResolveTaskFactory {
	return func() *task.Task[broker.DataSourceResolveParams] {
		return &task.Task[broker.DataSourceResolveParams]{
			OnPrepare: func(params *broker.DataSourceResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataSourceResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeResolveTaskPretendStub(flag *bool) WorkspaceNodeResolveTaskFactory {
	return func() *task.Task[broker.WorkspaceNodeResolveParams] {
		return &task.Task[broker.WorkspaceNodeResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newWorkspaceNodeSourceTaskPretendStub(flag *bool) func() *task.Task[broker.WorkspaceNodeSourceParams] {
	return func() *task.Task[broker.WorkspaceNodeSourceParams] {
		return &task.Task[broker.WorkspaceNodeSourceParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSourceParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeSourceParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newWorkspaceNodeResolveTaskProceedCapture(capture *broker.WorkspaceNodeResolveParams) WorkspaceNodeResolveTaskFactory {
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

func newWorkspaceNodeSourceTaskProceedCapture(capture *broker.WorkspaceNodeSourceParams) func() *task.Task[broker.WorkspaceNodeSourceParams] {
	return func() *task.Task[broker.WorkspaceNodeSourceParams] {
		return &task.Task[broker.WorkspaceNodeSourceParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSourceParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeSourceParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
