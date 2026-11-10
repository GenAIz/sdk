package run

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubListExecutor struct {
	listWorkspace string
	listError     error
}

func (sce *stubListExecutor) List(workspace string) error {
	sce.listWorkspace = workspace
	return sce.listError
}

type stubWorkspaceBridge struct {
	completions []cobra.Completion
	directive   cobra.ShellCompDirective
}

func (s stubWorkspaceBridge) Bridge(string) ([]cobra.Completion, cobra.ShellCompDirective) {
	return s.completions, s.directive
}

func TestListOptions_AccountParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewListOptions()
	var testParams = testOptions.AccountParams(testLedger)
	var expectedHost = "host"

	testViper.Set(testOptions.optionAccount.Key, expectedHost)
	actual := testParams.BrokerParams()
	assert.Equal(t, expectedHost, actual.HostAddr)
}

func TestListOptions_ListParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testBroker = &broker.Broker{}
	var testReadParams = &workspace.FlowReadParams{}
	var testOptions = NewListOptions()
	var expectedSince = timez.NewMonthTime()
	var testParams *broker.RunListParams

	testViper.Set(testOptions.optionDateMonthly.Key, true)
	testParams = testOptions.ListParams(testLedger, testBroker, testReadParams)
	assert.Equal(t, expectedSince, testParams.Since)
}

func TestListOptions_ListParams_Weekly(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testBroker = &broker.Broker{}
	var testReadParams = &workspace.FlowReadParams{}
	var testOptions = NewListOptions()
	var expectedSince = timez.NewWeekTime()
	var testParams *broker.RunListParams

	testViper.Set(testOptions.optionDateWeekly.Key, true)
	testParams = testOptions.ListParams(testLedger, testBroker, testReadParams)
	assert.Equal(t, expectedSince, testParams.Since)
}

func TestListOptions_ListParams_Today(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testBroker = &broker.Broker{}
	var testReadParams = &workspace.FlowReadParams{}
	var testOptions = NewListOptions()
	var expectedSince = timez.NewTodayTime()
	var testParams *broker.RunListParams

	testViper.Set(testOptions.optionDateToday.Key, true)
	testParams = testOptions.ListParams(testLedger, testBroker, testReadParams)
	assert.Equal(t, expectedSince, testParams.Since)
}

func TestListOptions_Options(t *testing.T) {
	var testOptions = NewListOptions()

	assert.Equal(t, 5, len(testOptions.Options()))
}

func TestListOptions_PrinterParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewListOptions()
	var testParams = testOptions.PrinterParams(testLedger)

	testViper.Set(testOptions.optionJsonPrinter.Key, true)
	assert.False(t, testParams.IsDefault())
}

func TestListAutoBridge(t *testing.T) {
	var stubBridge = &stubWorkspaceBridge{
		completions: []cobra.Completion{
			"testCompletion",
		},
		directive: cobra.ShellCompDirectiveFilterFileExt,
	}
	var testBridge = ListAutoBridge{
		workspaces: stubBridge,
	}
	var actual, directive = testBridge.bridgeArguments(&cobra.Command{}, []string{}, "test")

	assert.Equal(t, stubBridge.completions, actual)
	assert.Equal(t, cobra.ShellCompDirectiveFilterFileExt, directive)
}

func TestListAutoBridge_BridgeMoreArgs(t *testing.T) {
	var stubBridge = &stubWorkspaceBridge{}
	var testBridge = ListAutoBridge{
		workspaces: stubBridge,
	}
	var actual, directive = testBridge.bridgeArguments(&cobra.Command{}, []string{"test"}, "test")

	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestNewList(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &stubListExecutor{}
	var testExecutorFactory = func(command *cobra.Command) ListExecutor { return testExecutor }
	var testCmd = NewList(testLedger, testOptions, testExecutorFactory)
	var expectedWorkspace = "workspace"

	testCmd.Run(testCmd, []string{expectedWorkspace})
	assert.Equal(t, expectedWorkspace, testExecutor.listWorkspace)
}

func TestNewList_HandleError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("error")
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testOptions = NewListOptions()
	var testExecutor = &stubListExecutor{
		listError: expectedError,
	}
	var testExecutorFactory = func(command *cobra.Command) ListExecutor { return testExecutor }
	var testCmd = NewList(testLedger, testOptions, testExecutorFactory)
	var expectedWorkspace = "workspace"

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkspace})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}
