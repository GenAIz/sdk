package lang

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRef struct {
	A string
}

func TestAssists(t *testing.T) {
	var testOut string
	var expectedFunction = func(a string, b string, c string) error {
		testOut = fmt.Sprintf("%s%s%s", a, b, c)
		return nil
	}
	var testFunction = Assists("c", expectedFunction)

	assert.NoError(t, testFunction("a", "b"))
	assert.EqualValues(t, "cab", testOut)
}

func TestHandleExit(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})

	defer patch.Unpatch()
	HandleExit(nil)
	assert.False(t, patch.Called)
	HandleExit("should exit")
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestRef(t *testing.T) {
	var expectedString = "test"

	assert.Equal(t, expectedString, *Ref(expectedString))
}

func TestRefs(t *testing.T) {
	var testRef = stubRef{}
	var testRefs = []stubRef{testRef}

	actual := Refs(testRefs)
	assert.Equal(t, testRef, *actual[0])
}

func TestSupplier(t *testing.T) {
	var expectedValue = "value"

	assert.EqualValues(t, expectedValue, *Supplier(&expectedValue)())
}
