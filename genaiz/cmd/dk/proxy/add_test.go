package proxy

import (
	"fmt"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubAddExecutor struct {
	addLink *string
	addHost *string
	addPort *int
}

func (sa *stubAddExecutor) Add(addLink, addHost string, addPort int) error {
	*sa.addLink = addLink
	*sa.addHost = addHost
	*sa.addPort = addPort
	return nil
}

func TestNewAddProxy(t *testing.T) {
	var actualLink, actualHost string
	var actualPort int
	var expectedLink = "testLink"
	var expectedHost = "testHost"
	var expectedPort = "37"
	var testAddSpec = NewAddProxy(newAddFactory(&actualLink, &actualHost, &actualPort))

	testAddSpec.SetArgs([]string{expectedLink, fmt.Sprintf("%s:%s", expectedHost, expectedPort)})
	assert.NoError(t, testAddSpec.Execute())
	assert.Equal(t, expectedLink, actualLink)
	assert.Equal(t, expectedHost, actualHost)
	assert.Equal(t, cast.ToInt(expectedPort), actualPort)
}

func newAddFactory(actualLink, actualHost *string, actualPort *int) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addLink: actualLink,
			addHost: actualHost,
			addPort: actualPort,
		}
	}
}
