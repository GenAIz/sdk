//go:build dev

package env

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDefaultProtocolPrefix(t *testing.T) {
	assert.Equal(t, "http:/", DefaultProtocolPrefix("localhost:31337"))
	assert.Equal(t, "https:/", DefaultProtocolPrefix("test"))
}

func TestGetVersionTag(t *testing.T) {
	assert.Equal(t, "-dev", GetVersionTag())
}

func TestIsAllowedProtocol(t *testing.T) {
	assert.True(t, IsAllowedProtocol("anything_goes"))
}
