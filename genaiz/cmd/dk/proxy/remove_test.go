package proxy

import (
	"fmt"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	removeLink *string
	removeHost *string
	removePort *int
}

func (sa *stubRemoveExecutor) Remove(removeLink, removeHost string, removePort int) error {
	*sa.removeLink = removeLink
	*sa.removeHost = removeHost
	*sa.removePort = removePort
	return nil
}

func TestNewRemoveSpec(t *testing.T) {
	var actualLink, actualHost string
	var actualPort int
	var expectedHost = "testHost"
	var expectedLink = "datalink"
	var expectedPort = "37"
	var testRmSpec = NewRemoveProxy(newRemoveFactory(&actualLink, &actualHost, &actualPort))

	testRmSpec.SetArgs([]string{expectedLink, fmt.Sprintf("%s:%s", expectedHost, expectedPort)})
	assert.NoError(t, testRmSpec.Execute())
	assert.Equal(t, expectedLink, actualLink)
	assert.Equal(t, expectedHost, actualHost)
	assert.Equal(t, cast.ToInt(expectedPort), actualPort)
}

func newRemoveFactory(actualLink, actualHost *string, actualPort *int) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			removeLink: actualLink,
			removeHost: actualHost,
			removePort: actualPort,
		}
	}
}
