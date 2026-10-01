package slicez

import (
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
)

func TestFilter(t *testing.T) {
	var elements = []string{
		"element 1",
		"not element",
		"element 2",
	}

	actual := Filter(elements, func(s string) bool {
		return strings.HasPrefix(s, "element")
	})
	assert.Equal(t, 2, len(actual))
	assert.Equal(t, elements[0], actual[0])
	assert.Equal(t, elements[2], actual[1])
}

func TestTransform(t *testing.T) {
	var elements = []int{1, 2, 3, 4, 5}
	var expected = []string{"1", "2", "3", "4", "5"}

	actual := Transform(elements, func(i int) string {
		return cast.ToString(i)
	})
	assert.Equal(t, expected, actual)
}
