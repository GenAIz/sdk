package ws

import (
	"bytes"
	"io"
	"regexp"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserWorkspacesFacade struct {
	getWorkspaces []mgmt.UserWorkspace
	getError      task.Error
	filter        string
	logger        *logrus.Logger
	params        *broker.WorkspaceListParams
}

func (s *stubUserWorkspacesFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkspace] {
	s.filter = filter
	return s
}

func (s *stubUserWorkspacesFacade) Get() ([]mgmt.UserWorkspace, task.Error) {
	return s.getWorkspaces, s.getError
}

func (s *stubUserWorkspacesFacade) Provider() mgmt.Provider[[]mgmt.UserWorkspace] {
	return s
}

func (s *stubUserWorkspacesFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkspace, broker.WorkspaceListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkspacesFacade) WithParams(params *broker.WorkspaceListParams) mgmt.Facade[[]mgmt.UserWorkspace, broker.WorkspaceListParams] {
	s.params = params
	return s
}

func TestListExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`owner-only:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`rc-enabled:[\s\t]*false`), actual)
}

func TestListExecutor_Display_WithAccount(t *testing.T) {
	var expectedAccount = "account"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,
	}

	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`account:[\s\t]*`+expectedAccount), actual)
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`owner-only:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`rc-enabled:[\s\t]*false`), actual)
}

func TestListExecutor_Display_WithMonthly(t *testing.T) {
	var expectedDate = timez.NewMonthTime()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,
	}

	testViper.Set(testOptions.optionOwnerOnly.Key, true)
	testViper.Set(testOptions.optionDateMonthly.Key, true)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`from:[\s\t]*`+expectedDate.Format(time.DateTime)), actual)
	assert.Regexp(t, regexp.MustCompile(`owner-only:[\s\t]*true`), actual)
	assert.Regexp(t, regexp.MustCompile(`rc-enabled:[\s\t]*false`), actual)
}

func TestListExecutor_Display_WithToday(t *testing.T) {
	var expectedDate = timez.NewTodayTime()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,
	}

	testViper.Set(testOptions.optionDateToday.Key, true)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`from:[\s\t]*`+expectedDate.Format(time.DateTime)), actual)
	assert.Regexp(t, regexp.MustCompile(`owner-only:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`rc-enabled:[\s\t]*false`), actual)
}

func TestListExecutor_Display_WithWeekly(t *testing.T) {
	var expectedDate = timez.NewWeekTime()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,
	}

	testViper.Set(testOptions.optionDateWeekly.Key, true)
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`json:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`from:[\s\t]*`+expectedDate.Format(time.DateTime)), actual)
	assert.Regexp(t, regexp.MustCompile(`owner-only:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(`rc-enabled:[\s\t]*false`), actual)
}

func TestListExecutor_Pretend(t *testing.T) {
	var calledParams broker.WorkspaceListParams
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOptions = NewListOptions()
	var expectedAccount = "testAccount"
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams:            config.NewAccountParams(testLedger, testOptions.optionAccount),
		workspaceListTaskFactory: newWorkspaceListTaskPretendCapture(&calledParams),
	}

	testViper.Set(testOptions.optionRcEnabled.Key, true)
	testViper.Set(testOptions.optionOwnerOnly.Key, true)
	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.NotNil(t, calledParams)
	assert.Equal(t, expectedAccount, calledParams.HostAddr)
	assert.Equal(t, testLedger.AuthFile, calledParams.AuthFile)
	assert.True(t, calledParams.OwnerOnly)
	assert.True(t, calledParams.RcEnabled)
}

func TestListExecutor_Proceed(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPrinter = &stubPrinter{}
	var testWorkspacesFacade = &stubUserWorkspacesFacade{
		getWorkspaces: []mgmt.UserWorkspace{
			{
				Id:   37,
				Name: "expectedName",
			},
		},
	}
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		printerParams: &stubPrinterParametric{printer: testPrinter},
		workspaceFacadeProvider: func() mgmt.UserWorkspacesFacade {
			return testWorkspacesFacade
		},
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printOut.([]mgmt.UserWorkspace); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testWorkspacesFacade.getWorkspaces[0].Id, actual[0].Id)
		assert.Equal(t, testWorkspacesFacade.getWorkspaces[0].Name, actual[0].Name)
	} else {
		assert.Fail(t, "did not receive the workspaces")
	}
}

func TestListExecutor_Proceed_Error(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPrinter = &stubPrinter{}
	var testWorkspacesFacade = &stubUserWorkspacesFacade{
		getError: task.NewError("expected"),
	}
	var testOptions = NewListOptions()
	var testExecutor = &ListExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		ListOptions: testOptions,

		accountParams: config.NewAccountParams(testLedger, testOptions.optionAccount),
		printerParams: &stubPrinterParametric{printer: testPrinter},
		workspaceFacadeProvider: func() mgmt.UserWorkspacesFacade {
			return testWorkspacesFacade
		},
	}

	testExecutor.Proceed()

	if actual, ok := testPrinter.printError.(task.Error); ok {
		assert.ErrorIs(t, actual, testWorkspacesFacade.getError)
	} else {
		assert.Fail(t, "did not receive the error")
	}
}

func TestNewList(t *testing.T) {
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
	var testOptions = NewListOptions()
	var testCmd = NewList(testLedger, testCli)

	testViper.Set(testOptions.optionOwnerOnly.Key, false)
	testCmd.Run(testCmd, []string{})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOwnerOnly.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionRcEnabled.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionJsonPrinter.Param+`:[\s\t]*false`), actual)
}

func newWorkspaceListTaskPretendCapture(capture *broker.WorkspaceListParams) func() *task.Task[broker.WorkspaceListParams] {
	return func() *task.Task[broker.WorkspaceListParams] {
		return &task.Task[broker.WorkspaceListParams]{
			OnPrepare: func(params *broker.WorkspaceListParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *broker.WorkspaceListParams, state *task.State) error {
				*capture = *params
				return nil
			},
		}
	}
}
