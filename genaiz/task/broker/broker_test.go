package broker

import (
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task/shared"
)

func TestFunction_asIdentity(t *testing.T) {
	var actual *shared.Identity
	var function = Function{
		Id:      37,
		Digest:  "digest",
		Img:     "path",
		Version: "version",
	}

	actual = function.asIdentity()
	assert.Equal(t, function.Id, cast.ToInt(actual.Id))
	assert.Equal(t, function.Digest, actual.Hash)
	assert.Equal(t, function.Img, actual.Path)
	assert.Equal(t, function.Version, actual.Version)
}

func TestSolution_asIdentity(t *testing.T) {
	var actual *shared.Identity
	var solution = SolutionRemote{
		Solution: Solution{
			Version: "version",
		},
		Id:     37,
		Digest: "digest",
		Fqdn:   "path",
	}

	actual = solution.asIdentity()
	assert.Equal(t, solution.Id, cast.ToInt64(actual.Id))
	assert.Equal(t, solution.Digest, actual.Hash)
	assert.Equal(t, solution.Fqdn, actual.Path)
	assert.Equal(t, solution.Version, actual.Version)
}

func TestSolution_Merge(t *testing.T) {
	var expectedDescription = "description"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var testSolution = &Solution{
		Handle: expectedHandle,
		Oem:    expectedOem,
	}
	var testUpdate = &Solution{
		Description: expectedDescription,
		Name:        expectedName,
		Version:     expectedVersion,
	}
	var actual = testSolution.Merge(*testUpdate)

	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Empty(t, actual.Workflows)
}

func TestSolution_MergeWorkflows(t *testing.T) {
	var expectedDescription = "description"
	var expectedHandle = "handle"
	var expectedName = "name"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var mergedWorkflow = "merged"
	var sourceWorkflow = "source"
	var testSolution = &Solution{
		Handle:      expectedHandle,
		Oem:         expectedOem,
		Description: expectedDescription,
		Name:        expectedName,
		Version:     expectedVersion,
		Workflows: []Workflow{
			{
				Handle: sourceWorkflow,
			},
		},
	}
	var testUpdate = &Solution{
		Workflows: []Workflow{
			{
				Handle: mergedWorkflow,
			},
		},
	}
	var actual = testSolution.Merge(*testUpdate)

	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, 2, len(actual.Workflows))
	assert.Equal(t, sourceWorkflow, actual.Workflows[0].Handle)
	assert.Equal(t, mergedWorkflow, actual.Workflows[1].Handle)
}

func TestWorkflowHandlePredicate(t *testing.T) {
	var expectedHandle = "handle"
	var testWorkflow = Workflow{Handle: expectedHandle}

	assert.False(t, WorkflowHandlePredicate("notHandle")(testWorkflow))
	assert.True(t, WorkflowHandlePredicate(expectedHandle)(testWorkflow))
}

func TestWorkflowLink_Equals(t *testing.T) {
	var link1 = WorkflowLink{}
	var link2 = WorkflowLink{RhsNodePort: "port"}
	var link3 = WorkflowLink{RhsNode: "node"}
	var link4 = WorkflowLink{LhsNodePort: "otherPort"}

	assert.True(t, link1.Equals(link1))
	assert.False(t, link1.Equals(link2))
	assert.False(t, link1.Equals(link3))
	assert.False(t, link1.Equals(link4))
}

func TestWorkflowNamePredicate(t *testing.T) {
	var expectedName = "name"
	var testWorkflow = Workflow{Name: expectedName}

	assert.False(t, WorkflowNamePredicate("notName")(testWorkflow))
	assert.True(t, WorkflowNamePredicate(expectedName)(testWorkflow))
}

func TestWorkflowNode_Equals(t *testing.T) {
	var node1 = WorkflowNode{}
	var node2 = WorkflowNode{Handle: "handle"}

	assert.True(t, node1.Equals(node1))
	assert.False(t, node1.Equals(node2))
}
