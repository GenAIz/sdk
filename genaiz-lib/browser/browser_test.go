package browser

import (
	"os"
	"os/exec"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestOpenUrl(t *testing.T) {
	var restoredProviders = shellProviders

	defer func() {
		shellProviders = restoredProviders
	}()
	shellProviders = []string{"echo"}

	assert.NoError(t, OpenUrl("url", os.Stderr, os.Stdout))
}

func TestOpenUrl_NotFound(t *testing.T) {
	var testDir = t.TempDir()
	var restoredProviders = shellProviders

	defer func() {
		shellProviders = restoredProviders
	}()
	shellProviders = []string{testDir}
	t.Setenv("PATH", "")

	assert.ErrorIs(t, OpenUrl("url", os.Stderr, os.Stdout), exec.ErrNotFound)
}
