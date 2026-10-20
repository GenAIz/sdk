package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubWorkspaceClient struct {
	client
	createFlowError           error
	createFlowExpected        *WorkspaceFlow
	createFlowWorkspaceId     int64
	createFlowName            string
	createFlowDescription     string
	createWorkspaceError      error
	createWorkspaceParam      *Workspace
	createWorkspace           *Workspace
	findSolutionError         error
	findSolutionExpected      *Solution
	findSolutionOem           string
	findSolutionHandle        string
	findSolutionVersion       string
	getFunction               *Function
	getFunctionError          error
	getFunctionId             int64
	getNode                   *WorkspaceNode
	getNodeError              error
	getNodeId                 int64
	inspectNode               *WorkspaceNode
	inspectNodeError          error
	inspectNodeId             int64
	listFlows                 []WorkspaceFlow
	listFlowsError            error
	listFlowsFlags            int
	listFlowsMask             int
	listFlowsWsId             int64
	listNodes                 []WorkspaceNode
	listNodesError            error
	listNodesFlowId           int64
	listWorkspaceError        error
	listWorkspaceMask         int
	listWorkspaceFlags        int
	listWorkspaces            []Workspace
	liseWorkspacesUserId      int
	updateNode                *WorkspaceNode
	updateNodeError           error
	updateNodeNode            *WorkspaceNode
	updateNodeInput           *WorkspaceNode
	updateNodeInputError      error
	updateNodeInputNodeId     int64
	updateNodeInputDataSetId  int64
	updateNodeInputPortHandle string
}

func (swc *stubWorkspaceClient) CreateWorkspace(workspace *Workspace) (*Workspace, error) {
	swc.createWorkspaceParam = workspace
	return swc.createWorkspace, swc.createWorkspaceError
}

func (swc *stubWorkspaceClient) CreateWorkspaceFlow(wsId, wfId int64, name, description string) (*WorkspaceFlow, error) {
	swc.createFlowWorkspaceId = wsId
	swc.createFlowWorkspaceId = wfId
	swc.createFlowName = name
	swc.createFlowDescription = description
	return swc.createFlowExpected, swc.createFlowError
}

func (swc *stubWorkspaceClient) FindSolution(oem, handle, vers string) (*Solution, error) {
	swc.findSolutionOem = oem
	swc.findSolutionHandle = handle
	swc.findSolutionVersion = vers
	return swc.findSolutionExpected, swc.findSolutionError
}

func (swc *stubWorkspaceClient) GetFunction(functionId int64) (*Function, error) {
	swc.getFunctionId = functionId
	return swc.getFunction, swc.getFunctionError
}

func (swc *stubWorkspaceClient) GetNode(nodeId int64) (*WorkspaceNode, error) {
	swc.getNodeId = nodeId
	return swc.getNode, swc.getNodeError
}

func (swc *stubWorkspaceClient) GetUserId() int {
	return swc.liseWorkspacesUserId
}

func (swc *stubWorkspaceClient) InspectNode(id int64) (*WorkspaceNode, error) {
	swc.inspectNodeId = id
	return swc.inspectNode, swc.inspectNodeError
}

func (swc *stubWorkspaceClient) ListWorkspaces(mask, flags int) ([]Workspace, error) {
	swc.listWorkspaceMask = mask
	swc.listWorkspaceFlags = flags
	return swc.listWorkspaces, swc.listWorkspaceError
}

func (swc *stubWorkspaceClient) ListWorkspaceFlows(wsId int64, mask, flags int) ([]WorkspaceFlow, error) {
	swc.listFlowsWsId = wsId
	swc.listFlowsMask = mask
	swc.listFlowsFlags = flags
	return swc.listFlows, swc.listFlowsError
}

func (swc *stubWorkspaceClient) ListWorkspaceNodes(flowId int64) ([]WorkspaceNode, error) {
	swc.listNodesFlowId = flowId
	return swc.listNodes, swc.listNodesError
}

func (swc *stubWorkspaceClient) UpdateNode(node *WorkspaceNode) (*WorkspaceNode, error) {
	swc.updateNodeNode = node
	return swc.updateNode, swc.updateNodeError
}

func (swc *stubWorkspaceClient) UpdateNodeInput(nodeId, dataSetId int64, portHandle string) (*WorkspaceNode, error) {
	swc.updateNodeInputNodeId = nodeId
	swc.updateNodeInputDataSetId = dataSetId
	swc.updateNodeInputPortHandle = portHandle
	return swc.updateNodeInput, swc.updateNodeInputError
}

func TestWorkspaceTasks_AddSourceNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.AddSourceNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_AddStoreNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.AddStoreNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_CreateFlowTask(t *testing.T) {
	var testTask = WorkspaceTasks.CreateFlowTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_CreateTask(t *testing.T) {
	var testTask = WorkspaceTasks.CreateTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
	assert.Nil(t, testTask.OnIncomplete)
}

func TestWorkspaceTasks_ListFLowTask(t *testing.T) {
	var testTask = WorkspaceTasks.ListFlowTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ListNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.ListNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ListSetNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.ListSetNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ListTask(t *testing.T) {
	var testTask = WorkspaceTasks.ListTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
	assert.Nil(t, testTask.OnIncomplete)
}

func TestWorkspaceTasks_ResolveSetNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.ResolveSetNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_RemoveSourceNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.RemoveSourceNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_RemoveStoreNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.RemoveStoreNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.Nil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ResolveFlowTask(t *testing.T) {
	var testTask = WorkspaceTasks.ResolveFlowTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ResolveNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.ResolveNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_ResolveTask(t *testing.T) {
	var testTask = WorkspaceTasks.ResolveTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestWorkspaceTasks_UpdateSetNodeTask(t *testing.T) {
	var testTask = WorkspaceTasks.UpdateSetNodeTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleWorkspaceCreateContext(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{
			Name:       "expectedName",
			RcEnabled:  true,
			Visibility: VisibilityOrg,
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceCreateContext(testParams, testState))
	assert.Equal(t, 1, len(hook.Entries))
	assert.Contains(t, hook.Entries[0].Message, testParams.Workspace.Name)
	assert.Contains(t, hook.Entries[0].Message, "development")
}

func Test_handleWorkspaceCreateContext_CheckedOutput(t *testing.T) {
	assert.NoError(t, handleWorkspaceCreateContext(&WorkspaceCreateParams{}, &task.State{Output: "checked"}))
}

func Test_handleWorkspaceCreateContext_EmptyName(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{
			RcEnabled:  true,
			Visibility: VisibilityPrivate,
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceCreateContext(testParams, testState))
	assert.Equal(t, 2, len(hook.Entries))
	assert.Equal(t, logrus.WarnLevel, hook.Entries[0].Level)
}

func Test_handleWorkspaceCreateContext_EmptyVisibility(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{},
	}

	assert.ErrorIs(t, handleWorkspaceCreateContext(testParams, testState),
		workspace.Errors.WorkspaceVisibilityRequired())
}

func Test_handleWorkspaceCreateContext_EmptyWorkspace(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{}

	assert.ErrorIs(t, handleWorkspaceCreateContext(testParams, testState),
		workspace.Errors.WorkspaceDefInvalid())
}

func Test_handleWorkspaceCreateContext_RcDisabled(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{
			Name:       "expectedName",
			RcEnabled:  false,
			Visibility: VisibilityPrivate,
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceCreateContext(testParams, testState))
	assert.Equal(t, 1, len(hook.Entries))
	assert.Contains(t, hook.Entries[0].Message, testParams.Workspace.Name)
	assert.Contains(t, hook.Entries[0].Message, "production")
}

func Test_handleWorkspaceCreateComplete(t *testing.T) {
	var expectedWorkspace = &Workspace{Id: int64(37)}
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Workspace: &Workspace{},
	}
	var testClient = &stubWorkspaceClient{
		createWorkspace: expectedWorkspace,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceCreateComplete(testParams, testState))
	assert.NotEmpty(t, testState.Reports)
	assert.Equal(t, expectedWorkspace.Id, cast.ToInt64(testState.Output))
	assert.Equal(t, expectedWorkspace, testState.Internal)
}

func Test_handleWorkspaceCreateComplete_CreateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Workspace: &Workspace{},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			createWorkspaceError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceCreateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceCreateComplete_EmptyWorkspace(t *testing.T) {
	var testParams = &WorkspaceCreateParams{}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceCreateComplete(testParams, testState),
		workspace.Errors.WorkspaceDefInvalid())
}

func Test_handleWorkspaceCreateComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Workspace: &Workspace{},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceCreateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceCreatePretend(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Workspace: &Workspace{
			Name:        "expectedName",
			Description: "expectedDesc",
			Visibility:  VisibilityPrivate,
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceCreatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.Workspace.Name)
	assert.Contains(t, output, testParams.Workspace.Visibility)
	assert.Contains(t, output, testParams.Workspace.Description)
	assert.Contains(t, output, cast.ToString(testParams.Workspace.RcEnabled))
}

func Test_handleWorkspaceCreatePretend_EmptyWorkspace(t *testing.T) {
	var testParams = &WorkspaceCreateParams{}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceCreatePretend(testParams, testState),
		workspace.Errors.WorkspaceDefInvalid())
}

func Test_handleWorkspaceCreatePretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Workspace: &Workspace{},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceCreatePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowCreateContext(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkspaceId: new(int64(37)),
		WorkflowId:  new(int64(43)),
	}

	assert.NoError(t, handleWorkspaceFlowCreateContext(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowCreateContext_NoWorkflowId(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkspaceId: new(int64(37)),
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateContext(testParams, &task.State{}),
		workspace.Errors.WorkflowIdRequired())
}

func Test_handleWorkspaceFlowCreateContext_NoWorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkflowId: new(int64(43)),
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceIdRequired())
}

