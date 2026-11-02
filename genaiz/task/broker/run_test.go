package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/run"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubRunClient struct {
	client
	createRun           *WorkspaceRun
	createRunError      error
	createRunFlowId     int64
	createRunName       string
	createRunStart      bool
	listFlows           []WorkspaceFlow
	listFlowsError      error
	listFlowsFlags      int
	listFlowsMask       int
	listFlowsWsId       int64
	listRuns            []WorkspaceRun
	listRunsError       error
	listRunsFlags       int
	listRunsMask        int
	listRunsWorkspaceId int64
	startRun            *WorkspaceRun
	startRunError       error
	startRunId          int64
}

func (src *stubRunClient) CreateWorkspaceRun(flowId int64, name string, start bool) (*WorkspaceRun, error) {
	src.createRunFlowId = flowId
	src.createRunName = name
	src.createRunStart = start
	return src.createRun, src.createRunError
}

func (src *stubRunClient) ListWorkspaceFlows(wsId int64, mask, flags int) ([]WorkspaceFlow, error) {
	src.listFlowsWsId = wsId
	src.listFlowsMask = mask
	src.listFlowsFlags = flags
	return src.listFlows, src.listFlowsError
}

func (src *stubRunClient) ListWorkspaceRuns(wsId int64, mask, flags int) ([]WorkspaceRun, error) {
	src.listRunsWorkspaceId = wsId
	src.listRunsMask = mask
	src.listRunsFlags = flags
	return src.listRuns, src.listRunsError
}

func (src *stubRunClient) StartWorkspaceRun(runId int64) (*WorkspaceRun, error) {
	src.startRunId = runId
	return src.startRun, src.startRunError
}

func TestRunResolveParams_GetFlowId(t *testing.T) {
	var expectedFlowId = new(int64(73))
	var testParams = &RunResolveParams{}

	assert.Nil(t, testParams.GetFlowId())
	testParams.FlowReadParams = &workspace.FlowReadParams{
		FlowId: expectedFlowId,
	}
	assert.Equal(t, expectedFlowId, testParams.GetFlowId())
}

func TestRunResolveParams_SetFlowId(t *testing.T) {
	var expectedId = int64(37)
	var testParams = &RunResolveParams{}

	testParams.SetFlowId(&expectedId)
	assert.Equal(t, expectedId, *testParams.FlowId)
	testParams.SetFlowId(new(int64(42)))
	assert.NotEqual(t, expectedId, *testParams.FlowId)
}

func Test_handleRunResolveComplete(t *testing.T) {
	var expectedHandle = "handle"
	var expectedId = int64(73)
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
		WorkflowHandle: expectedHandle,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns: []WorkspaceRun{
				{
					Status: RunSuccess,
				},
				{
					Status: RunPaused,
					// No workflow
				},
			},
			listFlows: []WorkspaceFlow{
				{
					Id: int64(69),
				},
				{
					Id: expectedId,
					Solution: &Solution{
						Workflows: []Workflow{
							{
								Id:     new(int64(42)),
								Handle: expectedHandle,
							},
						},
					},
				},
			},
		}, nil
	}

	assert.NoError(t, handleRunResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.FlowId)
}

func Test_handleRunResolveComplete_FlowUnknown(t *testing.T) {
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns:  []WorkspaceRun{},
			listFlows: []WorkspaceFlow{},
		}, nil
	}

	assert.ErrorIs(t, handleRunResolveComplete(testParams, testState), run.Errors.RunFlowUnknown())
}

func Test_handleRunResolveComplete_ListFlowsError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns:       []WorkspaceRun{},
			listFlowsError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleRunResolveComplete(testParams, testState), expectedError)
}

func Test_handleRunResolveComplete_ListRunsError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRunsError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleRunResolveComplete(testParams, testState), expectedError)
}

func Test_handleRunResolveComplete_NoOutput(t *testing.T) {
	var testState = newTestState()

	assert.ErrorIs(t, handleRunResolveComplete(&RunResolveParams{}, testState),
		run.Errors.RunFlowUnknown())
}

func Test_handleRunResolveComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleRunResolveComplete(testParams, testState), expectedError)
}

