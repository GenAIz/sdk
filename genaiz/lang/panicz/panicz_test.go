package panicz

import (
	"errors"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestPanicIfError(t *testing.T) {
	assert.Panics(t, func() {
		PanicIfError(errors.New("error"))
	})
}

func TestPanicIfErrorNone(t *testing.T) {
	assert.NotPanics(t, func() {
		PanicIfError(nil)
	})
}

func TestRequiresNotNil(t *testing.T) {
	assert.Panics(t, func() {
		RequiresNotNil("value", nil)
	})
}

func TestRequiresNotNilNone(t *testing.T) {
	var testValue = "value"

	assert.NotPanics(t, func() {
		RequiresNotNil("value", &testValue)
	})
}
