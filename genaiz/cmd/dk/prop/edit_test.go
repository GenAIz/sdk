package prop

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubEditExecutor struct {
	editKey  *string
	editLink *string
}

func (sa *stubEditExecutor) Edit(editLink, editKey string) error {
	*sa.editKey = editKey
	*sa.editLink = editLink
	return nil
}

func TestNewEditSpec(t *testing.T) {
	var actualLink, actualKey string
	var expectedLink = "datalink"
	var expectedKey = "propSpecKey"
	var testEditSpec = NewEditSpec(newEditFactory(&actualLink, &actualKey))

	testEditSpec.SetArgs([]string{expectedLink, expectedKey})
	assert.NoError(t, testEditSpec.Execute())
	assert.Equal(t, expectedKey, actualKey)
	assert.Equal(t, expectedLink, actualLink)
}

func newEditFactory(actualLink, actualKey *string) EditExecutorFactory {
	return func(command *cobra.Command) EditExecutor {
		return &stubEditExecutor{
			editKey:  actualKey,
			editLink: actualLink,
		}
	}
}
