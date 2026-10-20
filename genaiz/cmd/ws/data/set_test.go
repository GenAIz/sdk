package data

import (
	"errors"
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubDataSetsBridge struct {
	stubBridge
	workspace string
	flow      string
	node      string
}

func (s *stubDataSetsBridge) WithWorkspaceNode(workspace, flow, node string) auto.DataSetAutoBridge {
	s.workspace = workspace
	s.flow = flow
	s.node = node
	return s
}

type stubSetExecutor struct {
	flow      string
	node      string
	workspace string
	dataSet   string
	filePath  string
	setError  error
}

func (s *stubSetExecutor) Set(workspace, flow, node, dataSet, filePath string) error {
	s.workspace = workspace
	s.flow = flow
	s.node = node
	s.dataSet = dataSet
	s.filePath = filePath
	return s.setError
}

func TestSetAutoBridge(t *testing.T) {
	var testBridge = &SetAutoBridge{}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd, []string{"too", "many", "args", "with", "file"}, "something")
	assert.Empty(t, actual)
	// can't prevent the user from trying to autocomplete things that don't exist, we can just block further auto-completion
	assert.Equal(t, cobra.ShellCompDirectiveNoFileComp, directive)
}

func TestSetAutoBridge_DataSet(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var expectedWorkspace = "my-workspace"
	var expectedWorkflow = "my-workflow"
	var expectedNode = "a-node"
	var stubDataSetBridge = &stubDataSetsBridge{
		stubBridge: stubBridge{
			completions: expectedCompletions,
			directive:   expectedDirective,
		},
	}
	var testBridge = &SetAutoBridge{
		dataSets: stubDataSetBridge,
	}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd,
		[]string{expectedWorkspace, expectedWorkflow, expectedNode},
		"something")
	assert.Equal(t, expectedCompletions, actual)
	assert.Equal(t, expectedDirective, directive)
	assert.Equal(t, expectedWorkspace, stubDataSetBridge.workspace)
	assert.Equal(t, expectedWorkflow, stubDataSetBridge.flow)
	assert.Equal(t, expectedNode, stubDataSetBridge.node)
}

func TestSetAutoBridge_FilePath(t *testing.T) {
	var stubDataSetBridge = &stubDataSetsBridge{
		stubBridge: stubBridge{},
	}
	var testBridge = &SetAutoBridge{
		dataSets: stubDataSetBridge,
	}
	var testCmd = &cobra.Command{}
	var actual []cobra.Completion
	var directive cobra.ShellCompDirective

	actual, directive = testBridge.bridgeArguments(testCmd,
		[]string{"my-workspace", "my-workflow", "a-node", "dataSet"},
		"something")
	assert.Empty(t, actual)
	assert.Equal(t, cobra.ShellCompDirectiveDefault, directive)
	assert.Empty(t, stubDataSetBridge.workspace)
	assert.Empty(t, stubDataSetBridge.flow)
	assert.Empty(t, stubDataSetBridge.node)
}

func TestSetAutoBridge_Node(t *testing.T) {
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
	var testBridge = &SetAutoBridge{
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

func TestSetAutoBridge_Workflow(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var expectedWorkspace = "my-workspace"
	var stubWorkflowBridge = &stubWorkspaceBridge{
		stubBridge: stubBridge{
			completions: expectedCompletions,
			directive:   expectedDirective,
		},
	}
	var testBridge = &SetAutoBridge{
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

func TestSetAutoBridge_Workspace(t *testing.T) {
	var expectedCompletions = []cobra.Completion{"expected"}
	var expectedDirective = cobra.ShellCompDirectiveNoSpace
	var testBridge = &SetAutoBridge{
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

func TestSetOptions_CreateSetParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), ".auth"),
	}
	var expectedNodeResolveParams = &broker.WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: "aName",
		},
	}
	var expectedSetResolveParams = &broker.WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "aPort",
		},
	}
	var expectedSetName = "setName"
	var expectedSetHandle = "setHandle"
	var expectedSetDesc = "setDesc"
	var testOptions = NewDataSetOptions()
	var actual *broker.DataSetCreateParams

	testViper.Set(testOptions.optionSetName.Key, expectedSetName)
	testViper.Set(testOptions.optionSetHandle.Key, expectedSetHandle)
	testViper.Set(testOptions.optionSetDesc.Key, expectedSetDesc)
	actual = testOptions.CreateSetParams(testLedger, expectedBroker, expectedNodeResolveParams, expectedSetResolveParams)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, expectedSetName, actual.Name)
	assert.Equal(t, expectedSetHandle, actual.Handle)
	assert.Equal(t, expectedSetDesc, actual.Description)
	assert.Equal(t, &expectedNodeResolveParams.NodeResolveParams, actual.NodeResolveParams)
	assert.Equal(t, &expectedSetResolveParams.NodeSetParams, actual.NodeSetParams)
}

func TestSetOptions_IsOverwrite(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewDataSetOptions()

	assert.False(t, testOptions.IsOverwrite(testLedger))
	testViper.Set(testOptions.optionOverwrite.Key, true)
	assert.True(t, testOptions.IsOverwrite(testLedger))
}