func Test_handleWorkspaceFlowCreateContext_OutputCheck(t *testing.T) {
	var testState = &task.State{Output: "check"}

	assert.NoError(t, handleWorkspaceFlowCreateContext(&WorkspaceFlowCreateParams{}, testState))
}

func Test_handleWorkspaceFlowCreateComplete(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testClient = &stubWorkspaceClient{
		createFlowExpected: &WorkspaceFlow{
			WorkspaceId: int64(37),
			WorkflowId:  int64(42),
			Name:        "expectedName",
			Description: "expectedDesc",
		},
	}
	var testParams = &WorkspaceFlowCreateParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: &testClient.createFlowExpected.WorkspaceId,
		WorkflowId:  &testClient.createFlowExpected.WorkflowId,
		Name:        testClient.createFlowExpected.Name,
	}

	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceFlowCreateComplete(testParams, testState))
	assert.NotEmpty(t, testState.Reports)
	assert.Empty(t, testState.Output)

	if actual, ok := testState.Internal.(*WorkspaceFlow); ok {
		assert.Equal(t, testClient.createFlowExpected, actual)
	} else {
		assert.Fail(t, "expected internal workspace flow")
	}
}

func Test_handleWorkspaceFlowCreateComplete_CreateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowCreateParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
		WorkflowId:  new(int64(42)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			createFlowError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowCreateComplete_InvalidParams(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkspaceId: new(int64(37)),
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateComplete(testParams, &task.State{}),
		workspace.Errors.WorkspaceFlowDefInvalid())
}

func Test_handleWorkspaceFlowCreateComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowCreateParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
		WorkflowId:  new(int64(42)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowCreatePretend(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowCreateParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
		WorkflowId:  new(int64(69)),
		Name:        "expectedName",
		Description: "expectedDesc",
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceFlowCreatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.WorkspaceId))
	assert.Contains(t, output, cast.ToString(*testParams.WorkflowId))
	assert.Contains(t, output, testParams.Name)
	assert.Contains(t, output, testParams.Description)
}

func Test_handleWorkspaceFlowCreatePretend_InvalidParams(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkspaceId: new(int64(37)),
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreatePretend(testParams, &task.State{}),
		workspace.Errors.WorkspaceFlowDefInvalid())
}

func Test_handleWorkspaceFlowCreatePretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowCreateParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
		WorkflowId:  new(int64(42)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreatePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowListComplete(t *testing.T) {
	var expectedFlow = &WorkspaceFlow{
		Id:          int64(37),
		Name:        "expectedName",
		Description: "expectedDesc",
		WorkflowId:  int64(38),
		SolutionId:  int64(42),
		Created:     1,
		Flags:       new(WorkspaceFlowFlags.Active),
	}
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				Broker: &Broker{
					AuthFile: "file",
					HostAddr: "hostAddr",
				},
				WorkspaceId: new(int64(37)),
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: []WorkspaceFlow{*expectedFlow},
		}, nil
	}

	assert.NoError(t, handleWorkspaceFlowListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]WorkspaceFlow); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, expectedFlow.Id, actual[0].Id)
		assert.Equal(t, expectedFlow.Name, actual[0].Name)
		assert.Equal(t, expectedFlow.Description, actual[0].Description)
		assert.Equal(t, expectedFlow.WorkflowId, actual[0].WorkflowId)
		assert.Equal(t, expectedFlow.SolutionId, actual[0].SolutionId)
		assert.Equal(t, expectedFlow.Created, actual[0].Created)
		assert.Equal(t, expectedFlow.Flags, actual[0].Flags)
		return
	}

	assert.Fail(t, "expected a slice of workspace flows")
}

func Test_handleWorkspaceFlowListComplete_InvalidParams(t *testing.T) {
	var testParams = &WorkspaceFlowListParams{}

	assert.ErrorIs(t, handleWorkspaceFlowListComplete(testParams, &task.State{}),
		workspace.Errors.WorkspaceIdRequired())
}

