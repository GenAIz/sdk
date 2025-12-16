package proxy

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubRemoveExecutor struct {
	rmHost *string
	rmPort *int
}

func (sa *stubRemoveExecutor) Remove(rmHost string, rmPort int) {
	*sa.rmHost = rmHost
	*sa.rmPort = rmPort
}

func TestNewRemoveProxy(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testHost string
	var testPort int
	var testCmd = NewRemoveProxy(newRemoveFactory(&testHost, &testPort))

	defer patch.Unpatch()
	testCmd.SetArgs([]string{"host:1133"})
	assert.NoError(t, testCmd.Execute())
	assert.Empty(t, patch.CalledWith)
	assert.Equal(t, testHost, "host")
	assert.Equal(t, testPort, 1133)
}

func TestNewRemoveProxy_InvalidHostPort(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testHost string
	var testPort int
	var testCmd = NewRemoveProxy(newRemoveFactory(&testHost, &testPort))

	defer patch.Unpatch()
	testCmd.SetArgs([]string{"host:-1"})
	assert.NoError(t, testCmd.Execute())
	assert.Empty(t, testHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newRemoveFactory(expectedHost *string, expectedPort *int) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			rmHost: expectedHost,
			rmPort: expectedPort,
		}
	}
}
