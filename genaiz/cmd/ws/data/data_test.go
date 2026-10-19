package data

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
)

type stubBridge struct {
	completions []cobra.Completion
	directive   cobra.ShellCompDirective
	toComplete  string
}

func (s *stubBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	s.toComplete = toComplete
	return s.completions, s.directive
}

type stubWorkspaceBridge struct {
	stubBridge
	readyOnly *config.BoolOption
	workspace string
}

func (s *stubWorkspaceBridge) WithReadyOnly(option *config.BoolOption) auto.WorkspaceBridge {
	s.readyOnly = option
	return s
}

func (s *stubWorkspaceBridge) WithWorkspace(workspace string) auto.WorkspaceBridge {
	s.workspace = workspace
	return s
}

type stubWorkspaceFlowBridge struct {
	stubBridge
	workspace string
	flow      string
}

func (s *stubWorkspaceFlowBridge) WithWorkspaceFlow(workspace string, flow string) auto.WorkspaceFlowBridge {
	s.workspace = workspace
	s.flow = flow
	return s
}

func TestBaseOptions_AccountParams(t *testing.T) {
	var expectedAccount = "account"
	var testOption = &config.StringOption{
		Option: config.Option{Key: "key"},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOptions = &baseOptions{
		optionAccount: testOption,
	}
	var testBroker *broker.Broker

	testViper.Set(testOption.Key, expectedAccount)
	actual := testOptions.AccountParams(testLedger)
	testBroker = actual.BrokerParams()
	assert.NotNil(t, testBroker)
	assert.Equal(t, expectedAccount, testBroker.HostAddr)
}

func TestBaseOptions_Definers(t *testing.T) {
	var expectedAccountOption = &config.StringOption{}
	var expectedJsonOption = &config.BoolOption{}
	var testOptions = &baseOptions{
		optionAccount:     expectedAccountOption,
		optionJsonPrinter: expectedJsonOption,
	}

	actual := testOptions.Definers()
	assert.Equal(t, 2, len(actual))
	assert.Same(t, expectedAccountOption, actual[0])
	assert.Same(t, expectedJsonOption, actual[1])
}

func TestBaseOptions_Options(t *testing.T) {
	var expectedAccountOption = &config.Option{Key: "account"}
	var expectedJsonOption = &config.Option{Key: "json"}
	var testOptions = &baseOptions{
		optionAccount:     &config.StringOption{Option: *expectedAccountOption},
		optionJsonPrinter: &config.BoolOption{Option: *expectedJsonOption},
	}

	actual := testOptions.Options()
	assert.Equal(t, 2, len(actual))
	assert.Equal(t, expectedAccountOption.Key, actual[0].Key)
	assert.Equal(t, expectedJsonOption.Key, actual[1].Key)
}

func TestBaseOptions_PrinterParams(t *testing.T) {
	var testOption = &config.BoolOption{
		Option: config.Option{Key: "key"},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOptions = &baseOptions{
		optionJsonPrinter: testOption,
	}
	var actual cli.PrinterParametric

	actual = testOptions.PrinterParams(testLedger)
	assert.True(t, actual.IsDefault())
	assert.NotNil(t, actual.Printer())
	testViper.Set(testOption.Key, true)
	actual = testOptions.PrinterParams(testLedger)
	assert.False(t, actual.IsDefault())
	assert.NotNil(t, actual.Printer())
}