func Test_handleWorkspaceFlowListComplete_ListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				Broker: &Broker{
					AuthFile: "file",
					HostAddr: "hostAddr",
				},
				WorkspaceId: new(int64(37)),
			},
		},
		ReadyOnly: true,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlowsError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceFlowListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowListComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				Broker: &Broker{
					AuthFile: "file",
					HostAddr: "hostAddr",
				},
				WorkspaceId: new(int64(37)),
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowListContext(t *testing.T) {
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				WorkspaceId: new(int64(37)),
			},
		},
	}

	assert.NoError(t, handleWorkspaceFlowListContext(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowListContext_CheckedOutput(t *testing.T) {
	var testState = &task.State{Output: "checked"}

	assert.NoError(t, handleWorkspaceFlowListContext(&WorkspaceFlowListParams{}, testState))
}

func Test_handleWorkspaceFlowListContext_InvalidWorkspace(t *testing.T) {
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{},
	}

	assert.ErrorIs(t, handleWorkspaceFlowListContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceIdRequired())
}

func Test_handleWorkspaceFlowListPretend(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				Broker: &Broker{
					AuthFile: "file",
					HostAddr: "hostAddr",
				},
				WorkspaceId: new(int64(37)),
			},
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceFlowListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	mask, flags := testParams.GetMaskFlags()
	assert.Contains(t, output, cast.ToString(*testParams.WorkspaceId))
	assert.Contains(t, output, fmt.Sprintf("mask=%d", mask))
	assert.Contains(t, output, fmt.Sprintf("flags=%d", flags))
}

func Test_handleWorkspaceFlowListPretend_InvalidParams(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceFlowListPretend(&WorkspaceFlowListParams{}, &task.State{}),
		workspace.Errors.WorkspaceIdRequired())
}

func Test_handleWorkspaceFlowListPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
				Broker: &Broker{
					AuthFile: "file",
					HostAddr: "hostAddr",
				},
				WorkspaceId: new(int64(37)),
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowListPretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolveComplete(t *testing.T) {
	var expectedId = int64(37)
	var testState = &task.State{Logger: logrus.New()}
	var testClient = &stubWorkspaceClient{
		listWorkspaces: []Workspace{
			{
				Id:   expectedId,
				Name: "expected",
			},
		},
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		WorkspaceName: testClient.listWorkspaces[0].Name,
		RcEnabled:     true,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.WorkspaceId)
}

func Test_handleWorkspaceFlowResolveComplete_ListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaceError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolveComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolveComplete_WorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceResolveComplete(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowResolveContext(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceName: "name",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkspaceResolveContext(testParams, testState))
}

func Test_handleWorkspaceFlowResolveContext_NoName(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkspaceResolveContext(testParams, testState),
		workspace.Errors.WorkspaceNameRequired())
}

func Test_handleWorkspaceFlowResolveContext_OutputCheck(t *testing.T) {
	var testState = &task.State{Output: "check"}

	assert.NoError(t, handleWorkspaceResolveContext(&WorkspaceFlowResolveParams{}, testState))
}

func Test_handleWorkspaceFlowResolveContext_WorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(937)),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkspaceResolveContext(testParams, testState),
		workspace.Errors.WorkspaceIdKnown())
}

func Test_handleWorkspaceFlowResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleWorkspaceResolveIncomplete(&WorkspaceFlowResolveParams{}, testState), testState.Error)
	assert.False(t, testState.Completed)
}

func Test_handleWorkspaceFlowResolveIncomplete_KnownId(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceIdKnown(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceResolveIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceFlowResolvePretend(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		RcEnabled: true,
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, fmt.Sprintf("mask=%d", WorkspaceFlags.RcEnabled|WorkspaceFlags.Active))
	assert.Contains(t, output, fmt.Sprintf("flags=%d", WorkspaceFlags.RcEnabled|WorkspaceFlags.Active))
}

func Test_handleWorkspaceFlowResolvePretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceResolvePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolvePretend_WorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(73)),
		},
	}

	assert.NoError(t, handleWorkspaceResolvePretend(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowSolutionComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testClient = &stubWorkspaceClient{
		findSolutionExpected: &Solution{
			Id: new(int64(37)),
			Workflows: []Workflow{
				{
					Id:     new(int64(37)),
					Handle: "expected",
				},
			},
		},
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		WorkflowHandle: "expected",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceFlowSolutionComplete(testParams, testState))
	assert.Equal(t, testClient.findSolutionExpected.Workflows[0].Id, testParams.WorkflowId)
}

func Test_handleWorkspaceFlowSolutionComplete_FindError(t *testing.T) {
	var testState = &task.State{}
	var testClient = &stubWorkspaceClient{
		findSolutionError: errors.New("expected"),
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionComplete(testParams, testState), testClient.findSolutionError)
}

func Test_handleWorkspaceFlowSolutionComplete_FindEmpty(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testClient = &stubWorkspaceClient{
		findSolutionExpected: &Solution{
			Id:        new(int64(37)),
			Workflows: make([]Workflow, 0),
		},
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionComplete(testParams, testState), ErrorWorkflowNotFound)
}

func Test_handleWorkspaceFlowSolutionComplete_NoMatches(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testClient = &stubWorkspaceClient{
		findSolutionExpected: &Solution{
			Id: new(int64(37)),
			Workflows: []Workflow{
				{
					Handle: "someHandle",
				},
			},
		},
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		WorkflowHandle: "anotherHandle",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionComplete(testParams, testState), ErrorWorkflowNotFound)
}

func Test_handleWorkspaceFlowSolutionComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowSolutionComplete_WorkflowId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkflowId: new(int64(73)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowSolutionComplete(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowSolutionContext(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionHandle:  "handle",
		SolutionOem:     "oem",
		SolutionVersion: "version",
	}

	assert.NoError(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowSolutionContext_OutputCheck(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{}
	var testState = &task.State{
		Output: "check",
	}

	assert.NoError(t, handleWorkspaceFlowSolutionContext(testParams, testState))
}

func Test_handleWorkspaceFlowSolutionContext_NoSolutionHandle(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionOem:     "oem",
		SolutionVersion: "version",
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}),
		workspace.Errors.WorkflowHandleRequired())
}

func Test_handleWorkspaceFlowSolutionContext_NoSolutionOem(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionHandle:  "handle",
		SolutionVersion: "version",
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}), workspace.Errors.WorkflowOemRequired())
}

func Test_handleWorkspaceFlowSolutionContext_NoSolutionVersion(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionHandle: "handle",
		SolutionOem:    "oem",
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}), workspace.Errors.WorkflowVersionRequired())
}

func Test_handleWorkspaceFlowSolutionContext_WorkflowId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkflowId: new(int64(37)),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, testState), workspace.Errors.WorkflowIdKnown())
}

func Test_handleWorkspaceFlowSolutionIncomplete(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionIncomplete(&WorkspaceFlowResolveParams{}, testState), testState.Error)
	assert.False(t, testState.Completed)
}

func Test_handleWorkspaceFlowSolutionIncomplete_KnownId(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkflowIdKnown(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkflowId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowSolutionIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceFlowSolutionPretend(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		SolutionHandle:  "expectedHandle",
		SolutionOem:     "expectedOem",
		SolutionVersion: "expectedVers",
		RcEnabled:       true,
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceFlowSolutionPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.SolutionOem)
	assert.Contains(t, output, testParams.SolutionHandle)
	assert.Contains(t, output, testParams.SolutionVersion)
}

func Test_handleWorkspaceFlowSolutionPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			Broker: &Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionPretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowSolutionPretend_WorkflowId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkflowId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowSolutionPretend(testParams, &task.State{}))
}

func Test_handleWorkspaceListContext(t *testing.T) {
	assert.NoError(t, handleWorkspaceListContext(&WorkspaceListParams{}, &task.State{}))

}

