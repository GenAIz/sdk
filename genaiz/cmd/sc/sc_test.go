package sc

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNewSc(t *testing.T) {
	var testSc = NewSc()

	assert.NotEmpty(t, testSc.Use)
	assert.NotEmpty(t, testSc.Aliases)
	assert.NotEmpty(t, testSc.Short)
}