func Test_handleRunResolveComplete_Startable(t *testing.T) {
	var expectedId = int64(87)
	var expectedFlowId = int64(89)
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
			FlowId:      &expectedFlowId,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns: []WorkspaceRun{
				{
					Id:              &expectedId,
					Status:          RunPaused,
					WorkspaceFlowId: expectedFlowId,
				},
			},
		}, nil
	}

	assert.NoError(t, handleRunResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.RunId)
	assert.Empty(t, testState.Output)
	assert.Equal(t, 1, len(testState.Reports))
}

func Test_handleRunResolveComplete_Startable_WorkflowHandle(t *testing.T) {
	var expectedId = int64(87)
	var expectedFlowId = int64(89)
	var expectedHandle = "handle"
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
		WorkflowHandle: expectedHandle,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns: []WorkspaceRun{
				{
					Id:              new(int64(57)),
					Status:          RunNone,
					WorkspaceFlowId: int64(42),
				},
				{
					Id:              new(int64(67)),
					Status:          RunNone,
					WorkspaceFlowId: int64(69),
					Workflow:        &Workflow{},
				},
				{
					Id:              &expectedId,
					Status:          RunPaused,
					WorkspaceFlowId: expectedFlowId,
					Workflow: &Workflow{
						Handle: expectedHandle,
					},
				},
			},
		}, nil
	}

	assert.NoError(t, handleRunResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.RunId)
	assert.Empty(t, testState.Output)
	assert.Equal(t, 1, len(testState.Reports))
}

func Test_handleRunResolveComplete_StartedError(t *testing.T) {
	var expectedId = int64(87)
	var expectedFlowId = int64(89)
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
			FlowId:      &expectedFlowId,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns: []WorkspaceRun{
				{
					Id:              &expectedId,
					Status:          RunQueued,
					WorkspaceFlowId: expectedFlowId,
				},
			},
		}, nil
	}

	actual := handleRunResolveComplete(testParams, testState)
	assert.True(t, run.Errors.IsRunDuplicated(actual))
}

func Test_handleRunResolveComplete_StartedError_WorkflowHandle(t *testing.T) {
	var expectedId = int64(87)
	var expectedFlowId = int64(89)
	var expectedHandle = "handle"
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
		WorkflowHandle: expectedHandle,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			listRuns: []WorkspaceRun{
				{
					Id:              new(int64(57)),
					Status:          RunScheduled,
					WorkspaceFlowId: int64(42),
				},
				{
					Id:              new(int64(67)),
					Status:          RunFailure,
					WorkspaceFlowId: int64(69),
					Workflow:        &Workflow{},
				},
				{
					Id:              &expectedId,
					Status:          RunQueued,
					WorkspaceFlowId: expectedFlowId,
					Workflow: &Workflow{
						Handle: expectedHandle,
					},
				},
			},
		}, nil
	}

	actual := handleRunResolveComplete(testParams, testState)
	assert.True(t, run.Errors.IsRunDuplicated(actual))
}

func Test_handleRunResolveContext(t *testing.T) {
	var expectedHandle = "handle"
	var testState = newTestState()
	var testParams = &RunResolveParams{
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(42)),
		},
		WorkflowHandle: expectedHandle,
	}

	assert.NoError(t, handleRunResolveContext(testParams, testState))
	assert.Equal(t, expectedHandle, testState.Output)
}

func Test_handleRunResolveContext_FlowId(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunResolveParams{
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(42)),
			FlowId:      new(int64(73)),
		},
	}

	assert.NoError(t, handleRunResolveContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.FlowId), testState.Output)
}

func Test_handleRunResolveContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleRunResolveContext(&RunResolveParams{}, testState))
}

func Test_handleRunResolveContext_RunFlowUnknown(t *testing.T) {
	var testParams = &RunResolveParams{
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleRunResolveContext(testParams, &task.State{}), run.Errors.RunFlowUnknown())
}

func Test_handleRunResolveContext_RunIdKnown(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunResolveParams{
		StartParams: &run.StartParams{
			RunId: new(int64(42)),
		},
	}

	assert.ErrorIs(t, handleRunResolveContext(testParams, testState), run.Errors.RunIdKnown())
}

