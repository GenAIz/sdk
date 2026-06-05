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
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
)

type stubWorkspaceClient struct {
	client
	createWorkspaceError error
	createWorkspaceParam *Workspace
	createWorkspace      *Workspace
	listWorkspaceError   error
	listWorkspaceMask    int
	listWorkspaceFlags   int
	listWorkspaces       []Workspace
	liseWorkspacesUserId int
}

func (swc *stubWorkspaceClient) CreateWorkspace(workspace *Workspace) (*Workspace, error) {
	swc.createWorkspaceParam = workspace

	if swc.createWorkspace != nil {
		return swc.createWorkspace, nil
	}

	return nil, swc.createWorkspaceError
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
		FromDate: lang.Ref(time.UnixMilli(1)),
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
		FromDate: lang.Ref(time.UnixMilli(1)),
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
