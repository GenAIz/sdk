//go:build prod

package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultProtocolPrefix(t *testing.T) {
	// prod will not allow http: under any circumstances
	assert.Equal(t, "https:/", DefaultProtocolPrefix("localhost:31337"))
}

func TestGetVersionTag(t *testing.T) {
	// prod versions are not qualified by the compiler
	assert.Empty(t, GetVersionTag())
}

func TestIsAllowedProtocol(t *testing.T) {
	assert.False(t, IsAllowedProtocol("anything_goes"))
	assert.False(t, IsAllowedProtocol("http://localhost:8081"))
	assert.True(t, IsAllowedProtocol("https://localhost:8081"))
}
