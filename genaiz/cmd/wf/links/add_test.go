package links

import (
	"errors"
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

var (
	errorInvalid = errors.New("test error")
)

type stubAddExecutor struct {
	actualWorkflow *string
	actualLinks    *[]string
}

func (sa *stubAddExecutor) Add(wf string, links []string) {
	*sa.actualWorkflow = wf
	*sa.actualLinks = links
}

func TestNewAddLinks(t *testing.T) {
	var actualWorkflow string
	var actualLinks []string
	var expectedWorkflow = "_workflow"
	var expectedLinks = "_add-handle"
	var testAddLinks = NewAddLinks(
		newAddFactory(&actualWorkflow, &actualLinks),
		newAddValidator(true))

	testAddLinks.SetArgs([]string{expectedWorkflow, expectedLinks})
	assert.NoError(t, testAddLinks.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedLinks, actualLinks[0])
}

func TestNewAddLinksInvalidArgs(t *testing.T) {
	var testAddLinks = NewAddLinks(nil, newAddValidator(false))

	testAddLinks.SetArgs([]string{"_workflow", "_add-handle"})
	assert.ErrorIs(t, testAddLinks.Execute(), errorInvalid)
}

func newAddFactory(actualWorkflow *string, actualLinks *[]string) AddExecutorFactory {
	return func(command *cobra.Command) AddExecutor {
		return &stubAddExecutor{
			actualWorkflow: actualWorkflow,
			actualLinks:    actualLinks,
		}
	}
}

func newAddValidator(valid bool) AddValidator {
	return func(args []string) error {
		if valid {
			return nil
		} else {
			return errorInvalid
		}
	}
}
