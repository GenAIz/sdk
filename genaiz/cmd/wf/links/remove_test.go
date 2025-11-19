package links

import (
	"testing"

	"github.com/spf13/cobra"
	"github.com/stretchr/testify/assert"
)

type stubRemoveExecutor struct {
	initError    error
	initLinks    *[]string
	initResult   []string
	initWorkflow *string
	rmLinks      *[]string
	rmWorkflow   *string
}

func (sr *stubRemoveExecutor) Init(wf string, links []string) ([]string, error) {
	if sr.initWorkflow != nil {
		*sr.initWorkflow = wf
	}

	if sr.initLinks != nil {
		*sr.initLinks = links
	}

	return sr.initResult, sr.initError
}

func (sr *stubRemoveExecutor) Remove(wf string, links []string) {
	*sr.rmWorkflow = wf
	*sr.rmLinks = links
}

func TestRemoveAddLinks(t *testing.T) {
	var actualWorkflow string
	var actualLinks []string
	var expectedWorkflow = "_workflow"
	var expectedLinks = "_add-handle"
	var testRmLinks = NewRemoveLinks(
		func(command *cobra.Command) RemoveExecutor {
			return &stubRemoveExecutor{
				rmWorkflow: &actualWorkflow,
				rmLinks:    &actualLinks,
				initResult: []string{expectedLinks},
			}
		},
		newRemoveValidator(true))

	testRmLinks.SetArgs([]string{expectedWorkflow, expectedLinks})
	assert.NoError(t, testRmLinks.Execute())
	assert.Equal(t, expectedWorkflow, actualWorkflow)
	assert.Equal(t, expectedLinks, actualLinks[0])
}

func TestNewRemoveLinksInvalidArgs(t *testing.T) {
	var testRmLinks = NewRemoveLinks(nil, newRemoveValidator(false))

	testRmLinks.SetArgs([]string{"_workflow", "_rm-handle"})
	assert.ErrorIs(t, testRmLinks.Execute(), errorInvalid)
}

func newRemoveFactory(actualWorkflow *string, actualLinks *[]string) RemoveExecutorFactory {
	return func(command *cobra.Command) RemoveExecutor {
		return &stubRemoveExecutor{
			rmWorkflow: actualWorkflow,
			rmLinks:    actualLinks,
		}
	}
}

func newRemoveValidator(valid bool) RemoveValidator {
	return func(args []string) error {
		if valid {
			return nil
		} else {
			return errorInvalid
		}
	}
}
