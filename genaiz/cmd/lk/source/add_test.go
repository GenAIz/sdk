package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubAddExecutor struct {
	addError  error
	addHandle string
	addLink   string
}

func (sae *stubAddExecutor) Add(handle string, link string) error {
	sae.addHandle = handle
	sae.addLink = link
	return sae.addError
}

func TestNewAddSource(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var testFactory = newStubAddExecutorFactory(expectedError)
	var testCmd = NewAddSource(testFactory)
	var expectedHandle = "handle"
	var expectedLink = "link"

	defer patch.Unpatch()
	testCmd.SetArgs([]string{expectedHandle, expectedLink})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newStubAddExecutorFactory(expected error) AddExecutorFactory {
	return func(*cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addError: expected,
		}
	}
}
