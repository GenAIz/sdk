package data

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubListExecutor struct {
	flow      string
	node      string
	workspace string
	listError error
}

func (s *stubListExecutor) List(workspace string, flow string, node string) error {
	s.workspace = workspace
	s.flow = flow
	s.node = node
	return s.listError
}

func TestListAutoBridge(t *testing.T) {
	var testBridge = &ListAutoBridge{}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd, []string{"too", "many", "args"}, "something")
	assert.Empty(t, actual)
	// can't prevent the user from trying to autocomplete things that don't exist, we can just block further auto-completion
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestListAutoBridge_Node(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var expectedWorkspace = "my-workspace"
	var expectedWorkflow = "my-workflow"
	var stubWorkflowBridge = &stubWorkspaceFlowBridge{
		stubBridge: stubBridge{
			completions: expectedCompletions,
			directive:   expectedDirective,
		},
	}
	var testBridge = &ListAutoBridge{
		workspaceNodes: stubWorkflowBridge,
	}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd,
		[]string{expectedWorkspace, expectedWorkflow}, "something")
	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubWorkflowBridge.workspace)
	assert.Equal(t, expectedWorkflow, stubWorkflowBridge.flow)
}

func TestListAutoBridge_Workflow(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var expectedWorkspace = "my-workspace"
	var stubWorkflowBridge = &stubWorkspaceBridge{
		stubBridge: stubBridge{
			completions: expectedCompletions,
			directive:   expectedDirective,
		},
	}
	var testBridge = &ListAutoBridge{
		workspaceFlows: stubWorkflowBridge,
	}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd, []string{expectedWorkspace}, "something")
	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubWorkflowBridge.workspace)
}

func TestListAutoBridge_Workspace(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var testBridge = &ListAutoBridge{
		workspaces: &stubBridge{
			completions: expectedCompletions,
			directive:   expectedDirective,
		},
	}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd, []string{}, "something")
	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, expectedDirective, directive)
}

func TestListOptions_ListParams(t *testing.T) {
	var testOptions = NewDataListOptions()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), "myAuth"),
	}
	var expectedResolve = &broker.WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: "myWorkspace",
		},
	}
	var actual *broker.WorkspaceNodeSetListParams

	actual = testOptions.ListParams(expectedBroker, expectedResolve)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, &expectedResolve.NodeResolveParams, actual.NodeResolveParams)
}

func TestNewDataListAutoBridge(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()

	assert.NotNil(t, NewDataListAutoBridge(testLedger))
}

func TestNewDataList(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewDataListOptions()
	var testExecutor = &stubListExecutor{}
	var testCmd = NewDataList(testLedger, testOptions, func(command *cobra.Command) ListExecutor {
		return testExecutor
	})
	var expectedWorkspace = "spacers"
	var expectedFlow = "flow"
	var expectedNode = "node"

	testCmd.Run(testCmd, []string{expectedWorkspace, expectedFlow, expectedNode})
	assert.Equal(t, expectedWorkspace, testExecutor.workspace)
	assert.Equal(t, expectedFlow, testExecutor.flow)
	assert.Equal(t, expectedNode, testExecutor.node)
}

func TestNewDataList_HandleError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewDataListOptions()
	var testCmd = NewDataList(testLedger, testOptions, func(command *cobra.Command) ListExecutor {
		return &stubListExecutor{
			listError: errors.New("exit"),
		}
	})

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{"0", "1", "2"})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}
