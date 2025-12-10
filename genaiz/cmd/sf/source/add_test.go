package source

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

type stubAddExecutor struct {
	addKey   *string
	addError error
}

func (sa *stubAddExecutor) Add(addKey string) error {
	*sa.addKey = addKey
	return sa.addError
}

func TestNewAddSource(t *testing.T) {
	var actualFqdnv string
	var expectedFqdnv = "fqdnv"
	var testAddSource = NewAddSource(newAddFactory(&actualFqdnv, nil))

	testAddSource.SetArgs([]string{expectedFqdnv})
	assert.NoError(t, testAddSource.Execute())
	assert.Equal(t, expectedFqdnv, actualFqdnv)
}

func TestNewAddSource_Error(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedError = errors.New("expected")
	var actualFqdnv string
	var expectedFqdnv = "fqdnv"
	var testAddSpec = NewAddSource(newAddFactory(&actualFqdnv, expectedError))

	defer patch.Unpatch()
	testAddSpec.SetArgs([]string{expectedFqdnv})
	assert.NoError(t, testAddSpec.Execute())
	assert.Equal(t, expectedFqdnv, actualFqdnv)
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func newAddFactory(expectedKey *string, expectedError error) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addKey:   expectedKey,
			addError: expectedError,
		}
	}
}
