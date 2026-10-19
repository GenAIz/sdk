package ws

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/args"
	"genaiz.com/genaiz/cmd/ws/data"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/workspaces"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubWorkspaceNodeDataSetFactory struct {
	mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []workspaces.UserWorkspaceDataSet]
}

type stubWorkspaceNodeDataSetProvider struct {
	dataSets []workspaces.UserWorkspaceDataSet
	getError task.Error
}

func (s stubWorkspaceNodeDataSetProvider) Get() ([]workspaces.UserWorkspaceDataSet, task.Error) {
	return s.dataSets, s.getError
}

func TestDataListExecutor_Display(t *testing.T) {
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(viper.New()).
		Build()
	var testOptions = data.NewDataListOptions()
	var testExecutor = &DataListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedFlow, expectedNode),
	}

	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspace), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedFlow), actual)
	assert.Regexp(t, regexp.MustCompile(`function handle:[\s\t]*`+expectedNode), actual)
}

func TestDataListExecutor_List(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(viper.New()).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = data.NewDataListOptions()
	var testFactory = newDataListExecutorFactory(testLedger, testCli, testOptions)
	var testExecutor = testFactory(&cobra.Command{})
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedNode = "node"

	assert.NoError(t, testExecutor.List(expectedWorkspace, expectedFlow, expectedNode))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspace), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedFlow), actual)
	assert.Regexp(t, regexp.MustCompile(`function handle:[\s\t]*`+expectedNode), actual)
}

func TestDataListExecutor_Pretend(t *testing.T) {
	var capturedList broker.WorkspaceNodeSetListParams
	var capturedResolve broker.WorkspaceNodeResolveParams
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataListOptions()
	var testExecutor = &DataListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		listNodeSetTaskFactory: newWorkspaceNodeSetListPretendCapture(&capturedList),
		resolveNodeTaskFactory: newWorkspaceNodeResolvePretendCapture(&capturedResolve),
	}

	testExecutor.Pretend()
	assert.NotEmpty(t, capturedList)
	assert.NotEmpty(t, capturedResolve)
	assert.Equal(t, expectedWorkspace, capturedResolve.WorkspaceName)
	assert.Equal(t, expectedWorkflow, capturedResolve.WorkflowHandle)
	assert.Equal(t, expectedNodeId, cast.ToString(*capturedResolve.NodeId))
}

func TestDataListExecutor_Proceed(t *testing.T) {
	var expectedDataSetId = int64(73)
	var expectedPort = "input"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataListOptions()
	var testPrinter = &stubPrinter{}
	var testFactory = &stubWorkspaceNodeDataSetFactory{}
	var testProvider = &stubWorkspaceNodeDataSetProvider{
		dataSets: []workspaces.UserWorkspaceDataSet{
			{
				Id:        &expectedDataSetId,
				InputPort: expectedPort,
			},
		},
	}
	var testExecutor = &DataListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		nodeArguments: args.NewNodeResolution("37", "42", "function"),

		workspaceNodeDataSetsFactory: testFactory.WithProvider(
			func(p workspaces.UserWorkspaceDataSetsProviderFactory) workspaces.UserWorkspaceDataSetsProvider {
				return testProvider
			}),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printOut.([]workspaces.UserWorkspaceDataSet); ok {
		assert.Equal(t, testProvider.dataSets, actual)
	} else {
		assert.Fail(t, "expected a slice of UserWorkspaceDataSet")
	}
}

func TestDataListExecutor_Proceed_ProviderError(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataListOptions()
	var testPrinter = &stubPrinter{}
	var testFactory = &stubWorkspaceNodeDataSetFactory{}
	var testProvider = &stubWorkspaceNodeDataSetProvider{
		getError: expectedError,
	}
	var testExecutor = &DataListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},

		nodeArguments: args.NewNodeResolution("37", "42", "function"),

		workspaceNodeDataSetsFactory: testFactory.WithProvider(
			func(p workspaces.UserWorkspaceDataSetsProviderFactory) workspaces.UserWorkspaceDataSetsProvider {
				return testProvider
			}),
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printOut.([]workspaces.UserWorkspaceDataSet); ok {
		assert.Fail(t, "expected an error")
	} else {
		assert.Empty(t, actual)
		assert.Equal(t, expectedError, testPrinter.printError)
	}
}

