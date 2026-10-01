package node

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
)

type stubListWorkspaceBridge struct {
	directive  cobra.ShellCompDirective
	results    []cobra.Completion
	toComplete string
}

func (s *stubListWorkspaceBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	s.toComplete = toComplete
	return s.results, s.directive
}

type stubListWorkspaceFlowsBridge struct {
	directive  cobra.ShellCompDirective
	results    []cobra.Completion
	toComplete string
	workspace  string
}

func (s *stubListWorkspaceFlowsBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	s.toComplete = toComplete
	return s.results, s.directive
}

func (s *stubListWorkspaceFlowsBridge) WithReadyOnly(*config.BoolOption) auto.WorkspaceBridge {
	panic("unexpected call")
}

func (s *stubListWorkspaceFlowsBridge) WithWorkspace(workspace string) auto.WorkspaceBridge {
	s.workspace = workspace
	return s
}

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
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var expectedWorkspace = "myWorkspace"
	var stubBridge = &stubListWorkspaceFlowsBridge{
		directive: expectedDirective,
	}
	var testAuto = &ListAutoBridge{
		workspaceFlows: stubBridge,
	}
	var testCmd = &cobra.Command{}

	actual, directive := testAuto.bridgeArguments(testCmd, []string{expectedWorkspace}, "myFlow")
	assert.Empty(t, actual)
	assert.Equal(t, expectedWorkspace, stubBridge.workspace)
	assert.Equal(t, expectedDirective, directive)
}

func TestNewListAuto_bridgeWorkspaces(t *testing.T) {
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var testAuto = &ListAutoBridge{
		workspaces: &stubListWorkspaceBridge{
			directive: expectedDirective,
		},
	}
	var testCmd = &cobra.Command{}

	actual, directive := testAuto.bridgeArguments(testCmd, []string{}, "myWorkspace")
	assert.Empty(t, actual)
	assert.Equal(t, expectedDirective, directive)
}
