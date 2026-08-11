package flow

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

type stubCreateExecutor struct {
	createSolution  string
	createWorkflow  string
	createWorkspace string
	createError     error
}

func (sce *stubCreateExecutor) Create(workspace, solution, workflow string) error {
	sce.createSolution = solution
	sce.createWorkflow = workflow
	sce.createWorkspace = workspace
	return sce.createError
}

type stubUserSolutionFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.SolutionListParams
	provider mgmt.Provider[[]mgmt.UserSolution]
}

func (s *stubUserSolutionFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserSolution] {
	s.filter = filter
	return s.provider
}

func (s *stubUserSolutionFacade) Provider() mgmt.Provider[[]mgmt.UserSolution] {
	return s.provider
}

func (s *stubUserSolutionFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserSolution, broker.SolutionListParams] {
	s.logger = logger
	return s
}

func (s *stubUserSolutionFacade) WithParams(params *broker.SolutionListParams) mgmt.Facade[[]mgmt.UserSolution, broker.SolutionListParams] {
	s.params = params
	return s
}

type stubUserSolutionProvider struct {
	solutions []mgmt.UserSolution
	getError  task.Error
}

func (s stubUserSolutionProvider) Get() ([]mgmt.UserSolution, task.Error) {
	return s.solutions, s.getError
}

type stubUserWorkflowFacade struct {
	filter   string
	graphers map[string]broker.SolutionGrapher
	logger   *logrus.Logger
	params   *broker.WorkflowListParams
	path     string
	provider mgmt.Provider[[]mgmt.UserWorkflow]
}

func (s *stubUserWorkflowFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkflow] {
	s.filter = filter
	return s.provider
}

func (s *stubUserWorkflowFacade) Provider() mgmt.Provider[[]mgmt.UserWorkflow] {
	return s.provider
}

func (s *stubUserWorkflowFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkflow, broker.WorkflowListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkflowFacade) WithParams(params *broker.WorkflowListParams) mgmt.Facade[[]mgmt.UserWorkflow, broker.WorkflowListParams] {
	s.params = params
	return s
}

func (s *stubUserWorkflowFacade) WithPathGraphers(path string, graphers map[string]broker.SolutionGrapher) mgmt.UserWorkflowFacade {
	s.path = path
	s.graphers = graphers
	return s
}

type stubUserWorkflowProvider struct {
	workflows []mgmt.UserWorkflow
	getError  task.Error
}

func (s stubUserWorkflowProvider) Get() ([]mgmt.UserWorkflow, task.Error) {
	return s.workflows, s.getError
}

type stubWorkspaceBridge struct {
	completions []cobra.Completion
	directive   cobra.ShellCompDirective
}

func (s stubWorkspaceBridge) Bridge(string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return s.completions, s.directive
}

func TestNewCreate(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewCreateOptions()
	var testExecutor = &stubCreateExecutor{}
	var testExecutorFactory = func(command *cobra.Command) CreateExecutor { return testExecutor }
	var testCreate = NewCreate(testLedger, testOptions, testExecutorFactory)
	var expectedWorkspaceName = "workspaceName"
	var expectedSolutionFqdn = "solutionFqdn"
	var expectedWorkflowHandle = "workflowHandle"

	testCreate.Run(testCreate, []string{expectedWorkspaceName, expectedSolutionFqdn, expectedWorkflowHandle})
	assert.Equal(t, expectedWorkspaceName, testExecutor.createWorkspace)
	assert.Equal(t, expectedSolutionFqdn, testExecutor.createSolution)
	assert.Equal(t, expectedWorkflowHandle, testExecutor.createWorkflow)
}

