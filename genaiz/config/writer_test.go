package config

import (
	"os"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/broker"
)

func TestSolutionWriter_BuildSolution(t *testing.T) {
	var expectedName = "name"
	var testWriter = &SolutionWriter{
		BaseWriter: BaseWriter{
			current: &broker.Solution{
				Handle: "handle",
			},
		},
		updated: &broker.Solution{
			Handle: "handle",
			Name:   expectedName,
		},
	}
	var key, actual = testWriter.BuildSolution()

	assert.NotEmpty(t, key)
	assert.Equal(t, testWriter.current.Handle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
}

func TestSolutionWriter_BuildSolution_CurrentEmpty(t *testing.T) {
	var expectedHandle = "updated"
	var testWriter = &SolutionWriter{}
	var key, actual = testWriter.WithSolution(&broker.Solution{Handle: expectedHandle}).
		BuildSolution()

	assert.NotEmpty(t, key)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestSolutionWriter_BuildSolution_UpdateEmpty(t *testing.T) {
	var expectedHandle = "current"
	var testWriter = &SolutionWriter{}
	var key, actual = testWriter.WithCurrent(&broker.Solution{Handle: expectedHandle}).
		BuildSolution()

	assert.NotEmpty(t, key)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestSolutionWriter_Read(t *testing.T) {
	var testViper = viper.New()
	var testLedger = NewBuilder().WithViper(testViper).Build()
	var testWriter = &SolutionWriter{}
	var actual = testWriter.Read(testLedger, "/_not_exist")

	assert.Empty(t, actual.current)
}

func TestSolutionWriter_WithWorkflow(t *testing.T) {
	var testCurrent = &broker.Solution{
		Workflows: []broker.Workflow{
			{Handle: "existing"},
		},
	}
	var testWorkflow = &broker.Workflow{Handle: "handle"}
	var testWriter = NewSolutionWriter()
	var key, actual = testWriter.WithCurrent(testCurrent).
		WithWorkflow(testWorkflow).
		WithWorkflow(testWorkflow).
		BuildSolution()

	assert.NotEmpty(t, key)
	assert.Equal(t, 2, len(actual.Workflows))
	assert.Equal(t, testWorkflow.Handle, actual.Workflows[1].Handle)
}

func TestSolutionWriter_Write(t *testing.T) {
	var expectedHandle = "handle"
	var testLedger = NewBuilder().WithViper(viper.New()).Build()
	var testWriter = NewSolutionWriter()
	var testFile, err = os.CreateTemp("", ".genaiz_solutionWriter*.yaml")
	var actual broker.Solution

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	testWriter.WithCurrent(&broker.Solution{
		Handle: expectedHandle,
	})
	assert.NoError(t, testWriter.Write(testFile.Name()))
	_, actual = testWriter.Read(testLedger, testFile.Name()).
		BuildSolution()
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestSolutionWriter_WritePermissionsError(t *testing.T) {
	var invalidPath = "/invalid/path.yaml"
	var testWriter = &SolutionWriter{}

	assert.Error(t, testWriter.Write(invalidPath))
}

func TestSolutionWriter_WriteYamlError(t *testing.T) {
	var testWriter = &SolutionWriter{}
	var testFile, err = os.CreateTemp("", "genaiz_solutionWriter*.yaml")

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	assert.NoError(t, os.WriteFile(testFile.Name(), []byte("--"), 0660))
	assert.Error(t, testWriter.Write(testFile.Name()))
}

func TestWorkflowWriter_BuildWorkflows(t *testing.T) {
	var expectedRoot = "root"
	var expectedHandle = "handle"
	var expectedLink = "link"
	var expectedNode = "node"
	var testWriter = &WorkflowWriter{
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
	var testWriter = &WorkflowWriter{
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	assert.Empty(t, testWriter.GetWorkflows())
	testWriter.WithCurrent(&broker.Solution{
		Workflows: []broker.Workflow{
			{
				Handle: expectedHandle,
			},
		},
	})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestWorkflowWriter_GetWorkflowByHandle_NotFound(t *testing.T) {
	var expectedHandle = "handle"
	var testWriter = &WorkflowWriter{
		workflows: make(map[string]broker.Workflow),
	}
	var actual *broker.Workflow
	var err error

	assert.Empty(t, testWriter.GetWorkflows())
	testWriter.current = &broker.Solution{
		Workflows: []broker.Workflow{
			{
				Handle: "otherHandle",
			},
		},
	}
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.Error(t, err)
	assert.Empty(t, actual)
}

func TestWorkflowWriter_Write(t *testing.T) {
	var expectedHandle = "handle"
	var testLedger = NewBuilder().WithViper(viper.New()).Build()
	var testWriter = NewWorkflowWriter()
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")
	var actual *broker.Workflow

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	testWriter.WithWorkflow(&broker.Workflow{
		Handle: expectedHandle,
	})
	assert.NoError(t, testWriter.Write(testFile.Name()))
	actual, err = testWriter.Read(testLedger, testFile.Name()).
		GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func TestWorkflowWriter_WritePermissionsError(t *testing.T) {
	var invalidPath = "/invalid/path.yaml"
	var testWriter = &WorkflowWriter{}

	assert.Error(t, testWriter.Write(invalidPath))
}

func TestWorkflowWriter_WriteYamlError(t *testing.T) {
	var testWriter = &WorkflowWriter{}
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	assert.NoError(t, os.WriteFile(testFile.Name(), []byte("--"), 0660))
	assert.Error(t, testWriter.Write(testFile.Name()))
}

func Test_newWorkflowWriterYamlError(t *testing.T) {
	var testLedger = NewBuilder().WithViper(viper.New()).Build()
	var testFile, err = os.CreateTemp("", "genaiz_workflowWriter*.yaml")
	var testWriter *WorkflowWriter

	assert.NoError(t, err)
	defer filez.RemoveSilently(testFile.Name())
	assert.NoError(t, os.WriteFile(testFile.Name(), []byte("--"), 0660))
	testLedger.InitLogging()
	testWriter = NewWorkflowWriter().Read(testLedger, testFile.Name())
	assert.Error(t, testWriter.Write(testFile.Name()))
	assert.Empty(t, testWriter.GetWorkflows())
}
