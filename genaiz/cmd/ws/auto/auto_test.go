package auto

import (
	"fmt"
	"testing"

	"github.com/sirupsen/logrus"
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
		Ledger: testLedger,
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
		Ledger: testLedger,
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
		Ledger: testLedger,
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
	assert.Same(t, testLedger, testBridge.Ledger)
}
