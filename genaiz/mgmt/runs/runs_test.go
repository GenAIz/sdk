package runs

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/workspace"
)

func TestUserRun_MarshalJSON(t *testing.T) {
	var testUserRun = &UserRun{
		Id:              int64(37),
		Created:         "created",
		Modified:        "modified",
		SolutionId:      int64(42),
		WorkflowId:      int64(69),
		WorkspaceId:     int64(73),
		WorkspaceFlowId: int64(67),
		Flags:           13,
		Progress:        120,
		Status:          "status",
	}
	var bytes, err = testUserRun.MarshalJSON()

	assert.NoError(t, err)
	actual := string(bytes)
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "id", testUserRun.Id))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":\"%s\"", "created", testUserRun.Created))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":\"%s\"", "modified", testUserRun.Modified))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "solutionId", testUserRun.SolutionId))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "workflowId", testUserRun.WorkflowId))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "workspaceId", testUserRun.WorkspaceId))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "workspaceFlowId", testUserRun.WorkspaceFlowId))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "flags", testUserRun.Flags))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":%d", "progress", testUserRun.Progress))
	assert.Contains(t, actual, fmt.Sprintf("\"%s\":\"%s\"", "status", testUserRun.Status))
}

func TestUserRun_MarshalSlice(t *testing.T) {
	var testUserRun = &UserRun{
		Id:             int64(37),
		Solution:       "solution",
		WorkflowHandle: "workflowHandle",
		Name:           "name",
		Created:        "created",
		Status:         "status",
	}
	var actual, err = testUserRun.MarshalSlice()

	assert.NoError(t, err)
	assert.Equal(t, cast.ToString(testUserRun.Id), actual[0])
	assert.Equal(t, testUserRun.Solution, actual[1])
	assert.Equal(t, testUserRun.WorkflowHandle, actual[2])
	assert.Equal(t, testUserRun.Name, actual[3])
	assert.Equal(t, testUserRun.Created, actual[4])
	assert.Equal(t, testUserRun.Status, actual[5])
}

func TestUserRunsProvider_Get(t *testing.T) {
	var testLogger, _ = test.NewNullLogger()
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var expectedSequence = 3
	var expectedRuns = []broker.WorkspaceRun{
		{
			Id: new(int64(72)),
		},
		{
			Id:      new(int64(37)),
			Created: new(time.Now().UnixMilli()),
		},
		{
			Id:       new(int64(42)),
			Modified: new(time.Now().Add(-1 * time.Hour).UnixMilli()),
			Solution: &broker.Solution{
				Oem:     expectedOem,
				Handle:  expectedHandle,
				Version: expectedVersion,
				Seq:     new(expectedSequence),
			},
		},
	}
	var expectedWorkspace = "expectedWorkspace"
	var expectedWorkspaceId = int64(37)
	var testProvider = &userRunsProvider{
		Plan: task.Plan{
			Logger: testLogger,
		},

		format:    timez.NewMidnightFormatter(),
		params:    &broker.RunListParams{},
		workspace: expectedWorkspace,

		runListTaskFactory:          newRunListTaskCompleteStub(expectedRuns),
		workspaceResolveTaskFactory: newWorkspaceResolveTaskStub(expectedWorkspaceId),
	}

	actual, err := testProvider.Get()
	assert.NoError(t, err)
	assert.Equal(t, 3, len(actual))
	assert.Equal(t, fmt.Sprintf("%s/%s:%s-rc-%d", expectedOem, expectedHandle, expectedVersion, expectedSequence),
		actual[2].Solution)
}

func TestUserRunsProvider_Get_Error(t *testing.T) {
	var testLogger, _ = test.NewNullLogger()
	var expectedError = task.NewError("expected")
	var testProvider = &userRunsProvider{
		Plan: task.Plan{
			Logger: testLogger,
		},

		params: &broker.RunListParams{
			FlowReadParams: &workspace.FlowReadParams{
				WorkspaceId: new(int64(37)),
			},
		},

		runListTaskFactory: newRunListTaskCompleteError(expectedError),
	}

	actual, err := testProvider.Get()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, expectedError)
}

func newRunListTaskCompleteError(err error) broker.RunListTaskFactory {
	return func() broker.RunListTask {
		return &task.Task[broker.RunListParams]{
			OnComplete: func(params *broker.RunListParams, state *task.State) error {
				return err
			},
			OnPrepare: func(params *broker.RunListParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newRunListTaskCompleteStub(runs []broker.WorkspaceRun) broker.RunListTaskFactory {
	return func() broker.RunListTask {
		return &task.Task[broker.RunListParams]{
			OnComplete: func(params *broker.RunListParams, state *task.State) error {
				state.Internal = runs
				return nil
			},
			OnPrepare: func(params *broker.RunListParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newWorkspaceResolveTaskStub(id int64) broker.WorkspaceResolveTaskFactory {
	return func() broker.WorkspaceResolveTask {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				params.SetWorkspaceId(id)
				return nil
			},
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
		}
	}
}