func TestDataSetExecutor_Display(t *testing.T) {
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var expectedPort = "port"
	var expectedPath = "path"
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(viper.New()).
		Build()
	var testOptions = data.NewDataSetOptions()
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,

		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedFlow, expectedNode),
		port:          expectedPort,
		path:          expectedPath,
	}

	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspace), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedFlow), actual)
	assert.Regexp(t, regexp.MustCompile(`function handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`input port:[\s\t]*`+expectedPort), actual)
	assert.Regexp(t, regexp.MustCompile(`file path:[\s\t]*`+expectedPath), actual)
}

func TestDataSetExecutor_Pretend(t *testing.T) {
	var captureCreate broker.DataSetCreateParams
	var captureListSet broker.DataSetListFilesParams
	var captureResolveSet broker.WorkspaceNodeSetResolveParams
	var captureResolveNode broker.WorkspaceNodeResolveParams
	var captureUpdateSet broker.WorkspaceNodeSetResolveParams
	var captureUploadSet broker.DataSetUploadParams
	var expectedFilePath = filepath.Join(t.TempDir(), "config.json")
	var expectedPort = "port"
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataSetOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,
		path:       expectedFilePath,
		port:       expectedPort,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		createDataSetTaskFactory:  newDataSetCreatePretendCapture(&captureCreate),
		listDataFilesTaskFactory:  newDataSetListFilesPretendCapture(&captureListSet),
		resolveDataSetTaskFactory: newWorkspaceNodeSetResolvePretendCapture(&captureResolveSet),
		resolveNodeTaskFactory:    newWorkspaceNodeResolvePretendCapture(&captureResolveNode),
		updateNodeSetTaskFactory:  newWorkspaceNodeSetUpdatePretendCapture(&captureUpdateSet),
		uploadDataSetTaskFactory:  newDataSetUploadPretendCapture(&captureUploadSet),
	}

	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureListSet)
	assert.NotEmpty(t, captureResolveSet)
	assert.NotEmpty(t, captureResolveNode)
	assert.NotEmpty(t, captureUpdateSet)
	assert.NotEmpty(t, captureUploadSet)
}

func TestDataSetExecutor_Pretend_Overwrite(t *testing.T) {
	var captureCreate broker.DataSetCreateParams
	var captureResolveSet broker.WorkspaceNodeSetResolveParams
	var captureResolveNode broker.WorkspaceNodeResolveParams
	var captureUpdateSet broker.WorkspaceNodeSetResolveParams
	var captureUploadSet broker.DataSetUploadParams
	var expectedFilePath = filepath.Join(t.TempDir(), "config.json")
	var expectedPort = "port"
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = data.NewDataSetOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,
		path:       expectedFilePath,
		port:       expectedPort,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		createDataSetTaskFactory:  newDataSetCreatePretendCapture(&captureCreate),
		resolveDataSetTaskFactory: newWorkspaceNodeSetResolvePretendCapture(&captureResolveSet),
		resolveNodeTaskFactory:    newWorkspaceNodeResolvePretendCapture(&captureResolveNode),
		updateNodeSetTaskFactory:  newWorkspaceNodeSetUpdatePretendCapture(&captureUpdateSet),
		uploadDataSetTaskFactory:  newDataSetUploadPretendCapture(&captureUploadSet),
	}

	testViper.Set(schema.Genaiz.Workspace.Node.Data.Set.Overwrite.Doc, true)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureResolveSet)
	assert.NotEmpty(t, captureResolveNode)
	assert.NotEmpty(t, captureUpdateSet)
	assert.NotEmpty(t, captureUploadSet)
}

func TestDataSetExecutor_Proceed(t *testing.T) {
	var captureCreate broker.DataSetCreateParams
	var captureListSet broker.DataSetListFilesParams
	var captureResolveSet broker.WorkspaceNodeSetResolveParams
	var captureResolveNode broker.WorkspaceNodeResolveParams
	var captureUpdateSet broker.WorkspaceNodeSetResolveParams
	var captureUploadSet broker.DataSetUploadParams
	var expectedFilePath = filepath.Join(t.TempDir(), "config.json")
	var expectedPort = "port"
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataSetOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,
		path:       expectedFilePath,
		port:       expectedPort,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		createDataSetTaskFactory:  newDataSetCreateProceedCapture(&captureCreate),
		listDataFilesTaskFactory:  newDataSetListFilesProceedCapture(&captureListSet),
		resolveDataSetTaskFactory: newWorkspaceNodeSetResolveProceedCapture(&captureResolveSet),
		resolveNodeTaskFactory:    newWorkspaceNodeResolveProceedCapture(&captureResolveNode),
		updateNodeSetTaskFactory:  newWorkspaceNodeSetUpdateProceedCapture(&captureUpdateSet),
		uploadDataSetTaskFactory:  newDataSetUploadProceedCapture(&captureUploadSet),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureListSet)
	assert.NotEmpty(t, captureResolveSet)
	assert.NotEmpty(t, captureResolveNode)
	assert.NotEmpty(t, captureUpdateSet)
	assert.NotEmpty(t, captureUploadSet)
}

func TestDataSetExecutor_Proceed_DefaultPrinter(t *testing.T) {
	var captureCreate broker.DataSetCreateParams
	var captureListSet broker.DataSetListFilesParams
	var captureResolveSet broker.WorkspaceNodeSetResolveParams
	var captureResolveNode broker.WorkspaceNodeResolveParams
	var captureUpdateSet broker.WorkspaceNodeSetResolveParams
	var captureUploadSet broker.DataSetUploadParams
	var expectedFilePath = filepath.Join(t.TempDir(), "config.json")
	var expectedPort = "port"
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = data.NewDataSetOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,
		path:       expectedFilePath,
		port:       expectedPort,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer:        testPrinter,
			defaultPrinter: true,
		},
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		createDataSetTaskFactory:  newDataSetCreateProceedCapture(&captureCreate),
		listDataFilesTaskFactory:  newDataSetListFilesProceedCapture(&captureListSet),
		resolveDataSetTaskFactory: newWorkspaceNodeSetResolveProceedCapture(&captureResolveSet),
		resolveNodeTaskFactory:    newWorkspaceNodeResolveProceedCapture(&captureResolveNode),
		updateNodeSetTaskFactory:  newWorkspaceNodeSetUpdateProceedCapture(&captureUpdateSet),
		uploadDataSetTaskFactory:  newDataSetUploadProceedCapture(&captureUploadSet),
	}

	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureListSet)
	assert.NotEmpty(t, captureResolveSet)
	assert.NotEmpty(t, captureResolveNode)
	assert.NotEmpty(t, captureUpdateSet)
	assert.NotEmpty(t, captureUploadSet)
}

func TestDataSetExecutor_Proceed_Overwrite(t *testing.T) {
	var captureCreate broker.DataSetCreateParams
	var captureResolveSet broker.WorkspaceNodeSetResolveParams
	var captureResolveNode broker.WorkspaceNodeResolveParams
	var captureUpdateSet broker.WorkspaceNodeSetResolveParams
	var captureUploadSet broker.DataSetUploadParams
	var expectedFilePath = filepath.Join(t.TempDir(), "config.json")
	var expectedPort = "port"
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var expectedNodeId = "23"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = data.NewDataSetOptions()
	var testPrinter = &stubPrinter{}
	var testExecutor = &DataSetExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		SetOptions: testOptions,
		path:       expectedFilePath,
		port:       expectedPort,

		accountParams: testOptions.AccountParams(testLedger),
		printerParams: &stubPrinterParametric{
			printer: testPrinter,
		},
		nodeArguments: args.NewNodeResolution(expectedWorkspace, expectedWorkflow, expectedNodeId),

		createDataSetTaskFactory:  newDataSetCreateProceedCapture(&captureCreate),
		resolveDataSetTaskFactory: newWorkspaceNodeSetResolveProceedCapture(&captureResolveSet),
		resolveNodeTaskFactory:    newWorkspaceNodeResolveProceedCapture(&captureResolveNode),
		updateNodeSetTaskFactory:  newWorkspaceNodeSetUpdateProceedCapture(&captureUpdateSet),
		uploadDataSetTaskFactory:  newDataSetUploadProceedCapture(&captureUploadSet),
	}

	testViper.Set(schema.Genaiz.Workspace.Node.Data.Set.Overwrite.Doc, true)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotEmpty(t, captureCreate)
	assert.NotEmpty(t, captureResolveSet)
	assert.NotEmpty(t, captureResolveNode)
	assert.NotEmpty(t, captureUpdateSet)
	assert.NotEmpty(t, captureUploadSet)
}

func TestDataSetExecutor_Set(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "config.json")
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(viper.New()).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = data.NewDataSetOptions()
	var testFactory = newDataSetExecutorFactory(testLedger, testCli, testOptions)
	var testExecutor = testFactory(&cobra.Command{})
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var expectedPort = "37"
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		assert.NoError(t, testExecutor.Set(expectedWorkspace, expectedFlow, expectedNode, expectedPort, testFile))
		actual := testOutput.String()
		assert.Regexp(t, regexp.MustCompile(`workspace name:[\s\t]*`+expectedWorkspace), actual)
		assert.Regexp(t, regexp.MustCompile(`workflow handle:[\s\t]*`+expectedFlow), actual)
		assert.Regexp(t, regexp.MustCompile(`function handle:[\s\t]*`+expectedNode), actual)
		assert.Regexp(t, regexp.MustCompile(`data set id:[\s\t]*`+expectedPort), actual)
		assert.Regexp(t, regexp.MustCompile(`file path:[\s\t]*`+testFile), actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestDataSetExecutor_Set_FileUnreadable(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "config.json")
	var testOutput = new(bytes.Buffer)
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(viper.New()).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testOptions = data.NewDataSetOptions()
	var testFactory = newDataSetExecutorFactory(testLedger, testCli, testOptions)
	var testExecutor = testFactory(&cobra.Command{})
	var expectedWorkspace = "workspace"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var expectedPort = "port"

	assert.Error(t, testExecutor.Set(expectedWorkspace, expectedFlow, expectedNode, expectedPort, testFile))
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestNewData(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCli = &Cli{}
	var testData = NewData(testLedger, testCli)

	assert.Equal(t, 4, len(testData.Commands()))
}

func newDataSetCreatePretendCapture(capture *broker.DataSetCreateParams) broker.DataSetCreateTaskFactory {
	return func() broker.DataSetCreateTask {
		return &task.Task[broker.DataSetCreateParams]{
			OnPrepare: func(params *broker.DataSetCreateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataSetCreateParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newDataSetListFilesPretendCapture(capture *broker.DataSetListFilesParams) broker.DataSetListFilesTaskFactory {
	return func() broker.DataSetListFilesTask {
		return &task.Task[broker.DataSetListFilesParams]{
			OnPrepare: func(params *broker.DataSetListFilesParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataSetListFilesParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newDataSetUploadPretendCapture(capture *broker.DataSetUploadParams) broker.DataSetUploadTaskFactory {
	return func() broker.DataSetUploadTask {
		return &task.Task[broker.DataSetUploadParams]{
			OnPrepare: func(params *broker.DataSetUploadParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.DataSetUploadParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetResolvePretendCapture(capture *broker.WorkspaceNodeSetResolveParams) broker.WorkspaceNodeSetResolveTaskFactory {
	return func() broker.WorkspaceNodeSetResolveTask {
		return &task.Task[broker.WorkspaceNodeSetResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetUpdatePretendCapture(capture *broker.WorkspaceNodeSetResolveParams) broker.WorkspaceNodeSetUpdateTaskFactory {
	return func() broker.WorkspaceNodeSetResolveTask {
		return &task.Task[broker.WorkspaceNodeSetResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newDataSetCreateProceedCapture(capture *broker.DataSetCreateParams) broker.DataSetCreateTaskFactory {
	return func() broker.DataSetCreateTask {
		return &task.Task[broker.DataSetCreateParams]{
			OnPrepare: func(params *broker.DataSetCreateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataSetCreateParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newDataSetListFilesProceedCapture(capture *broker.DataSetListFilesParams) broker.DataSetListFilesTaskFactory {
	return func() broker.DataSetListFilesTask {
		return &task.Task[broker.DataSetListFilesParams]{
			OnPrepare: func(params *broker.DataSetListFilesParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataSetListFilesParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newDataSetUploadProceedCapture(capture *broker.DataSetUploadParams) broker.DataSetUploadTaskFactory {
	return func() broker.DataSetUploadTask {
		return &task.Task[broker.DataSetUploadParams]{
			OnPrepare: func(params *broker.DataSetUploadParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataSetUploadParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetResolveProceedCapture(capture *broker.WorkspaceNodeSetResolveParams) broker.WorkspaceNodeSetResolveTaskFactory {
	return func() broker.WorkspaceNodeSetResolveTask {
		return &task.Task[broker.WorkspaceNodeSetResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetUpdateProceedCapture(capture *broker.WorkspaceNodeSetResolveParams) broker.WorkspaceNodeSetUpdateTaskFactory {
	return func() broker.WorkspaceNodeSetResolveTask {
		return &task.Task[broker.WorkspaceNodeSetResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeSetResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeSetListPretendCapture(capture *broker.WorkspaceNodeSetListParams) broker.WorkspaceNodeSetListTaskFactory {
	return func() broker.WorkspaceNodeSetListTask {
		return &task.Task[broker.WorkspaceNodeSetListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeSetListParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeSetListParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeResolvePretendCapture(capture *broker.WorkspaceNodeResolveParams) broker.WorkspaceNodeResolveTaskFactory {
	return func() broker.WorkspaceNodeResolveTask {
		return &task.Task[broker.WorkspaceNodeResolveParams]{
			OnPrepare: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceNodeResolveParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