func Test_handleWorkspaceListContext_CheckedOutput(t *testing.T) {
	assert.NoError(t, handleWorkspaceListContext(&WorkspaceListParams{}, &task.State{Output: "checked"}))
}

func Test_handleWorkspaceListContext_InvalidNco(t *testing.T) {
	var invalidTime = time.Now().AddDate(1, 0, 0)
	var testParams = &WorkspaceListParams{
		FromDate: &invalidTime,
	}

	assert.ErrorIs(t, handleWorkspaceListContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceInvalidNco())
}

func Test_handleWorkspaceListContext_InvalidOwner(t *testing.T) {
	var expectedHost = "host"
	var testAuth = &AuthData{
		Active: 0,
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedHost,
				AuthSession: &AuthSession{
					UserId: -1,
					Expiry: -1,
				},
			},
		},
	}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: expectedHost,
		},
		OwnerOnly: true,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.ErrorIs(t, handleWorkspaceListContext(testParams, &task.State{}),
					workspace.Errors.WorkspaceInvalidOwner())
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleWorkspaceListContext_NoLogin(t *testing.T) {
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
		OwnerOnly: true,
	}

	assert.ErrorIs(t, handleWorkspaceListContext(testParams, &task.State{}), ErrorNoLogin)
}

func Test_handleWorkspaceListContext_OwnerOnly(t *testing.T) {
	var testAuth = &AuthData{
		Active: 0,
		Accounts: []*AuthAccount{
			{
				AuthSession: &AuthSession{
					UserId: 100,
					Expiry: -1,
				},
			},
		},
	}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
		OwnerOnly: true,
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleWorkspaceListContext(testParams, &task.State{}))
				return
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleWorkspaceListComplete(t *testing.T) {
	var expectedId = int64(37)
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		RcEnabled: true,
	}
	var testClient = &stubWorkspaceClient{
		listWorkspaces: []Workspace{
			{
				Id: expectedId,
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]Workspace); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, expectedId, actual[0].Id)
		assert.Equal(t, 3, testClient.listWorkspaceMask)
		assert.Equal(t, 3, testClient.listWorkspaceFlags)
		return
	}

	assert.Fail(t, "did not receive the list of workspaces")
}

func Test_handleWorkspaceListComplete_DateFilter(t *testing.T) {
	var expectedId = int64(37)
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		FromDate: new(time.UnixMilli(1)),
	}
	var testWorkspaces = []Workspace{
		{
			Id:      expectedId,
			Created: 2,
		},
		{
			Id:      42,
			Created: 1,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: testWorkspaces,
		}, nil
	}

	assert.NoError(t, handleWorkspaceListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]Workspace); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, expectedId, actual[0].Id)
		return
	}

	assert.Fail(t, "did not receive the list of workspaces")
}

func Test_handleWorkspaceListComplete_ListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaceError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceListComplete_NoLogin(t *testing.T) {
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}

	assert.ErrorIs(t, handleWorkspaceListComplete(testParams, &task.State{}), ErrorNoLogin)
}

func Test_handleWorkspaceListComplete_OwnerFilter(t *testing.T) {
	var expectedId = int64(37)
	var expectedHost = "hostAddr"
	var expectedOwnerId = 42
	var testAuth = &AuthData{
		Active: 0,
		Accounts: []*AuthAccount{
			{
				HostAddr: expectedHost,
				AuthSession: &AuthSession{
					UserId: expectedOwnerId,
					Expiry: -1,
					Token:  "token",
				},
			},
		},
	}
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: expectedHost,
		},
		OwnerOnly: true,
	}
	var testWorkspaces = []Workspace{
		{
			Id:          expectedId,
			OwnerUserId: expectedOwnerId,
		},
		{
			Id:          42,
			OwnerUserId: 1337,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces:       testWorkspaces,
			liseWorkspacesUserId: expectedOwnerId,
		}, nil
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testParams.AuthFile); err == nil {
		var bytes []byte

		defer filez.CloseSilently(fd)

		if bytes, err = yaml.Marshal(testAuth); err == nil {
			if _, err = fd.Write(bytes); err == nil {
				assert.NoError(t, handleWorkspaceListComplete(testParams, testState))

				if actual, ok := testState.Internal.([]Workspace); ok {
					assert.Equal(t, 1, len(actual))
					assert.Equal(t, expectedId, actual[0].Id)
					return
				}

				assert.Fail(t, "did not receive the list of workspaces")
			}
		}
	}

	assert.NoError(t, err)
}

func Test_handleWorkspaceListPretend(t *testing.T) {
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
	}
	var expectedMask, expectedFlags = testParams.GetMaskFlags()
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testClient = &stubWorkspaceClient{}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, fmt.Sprintf("%s?mask=%d&flags=%d", testClient.ListWorkspacesUrl(), expectedMask, expectedFlags))
}

func Test_handleWorkspaceListPretend_FromDate(t *testing.T) {
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
			HostAddr: "hostAddr",
		},
		FromDate: new(time.UnixMilli(1)),
	}
	var expectedMask, expectedFlags = testParams.GetMaskFlags()
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testClient = &stubWorkspaceClient{}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, fmt.Sprintf("%s?mask=%d&flags=%d", testClient.ListWorkspacesUrl(), expectedMask, expectedFlags))
	assert.Equal(t, 1, len(hook.Entries))
	assert.Contains(t, hook.Entries[0].Message, "after date")
}

func Test_handleWorkspaceListPretend_NoLogin(t *testing.T) {
	var testParams = &WorkspaceListParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}

	assert.ErrorIs(t, handleWorkspaceListPretend(testParams, &task.State{}), ErrorNoLogin)
}

func Test_handleWorkspaceNodeListComplete(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:              37,
		WorkspaceId:     1337,
		WorkspaceFlowId: 42,
		WorkflowNodeId:  69,
		SmartFunctionId: 31337,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:     new(expectedNode.WorkspaceId),
		WorkspaceFlowId: new(expectedNode.WorkspaceFlowId),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listNodes: []WorkspaceNode{*expectedNode},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]WorkspaceNode); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, actual[0].Id, expectedNode.Id)
		assert.Equal(t, actual[0].WorkspaceId, expectedNode.WorkspaceId)
		assert.Equal(t, actual[0].WorkspaceFlowId, expectedNode.WorkspaceFlowId)
		assert.Equal(t, actual[0].WorkflowNodeId, expectedNode.WorkflowNodeId)
		assert.Equal(t, actual[0].SmartFunctionId, expectedNode.SmartFunctionId)
		return
	}

	assert.Fail(t, "expected a slice of nodes")
}

