package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	removeKey  *string
	removeLink *string
}

func (sa *stubRemoveExecutor) Remove(removeLink, removeKey string) error {
	*sa.removeKey = removeKey
	*sa.removeLink = removeLink
	return nil
}

func TestNewRemoveSpec(t *testing.T) {
	var actualLink, actualKey string
	var expectedKey = "propSpecKey"
	var expectedLink = "datalink"
	var testRmSpec = NewRemoveSpec(newRemoveFactory(&actualLink, &actualKey))

	testRmSpec.SetArgs([]string{expectedLink, expectedKey})
	assert.NoError(t, testRmSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
	assert.Equal(t, expectedLink, actualLink)
}

func newRemoveFactory(actualLink, actualKey *string) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			removeKey:  actualKey,
			removeLink: actualLink,
		}
	}
}
