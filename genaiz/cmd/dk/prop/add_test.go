package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubAddExecutor struct {
	addKey  *string
	addLink *string
}

func (sa *stubAddExecutor) Add(testLink, addKey string) error {
	*sa.addKey = addKey
	*sa.addLink = testLink
	return nil
}

func TestNewAddSpec(t *testing.T) {
	var actualLink, actualKey string
	var expectedLink = "testLink"
	var expectedKey = "propSpecKey"
	var testAddSpec = NewAddSpec(newAddFactory(&actualLink, &actualKey), newAddValidator(nil))

	testAddSpec.SetArgs([]string{expectedLink, expectedKey})
	assert.NoError(t, testAddSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
	assert.Equal(t, expectedLink, actualLink)
}

func newAddFactory(actualLink, actualKey *string) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			addKey:  actualKey,
			addLink: actualLink,
		}
	}
}

func newAddValidator(err error) AddValidator {
	return func(key string) error {
		return err
	}
}