func Test_handleWorkspaceNodeListComplete_InvalidParams(t *testing.T) {
	var testState = &task.State{}
	var testParams = &WorkspaceNodeListParams{}

	assert.ErrorIs(t, handleWorkspaceNodeListComplete(testParams, testState),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeListComplete_ListError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:     new(int64(37)),
		WorkspaceFlowId: new(int64(42)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listNodesError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListComplete_NoFunction(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:              37,
		WorkspaceId:     1337,
		WorkspaceFlowId: 42,
		WorkflowNodeId:  69,
		SmartFunctionId: 31337,
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:     new(expectedNode.WorkspaceId),
		WorkspaceFlowId: new(expectedNode.WorkspaceFlowId),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getFunctionError: errors.New("notFound"),
			listNodes:        []WorkspaceNode{*expectedNode},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]WorkspaceNode); ok {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, actual[0].Id, expectedNode.Id)
		assert.Equal(t, actual[0].WorkspaceId, expectedNode.WorkspaceId)
		assert.Equal(t, actual[0].WorkspaceFlowId, expectedNode.WorkspaceFlowId)
		assert.Equal(t, actual[0].WorkflowNodeId, expectedNode.WorkflowNodeId)
		assert.Equal(t, actual[0].SmartFunctionId, expectedNode.SmartFunctionId)
		assert.Empty(t, actual[0].SmartFunction)
		return
	}

	assert.Fail(t, "expected a slice of nodes")
}

func Test_handleWorkspaceNodeListComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceFlowId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListContext(t *testing.T) {
	var testParams = &WorkspaceNodeListParams{
		WorkspaceFlowId: new(int64(37)),
		WorkflowHandle:  "expectedHandle",
	}

	assert.NoError(t, handleWorkspaceNodeListContext(testParams, &task.State{}))
}

func Test_handleWorkspaceNodeListContext_CheckOutput(t *testing.T) {
	var testState = &task.State{Output: "checked"}

	assert.NoError(t, handleWorkspaceNodeListContext(&WorkspaceNodeListParams{}, testState))
}

func Test_handleWorkspaceNodeListContext_NoWorkflowHandle(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeListContext(&WorkspaceNodeListParams{}, &task.State{}),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeListContext_NoWorkflowId(t *testing.T) {
	var testParams = &WorkspaceNodeListParams{
		WorkflowHandle: "expectedHandle",
	}

	assert.ErrorIs(t, handleWorkspaceNodeListContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceFlowUnresolved())
}

func Test_handleWorkspaceNodeListIncomplete(t *testing.T) {
	var expectedWorkspace = "expectedWorkspace"
	var expectedWorkflow = "expectedWorkflow"
	var expectedFlow = &WorkspaceFlow{
		Id:         38,
		Name:       "flow name",
		WorkflowId: 40,
		Solution: &Solution{
			Id: new(int64(39)),
			Workflows: []Workflow{
				{
					Id:     new(int64(40)),
					Handle: expectedWorkflow,
				},
			},
		},
	}
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceName:  expectedWorkspace,
		WorkflowHandle: expectedWorkflow,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: []Workspace{
				{
					Id:   37,
					Name: expectedWorkspace,
				},
			},
			listFlows: []WorkspaceFlow{*expectedFlow},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeListIncomplete(testParams, testState))
	assert.Equal(t, expectedFlow.Id, *testParams.WorkspaceFlowId)
	assert.False(t, testState.Completed)
}

func Test_handleWorkspaceNodeListIncomplete_ListFlows_ConflictResults(t *testing.T) {
	var expectedHandle = "wanted"
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:    new(int64(37)),
		WorkflowHandle: expectedHandle,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: []WorkspaceFlow{
				{
					Id:         42,
					WorkflowId: 44,
					Solution: &Solution{
						Workflows: []Workflow{
							{
								Id:     new(int64(44)),
								Handle: expectedHandle,
							},
						},
					},
				},
				{
					Id:         43,
					WorkflowId: 44,
					Solution: &Solution{
						Workflows: []Workflow{
							{
								Id:     new(int64(44)),
								Handle: expectedHandle,
							},
						},
					},
				},
			},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceFlowConflict())
}

func Test_handleWorkspaceNodeListIncomplete_ListFlows_Error(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlowsError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListIncomplete_ListFlows_InvalidWorkflow(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:    new(int64(37)),
		WorkflowHandle: "wanted",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: []WorkspaceFlow{
				{
					Id: 42,
					// Just like with no solution, but the workflowHandle specified is not found
					Solution: &Solution{
						Workflows: []Workflow{},
					},
				},
			},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState), ErrorWorkflowNotFound)
}

func Test_handleWorkspaceNodeListIncomplete_ListFlows_NoResults(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: []WorkspaceFlow{},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeListIncomplete_ListFlows_NoSolutions(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: []WorkspaceFlow{
				{
					Id: 37,
					// Should be a solution otherwise the Flow is invalid
				},
			},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeListIncomplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceFlowId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListIncomplete_NoWorkspace(t *testing.T) {
	var expectedWorkspace = &Workspace{
		Id:   37,
		Name: "notTheOne",
	}
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceName: "wanted",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: []Workspace{*expectedWorkspace},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceNotFound())
}

func Test_handleWorkspaceNodeListIncomplete_NoWorkspace_Conflict(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceName: "wanted",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: []Workspace{
				{
					Id:   37,
					Name: "wanted",
				},
				{
					Id:   38,
					Name: "wanted",
				},
			},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceConflict())
}

func Test_handleWorkspaceNodeListIncomplete_NoWorkspace_Error(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceName: "wanted",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaceError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListIncomplete_NoWorkspace_NoResults(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceName: "wanted",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: []Workspace{},
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceNotFound())
}

func Test_handleWorkspaceNodeListIncomplete_ResolvedError(t *testing.T) {
	var testState = &task.State{}
	var testParams = &WorkspaceNodeListParams{}

	assert.ErrorIs(t, handleWorkspaceNodeListIncomplete(testParams, testState),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeListPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId:     new(int64(37)),
		WorkspaceFlowId: new(int64(42)),
	}
	var testClient = &stubWorkspaceClient{}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.NotContains(t, output, testClient.ListWorkspacesUrl())
	assert.NotContains(t, output, testClient.ListWorkspaceFlowsUrl())
	assert.Contains(t, output, testClient.ListWorkspaceNodesUrl())
}

func Test_handleWorkspaceNodeListPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceFlowId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeListPretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeListPretend_NoWorkspace(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testClient = &stubWorkspaceClient{}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testClient.ListWorkspacesUrl())
	assert.Contains(t, output, testClient.ListWorkspaceFlowsUrl())
	assert.Contains(t, output, testClient.ListWorkspaceNodesUrl())
}

