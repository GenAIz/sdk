package wf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestBaseExecutor_findFunctionByOemHandle(t *testing.T) {
	var expectedNodePath = "node"
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var testLedger = config.NewBuilder().
		WithWorkDir(t.TempDir()).
		Build()
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}
	var err error

	if err = os.MkdirAll(filepath.Join(testLedger.WorkDir, expectedNodePath), 0750); err == nil {
		var fd *os.File

		if fd, err = os.Create(filepath.Join(testLedger.WorkDir, ".notGenaiz.yaml")); err == nil {
			var vp = viper.New()

			filez.CloseSilently(fd)
			vp.Set(schema.Genaiz.Function.Publish.Internal.Doc, &broker.Function{
				Oem:     expectedOem,
				Handle:  expectedHandle,
				Version: "1.0.0",
			})

			if err = vp.WriteConfigAs(filepath.Join(testLedger.WorkDir, expectedNodePath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
				var actual *broker.Function

				t.Chdir(testLedger.WorkDir)
				actual, err = testExecutor.findFunctionByOemHandle(testLedger.WorkDir, expectedOem, expectedHandle)
				assert.NotNil(t, actual)
				assert.NoError(t, err)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestBaseExecutor_findFunctionByOemHandle_Empty(t *testing.T) {
	var testPath = t.TempDir()
	var testExecutor = &BaseExecutor{}
	var actual, err = testExecutor.findFunctionByOemHandle(testPath, "oem", "handle")

	assert.Nil(t, actual)
	assert.ErrorIs(t, err, errorNoFunction)
}

func TestBaseExecutor_findFunctionByOemHandle_Invalid(t *testing.T) {
	var expectedNodePath = "node"
	var testLedger = config.NewBuilder().
		WithWorkDir(t.TempDir()).
		Build()
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}
	var err error

	if err = os.MkdirAll(filepath.Join(testLedger.WorkDir, expectedNodePath), 0750); err == nil {
		var testPath = filepath.Join(testLedger.WorkDir, expectedNodePath, testLedger.ConfigName+"."+shared.ConfigTypeYaml)
		var fd *os.File

		if fd, err = os.Create(testPath); err == nil {
			var actual *broker.Function

			filez.CloseSilently(fd)
			t.Chdir(testLedger.WorkDir)
			actual, err = testExecutor.findFunctionByOemHandle(testLedger.WorkDir, "oem", "handle")
			assert.Nil(t, actual)
			assert.Error(t, err)
			return
		}
	}

	assert.NoError(t, err)
}

func TestCli_WorkingConfigType(t *testing.T) {
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+".toml")); err == nil {
		var testFunc = testCli.WorkingConfigType()

		defer filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		assert.Equal(t, shared.ConfigTypeToml, testFunc(testLedger))
		return
	}

	assert.NoError(t, err)
}

func TestCli_WorkingConfigType_SyntaxError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = &Cli{}
	var fd *os.File
	var err error

	defer patch.Unpatch()

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+".json")); err == nil {
		var testFunc = testCli.WorkingConfigType()

		defer filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		assert.Nil(t, testFunc(testLedger))
		assert.EqualValues(t, 1, patch.CalledWith)
		return
	}

	assert.NoError(t, err)
}

func TestNewWf(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWf(testLedger, nil, nil, nil)

	assert.Equal(t, 5, len(testCmd.Commands()))
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
	_, _ = testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
	_, _ = testWriter.addLinks(expectedHandle, []broker.WorkflowLink{{LhsNode: expectedLink}})
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

func newWorkflowWriterFactory(stubSolution *broker.Solution) workflowWriterFactory {
	return func(*config.Ledger, string) *workflowWriter {
		var stub = &workflowWriter{
			WorkflowWriter: config.NewWorkflowWriter(),
		}

		stub.WithCurrent(stubSolution)
		return stub
	}
}
