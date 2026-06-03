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
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubPrinter struct {
	err        error
	printError interface{}
	printOut   interface{}
}

func (s *stubPrinter) Error(i interface{}) error {
	s.printError = i
	return s.err
}

func (s *stubPrinter) Print(i interface{}) error {
	s.printOut = i
	return s.err
}

type stubPrinterParametric struct {
	defaultPrinter bool
	printer        cli.Printer
}

func (s stubPrinterParametric) IsDefault() bool {
	return s.defaultPrinter
}

func (s stubPrinterParametric) Printer() cli.Printer {
	return s.printer
}

func TestCreateExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewCreateOptions()
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,
		workspaceName: expectedName,
	}

	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityOrg)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDisallowRc.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionJsonPrinter.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVisibility.Param+`:[\s\t]*`+broker.WorkspaceVisibilityOrg), actual)
}

func TestCreateExecutor_Display_WithAccount(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewCreateOptions()
	var expectedAccount = "expectedAccount"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,
		workspaceName: expectedName,
	}

	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityOrg)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionAccount.Param+`:[\s\t]*`+expectedAccount), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDisallowRc.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionJsonPrinter.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVisibility.Param+`:[\s\t]*`+broker.WorkspaceVisibilityOrg), actual)
}

func TestCreateExecutor_Pretend(t *testing.T) {
	var calledParams broker.WorkspaceCreateParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOptions = NewCreateOptions()
	var expectedAccount = "testAccount"
	var expectedName = "testName"
	var expectedDescription = "testDescription"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		accountParams:              config.NewAccountParams(testLedger, testOptions.optionAccount),
		workspaceName:              expectedName,
		workspaceCreateTaskFactory: newWorkspaceCreateTaskPretendCapture(&calledParams),
	}

	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityPrivate)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotNil(t, calledParams)
	assert.Equal(t, expectedAccount, calledParams.HostAddr)
	assert.Equal(t, testLedger.AuthFile, calledParams.AuthFile)
	assert.Equal(t, expectedName, calledParams.Workspace.Name)
	assert.True(t, calledParams.Workspace.RcEnabled)
	assert.Equal(t, expectedDescription, calledParams.Workspace.Description)
	assert.Equal(t, broker.WorkspaceVisibilityPrivate, calledParams.Workspace.Visibility)
}

func TestCreateExecutor_Proceed_DefaultPrinter(t *testing.T) {
	var calledParams broker.WorkspaceCreateParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var expectedId = int64(37)
	var testOptions = NewCreateOptions()
	var expectedAccount = "testAccount"
	var expectedName = "testName"
	var expectedDescription = "testDescription"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		printerParams: cli.NewPrinterParam(testLedger, testOptions.optionJsonPrinter),
		workspaceName: expectedName,
		workspaceCreateTaskFactory: newWorkspaceCreateTaskCompleteCapture(&calledParams, &broker.Workspace{
			Id: expectedId,
		}),
	}

	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityPrivate)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, calledParams)
	assert.Equal(t, expectedAccount, calledParams.HostAddr)
	assert.Equal(t, testLedger.AuthFile, calledParams.AuthFile)
	assert.Equal(t, expectedName, calledParams.Workspace.Name)
	assert.True(t, calledParams.Workspace.RcEnabled)
	assert.Equal(t, expectedDescription, calledParams.Workspace.Description)
	assert.Equal(t, broker.WorkspaceVisibilityPrivate, calledParams.Workspace.Visibility)
}

func TestCreateExecutor_Proceed_JsonPrinter(t *testing.T) {
	var calledParams broker.WorkspaceCreateParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var expectedId = int64(37)
	var testOptions = NewCreateOptions()
	var testPrinter = &stubPrinter{}
	var expectedName = "testName"
	var testExecutor = &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		CreateOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		printerParams: &stubPrinterParametric{printer: testPrinter},
		workspaceName: expectedName,
		workspaceCreateTaskFactory: newWorkspaceCreateTaskCompleteCapture(&calledParams, &broker.Workspace{
			Id:   expectedId,
			Name: expectedName,
		}),
	}

	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityOrg)
	testLedger.InitLogging()
	testExecutor.Proceed()
	assert.NotNil(t, calledParams)
	assert.Equal(t, testLedger.AuthFile, calledParams.AuthFile)
	assert.Equal(t, expectedName, calledParams.Workspace.Name)
	assert.True(t, calledParams.Workspace.RcEnabled)
	assert.Equal(t, broker.WorkspaceVisibilityOrg, calledParams.Workspace.Visibility)

	if workspace, ok := testPrinter.printOut.(*broker.Workspace); ok {
		assert.Equal(t, expectedId, workspace.Id)
		assert.Equal(t, expectedName, workspace.Name)
	} else {
		assert.Fail(t, "not a workspace")
	}
}

func TestNewCreate(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		WithOutput(testOutput).
		Build()
	var testCli = NewWsCli(nil, func(ledger *config.Ledger) bool {
		return true
	}, nil)
	var testOptions = NewCreateOptions()
	var testCmd = NewCreate(testLedger, testCli, &Validation{})
	var expectedDescription = "description"
	var expectedName = "name"

	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionVisibility.Key, broker.WorkspaceVisibilityOrg)
	testCmd.Run(testCmd, []string{expectedName})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`name:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDisallowRc.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionJsonPrinter.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVisibility.Param+`:[\s\t]*`+broker.WorkspaceVisibilityOrg), actual)
}

func newWorkspaceCreateTaskCompleteCapture(capture *broker.WorkspaceCreateParams, expected *broker.Workspace) func() *task.Task[broker.WorkspaceCreateParams] {
	return func() *task.Task[broker.WorkspaceCreateParams] {
		return &task.Task[broker.WorkspaceCreateParams]{
			OnPrepare: func(params *broker.WorkspaceCreateParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceCreateParams, state *task.State) error {
				*capture = *params
				state.Internal = expected
				return nil
			},
		}
	}
}

func newWorkspaceCreateTaskPretendCapture(capture *broker.WorkspaceCreateParams) func() *task.Task[broker.WorkspaceCreateParams] {
	return func() *task.Task[broker.WorkspaceCreateParams] {
		return &task.Task[broker.WorkspaceCreateParams]{
			OnPrepare: func(params *broker.WorkspaceCreateParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceCreateParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