func TestSetOptions_ListFilesParams(t *testing.T) {
	var testOptions = NewDataSetOptions()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), "myAuth"),
	}
	var expectedResolve = &broker.WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "my-port",
		},
	}
	var actual *broker.DataSetListFilesParams

	actual = testOptions.ListFilesParams(expectedBroker, expectedResolve)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, &expectedResolve.NodeSetParams, actual.NodeSetParams)
}

func TestSetOptions_ResolveSetParams(t *testing.T) {
	var testOptions = NewDataSetOptions()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), "myAuth"),
	}
	var expectedNodeResolveParams = &broker.WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: "aName",
		},
	}
	var expectedDataSetId = int64(37)
	var actual *broker.WorkspaceNodeSetResolveParams

	actual = testOptions.ResolveSetParams(cast.ToString(expectedDataSetId), expectedBroker, expectedNodeResolveParams)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, &expectedDataSetId, actual.DataSetId)
	assert.Equal(t, &expectedNodeResolveParams.NodeResolveParams, actual.NodeResolveParams)
}

func TestSetOptions_ResolveSetParams_PortString(t *testing.T) {
	var testOptions = NewDataSetOptions()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), "myAuth"),
	}
	var expectedNodeResolveParams = &broker.WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: "aName",
		},
	}
	var expectedPortHandle = "port"
	var actual *broker.WorkspaceNodeSetResolveParams

	actual = testOptions.ResolveSetParams(expectedPortHandle, expectedBroker, expectedNodeResolveParams)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, expectedPortHandle, actual.PortHandle)
	assert.Equal(t, &expectedNodeResolveParams.NodeResolveParams, actual.NodeResolveParams)
}

func TestSetOptions_UploadFileParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewDataSetOptions()
	var expectedBroker = &broker.Broker{
		AuthFile: filepath.Join(t.TempDir(), "myAuth"),
	}
	var expectedResolve = &broker.WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "my-port",
		},
	}
	var expectedPath = filepath.Join(t.TempDir(), "subfolder")
	var expectedFileName = "fileName"
	var expectedDestination = "destination"
	var actual *broker.DataSetUploadParams

	testViper.Set(testOptions.optionFilename.Key, expectedFileName)
	testViper.Set(testOptions.optionDestination.Key, expectedDestination)
	actual = testOptions.UploadFileParams(testLedger, expectedPath, expectedBroker, expectedResolve)
	assert.Equal(t, expectedBroker, &actual.Broker)
	assert.Equal(t, expectedPath, actual.FilePath)
	assert.Equal(t, expectedFileName, actual.FileName)
	assert.Equal(t, expectedDestination, actual.DestinationPath)
	assert.Equal(t, &expectedResolve.NodeSetParams, actual.NodeSetParams)
}

func TestNewDataSet(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewDataSetOptions()
	var testExecutor = &stubSetExecutor{}
	var testCmd = NewDataSet(testLedger, testOptions, func(command *cobra.Command) SetExecutor {
		return testExecutor
	})
	var expectedWorkspace = "spacers"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var expectedSet = "dataSet"
	var expectedFile = "file"

	testCmd.Run(testCmd, []string{expectedWorkspace, expectedFlow, expectedNode, expectedSet, expectedFile})
	assert.Equal(t, expectedWorkspace, testExecutor.workspace)
	assert.Equal(t, expectedFlow, testExecutor.flow)
	assert.Equal(t, expectedNode, testExecutor.node)
	assert.Equal(t, expectedSet, testExecutor.dataSet)
	assert.Equal(t, expectedFile, testExecutor.filePath)
}

func TestNewDataSet_HandleExit(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testOptions = NewDataSetOptions()
	var testExecutor = &stubSetExecutor{
		setError: errors.New("exit"),
	}
	var testCmd = NewDataSet(testLedger, testOptions, func(command *cobra.Command) SetExecutor {
		return testExecutor
	})
	var expectedWorkspace = "spacers"
	var expectedFlow = "flow"
	var expectedNode = "node"
	var expectedSet = "dataSet"
	var expectedFile = "file"

	defer patch.Unpatch()
	testCmd.Run(testCmd, []string{expectedWorkspace, expectedFlow, expectedNode, expectedSet, expectedFile})
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewDataSetOptions_DefaultDesc(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewDataSetOptions()
	var expectedValue = "value"

	testViper.Set(testOptions.optionSetHandle.Key, expectedValue)
	assert.Equal(t, expectedValue, testLedger.GetString(testOptions.optionSetHandle))
	assert.Equal(t, expectedValue, testLedger.GetString(testOptions.optionSetName))
}

func TestNewDataSetOptions_DefaultName(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewDataSetOptions()
	var expectedValue = "value"

	testViper.Set(testOptions.optionSetHandle.Key, expectedValue)
	assert.Equal(t, expectedValue, testLedger.GetString(testOptions.optionSetHandle))
	assert.Equal(t, expectedValue, testLedger.GetString(testOptions.optionSetDesc))
}
