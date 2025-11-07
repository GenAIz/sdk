package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubAddExecutor struct {
	addKey *string
}

func (sa *stubAddExecutor) Add(addKey string) error {
	*sa.addKey = addKey
	return nil
}

func TestNewAddSpec(t *testing.T) {
	var actualKey string
	var expectedKey = "propSpecKey"
	var testAddSpec = NewAddSpec(newAddFactory(&actualKey), newAddValidator(nil))

	testAddSpec.SetArgs([]string{expectedKey})
	assert.NoError(t, testAddSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
}

func newAddFactory(expectedKey *string) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addKey: expectedKey,
		}
	}
}

func newAddValidator(err error) AddValidator {
	return func(key string) error {
		return err
	}
}
