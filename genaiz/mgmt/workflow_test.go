package mgmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestUserWorkflow_MarshalJSON(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserWorkflow = &UserWorkflow{
		Id:          new(int64(37)),
		Handle:      "expectedHandle",
		Name:        "expectedName",
		Description: "expectedDescription",
		Created:     testCreated.UnixMilli(),
		Flags:       new(1337),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserWorkflow.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", *testUserWorkflow.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testUserWorkflow.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserWorkflow.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserWorkflow.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", expectedCreated))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testUserWorkflow.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkflow_MarshalJSON_NoCreated(t *testing.T) {
	var testModified = time.Now()
	var expectedModified = timez.NewTodayFormatter().FormatMillis(testModified.UnixMilli())
	var testUserWorkflow = &UserWorkflow{
		Id:          new(int64(37)),
		Handle:      "expectedHandle",
		Name:        "expectedName",
		Description: "expectedDescription",
		Modified:    testModified.UnixMilli(),
		Flags:       new(1337),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserWorkflow.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", *testUserWorkflow.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testUserWorkflow.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserWorkflow.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserWorkflow.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"modified\":\"%s\"", expectedModified))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testUserWorkflow.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkflow_MarshalSlice(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserWorkflow = &UserWorkflow{
		Id:              new(int64(37)),
		SolutionFqdn:    "expectedFqdn",
		SolutionVersion: "expectedVersion",
		Handle:          "expectedHandle",
		Created:         testCreated.UnixMilli(),
		Nodes: []broker.WorkflowNode{
			{
				Id: new(int64(42)),
			},
		},
		Local: true,
	}
	var values []string
	var err error

	if values, err = testUserWorkflow.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(*testUserWorkflow.Id))
		assert.Equal(t, values[1], testUserWorkflow.SolutionFqdn)
		assert.Equal(t, values[2], testUserWorkflow.SolutionVersion)
		assert.Equal(t, values[3], testUserWorkflow.Handle)
		assert.Equal(t, values[4], expectedCreated)
		assert.Equal(t, values[5], "1")
		assert.Equal(t, values[6], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkflow_MarshalSlice_NoCreated(t *testing.T) {
	var testUserWorkflow = &UserWorkflow{
		SolutionFqdn:    "expectedFqdn",
		SolutionVersion: "expectedVersion",
		Handle:          "expectedHandle",
		Nodes: []broker.WorkflowNode{
			{
				Id: new(int64(42)),
			},
		},
	}
	var values []string
	var err error

	if values, err = testUserWorkflow.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Empty(t, values[0])
		assert.Equal(t, values[1], testUserWorkflow.SolutionFqdn)
		assert.Equal(t, values[2], testUserWorkflow.SolutionVersion)
		assert.Equal(t, values[3], testUserWorkflow.Handle)
		assert.Equal(t, values[4], "-")
		assert.Equal(t, values[5], "1")
		assert.Equal(t, values[6], "no")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkflowFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var expectedGraphers = map[string]broker.SolutionGrapher{}
	var testProvider = NewUserWorkflowFacade().
		WithPathGraphers("path", expectedGraphers).
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)

}

func TestUserWorkflowFacade_Provider(t *testing.T) {
	var testProvider = NewUserWorkflowFacade().
		WithParams(&broker.WorkflowListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkflowProvider_Get(t *testing.T) {
	var testParams = &broker.WorkflowListParams{Oem: "oem"}
	var expectedSolutions = []broker.Solution{
		{
			Fqdn:    new("expectedFqdn"),
			Version: "expectedVersion",

			Workflows: []broker.Workflow{
				{
					Handle: "expectedWorkflowHandle",
				},
			},
		},
	}
	var testProvider = &userWorkflowProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                  testParams,
		workflowListTaskFactory: newWorkflowListTaskComplete(expectedSolutions),
	}
	var actual []UserWorkflow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, *expectedSolutions[0].Fqdn, actual[0].SolutionFqdn)
		assert.Equal(t, expectedSolutions[0].Version, actual[0].SolutionVersion)
		return
	}

	assert.NoError(t, err)
}

func TestUserWorkflowProvider_Get_Failure(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testParams = &broker.WorkflowListParams{Oem: "oem"}
	var testProvider = &userWorkflowProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                    testParams,
		path:                      t.TempDir(),
		solutionCollectTasFactory: newSolutionCollectTaskCompleteError(expectedError),
	}
	var err error

	if _, err = testProvider.Get(); err != nil {
		assert.ErrorIs(t, err, expectedError)
		return
	}

	assert.Fail(t, "expected an error")
}

func newSolutionCollectTaskCompleteError(expectedError error) solutionCollectTaskFactory {
	return func() *task.Task[broker.SolutionCollectParams] {
		return &task.Task[broker.SolutionCollectParams]{
			OnPrepare: func(params *broker.SolutionCollectParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.SolutionCollectParams, state *task.State) error {
				return expectedError
			},
		}
	}
}

func newWorkflowListTaskComplete(seeded []broker.Solution) func() *task.Task[broker.WorkflowListParams] {
	return func() *task.Task[broker.WorkflowListParams] {
		return &task.Task[broker.WorkflowListParams]{
			OnPrepare: func(params *broker.WorkflowListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkflowListParams, state *task.State) error {
				state.Internal = seeded
				return nil
			},
		}
	}
}
