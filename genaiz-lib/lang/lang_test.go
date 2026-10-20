package lang

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

type testInterface interface {
	Test()
}

type stubInterface struct {
}

func (si stubInterface) Test() {
}

func TestFirstNonNil(t *testing.T) {
	var expectedValue = new("stringValue")
	var nilRef *string
	var nilValue string

	assert.Same(t, expectedValue, FirstNonNil(nilRef, expectedValue))
	assert.NotSame(t, expectedValue, FirstNonNil(&nilValue, expectedValue))
	assert.Nil(t, FirstNonNil(nilRef, nilRef, nilRef))
}

func TestFirstNonNil_Interface(t *testing.T) {
	var expectedInterface testInterface
	var nilInterface *testInterface

	expectedInterface = &stubInterface{}
	assert.NotSame(t, expectedInterface, FirstNonNil(nilInterface, &expectedInterface))
}
