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
	addLinks     *[]string
	addWorkflow  *string
	initError    error
	initLinks    *[]string
	initResult   []string
	initWorkflow *string
}

func (sa *stubAddExecutor) Add(wf string, links []string) {
	*sa.addWorkflow = wf
	*sa.addLinks = links
}

func (sa *stubAddExecutor) Init(wf string, links []string) ([]string, error) {
	if sa.initWorkflow != nil {
		*sa.initWorkflow = wf
	}

	if sa.initLinks != nil {
		*sa.initLinks = links
	}

	return sa.initResult, sa.initError
}

func TestNewAddLinks(t *testing.T) {
	var actualWorkflow string
	var actualLinks []string
	var expectedWorkflow = "_workflow"
	var expectedLinks = "_add-handle"
	var testAddLinks = NewAddLinks(
		func(command *cobra.Command) AddExecutor {
			return &stubAddExecutor{
				addWorkflow: &actualWorkflow,
				addLinks:    &actualLinks,
				initResult:  []string{expectedLinks},
			}
		},
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

func newAddValidator(valid bool) AddValidator {
	return func(args []string) error {
		if valid {
			return nil
		} else {
			return errorInvalid
		}
	}
}
