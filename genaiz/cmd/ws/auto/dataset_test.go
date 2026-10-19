package auto

import (
	"fmt"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/mgmt/workspaces"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type stubUserWorkspaceDataSetsFactory struct {
	mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []workspaces.UserWorkspaceDataSet]
}

type stubUserWorkspaceDataSetsProvider struct {
	dataSets []workspaces.UserWorkspaceDataSet
	getError task.Error
}

func (s stubUserWorkspaceDataSetsProvider) Get() ([]workspaces.UserWorkspaceDataSet, task.Error) {
	return s.dataSets, s.getError
}

func TestDataSetAutoBridge_Bridge(t *testing.T) {
	var expectedId = int64(37)
	var expectedPort = "port"
	var testFactory = &stubUserWorkspaceDataSetsFactory{}
	var testProvider = stubUserWorkspaceDataSetsProvider{
		dataSets: []workspaces.UserWorkspaceDataSet{
			{
				Id:        &expectedId,
				InputPort: expectedPort,
			},
		},
	}
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testBridge = &dataSetAutoBridge{
		ledger: testLedger,
		workspaceNodeDataSetsFactory: testFactory.WithProvider(
			func(p workspaces.UserWorkspaceDataSetsProviderFactory) workspaces.UserWorkspaceDataSetsProvider {
				return testProvider
			}),
	}

	completions, directive := testBridge.Bridge("something")
	assert.Equal(t, 1, len(completions))
	assert.Equal(t, fmt.Sprintf("%s\t%d", expectedPort, expectedId), completions[0])
	assert.Equal(t, cobra.ShellCompDirectiveKeepOrder, directive)
}

func TestDataSetAutoBridge_Bridge_EmptySets(t *testing.T) {
	var testFactory = &stubUserWorkspaceDataSetsFactory{}
	var testProvider = stubUserWorkspaceDataSetsProvider{}
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testBridge = &dataSetAutoBridge{
		ledger: testLedger,
		workspaceNodeDataSetsFactory: testFactory.WithProvider(
			func(p workspaces.UserWorkspaceDataSetsProviderFactory) workspaces.UserWorkspaceDataSetsProvider {
				return testProvider
			}),
	}

	completions, directive := testBridge.Bridge("something")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestDataSetAutoBridge_Bridge_GetError(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testFactory = &stubUserWorkspaceDataSetsFactory{}
	var testProvider = stubUserWorkspaceDataSetsProvider{
		getError: expectedError,
	}
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testBridge = &dataSetAutoBridge{
		ledger: testLedger,
		workspaceNodeDataSetsFactory: testFactory.WithProvider(
			func(p workspaces.UserWorkspaceDataSetsProviderFactory) workspaces.UserWorkspaceDataSetsProvider {
				return testProvider
			}),
	}

	completions, directive := testBridge.Bridge("something")
	assert.Empty(t, completions)
	assert.Equal(t, cobra.ShellCompDirectiveError, directive)
}

func TestDataSetAutoBridge_WithWorkspaceNode(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testBridge = NewDataSetAutoBridge(testLedger)
	var expectedWorkspace = "workspace"
	var expectedFlow = "workflow"
	var expectedNode = "node"

	actual := testBridge.WithWorkspaceNode(expectedWorkspace, expectedFlow, expectedNode)
	assert.Equal(t, expectedWorkspace, actual.(*dataSetAutoBridge).workspace)
	assert.Equal(t, expectedFlow, actual.(*dataSetAutoBridge).flow)
	assert.Equal(t, expectedNode, actual.(*dataSetAutoBridge).node)
}