func Test_handleRunResolveContext_RunWorkspaceUnknown(t *testing.T) {
	assert.ErrorIs(t, handleRunResolveContext(&RunResolveParams{}, &task.State{}), run.Errors.RunWorkspaceUnknown())
}

func Test_handleRunResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Logger: testLogger,
		Error:  run.Errors.RunIdKnown(),
		Output: "output",
	}
	var testParams = &RunResolveParams{
		StartParams: &run.StartParams{
			RunId: new(int64(42)),
		},
	}

	assert.NoError(t, handleRunResolveIncomplete(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleRunResolveIncomplete_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleRunResolveIncomplete(&RunResolveParams{}, testState), expectedError)
}

func Test_handleRunResolvePretend(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
			FlowId:      new(int64(67)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{}, nil
	}

	assert.NoError(t, handleRunResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.WorkspaceId))
	assert.Contains(t, output, ".runs")
	assert.NotContains(t, output, ".workspaceFlows")
}

func Test_handleRunResolvePretend_NoFlowId(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{}, nil
	}

	assert.NoError(t, handleRunResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.WorkspaceId))
	assert.Contains(t, output, ".runs")
	assert.Contains(t, output, ".workspaceFlows")
}

func Test_handleRunResolvePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = newTestState()
	var testParams = &RunResolveParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleRunResolvePretend(testParams, testState), expectedError)
}

func Test_handleRunStartComplete(t *testing.T) {
	var expectedRun = &WorkspaceRun{
		Id:   new(int64(37)),
		Name: "myName",
	}
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		StartParams: &run.StartParams{
			RunId: expectedRun.Id,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			startRun: expectedRun,
		}, nil
	}

	assert.NoError(t, handleRunStartComplete(testParams, testState))
	assert.Empty(t, testState.Output)
	assert.Equal(t, 1, len(testState.Reports))
}

func Test_handleRunStartComplete_NoOutput(t *testing.T) {
	var testState = newTestState()

	assert.ErrorIs(t, handleRunStartComplete(&RunStartParams{}, testState),
		run.Errors.RunUnprepared())
}

func Test_handleRunStartComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleRunStartComplete(testParams, testState), expectedError)
}

func Test_handleRunStartComplete_StartError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Output: "output",
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		StartParams: &run.StartParams{
			RunId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			startRunError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleRunStartComplete(testParams, testState), expectedError)
}

func Test_handleRunStartContext(t *testing.T) {
	var expectedId = int64(37)
	var expectedName = "myName"
	var testState = &task.State{
		Logger: testLogger,
		Internal: WorkspaceFlow{
			WorkflowId: expectedId,
			Solution: &Solution{
				Workflows: []Workflow{
					{
						Id:     &expectedId,
						Handle: "myHandle",
					},
				},
			},
		},
	}
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
		StartParams: &run.StartParams{
			RunId: new(int64(42)),
		},
		Name: expectedName,
	}

	assert.NoError(t, handleRunStartContext(testParams, testState))
	assert.Equal(t, expectedName, testState.Output)
}

func Test_handleRunStartContext_NoName(t *testing.T) {
	var expectedId = int64(37)
	var expectedHandle = "myHandle"
	var testState = &task.State{
		Logger: testLogger,
		Internal: WorkspaceFlow{
			WorkflowId: expectedId,
			Solution: &Solution{
				Workflows: []Workflow{
					{
						Id:     &expectedId,
						Handle: expectedHandle,
					},
				},
			},
		},
	}
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
		StartParams: &run.StartParams{
			RunId: new(int64(42)),
		},
	}

	assert.NoError(t, handleRunStartContext(testParams, testState))
	assert.True(t, strings.HasPrefix(testState.Output, expectedHandle))
}

func Test_handleRunStartContext_NoNameNoWorkflow(t *testing.T) {
	var testState = &task.State{
		Logger: testLogger,
		Internal: WorkspaceFlow{
			Solution: &Solution{},
		},
	}
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleRunStartContext(testParams, testState), run.Errors.RunWorkflowUnknown())
}

