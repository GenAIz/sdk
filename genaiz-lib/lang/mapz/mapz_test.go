package mapz

import (
	"slices"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetOrDefault(t *testing.T) {
	var expectedDefault = "default"
	var expectedValue = "value"
	var testKey = "key"
	var testAbsentKey = "absent"
	var testSupplier = func() *string {
		return &expectedDefault
	}
	var testMap = map[string]*string{
		testKey:       &expectedValue,
		testAbsentKey: nil,
	}

	assert.EqualValues(t, expectedValue, *GetOrDefault(testMap, testKey, testSupplier))
	assert.EqualValues(t, expectedDefault, *GetOrDefault(testMap, "unexpectedKey", testSupplier))
	assert.Empty(t, GetOrDefault(testMap, testAbsentKey, testSupplier))
}

func TestMapped(t *testing.T) {
	var testKey = "key"
	var testSlice = []string{"value"}
	var testKeySupplier = func(v string) string { return testKey }
	var testMap = Mapped(testSlice, testKeySupplier)

	assert.EqualValues(t, testSlice[0], testMap[testKey])
}

func TestMappedInt64(t *testing.T) {
	var testKey = int64(37)
	var testSlice = []string{"value"}
	var testKeySupplier = func(v string) int64 { return testKey }
	var testMap = MappedInt64(testSlice, testKeySupplier)

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
