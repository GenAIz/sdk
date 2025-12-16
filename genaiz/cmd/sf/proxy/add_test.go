package proxy

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubAddExecutor struct {
	addHost  *string
	addPort  *int
	addError error
}

func (sa *stubAddExecutor) Add(addHost string, addPort int) error {
	*sa.addHost = addHost
	*sa.addPort = addPort
	return sa.addError
}

func TestNewAddProxy(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testHost string
	var testPort int
	var testCmd = NewAddProxy(newAddFactory(&testHost, &testPort, nil))

	defer patch.Unpatch()
	testCmd.SetArgs([]string{"host:1133"})
	assert.NoError(t, testCmd.Execute())
	assert.Empty(t, patch.CalledWith)
	assert.Equal(t, testHost, "host")
	assert.Equal(t, testPort, 1133)
}

func TestNewAddProxy_AddError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expectedError")
	var testHost string
	var testPort int
	var testCmd = NewAddProxy(newAddFactory(&testHost, &testPort, expectedError))

	defer patch.Unpatch()
	testCmd.SetArgs([]string{"host:1133"})
	assert.NoError(t, testCmd.Execute())
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewAddProxy_InvalidHostPort(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testHost string
	var testPort int
	var testCmd = NewAddProxy(newAddFactory(&testHost, &testPort, nil))

	defer patch.Unpatch()
	testCmd.SetArgs([]string{"host:-1"})
	assert.NoError(t, testCmd.Execute())
	assert.Empty(t, testHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newAddFactory(expectedHost *string, expectedPort *int, expectedError error) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addHost:  expectedHost,
			addPort:  expectedPort,
			addError: expectedError,
		}
	}
}