func Test_handleRunStartContext_NoNameNoSolution(t *testing.T) {
	var testState = &task.State{
		Logger:   testLogger,
		Internal: WorkspaceFlow{},
	}
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleRunStartContext(testParams, testState), run.Errors.RunSolutionUnknown())
}

func Test_handleRunStartContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleRunStartContext(&RunStartParams{}, testState))
}

func Test_handleRunStartContext_RunFlowUnknown(t *testing.T) {
	assert.ErrorIs(t, handleRunStartContext(&RunStartParams{}, &task.State{}), run.Errors.RunFlowUnknown())
}

func Test_handleRunStartContext_RunIdRequired(t *testing.T) {
	var testState = &task.State{
		Internal: WorkspaceFlow{},
	}
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
		Name: "expectedName",
	}

	assert.ErrorIs(t, handleRunStartContext(testParams, testState), run.Errors.RunIdRequired())
}

func Test_handleRunStartContext_RunUnprepared(t *testing.T) {
	var testParams = &RunStartParams{
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleRunStartContext(testParams, &task.State{}), run.Errors.RunUnprepared())
}

func Test_handleRunStartIncomplete(t *testing.T) {
	var expectedRun = &WorkspaceRun{
		Id:   new(int64(42)),
		Name: "expectedName",
	}
	var testState = &task.State{
		Logger: testLogger,
		Error:  run.Errors.RunIdRequired(),
		Output: expectedRun.Name,
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}
	var testClient = &stubRunClient{createRun: expectedRun}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleRunStartIncomplete(testParams, testState))
	assert.Equal(t, expectedRun.Name, testClient.createRunName)
	assert.True(t, testClient.createRunStart)
	assert.Equal(t, *testParams.FlowId, testClient.createRunFlowId)
	assert.Empty(t, testState.Output)
}

func Test_handleRunStartIncomplete_CreateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Error:  run.Errors.RunIdRequired(),
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{
			createRunError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleRunStartIncomplete(testParams, testState), expectedError)
}

func Test_handleRunStartIncomplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: testLogger,
		Error:  run.Errors.RunIdRequired(),
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleRunStartIncomplete(testParams, testState), expectedError)
}

func Test_handleRunStartIncomplete_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleRunStartIncomplete(&RunStartParams{}, testState), expectedError)
}

func Test_handleRunStartPretend(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		StartParams: &run.StartParams{
			RunId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{}, nil
	}

	assert.NoError(t, handleRunStartPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, fmt.Sprintf("-f id=%d", *testParams.RunId))
	assert.Contains(t, output, "-r .run")
}

func Test_handleRunStartPretend_Create(t *testing.T) {
	var testState = &task.State{
		Logger: testLogger,
		Output: "outputName",
	}
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
		FlowReadParams: &workspace.FlowReadParams{
			FlowId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{}, nil
	}

	assert.NoError(t, handleRunStartPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, fmt.Sprintf("-f workspaceFlowId=%d", *testParams.FlowId))
	assert.Contains(t, output, fmt.Sprintf("-f name=%s", testState.Output))
	assert.Contains(t, output, "-f start=true")
	assert.Contains(t, output, "-r .run")
}

func Test_handleRunStartPretend_RunUnpreparedError(t *testing.T) {
	var testState = newTestState()
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubRunClient{}, nil
	}

	assert.ErrorIs(t, handleRunStartPretend(testParams, testState), run.Errors.RunUnprepared())
}

func Test_handleRunStartPretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = newTestState()
	var testParams = &RunStartParams{
		Broker: makeTestBroker(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleRunStartPretend(testParams, testState), expectedError)
}

func Test_handleRunStartPretend_StateError(t *testing.T) {
	var expectedErrors = errors.New("expected")
	var testState = &task.State{
		Error: expectedErrors,
	}

	assert.ErrorIs(t, handleRunStartPretend(&RunStartParams{}, testState), expectedErrors)
}

func TestRunTasks_ResolveRunTask(t *testing.T) {
	var testTask = RunTasks.ResolveRunTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestRunTasks_StartRunTask(t *testing.T) {
	var testTask = RunTasks.StartRunTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}