func TestNewCreate_Error(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewCreateOptions()
	var testExecutor = &stubCreateExecutor{
		createError: errors.New("expected"),
	}
	var testExecutorFactory = func(command *cobra.Command) CreateExecutor { return testExecutor }
	var testCreate = NewCreate(testLedger, testOptions, testExecutorFactory)

	defer patch.Unpatch()
	testCreate.Run(testCreate, []string{"", "", ""})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewCreate_NoSolution(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewCreateOptions()
	var testExecutor = &stubCreateExecutor{}
	var testExecutorFactory = func(command *cobra.Command) CreateExecutor { return testExecutor }
	var testCreate = NewCreate(testLedger, testOptions, testExecutorFactory)
	var expectedWorkspaceName = "workspaceName"
	var expectedWorkflowHandle = "workflowHandle"

	testCreate.Run(testCreate, []string{expectedWorkspaceName, expectedWorkflowHandle})
	assert.Equal(t, expectedWorkspaceName, testExecutor.createWorkspace)
	assert.Empty(t, testExecutor.createSolution)
	assert.Equal(t, expectedWorkflowHandle, testExecutor.createWorkflow)
}

func TestNewCreateAuto_bridgeArguments(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCmd = &cobra.Command{}
	var testAuto = NewCreateAuto(testLedger)
	var testArgs = []string{"1", "2", "3", "4"}

	if actual, directive := testAuto.bridgeArguments(testCmd, testArgs, ""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewCreateAuto_bridgeSolutions(t *testing.T) {
	var expectedFqdn = "oem/handle"
	var expectedVersion = "version"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserSolutionProvider{
		solutions: []mgmt.UserSolution{
			{
				Fqdn:    expectedFqdn,
				Version: expectedVersion,
			},
		},
	}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		solutionFacadeProvider: func() mgmt.UserSolutionFacade {
			return &stubUserSolutionFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "oem/"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName"}, testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%s:%s", testProvider.solutions[0].Fqdn, testProvider.solutions[0].Version), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual solution results")
	}
}

func TestNewCreateAuto_bridgeSolutions_Empty(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserSolutionProvider{}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		solutionFacadeProvider: func() mgmt.UserSolutionFacade {
			return &stubUserSolutionFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "oem/"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName"}, testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no solution results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewCreateAuto_bridgeSolutions_Error(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserSolutionProvider{
		getError: task.NewError("expected"),
	}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		solutionFacadeProvider: func() mgmt.UserSolutionFacade {
			return &stubUserSolutionFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "oem/"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName"}, testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no solution results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestNewCreateAuto_bridgeWorkflows(t *testing.T) {
	var expectedId = int64(37)
	var expectedHandle = "workflowHandle"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkflowProvider{
		workflows: []mgmt.UserWorkflow{
			{
				Id:     &expectedId,
				Handle: expectedHandle,
			},
		},
	}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		workflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return &stubUserWorkflowFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "37"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName", "oem/handle"}, testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%d\t%s", *testProvider.workflows[0].Id, testProvider.workflows[0].Handle), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual workflow results")
	}
}

func TestNewCreateAuto_bridgeWorkflows_Empty(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkflowProvider{}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		workflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return &stubUserWorkflowFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "37"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName", "oem/handle"}, testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no workflow results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewCreateAuto_bridgeWorkflows_Error(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkflowProvider{
		getError: task.NewError("expected"),
	}
	var testAuto = &CreateAutoBridge{
		ledger: testLedger,
		workflowFacadeProvider: func() mgmt.UserWorkflowFacade {
			return &stubUserWorkflowFacade{
				provider: testProvider,
			}
		},
	}
	var testCmd = &cobra.Command{}
	var testCompletable = "37"

	if actual, directive := testAuto.bridgeArguments(testCmd, []string{"workspaceName", "oem/handle"}, testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no workflow results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestNewCreateAuto_bridgeWorkspaces(t *testing.T) {
	var expectedCompletion = "expected"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testAuto = &CreateAutoBridge{
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

func TestNewCreateOptions(t *testing.T) {
	var testOptions = NewCreateOptions()
	var actual = testOptions.allDefiners()

	assert.Contains(t, actual, testOptions.OptionAccount)
	assert.Contains(t, actual, testOptions.OptionDescription)
	assert.Contains(t, actual, testOptions.OptionJsonPrinter)
	assert.Contains(t, actual, testOptions.OptionName)
}