func Test_handleWorkspaceNodeListPretend_Unresolved(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceFlowUnresolved(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeListParams{
		Broker: &Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		WorkspaceId: new(int64(37)),
	}
	var testClient = &stubWorkspaceClient{}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return testClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.NotContains(t, output, testClient.ListWorkspacesUrl())
	assert.Contains(t, output, testClient.ListWorkspaceFlowsUrl())
	assert.Contains(t, output, testClient.ListWorkspaceNodesUrl())
}

func Test_handleWorkspaceNodeResolveComplete(t *testing.T) {
	var expectedHandle = "myHandle"
	var testNodes = []WorkspaceNode{
		{
			Id: int64(37),
			SmartFunction: &Function{
				Handle: "notTheHandle",
			},
		},
		{
			Id: int64(39),
			SmartFunction: &Function{
				Handle: expectedHandle,
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: expectedHandle,
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceId: new(int64(77)),
			FlowId:      new(int64(78)),
			FnHandle:    expectedHandle,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listNodes: testNodes,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeResolveComplete(testParams, testState))
	assert.Equal(t, testNodes[1].Id, *testParams.NodeId)
	assert.Empty(t, testState.Output)
}

func Test_handleWorkspaceNodeResolveComplete_NoNode(t *testing.T) {
	var expectedHandle = "handle"
	var expectedWorkspace = "workspace"
	var testWorkspaces = []Workspace{
		{
			Id:   int64(67),
			Name: "notTheWorkspaceYouAreLookingFor",
		},
		{
			Id:   int64(69),
			Name: expectedWorkspace,
		},
	}
	var testFlows = []WorkspaceFlow{
		{
			Id: int64(73),
			Solution: &Solution{
				Workflows: []Workflow{
					{
						Id:     new(int64(74)),
						Handle: "notTheRightWorkflow",
					},
					{
						Id:     new(int64(75)),
						Handle: expectedHandle,
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName:  expectedWorkspace,
			WorkflowHandle: expectedHandle,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: testWorkspaces,
			listFlows:      testFlows,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeResolveComplete(testParams, testState))
	assert.Empty(t, testState.Output)
}

func Test_handleWorkspaceNodeResolveComplete_FlowRequiredError(t *testing.T) {
	var testFlows = []WorkspaceFlow{
		{
			Id: int64(73),
			Solution: &Solution{
				Workflows: []Workflow{
					{
						Id:     new(int64(74)),
						Handle: "notTheRightWorkflow",
					},
				},
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceId:    new(int64(67)),
			WorkflowHandle: "myWorkflow",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listFlows: testFlows,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState),
		workspace.Errors.WorkspaceFlowRequired())
}

func Test_handleWorkspaceNodeResolveComplete_ListFlowsError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedName = "myWorkspace"
	var testWorkspaces = []Workspace{
		{
			Id:   int64(67),
			Name: "notTheWorkspaceYouAreLookingFor",
		},
		{
			Id:   int64(69),
			Name: expectedName,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: expectedName,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: testWorkspaces,
			listFlowsError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeResolveComplete_ListNodesError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceId: new(int64(73)),
			FlowId:      new(int64(74)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listNodesError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeResolveComplete_ListWorkspacesError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaceError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeResolveComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeResolveComplete_SfHandleError(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(&WorkspaceNodeResolveParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeFnHandleRequired())
}

func Test_handleWorkspaceNodeResolveComplete_WorkspaceNotFound(t *testing.T) {
	var testWorkspaces = []Workspace{
		{
			Id:   int64(67),
			Name: "notTheWorkspaceYouAreLookingFor",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "myFunction",
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceName: "myWorkspace",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			listWorkspaces: testWorkspaces,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveComplete(testParams, testState),
		workspace.Errors.WorkspaceNotFound())
}

func Test_handleWorkspaceNodeResolveContext(t *testing.T) {
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceId: new(int64(37)),
			FlowId:      new(int64(73)),
			FnHandle:    "myHandle",
		},
	}
	var testState = &task.State{}

	assert.NoError(t, handleWorkspaceNodeResolveContext(testParams, testState))
	assert.Equal(t, testParams.FnHandle, testState.Output)
}

func Test_handleWorkspaceNodeResolveContext_NodeKnownError(t *testing.T) {
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeResolveContext(testParams, testState),
		workspace.Errors.WorkspaceNodeKnown())
}

func Test_handleWorkspaceNodeResolveContext_SfHandleError(t *testing.T) {
	var testParams = &WorkspaceNodeResolveParams{}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeResolveContext(testParams, testState),
		workspace.Errors.WorkspaceNodeFnHandleRequired())
}

func Test_handleWorkspaceNodeResolveContext_WfHandleError(t *testing.T) {
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			FnHandle: "myFunction",
		},
	}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeResolveContext(testParams, testState),
		workspace.Errors.WorkflowHandleRequired())
}

func Test_handleWorkspaceNodeResolveContext_WsNameError(t *testing.T) {
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			FnHandle:       "myFunction",
			WorkflowHandle: "myWorkflow",
		},
	}
	var testState = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeResolveContext(testParams, testState),
		workspace.Errors.WorkspaceNameRequired())
}

func Test_handleWorkspaceNodeResolveIncomplete(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolveIncomplete(&WorkspaceNodeResolveParams{}, testState), expectedError)
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceNodeResolveIncomplete_NodeKnownError(t *testing.T) {
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceNodeKnown(),
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceNodeResolveIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceNodeResolvePretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
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
		return &stubWorkspaceClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleWorkspaceNodeResolvePretend_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolvePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeResolvePretend_NodeKnownError(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  workspace.Errors.WorkspaceNodeKnown(),
		Logger: testLogger,
	}
	var testParams = &WorkspaceNodeResolveParams{
		NodeResolveParams: workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceNodeResolvePretend(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
}

func Test_handleWorkspaceNodeResolvePretend_NodeOnly(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: workspace.NodeResolveParams{
			WorkspaceId: new(int64(37)),
			FlowId:      new(int64(73)),
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
		return &stubWorkspaceClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, cast.ToString(*testParams.FlowId))
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleWorkspaceNodeResolvePretend_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleWorkspaceNodeResolvePretend(&WorkspaceNodeResolveParams{}, testState), expectedError)
}

func Test_handleWorkspaceNodeSetListComplete(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id: int64(37),
		InputPorts: []WorkspaceNodePort{
			{
				DataSetId:  new(int64(73)),
				PortHandle: "port",
			},
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetListParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetListComplete(testParams, testState))

	if actual, ok := testState.Internal.([]WorkspaceNodePort); ok {
		assert.Equal(t, expectedNode.InputPorts, actual)
	} else {
		assert.Fail(t, "expected a slice of WorkspaceNodePort")
	}
}

func Test_handleWorkspaceNodeSetListComplete_InspectError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetListParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetListComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetListComplete(&WorkspaceNodeSetListParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeSetListComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetListParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetListComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetListContext(t *testing.T) {
	var testParams = &WorkspaceNodeSetListParams{
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(42)),
		},
	}
	var testState = &task.State{}

	assert.NoError(t, handleWorkspaceNodeSetListContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.NodeId), testState.Output)
}

func Test_handleWorkspaceNodeSetListContext_InvalidNodeId(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetListContext(&WorkspaceNodeSetListParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeSetListContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleWorkspaceNodeSetListContext(&WorkspaceNodeSetListParams{}, testState))
}

func Test_handleWorkspaceNodeSetListPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetListParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
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
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeId))
}

func Test_handleWorkspaceNodeSetListPretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetListParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetListPretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetResolveComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "expected",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var expectedDataSet = &DataSet{
		Id:     new(int64(73)),
		Handle: "dsHandle",
	}
	var expectedNode = &WorkspaceNode{
		Id: int64(42),
		InputPorts: []WorkspaceNodePort{
			{
				DataSetId:  expectedDataSet.Id,
				PortHandle: "expected",
				DataSet:    expectedDataSet,
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetResolveComplete(testParams, testState))
	assert.Empty(t, testState.Output)

	if actual, ok := testState.Internal.(DataSet); ok {
		assert.Equal(t, *expectedDataSet, actual)
	} else {
		assert.Fail(t, "expected a DataSet")
	}
}

func Test_handleWorkspaceNodeSetResolveComplete_InspectError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetResolveComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetResolveComplete(&WorkspaceNodeSetResolveParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeFnPortRequired())
}

