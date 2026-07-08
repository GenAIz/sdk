package version

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetVersion(t *testing.T) {
	assert.NotEmpty(t, GetVersion())
}

func TestGetVersion_Head(t *testing.T) {
	Head = "expectedHead"
	suffix := fmt.Sprintf("(%s)", Head)
	assert.True(t, strings.HasSuffix(GetVersion(), suffix))
	Head = ""
}
