package secretz

import (
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
)

func TestFromEnclave(t *testing.T) {
	var expectedValue = "value"
	var testEnclave = memguard.NewEnclave([]byte(expectedValue))
	var actual = FromEnclave(testEnclave)

	defer actual.Destroy()
	assert.Equal(t, expectedValue, actual.String())
}

func TestFromEnclave_Empty(t *testing.T) {
	assert.Empty(t, FromEnclave(nil).String())
}
