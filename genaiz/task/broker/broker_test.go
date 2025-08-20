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
