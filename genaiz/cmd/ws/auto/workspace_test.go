package auto

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserWorkspaceFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.WorkspaceListParams
	provider mgmt.Provider[[]mgmt.UserWorkspace]
}

func (s *stubUserWorkspaceFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkspace] {
	s.filter = filter
	return s.provider
}

func (s *stubUserWorkspaceFacade) Provider() mgmt.Provider[[]mgmt.UserWorkspace] {
	return s.provider
}

func (s *stubUserWorkspaceFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkspace, broker.WorkspaceListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkspaceFacade) WithParams(params *broker.WorkspaceListParams) mgmt.Facade[[]mgmt.UserWorkspace, broker.WorkspaceListParams] {
	s.params = params
	return s
}

type stubUserWorkspaceProvider struct {
	workspaces []mgmt.UserWorkspace
	getError   task.Error
}

func (s stubUserWorkspaceProvider) Get() ([]mgmt.UserWorkspace, task.Error) {
	return s.workspaces, s.getError
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

type stubUserWorkspaceNodesFacade struct {
	filter   string
	logger   *logrus.Logger
	params   *broker.WorkspaceNodeListParams
	provider mgmt.Provider[[]mgmt.UserWorkspaceNode]
}

func (s *stubUserWorkspaceNodesFacade) Filtering(filter string) mgmt.Provider[[]mgmt.UserWorkspaceNode] {
	s.filter = filter
	return s.provider
}

func (s *stubUserWorkspaceNodesFacade) Provider() mgmt.Provider[[]mgmt.UserWorkspaceNode] {
	return s.provider
}

func (s *stubUserWorkspaceNodesFacade) WithLogger(logger *logrus.Logger) mgmt.Facade[[]mgmt.UserWorkspaceNode, broker.WorkspaceNodeListParams] {
	s.logger = logger
	return s
}

func (s *stubUserWorkspaceNodesFacade) WithParams(params *broker.WorkspaceNodeListParams) mgmt.Facade[[]mgmt.UserWorkspaceNode, broker.WorkspaceNodeListParams] {
	s.params = params
	return s
}

type stubUserWorkspaceNodesProvider struct {
	nodes    []mgmt.UserWorkspaceNode
	getError task.Error
}

func (s stubUserWorkspaceNodesProvider) Get() ([]mgmt.UserWorkspaceNode, task.Error) {
	return s.nodes, s.getError
}

func TestWorkspaceAutoBridge_Bridge(t *testing.T) {
	var expectedId = int64(37)
	var expectedName = "workspaceName"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceProvider{
		workspaces: []mgmt.UserWorkspace{
			{
				Id:   expectedId,
				Name: expectedName,
			},
		},
	}
	var testAuto = &WorkspaceAutoBridge{
		ledger: testLedger,
		workspaceFacadeProvider: func() mgmt.UserWorkspacesFacade {
			return &stubUserWorkspaceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%d\t%s", testProvider.workspaces[0].Id, testProvider.workspaces[0].Name), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual workspace results")
	}
}

func TestWorkspaceAutoBridge_Bridge_Empty(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceProvider{}
	var testAuto = &WorkspaceAutoBridge{
		ledger: testLedger,
		workspaceFacadeProvider: func() mgmt.UserWorkspacesFacade {
			return &stubUserWorkspaceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no workspace results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestWorkspaceAutoBridge_Bridge_Error(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceProvider{
		getError: task.NewError("expected"),
	}
	var testAuto = &WorkspaceAutoBridge{
		ledger: testLedger,
		workspaceFacadeProvider: func() mgmt.UserWorkspacesFacade {
			return &stubUserWorkspaceFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "expected no workspace results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestNewWorkspaceBridge(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testBridge = NewWorkspaceBridge(testLedger)

	assert.NotNil(t, testBridge.workspaceFacadeProvider)
	assert.Same(t, testLedger, testBridge.ledger)
}

func TestWorkspaceFlowAutoBridge_Bridge(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = &stubUserWorkspaceFlowsProvider{}
	var testAuto = &WorkspaceFlowAutoBridge{
		ledger:          testLedger,
		readyOnlyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspace:       "workspaceName",

		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}

	if actual, directive := testAuto.Bridge(""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestWorkspaceFlowAutoBridge_Bridge_GetError(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = &stubUserWorkspaceFlowsProvider{
		getError: task.NewError("expected"),
	}
	var testAuto = &WorkspaceFlowAutoBridge{
		ledger:          testLedger,
		readyOnlyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspace:       "workspaceName",

		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}

	if actual, directive := testAuto.Bridge(""); len(actual) > 0 {
		assert.Fail(t, "expected no results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestWorkspaceFlowAutoBridge_Bridge_WorkspaceId(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = &stubUserWorkspaceFlowsProvider{
		flows: []mgmt.UserWorkspaceFlow{
			{
				Id:             1337,
				WorkflowHandle: "handle",
			},
		},
	}
	var testAuto = &WorkspaceFlowAutoBridge{
		ledger:          testLedger,
		readyOnlyOption: &config.BoolOption{Option: config.Option{Key: "key"}},
		workspace:       "37",

		workspaceFlowFacadeProvider: func() mgmt.UserWorkspaceFlowsFacade {
			return &stubUserWorkspaceFlowsFacade{
				provider: testProvider,
			}
		},
	}

	if actual, directive := testAuto.Bridge(""); len(actual) == 0 {
		assert.Fail(t, "expected results")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
		assert.Equal(t, fmt.Sprintf("%d\t%s", testProvider.flows[0].Id, testProvider.flows[0].WorkflowHandle), actual[0])
	}
}

func TestNewWorkspaceFlowBridge(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOption = &config.BoolOption{
		Option: config.Option{Key: "key"},
	}
	var testBridge = NewWorkspaceFlowBridge(testLedger)

	assert.NotNil(t, testBridge)
	actual := testBridge.WithWorkspace("workspace")
	assert.NotNil(t, actual)
	actual = actual.WithReadyOnly(testOption)
	assert.NotNil(t, actual)
}

func TestWorkspaceNodeAutoBridge_Bridge(t *testing.T) {
	var expectedId = int64(37)
	var expectedHandle = "smartFunctionHandle"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceNodesProvider{
		nodes: []mgmt.UserWorkspaceNode{
			{
				Id:                  expectedId,
				SmartFunctionHandle: expectedHandle,
			},
		},
	}
	var testAuto = &WorkspaceNodeAutoBridge{
		ledger:    testLedger,
		flow:      "flow",
		workspace: "workspace",
		workspaceNodeFacadeProvider: func() mgmt.UserWorkspaceNodesFacade {
			return &stubUserWorkspaceNodesFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Equal(t, fmt.Sprintf("%d\t%s", testProvider.nodes[0].Id, testProvider.nodes[0].SmartFunctionHandle), actual[0])
		assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
	} else {
		assert.Fail(t, "expected actual workspace nodes")
	}
}

func TestWorkspaceNodeAutoBridge_Bridge_GetError(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceNodesProvider{
		getError: expectedError,
	}
	var testAuto = &WorkspaceNodeAutoBridge{
		ledger:    testLedger,
		flow:      cast.ToString(37),
		workspace: "workspace",
		workspaceNodeFacadeProvider: func() mgmt.UserWorkspaceNodesFacade {
			return &stubUserWorkspaceNodesFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect workspace nodes")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveError, directive)
	}
}

func TestWorkspaceNodeAutoBridge_Bridge_NoNodes(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testProvider = stubUserWorkspaceNodesProvider{}
	var testAuto = &WorkspaceNodeAutoBridge{
		ledger:    testLedger,
		flow:      "flow",
		workspace: cast.ToString(37),
		workspaceNodeFacadeProvider: func() mgmt.UserWorkspaceNodesFacade {
			return &stubUserWorkspaceNodesFacade{
				provider: testProvider,
			}
		},
	}
	var testCompletable = "37"

	if actual, directive := testAuto.Bridge(testCompletable); len(actual) > 0 {
		assert.Fail(t, "did not expect workspace nodes")
	} else {
		assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
	}
}

func TestNewWorkspaceNodeAutoBridge(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testBridge = NewWorkspaceNodeAutoBridge(testLedger)

	assert.NotNil(t, testBridge)
	assert.NotNil(t, testBridge.WithWorkspaceFlow("workspace", "flow"))
}
