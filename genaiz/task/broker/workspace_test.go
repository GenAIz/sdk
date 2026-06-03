package broker

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
)

type stubWorkspaceClient struct {
	client
	createWorkspaceError error
	createWorkspaceParam *Workspace
	createWorkspace      *Workspace
}

func (swc *stubWorkspaceClient) CreateWorkspace(workspace *Workspace) (*Workspace, error) {
	swc.createWorkspaceParam = workspace

	if swc.createWorkspace != nil {
		return swc.createWorkspace, nil
	}

	return nil, swc.createWorkspaceError
}

func TestNewWorkspaceCreateTask(t *testing.T) {
	var testTask = NewWorkspaceCreateTask()

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
