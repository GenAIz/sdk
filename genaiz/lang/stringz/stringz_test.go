package stringz

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestAllNonEmpty(t *testing.T) {
	var expectedValue = "value"
	var testStrings = []string{"test", "", "another", "value"}

	assert.EqualValues(t, expectedValue, AllNonEmpty(testStrings...)[2])
}

func TestFirstNonEmpty(t *testing.T) {
	var expectedValue = "value"

	assert.EqualValues(t, expectedValue, FirstNonEmpty("", "", expectedValue))
}

func TestMultiTagLabel(t *testing.T) {
	var expected = "label-tag"

	assert.EqualValues(t, expected, MultiTagLabel("label", "-", "tag"))
}

func TestMultiTagLabelEmptyDelimiter(t *testing.T) {
	assert.Empty(t, MultiTagLabel("", "", ""))
}

func TestMultiTagLabelEmptyTag(t *testing.T) {
	var expectedLabel = "label"

	assert.EqualValues(t, expectedLabel, MultiTagLabel(expectedLabel, "-", ""))
}

func TestSingleTagLabel(t *testing.T) {
	var expectedLabel = "label:tag"

	assert.EqualValues(t, expectedLabel, SingleTagLabel("label", ":", "tag"))

}

func TestSingleTagLabelEmptyDelimiter(t *testing.T) {
	var expectedLabel = "label"

	assert.EqualValues(t, expectedLabel, SingleTagLabel(expectedLabel, "", "tag"))
}

func TestSingleTagLabelExistingDelimiter(t *testing.T) {
	var expectedLabel = "label:tag"

	assert.EqualValues(t, expectedLabel, SingleTagLabel(expectedLabel, ":", "tag2"))
}

func TestSingleTagLabelEmptyTag(t *testing.T) {
	var expectedLabel = "label"

	assert.EqualValues(t, expectedLabel, SingleTagLabel(expectedLabel, ":", ""))
}
