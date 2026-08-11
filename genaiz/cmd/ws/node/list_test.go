package node

import (
	"errors"
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubListExecutor struct {
	workspaceArg string
	flowArg      string
	listError    error
}

func (sle *stubListExecutor) List(workspaceArg string, flowArg string) error {
	sle.workspaceArg = workspaceArg
	sle.flowArg = flowArg
	return sle.listError
}

type stubUserWorkspaceFlowsFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.WorkspaceFlowListParams
	provider mgmt.Provider[[]mgmt.UserWorkspaceFlow]
}

func (s *stubUserWorkspaceFlowsFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkspaceFlow] {
	s.filter = filter
	return s.provider
}

func (s *stubUserWorkspaceFlowsFacade) Provider() mgmt.Provider[[]mgmt.UserWorkspaceFlow] {
	return s.provider
}

func (s *stubUserWorkspaceFlowsFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkspaceFlow, broker.WorkspaceFlowListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkspaceFlowsFacade) WithParams(params *broker.WorkspaceFlowListParams) mgmt.Facade[[]mgmt.UserWorkspaceFlow, broker.WorkspaceFlowListParams] {
	s.params = params
	return s
}

type stubUserWorkspaceFlowsProvider struct {
	flows    []mgmt.UserWorkspaceFlow
	getError task.Error
}

func (s stubUserWorkspaceFlowsProvider) Get() ([]mgmt.UserWorkspaceFlow, task.Error) {
	return s.flows, s.getError
}

type stubWorkspaceBridge struct {
	completions []cobra.Completion
	directive   cobra.ShellCompDirective
}

func (s stubWorkspaceBridge) Bridge(string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return s.completions, s.directive
}

func TestNewList(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &stubListExecutor{}
	var testExecutorFactory = func(command *cobra.Command) ListExecutor { return testExecutor }
	var testList = NewList(testLedger, testOptions, testExecutorFactory)
	var expectedWorkspaceArg = "workspaceName"
	var expectedFlowArg = "flowId"

	testList.Run(testList, []string{expectedWorkspaceArg, expectedFlowArg})
	assert.Equal(t, expectedWorkspaceArg, testExecutor.workspaceArg)
	assert.Equal(t, expectedFlowArg, testExecutor.flowArg)
}

func TestNewList_Error(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &stubListExecutor{
		listError: errors.New("expected"),
	}
	var testExecutorFactory = func(command *cobra.Command) ListExecutor { return testExecutor }
	var testList = NewList(testLedger, testOptions, testExecutorFactory)

	defer patch.Unpatch()
	testList.Run(testList, []string{"workspace", "flow"})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewListAuto_bridgeArguments(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCmd = &cobra.Command{}
	var testOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testAuto = NewListAuto(testLedger, testOption)
	var testArgs = []string{"1", "2", "3", "4"}

	if actual, directive := testAuto.bridgeArguments(testCmd, testArgs, ""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewListAuto_bridgeFlows(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCmd = &cobra.Command{}
	var testProvider = &stubUserWorkspaceFlowsProvider{}
	var testAuto = &ListAutoBridge{
		ledger:      testLedger,
		readyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}
	var testArgs = []string{"workspaceName"}

	if actual, directive := testAuto.bridgeArguments(testCmd, testArgs, ""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewListAuto_bridgeFlows_GetError(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCmd = &cobra.Command{}
	var testProvider = &stubUserWorkspaceFlowsProvider{
		getError: task.NewError("expected"),
	}
	var testAuto = &ListAutoBridge{
		ledger:      testLedger,
		readyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}
	var testArgs = []string{"workspaceName"}

	if actual, directive := testAuto.bridgeArguments(testCmd, testArgs, ""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestNewListAuto_bridgeFlows_WorkspaceId(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCmd = &cobra.Command{}
	var testProvider = &stubUserWorkspaceFlowsProvider{
		flows: []mgmt.UserWorkspaceFlow{
			{
				Id:             1337,
				WorkflowHandle: "handle",
			},
		},
	}
	var testAuto = &ListAutoBridge{
		ledger:      testLedger,
		readyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}
	var testArgs = []string{"37"}

	if actual, directive := testAuto.bridgeArguments(testCmd, testArgs, ""); len(actual) == 0 {
		assert.Fail(t, "expected results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
		assert.Equal(t, fmt.Sprintf("%d\t%s", testProvider.flows[0].Id, testProvider.flows[0].WorkflowHandle), actual[0])
	}
}

func TestNewListAuto_bridgeWorkspaces(t *testing.T) {
	var expectedCompletion = "expected"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testAuto = &ListAutoBridge{
		ledger: testLedger,
		workspaces: &stubWorkspaceBridge{
			completions: []string{expectedCompletion},
			directive:   cobra.ShellCompDirectiveKeepOrder,
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "37"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{}, testCompletable); len(actual) > 0 {
		assert.Contains(t, actual, expectedCompletion)
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual workspace results")
	}
}
