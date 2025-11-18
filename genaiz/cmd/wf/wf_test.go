package wf

import (
	"os"
	"path/filepath"
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

	testViper.Set(testOption.Key, shared.ConfigTypeJson)
	actual, err = testExecutor.makeConfigParams(testOption)
	assert.NoError(t, err)
	assert.Equal(t, testLedger.ConfigName, actual.ConfigName)
	assert.Equal(t, shared.ConfigTypeJson, *actual.ConfigType)
}

func TestBaseExecutor_makeConfigParams_FileInvalidError(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, "Genaiz.txt")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionConfigType("test")
	var testExecutor = &BaseExecutor{Ledger: testLedger}
	var actual *shared.ConfigParams
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		t.Chdir(testDir)
		actual, err = testExecutor.makeConfigParams(testOption)
		assert.Error(t, err)
		assert.Empty(t, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBaseExecutor_makeConfigParams_TypeNone(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, "Genaiz.yaml")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionConfigType("test")
	var testExecutor = &BaseExecutor{Ledger: testLedger}
	var actual *shared.ConfigParams
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		t.Chdir(testDir)
		actual, err = testExecutor.makeConfigParams(testOption)
		assert.NoError(t, err)
		assert.Equal(t, testLedger.ConfigName, actual.ConfigName)
		assert.Equal(t, shared.ConfigTypeYaml, *actual.ConfigType)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestBaseExecutor_makeConfigParams_TypeNoneError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = newOptionConfigType("test")
	var testExecutor = &BaseExecutor{Ledger: testLedger}
	var actual *shared.ConfigParams
	var err error

	actual, err = testExecutor.makeConfigParams(testOption)
	assert.Error(t, err)
	assert.Empty(t, actual)
}

func TestNewWf(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWf(testLedger, nil, nil, nil)

	assert.Equal(t, 4, len(testCmd.Commands()))
}

func TestWorkflowWriter_addLinks(t *testing.T) {
	var expectedHandle = "handle"
	var expectedLink = "link"
	var testWriter = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}
	var actual *broker.Workflow
	var err error

	testWriter.WithCurrent(&broker.Solution{
		Workflows: []broker.Workflow{{Handle: expectedHandle}},
	})
	testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Links))
	assert.Equal(t, expectedLink, actual.Links[0].LhsNode)
}

func TestWorkflowWriter_addNodes(t *testing.T) {
	var expectedHandle = "handle"
	var expectedNode = "node"
	var testWriter = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}
	var actual *broker.Workflow
	var err error

	testWriter.WithCurrent(&broker.Solution{
		Workflows: []broker.Workflow{{Handle: expectedHandle}},
	})
	_, err = testWriter.addNodes(expectedHandle, &broker.WorkflowNode{Handle: expectedNode})
	assert.NoError(t, err)
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Nodes))
	assert.Equal(t, expectedNode, actual.Nodes[0].Handle)
}

func TestWorkflowWriter_removeLinks(t *testing.T) {
	var expectedHandle = "handle"
	var expectedLink = "link"
	var testWriter = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}
	var actual *broker.Workflow
	var err error

	testWriter.WithCurrent(&broker.Solution{
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
	})
	testWriter.removeLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	testWriter.removeLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Links))
}

func TestWorkflowWriter_removeNodes(t *testing.T) {
	var expectedHandle = "handle"
	var expectedNode = "node"
	var testWriter = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}
	var actual *broker.Workflow
	var err error

	testWriter.WithCurrent(&broker.Solution{
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
	})
	testWriter.removeNodes(expectedHandle, expectedNode)
	testWriter.removeNodes(expectedHandle, expectedNode)
	actual, err = testWriter.GetWorkflowByHandle(expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, 1, len(actual.Nodes))
}
