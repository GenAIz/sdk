package adapt

import (
	"testing"

	"github.com/moby/moby/api/types/container"
	"github.com/stretchr/testify/assert"
)

func TestRunArguments_WaitCondition(t *testing.T) {
	var testParams = &RunArguments{}

	assert.Equal(t, container.WaitConditionNextExit, testParams.WaitCondition())
	testParams.Dispose = true
	assert.Equal(t, container.WaitConditionRemoved, testParams.WaitCondition())
}
