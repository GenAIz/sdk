package wf

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestBaseExecutor_makeConfigParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionConfigType("test")
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}
	var actual *shared.ConfigParams
	var err error

	actual, err = testExecutor.makeConfigParams(testOption)
	assert.Error(t, err)
	testViper.Set(testOption.Key, shared.ConfigTypeJson)
	actual, err = testExecutor.makeConfigParams(testOption)
	assert.NoError(t, err)
	assert.Equal(t, testLedger.ConfigName, actual.ConfigName)
	assert.Equal(t, shared.ConfigTypeJson, *actual.ConfigType)
}

func TestNewWf(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWf(testLedger, nil, nil, nil)

	assert.Equal(t, 4, len(testCmd.Commands()))
}

func TestWorkflowWriter_BuildWorkflows(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedLink = "link"
	var expectedNode = "node"
	var testWriter = &workflowWriter{
		root:      expectedRoot,
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
	var root, actual = testWriter.WithWorkflow(&broker.Workflow{Handle: expectedHandle}).
		BuildWorkflows()

	assert.Equal(t, expectedRoot, root)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedHandle, actual[0].Handle)
	assert.Empty(t, actual[0].Links)
	assert.Empty(t, actual[0].Nodes)
	root, actual = testWriter.
		WithWorkflows([]broker.Workflow{{Handle: expectedHandle}}).
		WithWorkflowLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}}).
		BuildWorkflows()
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, 1, len(actual[0].Links))
	assert.Equal(t, expectedLink, actual[0].Links[0].LhsNode)
	root, actual = testWriter.
		WithWorkflowNodes(expectedHandle, []broker.WorkflowNode{{Handle: expectedNode}}).
		BuildWorkflows()
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, 1, len(actual[0].Nodes))
	assert.Equal(t, expectedNode, actual[0].Nodes[0].Handle)
}

func TestWorkflowWriter_GetWorkflowByHandle(t *testing.T) {
	var expectedHandle = "handle"
	var testWriter = &workflowWriter{
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	assert.Empty(t, testWriter.GetWorkflows())
	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{
			{
				Handle: expectedHandle,
			},
		},
	}
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestWorkflowWriter_addLinks(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedLink = "link"
	var testWriter = &workflowWriter{
		root:      expectedRoot,
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{{Handle: expectedHandle}},
	}
	testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Links))
	assert.Equal(t, expectedLink, actual.Links[0].LhsNode)
}

func TestWorkflowWriter_addNodes(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedNode = "node"
	var testWriter = &workflowWriter{
		root:      expectedRoot,
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{{Handle: expectedHandle}},
	}
	testWriter.addNodes(expectedHandle, &broker.WorkflowNode{Handle: expectedNode})
	testWriter.addNodes(expectedHandle, &broker.WorkflowNode{Handle: expectedNode})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Nodes))
	assert.Equal(t, expectedNode, actual.Nodes[0].Handle)
}

func TestWorkflowWriter_removeLinks(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedLink = "link"
	var testWriter = &workflowWriter{
		root:      expectedRoot,
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{{
			Handle: expectedHandle,
			Links: []broker.WorkflowLink{
				{
					LhsNode: expectedLink,
				},
				{
					LhsNode: "other",
				},
			},
		}},
	}
	testWriter.removeLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	testWriter.removeLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Links))
}

func TestWorkflowWriter_removeNodes(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedNode = "node"
	var testWriter = &workflowWriter{
		root:      expectedRoot,
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{{
			Handle: expectedHandle,
			Nodes: []broker.WorkflowNode{
				{
					Handle: expectedNode,
				},
				{
					Handle: "other",
				},
			},
		}},
	}
	testWriter.removeNodes(expectedHandle, expectedNode)
	testWriter.removeNodes(expectedHandle, expectedNode)
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Nodes))
}

func TestWorkflowWriter_Write(t *testing.T) {
	var expectedHandle = "handle"
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testWriter = &workflowWriter{
		root:      "solution.workflows",
		workflows: make(map[string]broker.Workflow),
	}
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")
	var actual *broker.Workflow

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	testWriter.WithWorkflow(&broker.Workflow{
		Handle: expectedHandle,
	})
	assert.NoError(t, testWriter.Write(testFile.Name()))
	testWriter = newWorkflowWriter(testLedger, testFile.Name())
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestWorkflowWriter_WritePermissionsError(t *testing.T) {
	var invalidPath = "/invalid/path.yaml"
	var testWriter = &workflowWriter{}

	assert.Error(t, testWriter.Write(invalidPath))
}

func TestWorkflowWriter_WriteYamlError(t *testing.T) {
	var testWriter = &workflowWriter{}
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	assert.NoError(t, os.WriteFile(testFile.Name(), []byte("--"), 0660))
	assert.Error(t, testWriter.Write(testFile.Name()))
}

func Test_newWorkflowWriterYamlError(t *testing.T) {
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")
	var testWriter *workflowWriter

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	assert.NoError(t, os.WriteFile(testFile.Name(), []byte("--"), 0660))
	testLedger.InitLogging()
	testWriter = newWorkflowWriter(testLedger, testFile.Name())
	assert.Error(t, testWriter.Write(testFile.Name()))
	assert.Empty(t, testWriter.GetWorkflows())
}
