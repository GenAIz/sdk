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
)

type stubWorkspaceClient struct {
	client
	createFlowError       error
	createFlowExpected    *WorkspaceFlow
	createFlowWorkspaceId int64
	createFlowWorkflowId  int64
	createFlowName        string
	createFlowDescription string
	createWorkspaceError  error
	createWorkspaceParam  *Workspace
	createWorkspace       *Workspace
	findSolutionError     error
	findSolutionExpected  *Solution
	findSolutionOem       string
	findSolutionHandle    string
	findSolutionVersion   string
	listWorkspaceError    error
	listWorkspaceMask     int
	listWorkspaceFlags    int
	listWorkspaces        []Workspace
	liseWorkspacesUserId  int
}

func (swc *stubWorkspaceClient) CreateWorkspace(workspace *Workspace) (*Workspace, error) {
	swc.createWorkspaceParam = workspace

	if swc.createWorkspace != nil {
		return swc.createWorkspace, nil
	}

	return nil, swc.createWorkspaceError
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

func (swc *stubWorkspaceClient) GetUserId() int {
	return swc.liseWorkspacesUserId
}

func (swc *stubWorkspaceClient) ListWorkspaces(mask, flags int) ([]Workspace, error) {
	swc.listWorkspaceMask = mask
	swc.listWorkspaceFlags = flags
	return swc.listWorkspaces, swc.listWorkspaceError
}

func TestNewWorkspaceCreateTask(t *testing.T) {
	var testTask = NewWorkspaceCreateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
	assert.Nil(t, testTask.OnIncomplete)
}

func TestNewWorkspaceFlowCreateTask(t *testing.T) {
	var testTask = NewWorkspaceFlowCreateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewWorkspaceFlowResolveTask(t *testing.T) {
	var testTask = NewWorkspaceFlowResolveTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewWorkspaceFlowSolutionTask(t *testing.T) {
	var testTask = NewWorkspaceFlowSolutionTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewWorkspaceListTask(t *testing.T) {
	var testTask = NewWorkspaceListTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
	assert.Nil(t, testTask.OnIncomplete)
}

func Test_handleWorkspaceCreateContext(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{
			Name:       "expectedName",
			RcEnabled:  true,
			Visibility: WorkspaceVisibilityOrg,
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
			Visibility: WorkspaceVisibilityPrivate,
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

	assert.ErrorIs(t, handleWorkspaceCreateContext(testParams, testState), ErrorWorkspaceVisibility)
}

func Test_handleWorkspaceCreateContext_EmptyWorkspace(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &WorkspaceCreateParams{}

	assert.ErrorIs(t, handleWorkspaceCreateContext(testParams, testState), ErrorWorkspaceEmpty)
}

func Test_handleWorkspaceCreateContext_RcDisabled(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testState = &task.State{Logger: testLogger}
	var testParams = &WorkspaceCreateParams{
		Workspace: &Workspace{
			Name:       "expectedName",
			RcEnabled:  false,
			Visibility: WorkspaceVisibilityPrivate,
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

	assert.ErrorIs(t, handleWorkspaceCreateComplete(testParams, testState), ErrorWorkspaceEmpty)
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
			Visibility:  WorkspaceVisibilityPrivate,
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

	assert.ErrorIs(t, handleWorkspaceCreatePretend(testParams, testState), ErrorWorkspaceEmpty)
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

	assert.ErrorIs(t, handleWorkspaceFlowCreateContext(testParams, &task.State{}), ErrorWorkflowIdRequired)
}

func Test_handleWorkspaceFlowCreateContext_NoWorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowCreateParams{
		WorkflowId: new(int64(43)),
	}

	assert.ErrorIs(t, handleWorkspaceFlowCreateContext(testParams, &task.State{}), ErrorWorkspaceIdRequired)
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

	assert.ErrorIs(t, handleWorkspaceFlowCreateComplete(testParams, &task.State{}), ErrorWorkspaceFlowInvalid)
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

	assert.ErrorIs(t, handleWorkspaceFlowCreatePretend(testParams, &task.State{}), ErrorWorkspaceFlowInvalid)
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

	assert.NoError(t, handleWorkspaceFlowResolveComplete(testParams, testState))
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

	assert.ErrorIs(t, handleWorkspaceFlowResolveComplete(testParams, testState), expectedError)
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

	assert.ErrorIs(t, handleWorkspaceFlowResolveComplete(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolveComplete_WorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowResolveComplete(testParams, &task.State{}))
}

func Test_handleWorkspaceFlowResolveContext(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceName: "name",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleWorkspaceFlowResolveContext(testParams, testState))
}

func Test_handleWorkspaceFlowResolveContext_NoName(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleWorkspaceFlowResolveContext(testParams, testState), ErrorWorkspaceNameRequired)
}

func Test_handleWorkspaceFlowResolveContext_OutputCheck(t *testing.T) {
	var testState = &task.State{Output: "check"}

	assert.NoError(t, handleWorkspaceFlowResolveContext(&WorkspaceFlowResolveParams{}, testState))
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

	assert.ErrorIs(t, handleWorkspaceFlowResolveContext(testParams, testState), ErrorWorkspaceIdKnown)
}

func Test_handleWorkspaceFlowResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleWorkspaceFlowResolveIncomplete(&WorkspaceFlowResolveParams{}, testState), testState.Error)
	assert.False(t, testState.Completed)
}

func Test_handleWorkspaceFlowResolveIncomplete_KnownId(t *testing.T) {
	var testState = &task.State{
		Error:  ErrorWorkspaceIdKnown,
		Logger: logrus.New(),
	}
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(37)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowResolveIncomplete(testParams, testState))
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

	assert.NoError(t, handleWorkspaceFlowResolvePretend(testParams, testState))

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

	assert.ErrorIs(t, handleWorkspaceFlowResolvePretend(testParams, testState), expectedError)
}

func Test_handleWorkspaceFlowResolvePretend_WorkspaceId(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		WorkspaceFlowCreateParams: &WorkspaceFlowCreateParams{
			WorkspaceId: new(int64(73)),
		},
	}

	assert.NoError(t, handleWorkspaceFlowResolvePretend(testParams, &task.State{}))
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

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}), ErrorWorkflowHandleRequired)
}

func Test_handleWorkspaceFlowSolutionContext_NoSolutionOem(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionHandle:  "handle",
		SolutionVersion: "version",
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}), ErrorWorkflowOemRequired)
}

func Test_handleWorkspaceFlowSolutionContext_NoSolutionVersion(t *testing.T) {
	var testParams = &WorkspaceFlowResolveParams{
		SolutionHandle: "handle",
		SolutionOem:    "oem",
	}

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, &task.State{}), ErrorWorkflowVersionRequired)
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

	assert.ErrorIs(t, handleWorkspaceFlowSolutionContext(testParams, testState), ErrorWorkflowIdKnown)
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
		Error:  ErrorWorkflowIdKnown,
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

	assert.ErrorIs(t, handleWorkspaceListContext(testParams, &task.State{}), ErrorWorkspaceInvalidNco)
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
				assert.ErrorIs(t, handleWorkspaceListContext(testParams, &task.State{}), ErrorWorkspaceInvalidOwner)
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
