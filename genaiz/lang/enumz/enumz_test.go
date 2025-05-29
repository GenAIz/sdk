package enumz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEnumType_AllFromStrings(t *testing.T) {
	var expectedValue = "value"
	var testType = NewEnumType("test", "another", expectedValue)
	var value, _ = testType.AllFromStrings(&[]string{expectedValue})

	assert.EqualValues(t, expectedValue, value[0])
}

func TestEnumType_AllFromStringsError(t *testing.T) {
	var testType = NewEnumType("test", "another")
	var _, err = testType.AllFromStrings(&[]string{"notValid"})

	assert.Error(t, err)
}

func TestEnumType_AllFromStringsNilLabels(t *testing.T) {
	var testType = NewEnumType("test", "another")

	assert.Panics(t, func() {
		_, _ = testType.AllFromStrings(nil)
	})
}

func TestEnumType_FromOrdinal(t *testing.T) {
	var expectedValue = "value"
	var testType = NewEnumType("test", "another", expectedValue)
	var value, _ = testType.FromOrdinal(2)

	assert.EqualValues(t, expectedValue, *value)
}

func TestEnumType_FromOrdinalError(t *testing.T) {
	var testType = NewEnumType("test", "another")
	var _, err = testType.FromOrdinal(2)

	assert.Error(t, err)
}

func TestEnumType_IsValid(t *testing.T) {
	var testType = NewEnumType("test", "another")

	assert.True(t, testType.IsValid("test"))
}