func Test_handleWorkspaceNodeSetResolveComplete_NoDataSet(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "expected",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var expectedNode = &WorkspaceNode{
		Id: int64(42),
		InputPorts: []WorkspaceNodePort{
			{
				PortHandle: "expected",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetResolveComplete(testParams, testState))
	assert.Nil(t, testState.Internal)
	assert.Empty(t, testState.Output)
}

func Test_handleWorkspaceNodeSetResolveComplete_PortNotFound(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "notFound",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var expectedNode = &WorkspaceNode{
		Id: int64(42),
		InputPorts: []WorkspaceNodePort{
			{
				PortHandle: "notTheOne",
			},
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			inspectNode: expectedNode,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveComplete(testParams, testState),
		workspace.Errors.WorkspaceNodeFnPortNotFound())
}

func Test_handleWorkspaceNodeSetResolveComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetResolveContext(t *testing.T) {
	var expectedHandle = "port"
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: expectedHandle,
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var testState = &task.State{}

	assert.NoError(t, handleWorkspaceNodeSetResolveContext(testParams, testState))
	assert.Equal(t, expectedHandle, testState.Output)
}

func Test_handleWorkspaceNodeSetResolveContext_DataSetKnown(t *testing.T) {
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeSetKnown())
}

func Test_handleWorkspaceNodeSetResolveContext_InvalidNode(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetResolveContext(&WorkspaceNodeSetResolveParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeSetResolveContext_InvalidPort(t *testing.T) {
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeFnPortRequired())
}

func Test_handleWorkspaceNodeSetResolveContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleWorkspaceNodeSetResolveContext(&WorkspaceNodeSetResolveParams{}, testState))
}

func Test_handleWorkspaceNodeSetResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  workspace.Errors.WorkspaceNodeSetKnown(),
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceNodeSetResolveIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceNodeSetResolveIncomplete_StateError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolveIncomplete(&WorkspaceNodeSetResolveParams{}, testState), expectedError)
	assert.True(t, testState.Completed)
}

func Test_handleWorkspaceNodeSetResolvePretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			PortHandle: "port",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
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
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetResolvePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeId))
}

func Test_handleWorkspaceNodeSetResolvePretend_DataSetKnown(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(42)),
		},
	}
	var testState = &task.State{
		Logger: testLogger,
		Error:  workspace.Errors.WorkspaceNodeSetKnown(),
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleWorkspaceNodeSetResolvePretend(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
}

func Test_handleWorkspaceNodeSetResolvePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolvePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetResolvePretend_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetResolvePretend(&WorkspaceNodeSetResolveParams{}, testState), expectedError)
}

func Test_handleWorkspaceNodeSetUpdateComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(32)),
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var expectedNode = &WorkspaceNode{
		Id:          int64(49),
		WorkspaceId: int64(37),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			updateNodeInput: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetUpdateComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); ok {
		assert.Equal(t, *expectedNode, actual)
	} else {
		assert.Fail(t, "expected a WorkspaceNode")
	}
}

func Test_handleWorkspaceNodeSetUpdateComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetUpdateComplete(&WorkspaceNodeSetResolveParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeSetRequired())
}

func Test_handleWorkspaceNodeSetUpdateComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetUpdateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetUpdateComplete_UpdateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(32)),
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			updateNodeInputError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetUpdateComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSetUpdateContext(t *testing.T) {
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(42)),
		},
	}
	var testState = &task.State{}

	assert.NoError(t, handleWorkspaceNodeSetUpdateContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.DataSetId), testState.Output)
}

func Test_handleWorkspaceNodeSetUpdateContext_InvalidDataSet(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSetUpdateContext(&WorkspaceNodeSetResolveParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeSetRequired())
}

func Test_handleWorkspaceNodeSetUpdateContext_InvalidNode(t *testing.T) {
	var testParams = &WorkspaceNodeSetResolveParams{
		NodeSetParams: workspace.NodeSetParams{
			DataSetId: new(int64(67)),
		},
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetUpdateContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeSetUpdateContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleWorkspaceNodeSetUpdateContext(&WorkspaceNodeSetResolveParams{}, testState))
}

func Test_handleWorkspaceNodeSetUpdatePretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeSetParams: workspace.NodeSetParams{
			DataSetId:  new(int64(32)),
			PortHandle: "expectedPort",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
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
		return &stubWorkspaceClient{}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSetUpdatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeId))
	assert.Contains(t, output, cast.ToString(*testParams.DataSetId))
	assert.Contains(t, output, testParams.PortHandle)

}

func Test_handleWorkspaceNodeSetUpdatePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceNodeSetResolveParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSetUpdatePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSourceContext(t *testing.T) {
	var testState = &task.State{}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			DataSourceId: new(int64(73)),
		},
		NodeId: new(int64(37)),
	}

	assert.NoError(t, handleWorkspaceNodeSourceContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.DataSourceId), testState.Output)
}

func Test_handleWorkspaceNodeSourceContext_DataSourceRequiredError(t *testing.T) {
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{},
		NodeId:                  new(int64(37)),
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeSourceRequired())
}

func Test_handleWorkspaceNodeSourceContext_NodeRequiredError(t *testing.T) {
	var testParams = &WorkspaceNodeSourceParams{}

	assert.ErrorIs(t, handleWorkspaceNodeSourceContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeSourceContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleWorkspaceNodeSourceContext(&WorkspaceNodeSourceParams{}, testState))
}

func Test_handleWorkspaceNodeSourceAddComplete(t *testing.T) {
	var expectedDsId = int64(37)
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{int64(42)},
	}
	var modifiedNode = expectedNode.AddDataSource(expectedDsId)
	var testState = &task.State{
		Logger: logrus.New(),
		Output: cast.ToString(expectedDsId),
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: &expectedDsId,
		},
		NodeId: new(int64(73)),
	}
	var stubClient = &stubWorkspaceClient{
		getNode:    expectedNode,
		updateNode: modifiedNode,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return stubClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeSourceAddComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *modifiedNode, actual)
	}
}

func Test_handleWorkspaceNodeSourceAddComplete_GetError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceAddComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeSourceAddComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSourceAddComplete(&WorkspaceNodeSourceParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeSourceRequired())
}

func Test_handleWorkspaceNodeSourceAddComplete_NoAssigned(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{int64(37)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSourceAddComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); ok {
		assert.Equal(t, *expectedNode, actual)
	} else {
		assert.Fail(t, "expected workspace node")
	}
}

func Test_handleWorkspaceNodeSourceAddComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceAddComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeSourceAddComplete_UpdateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedNode = &WorkspaceNode{
		Id: int64(42),
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode:         expectedNode,
			updateNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceAddComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSourcePretend(t *testing.T) {
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		NodeId: new(int64(37)),
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
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
		return &stubWorkspaceClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSourcePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, "-X GET")
	assert.Contains(t, output, "-X POST")
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleWorkspaceNodeSourcePretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeSourcePretend(&WorkspaceNodeSourceParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeSourceRequired())
}

func Test_handleWorkspaceNodeSourcePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var testState = &task.State{
		Output: "output",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourcePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeSourceRemoveComplete(t *testing.T) {
	var expectedDsId = int64(37)
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{expectedDsId},
	}
	var modifiedNode = expectedNode.RemoveDataSource(expectedDsId)
	var testState = &task.State{
		Logger: logrus.New(),
		Output: cast.ToString(expectedDsId),
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(expectedDsId),
		},
		NodeId: new(int64(73)),
	}
	var stubClient = &stubWorkspaceClient{
		getNode:    expectedNode,
		updateNode: modifiedNode,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return stubClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeSourceRemoveComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *modifiedNode, actual)
	}
}

func Test_handleWorkspaceNodeSourceRemoveComplete_GetError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceRemoveComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeSourceRemoveComplete_OutputError(t *testing.T) {
	var testTask = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeSourceRemoveComplete(&WorkspaceNodeSourceParams{}, testTask),
		workspace.Errors.WorkspaceNodeSourceRequired())
}

func Test_handleWorkspaceNodeSourceRemoveComplete_NotAssigned(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{int64(1337)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeSourceRemoveComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *expectedNode, actual)
	}
}

func Test_handleWorkspaceNodeSourceRemoveComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceRemoveComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeSourceRemoveComplete_UpdateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{int64(37)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeSourceParams{
		DataSourceResolveParams: &DataSourceResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataSourceId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode:         expectedNode,
			updateNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeSourceRemoveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeStoreContext(t *testing.T) {
	var testState = &task.State{}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			DataStoreId: new(int64(73)),
		},
		NodeId: new(int64(37)),
	}

	assert.NoError(t, handleWorkspaceNodeStoreContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.DataStoreId), testState.Output)
}

func Test_handleWorkspaceNodeStoreContext_DataStoreRequiredError(t *testing.T) {
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{},
		NodeId:                 new(int64(37)),
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeStoreRequired())
}

func Test_handleWorkspaceNodeStoreContext_NodeRequiredError(t *testing.T) {
	var testParams = &WorkspaceNodeStoreParams{}

	assert.ErrorIs(t, handleWorkspaceNodeStoreContext(testParams, &task.State{}),
		workspace.Errors.WorkspaceNodeRequired())
}

func Test_handleWorkspaceNodeStoreContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "output",
	}

	assert.NoError(t, handleWorkspaceNodeStoreContext(&WorkspaceNodeStoreParams{}, testState))
}

func Test_handleWorkspaceNodeStoreAddComplete(t *testing.T) {
	var expectedDsId = int64(37)
	var expectedNode = &WorkspaceNode{
		Id:            int64(42),
		DataSourceIds: []int64{int64(42)},
	}
	var modifiedNode = expectedNode.AddDataStore(expectedDsId)
	var testState = &task.State{
		Logger: logrus.New(),
		Output: cast.ToString(expectedDsId),
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: &expectedDsId,
		},
		NodeId: new(int64(73)),
	}
	var stubClient = &stubWorkspaceClient{
		getNode:    expectedNode,
		updateNode: modifiedNode,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return stubClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeStoreAddComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *modifiedNode, actual)
	}
}

func Test_handleWorkspaceNodeStoreAddComplete_GetError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreAddComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeStoreAddComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeStoreAddComplete(&WorkspaceNodeStoreParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeStoreRequired())
}

func Test_handleWorkspaceNodeStoreAddComplete_NoAssigned(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:           int64(42),
		DataStoreIds: []int64{int64(37)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeStoreAddComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); ok {
		assert.Equal(t, *expectedNode, actual)
	} else {
		assert.Fail(t, "expected workspace node")
	}
}

func Test_handleWorkspaceNodeStoreAddComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreAddComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeStoreAddComplete_UpdateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedNode = &WorkspaceNode{
		Id: int64(42),
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode:         expectedNode,
			updateNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreAddComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeStorePretend(t *testing.T) {
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
		NodeId: new(int64(37)),
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
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
		return &stubWorkspaceClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeStorePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, "-X GET")
	assert.Contains(t, output, "-X POST")
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleWorkspaceNodeStorePretend_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleWorkspaceNodeStorePretend(&WorkspaceNodeStoreParams{}, &task.State{}),
		workspace.Errors.WorkspaceNodeStoreRequired())
}

func Test_handleWorkspaceNodeStorePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				AuthFile: "file",
				HostAddr: "hostAddr",
			},
		},
	}
	var testState = &task.State{
		Output: "output",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeStorePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceNodeStoreRemoveComplete(t *testing.T) {
	var expectedDsId = int64(37)
	var expectedNode = &WorkspaceNode{
		Id:           int64(42),
		DataStoreIds: []int64{expectedDsId},
	}
	var modifiedNode = expectedNode.RemoveDataStore(expectedDsId)
	var testState = &task.State{
		Logger: logrus.New(),
		Output: cast.ToString(expectedDsId),
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(expectedDsId),
		},
		NodeId: new(int64(73)),
	}
	var stubClient = &stubWorkspaceClient{
		getNode:    expectedNode,
		updateNode: modifiedNode,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return stubClient, nil
	}

	assert.NoError(t, handleWorkspaceNodeStoreRemoveComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *modifiedNode, actual)
	}
}

func Test_handleWorkspaceNodeStoreRemoveComplete_GetError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreRemoveComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeStoreRemoveComplete_OutputError(t *testing.T) {
	var testTask = &task.State{}

	assert.ErrorIs(t, handleWorkspaceNodeStoreRemoveComplete(&WorkspaceNodeStoreParams{}, testTask),
		workspace.Errors.WorkspaceNodeStoreRequired())
}

func Test_handleWorkspaceNodeStoreRemoveComplete_NotAssigned(t *testing.T) {
	var expectedNode = &WorkspaceNode{
		Id:           int64(42),
		DataStoreIds: []int64{int64(1337)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode: expectedNode,
		}, nil
	}

	assert.NoError(t, handleWorkspaceNodeStoreRemoveComplete(testParams, testState))

	if actual, ok := testState.Internal.(WorkspaceNode); !ok {
		assert.Fail(t, "expected a workspace node")
	} else {
		assert.Equal(t, *expectedNode, actual)
	}
}

func Test_handleWorkspaceNodeStoreRemoveComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testTask = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreRemoveComplete(testParams, testTask), expectedError)
}

func Test_handleWorkspaceNodeStoreRemoveComplete_UpdateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var expectedNode = &WorkspaceNode{
		Id:           int64(42),
		DataStoreIds: []int64{int64(37)},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "37",
	}
	var testParams = &WorkspaceNodeStoreParams{
		DataStoreResolveParams: &DataStoreResolveParams{
			Broker: Broker{
				HostAddr: "hostAddr",
			},
			DataStoreId: new(int64(37)),
		},
		NodeId: new(int64(73)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubWorkspaceClient{
			getNode:         expectedNode,
			updateNodeError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleWorkspaceNodeStoreRemoveComplete(testParams, testState), expectedError)
}
