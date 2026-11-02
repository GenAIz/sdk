package workspace

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlowCreateParams_IsValid(t *testing.T) {
	var testParams = &FlowCreateParams{}

	assert.False(t, testParams.IsValid())
	testParams.WorkspaceId = new(int64(37))
	assert.False(t, testParams.IsValid())
	testParams.WorkflowId = new(int64(45))
	assert.True(t, testParams.IsValid())
	testParams.WorkspaceId = nil
	assert.False(t, testParams.IsValid())
}
