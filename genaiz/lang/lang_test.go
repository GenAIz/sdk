package lang

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-it/mock"
)

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

func TestSupplier(t *testing.T) {
	var expectedValue = "value"

	assert.EqualValues(t, expectedValue, *Supplier(&expectedValue)())
}
