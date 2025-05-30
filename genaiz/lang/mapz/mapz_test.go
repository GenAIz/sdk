package mapz

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMapped(t *testing.T) {
	var testKey = "key"
	var testSlice = []string{"value"}
	var testKeySupplier = func(v string) string { return testKey }
	var testMap = Mapped(testSlice, testKeySupplier)

	assert.EqualValues(t, testSlice[0], testMap[testKey])
}

func TestSorted(t *testing.T) {
	var testKey1 = "key9"
	var testKey2 = "key1"
	var testMap = map[string]string{testKey1: "value", testKey2: "anotherValue", "key3": "value3"}
	var testResults []string

	Sorted(testMap, func(key string) {
		testResults = append(testResults, key)
	})

	assert.True(t, slices.Index(testResults, testKey1) > slices.Index(testResults, testKey2))
}
