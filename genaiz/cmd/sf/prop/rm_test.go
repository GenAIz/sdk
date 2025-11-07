package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	removeKey *string
}

func (sa *stubRemoveExecutor) Remove(removeKey string) error {
	*sa.removeKey = removeKey
	return nil
}

func TestNewRemoveSpec(t *testing.T) {
	var actualKey string
	var expectedKey = "propSpecKey"
	var testRmSpec = NewRemoveSpec(newRemoveFactory(&actualKey))

	testRmSpec.SetArgs([]string{expectedKey})
	assert.NoError(t, testRmSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
}

func newRemoveFactory(expectedKey *string) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			removeKey: expectedKey,
		}
	}
}
