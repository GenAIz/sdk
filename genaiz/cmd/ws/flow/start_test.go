package flow

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubStartExecutor struct {
	startWorkflow  string
	startWorkspace string
	startError     error
}

func (sse *stubStartExecutor) Start(workspace, workflow string) error {
	sse.startWorkspace = workspace
	sse.startWorkflow = workflow
	return sse.startError
}

type stubWorkspaceFlowBridge struct {
	completions []cobra.Completion
	directive   cobra.ShellCompDirective

	readOnly   *config.BoolOption
	toComplete string
	workspace  string
}

func (s *stubWorkspaceFlowBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	s.toComplete = toComplete
	return s.completions, s.directive
}

func (s *stubWorkspaceFlowBridge) WithReadyOnly(option *config.BoolOption) auto.WorkspaceBridge {
	s.readOnly = option
	return s
}

func (s *stubWorkspaceFlowBridge) WithWorkspace(workspace string) auto.WorkspaceBridge {
	s.workspace = workspace
	return s
}

func TestStartAutoBridge_bridgeArguments(t *testing.T) {
	var testBridge = &StartAutoBridge{}
	var actual, directive = testBridge.bridgeArguments(&cobra.Command{}, []string{"workspace", "workflow"}, "")

	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestStartAutoBridge_bridgeArguments_Workspaces(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"myFlow"}
	var testWorkspaceBridge = &stubWorkspaceBridge{
		completions: expectedCompletions,
		directive:   cobra.ShellCompDirectiveFilterFileExt,
	}
	var testBridge = &StartAutoBridge{
		workspaces: testWorkspaceBridge,
	}
	var actual, directive = testBridge.bridgeArguments(&cobra.Command{}, []string{}, "")

	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, cobra.ShellCompDirectiveFilterFileExt, directive)
}

func TestStartAutoBridge_bridgeArguments_WorkspaceFlows(t *testing.T) {
	var expectedWorkspace = "workspace"
	var expectedCompletions = []cobra.Completion{"myFlow"}
	var testWorkspaceBridge = &stubWorkspaceFlowBridge{
		completions: expectedCompletions,
		directive:   cobra.ShellCompDirectiveFilterFileExt,
	}
	var testBridge = &StartAutoBridge{
		workspaceFlows: testWorkspaceBridge,
	}
	var actual, directive = testBridge.bridgeArguments(&cobra.Command{}, []string{expectedWorkspace}, "")

	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, cobra.ShellCompDirectiveFilterFileExt, directive)
	assert.Equal(t, expectedWorkspace, testWorkspaceBridge.workspace)
}

func TestStartOptions_AccountParams(t *testing.T) {
	var expectedAccount = "account"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewStartOptions()
	var testParametric = testOptions.AccountParams(testLedger)

	testViper.Set(testOptions.optionAccount.Key, expectedAccount)
	testParametric.BrokerParams()
}

func TestStartOptions_Options(t *testing.T) {
	var testOptions = NewStartOptions()

	assert.Equal(t, 3, len(testOptions.Options()))
}

func TestStartOptions_PrinterParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewStartOptions()
	var testParametric = testOptions.PrinterParams(testLedger)

	testViper.Set(testOptions.optionJsonPrinter.Key, true)
	assert.False(t, testParametric.IsDefault())
}

func TestStartOptions_StartParams(t *testing.T) {
	var expectedName = "name"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewStartOptions()
	var actual *broker.RunStartParams

	testViper.Set(testOptions.optionName.Key, expectedName)
	actual = testOptions.StartParams(testLedger, &broker.Broker{}, &workspace.FlowReadParams{})
	assert.Equal(t, expectedName, actual.Name)
	assert.NotNil(t, actual.Broker)
	assert.NotNil(t, actual.FlowReadParams)
	assert.NotNil(t, actual.StartParams)
}

func TestNewStartOptions_allDefiners(t *testing.T) {
	var testOptions = NewStartOptions()

	assert.Equal(t, 3, len(testOptions.allDefiners()))
}

func TestNewStart_HandleExit(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var expectedWorkspace = "workspace"
	var expectedWorkflow = "workflow"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewStartOptions()
	var testExecutor = &stubStartExecutor{
		startError: expectedError,
	}
	var testFactory = func(cmd *cobra.Command) StartExecutor {
		return testExecutor
	}
	var testCmd = NewStart(testLedger, testOptions, testFactory)

	defer patch.Unpatch()
	testCmd.Run(&cobra.Command{}, []string{expectedWorkspace, expectedWorkflow})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}
