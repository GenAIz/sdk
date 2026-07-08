package intz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestInt64ToDefault(t *testing.T) {
	var expectedValue = int64(64)

	assert.Equal(t, expectedValue, Int64ToDefault(&expectedValue, int64(37)))
}

func TestInt64ToDefault_Default(t *testing.T) {
	var expectedDefault = int64(37)

	assert.Equal(t, expectedDefault, Int64ToDefault(nil, expectedDefault))
}
