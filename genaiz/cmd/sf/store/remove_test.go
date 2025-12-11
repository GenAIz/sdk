package store

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	rmKey *string
}

func (sa *stubRemoveExecutor) Remove(rmKey string) {
	*sa.rmKey = rmKey
}

func TestNewRemoveStore(t *testing.T) {
	var actualFqdnv string
	var expectedFqdnv = "fqdnv"
	var testAddSource = NewRemoveStore(newRemoveFactory(&actualFqdnv))

	testAddSource.SetArgs([]string{expectedFqdnv})
	assert.NoError(t, testAddSource.Execute())
	assert.Equal(t, expectedFqdnv, actualFqdnv)
}

func newRemoveFactory(expectedKey *string) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			rmKey: expectedKey,
		}
	}
}
