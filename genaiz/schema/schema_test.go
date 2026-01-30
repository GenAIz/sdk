package schema

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

type stubStruct struct {
	A string
}

func TestKeys_GetString(t *testing.T) {
	var expectedValue = "value"
	var testViper = viper.New()
	var testKeys = &Keys{Doc: "key"}

	testViper.Set(testKeys.Doc, expectedValue)
	assert.Equal(t, expectedValue, testKeys.GetString(testViper))
}

func TestKeys_GetString_DefaultValue(t *testing.T) {
	var expectedDefault = []string{"default", "Value"}
	var testViper = viper.New()
	var testKeys = &Keys{Doc: "key", Pseudonyms: []string{"no", "yes"}}

	assert.Equal(t, "defaultValue", testKeys.GetString(testViper, expectedDefault...))
}

func TestKeys_GetString_Pseudonyms(t *testing.T) {
	var expectedPseudo = "yes"
	var expectedValue = "value"
	var testViper = viper.New()
	var testKeys = &Keys{Doc: "key", Pseudonyms: []string{"no", expectedPseudo}}

	testViper.Set(expectedPseudo, expectedValue)
	assert.Equal(t, expectedValue, testKeys.GetString(testViper))
}

func TestKeys_Unmarshall(t *testing.T) {
	var testKeys = Keys{Doc: "found"}
	var testViper = viper.New()
	var expected = stubStruct{A: "value"}
	var actual stubStruct

	testViper.Set(testKeys.Doc, expected)
	assert.NoError(t, testKeys.Unmarshall(testViper, &actual))
	assert.Equal(t, expected, actual)
}

func TestKeys_Unmarshall_NotFound(t *testing.T) {
	var testKeys = Keys{Doc: "notFound"}
	var testViper = viper.New()
	var actual struct{ A string }

	assert.Error(t, testKeys.Unmarshall(testViper, &actual))
	assert.Empty(t, actual)
}

func TestKeys_Unmarshall_Pseudonym(t *testing.T) {
	var expectedPseudo = "found"
	var testKeys = Keys{Doc: "notfound", Pseudonyms: []string{"stillNotFound", expectedPseudo}}
	var testViper = viper.New()
	var expected = stubStruct{A: "value"}
	var actual stubStruct

	testViper.Set(expectedPseudo, expected)
	assert.NoError(t, testKeys.Unmarshall(testViper, &actual))
	assert.Equal(t, expected, actual)
}

func TestKeys_Unmarshall_PseudonymNotFound(t *testing.T) {
	var testKeys = Keys{Doc: "notfound", Pseudonyms: []string{"stillNotFound", "neverFound"}}
	var testViper = viper.New()
	var tested = stubStruct{A: "value"}
	var actual stubStruct

	testViper.Set("found", tested)
	assert.Error(t, testKeys.Unmarshall(testViper, &actual))
	assert.Empty(t, actual)
}

func TestNormalize(t *testing.T) {
	var expectedAcValue = "expectedAcValue"
	var expectedSfValue = "expectedSfValue"
	var expectedSnValue = "expectedSnValue"
	var expectedWfValue = "expectedWfValue"
	var replacedSfValue = "replacedSfValue"
	var testViper = viper.New()

	testViper.Set("sf.test.key", replacedSfValue)
	testViper.Set("function.test.key", expectedSfValue)
	testViper.Set("ac.test.key", expectedAcValue)
	testViper.Set("sn.test.key", expectedSnValue)
	testViper.Set("wf.test.key", expectedWfValue)
	actualViper := Normalize(testViper)
	assert.Equal(t, expectedAcValue, actualViper.Get("account.test.key"))
	assert.Equal(t, expectedSfValue, actualViper.Get("function.test.key"))
	assert.Equal(t, expectedSnValue, actualViper.Get("solution.test.key"))
	assert.Equal(t, expectedWfValue, actualViper.Get("workflow.test.key"))
	assert.Nil(t, actualViper.Get("ac.test.key"))
	assert.Nil(t, actualViper.Get("sf.test.key"))
	assert.Nil(t, actualViper.Get("sn.test.key"))
	assert.Nil(t, actualViper.Get("wf.test.key"))
}
