package slicez

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	var elements = []string{
		"element 1",
		"not element",
		"element 2",
	}

	var actual = Filter(elements, func(s string) bool {
		return strings.HasPrefix(s, "element")
	})

	assert.Equal(t, 2, len(actual))
	assert.Equal(t, elements[0], actual[0])
	assert.Equal(t, elements[2], actual[1])
}
