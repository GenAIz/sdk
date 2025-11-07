package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubEditExecutor struct {
	editKey *string
}

func (sa *stubEditExecutor) Edit(editKey string) error {
	*sa.editKey = editKey
	return nil
}

func TestNewEditSpec(t *testing.T) {
	var actualKey string
	var expectedKey = "propSpecKey"
	var testEditSpec = NewEditSpec(newEditFactory(&actualKey))

	testEditSpec.SetArgs([]string{expectedKey})
	assert.NoError(t, testEditSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
}

func newEditFactory(expectedKey *string) EditExecutorFactory {
	return func(command *cobra.Command) EditExecutor {
		return &stubEditExecutor{
			editKey: expectedKey,
		}
	}
}
