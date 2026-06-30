package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type stubFunctionClient struct {
	client
	getFunction       *Function
	getFunctionError  error
	getFunctionId     int64
	provisionError    error
	provisionIdentity *shared.Identity
	provisionExtras   map[string]any
	publishError      error
	publishFunction   *Function
}

func (sfc *stubFunctionClient) GetFunction(id int64) (*Function, error) {
	sfc.getFunctionId = id
	return sfc.getFunction, sfc.getFunctionError
}

func (sfc *stubFunctionClient) ProvisionFunction(function *Function, extras map[string]any) (*shared.Identity, error) {
	sfc.provisionExtras = extras

	if sfc.provisionError != nil {
		return nil, sfc.provisionError
	}

	if sfc.provisionIdentity != nil {
		return sfc.provisionIdentity, nil
	}

	return function.asIdentity(), nil
}

func (sfc *stubFunctionClient) PublishFunction(*shared.Identity) (*Function, error) {
	if sfc.publishError != nil {
		return nil, sfc.publishError
	}

	if sfc.publishFunction != nil {
		return sfc.publishFunction, nil
	}

	return nil, nil
}

func TestProvisionParams_ToGetParams(t *testing.T) {
	var testParams = &ProvisionParams{
		Broker: Broker{
			HostAddr: "hostAddr",
		},
		GetParams: GetParams{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
	}
	var testGetParams = testParams.ToGetParams()

	assert.Equal(t, testParams.HostAddr, testGetParams.HostAddr)
	assert.Equal(t, testParams.Oem, testGetParams.Oem)
	assert.Equal(t, testParams.Handle, testGetParams.Handle)
	assert.Equal(t, testParams.Version, testGetParams.Version)
}

func TestProvisionParams_asFunction(t *testing.T) {
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:     "oem",
			Handle:  "handle",
			Version: "version",
		},
		Arches:      []string{"arch1", "arch2"},
		Description: "desc",
		Name:        "name",
		Type:        "type",
	}
	var testFunc = testParams.asFunction()

	assert.Equal(t, testParams.Arches, testFunc.Arches)
	assert.Equal(t, testParams.Description, testFunc.Description)
	assert.Equal(t, testParams.Handle, testFunc.Handle)
	assert.Equal(t, testParams.Name, testFunc.Name)
	assert.Equal(t, testParams.Oem, testFunc.Oem)
	assert.Equal(t, strings.ToUpper(testParams.Type), testFunc.Type)
	assert.Equal(t, testParams.Version, testFunc.Version)
}

func TestProvisionParams_validate_Connector(t *testing.T) {
	var testParams = &ProvisionParams{
		Type:        shared.FunctionTypeConnector,
		DataSources: []string{"source"},
		DataStores:  []string{"store"},
	}

	assert.Nil(t, testParams.validate())
}

func TestProvisionParams_validate_FunctionSources(t *testing.T) {
	var testParams = &ProvisionParams{
		Type:        shared.FunctionTypeFunction,
		DataSources: []string{"source"},
	}

	assert.Error(t, testParams.validate())
}

func TestNewFunctionGetTask(t *testing.T) {
	var testTask = NewFunctionGetTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewFunctionIdTask(t *testing.T) {
	var testTask = NewFunctionIdTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewFunctionProvisionTask(t *testing.T) {
	var testTask = NewFunctionProvisionTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewFunctionPublishTask(t *testing.T) {
	var testTask = NewFunctionPublishTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleFunctionGetComplete(t *testing.T) {
	var expectedFunction = &Function{Id: 37}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
			HostAddr: "hostAddr",
		},
		Id: new(expectedFunction.Id),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			getFunction: expectedFunction,
		}, nil
	}

	assert.NoError(t, handleFunctionGetComplete(testParams, testState))
	assert.Equal(t, expectedFunction, testState.Internal)
}

func Test_handleFunctionGetComplete_GetFunctionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
			HostAddr: "hostAddr",
		},
		Id: new(37),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			getFunctionError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleFunctionGetComplete(testParams, testState), expectedError)
}

func Test_handleFunctionGetComplete_NoId(t *testing.T) {
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{}, nil
	}

	assert.ErrorIs(t, handleFunctionGetComplete(testParams, &task.State{}), errorInvalidFunctionGet)
}

func Test_handleFunctionGetComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
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

	assert.ErrorIs(t, handleFunctionGetComplete(testParams, &task.State{}), expectedError)
}

func Test_handleFunctionGetContext(t *testing.T) {
	var testParams = &GetParams{Id: new(37)}

	assert.NoError(t, handleFunctionGetContext(testParams, &task.State{}))
}

func Test_handleFunctionGetContext_CheckOutput(t *testing.T) {
	var testState = &task.State{Output: "check"}

	assert.NoError(t, handleFunctionGetContext(&GetParams{}, testState))
}

func Test_handleFunctionGetContext_NoIdNoFqdn(t *testing.T) {
	assert.ErrorIs(t, handleFunctionGetContext(&GetParams{}, &task.State{}), errorInvalidFunctionGet)
}

func Test_handleFunctionGetContext_WithFqdn(t *testing.T) {
	var testParams = &GetParams{
		Oem:     "testOem",
		Handle:  "testHandle",
		Version: "testVersion",
	}

	assert.ErrorIs(t, handleFunctionGetContext(testParams, &task.State{}), errorUnsupportedFunctionGet)
}

func Test_handleFunctionGetPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
			HostAddr: "hostAddr",
		},
		Id: new(37),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{}, nil
	}

	assert.NoError(t, handleFunctionGetPretend(testParams, testState))
}

func Test_handleFunctionGetPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
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

	assert.ErrorIs(t, handleFunctionGetPretend(testParams, &task.State{}), expectedError)
}

func Test_handleFunctionGetPretend_StateError(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}
	var testParams = &GetParams{
		Broker: Broker{
			AuthFile: filepath.Join(t.TempDir(), "file"),
			HostAddr: "hostAddr",
		},
	}

	assert.ErrorIs(t, handleFunctionGetPretend(testParams, testState), testState.Error)
}

func Test_handleFunctionIdContext(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testParams = &GetParams{
		Oem:     "testOem",
		Handle:  "testHandle",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger:   testLogger,
		Internal: &Function{},
	}

	testLogger.Level = logrus.DebugLevel
	assert.NoError(t, handleFunctionIdContext(testParams, testState))
	assert.Equal(t, 1, len(hook.Entries))
	assert.Contains(t, hook.Entries[0].Message, testParams.Oem)
	assert.Contains(t, hook.Entries[0].Message, testParams.Handle)
	assert.Contains(t, hook.Entries[0].Message, testParams.Version)
}

func Test_handleFunctionIdContext_NoInternal(t *testing.T) {
	var testLogger, hook = test.NewNullLogger()
	var testParams = &GetParams{
		Oem:     "testOem",
		Handle:  "testHandle",
		Version: "testVersion",
	}
	var testState = &task.State{
		Logger: testLogger,
	}

	assert.ErrorIs(t, handleFunctionIdContext(testParams, testState), errorMissConfiguredGet)
	assert.Empty(t, hook.Entries)
}

func Test_handleFunctionIdPretend(t *testing.T) {
	assert.Nil(t, handleFunctionIdPretend(&GetParams{}, &task.State{}))
}

func Test_handleFunctionIdUpdate(t *testing.T) {
	var expectedFunction = &Function{Id: 37}
	var testParams = &GetParams{}
	var testState = &task.State{
		Internal: expectedFunction,
	}

	assert.NoError(t, handleFunctionIdUpdate(testParams, testState))

	if testParams.Id != nil {
		assert.Equal(t, expectedFunction.Id, *testParams.Id)
	} else {
		assert.Fail(t, "expected testParams.Id")
	}
}

func Test_handleFunctionIdUpdate_NotAFunction(t *testing.T) {
	var testParams = &GetParams{}

	assert.ErrorIs(t, handleFunctionIdUpdate(testParams, &task.State{}), errorMissConfiguredGet)
	assert.Nil(t, testParams.Id)
}

func Test_handleFunctionProvisionContext(t *testing.T) {
	var testLogger = logrus.New()
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:    "oem",
			Handle: "handle",
		},
	}

	assert.ErrorIs(t, handleFunctionProvisionContext(testParams, &task.State{Logger: testLogger}), errorNoBuildToProvision)
	assert.Error(t, handleFunctionProvisionContext(&ProvisionParams{}, &task.State{Internal: &shared.Identity{}}))
	assert.Error(t, handleFunctionProvisionContext(&ProvisionParams{}, &task.State{Internal: &shared.Identity{Id: "sha256:1"}}))
	assert.NoError(t, handleFunctionProvisionContext(testParams, &task.State{Logger: testLogger, Internal: &shared.Identity{Id: "sha256:1"}}))
	assert.NoError(t, handleFunctionProvisionContext(testParams, &task.State{Logger: testLogger, Internal: &shared.Identity{Id: "sha256:1", Hash: "hash"}}))
}

func Test_handleFunctionProvisionContext_InvalidFunctionStores(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: &shared.Identity{
			Id: "sha256:1",
		},
	}
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:    "oem",
			Handle: "handle",
		},
		Type:       shared.FunctionTypeFunction,
		DataStores: []string{"store"},
	}

	assert.Error(t, handleFunctionProvisionContext(testParams, testState))
}

func Test_handleFunctionProvisionComplete(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: &shared.Identity{},
	}
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:    "testOem",
			Handle: "testHandle",
		},
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testIdentity = &shared.Identity{Id: "id"}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			provisionIdentity: testIdentity,
		}, nil
	}

	assert.NoError(t, handleFunctionProvisionComplete(testParams, testState))
	assert.Empty(t, testState.Internal.(*shared.Identity).Hash)
	assert.Equal(t, testIdentity.Id, testState.Internal.(*shared.Identity).Id)
	assert.NotEmpty(t, testState.Reports)
}

func Test_handleFunctionProvisionComplete_ConflictingHashes(t *testing.T) {
	var expectedHash = "currentHash"
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: &shared.Identity{
			Hash: expectedHash,
		},
	}
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:    "testOem",
			Handle: "testHandle",
		},
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testIdentity = &shared.Identity{Id: "id", Hash: "identityHash"}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			provisionIdentity: testIdentity,
		}, nil
	}

	assert.NoError(t, handleFunctionProvisionComplete(testParams, testState))
	assert.Empty(t, testState.Internal.(*shared.Identity).Hash)
	assert.Equal(t, testIdentity.Id, testState.Internal.(*shared.Identity).Id)
}

func Test_handleFunctionProvisionComplete_CurrentHash(t *testing.T) {
	var expectedHash = "currentHash"
	var testState = &task.State{
		Logger: logrus.New(),
		Internal: &shared.Identity{
			Hash: expectedHash,
		},
	}
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:    "testOem",
			Handle: "testHandle",
		},
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testIdentity = &shared.Identity{Id: "id"}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			provisionIdentity: testIdentity,
		}, nil
	}

	assert.NoError(t, handleFunctionProvisionComplete(testParams, testState))
	assert.Equal(t, expectedHash, testState.Internal.(*shared.Identity).Hash)
	assert.Equal(t, testIdentity.Id, testState.Internal.(*shared.Identity).Id)
}

func Test_handleFunctionProvisionComplete_Failed(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: &shared.Identity{},
	}
	var testParams = &ProvisionParams{
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
		return &stubFunctionClient{
			provisionError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleFunctionProvisionComplete(testParams, testState), expectedError)
}

func Test_handleFunctionProvisionComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &ProvisionParams{
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
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleFunctionProvisionComplete(testParams, &task.State{}))
}

func Test_handleFunctionProvisionPretend(t *testing.T) {
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: &shared.Identity{},
	}
	var testParams = &ProvisionParams{
		GetParams: GetParams{
			Oem:     "testOem",
			Handle:  "testHandle",
			Version: "testVersion",
		},
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Arches:      []string{"arch1"},
		Description: "testDescription",
		Name:        "testName",
		Type:        "testType",
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
		return &stubLoginClient{
			session: &AuthSession{
				Token:    "token",
				Username: "user",
			},
		}, nil
	}

	assert.NoError(t, handleFunctionProvisionPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.Arches[0])
	assert.Contains(t, output, testParams.Description)
	assert.Contains(t, output, testParams.Handle)
	assert.Contains(t, output, testParams.Name)
	assert.Contains(t, output, testParams.Oem)
	assert.Contains(t, output, testParams.Type)
	assert.Contains(t, output, testParams.Version)
	assert.Equal(t, fmt.Sprintf("%s/%s", testParams.Oem, testParams.Handle), testState.Internal.(*shared.Identity).Path)
}

func Test_handleFunctionProvisionPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &ProvisionParams{
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
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleFunctionProvisionPretend(testParams, &task.State{}))
}

func Test_handleFunctionProvisionPretend_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error:  expectedError,
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleFunctionProvisionPretend(&ProvisionParams{}, testState), expectedError)
}

func Test_handleFunctionPublishContext(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{
			Id:    "id",
			Hash:  "hash",
			Flags: FunctionFlags.Provisioning,
		},
		Logger: logrus.New(),
	}

	assert.NoError(t, handleFunctionPublishContext(&PublishParams{}, testState))
}

func Test_handleFunctionPublishContext_NoCurrent(t *testing.T) {
	assert.ErrorIs(t, handleFunctionPublishContext(&PublishParams{}, &task.State{}), errorNoRepoIdentity)
}

func Test_handleFunctionPublishContext_NoRepoIdentifiers(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{
			Flags: FunctionFlags.Provisioning,
		},
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleFunctionPublishContext(&PublishParams{}, testState), errorNoRepoIdentity)
}

func Test_handleFunctionPublishContext_NotProvisioning(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{},
		Logger:   logrus.New(),
	}

	assert.ErrorIs(t, handleFunctionPublishContext(&PublishParams{}, testState), errorNoRepoProvisioning)
}

func Test_handleFunctionPublishContext_ProvisioningDuplicate(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{
			Hash: "someHash",
		},
		Logger: logrus.New(),
	}

	assert.ErrorIs(t, handleFunctionPublishContext(&PublishParams{}, testState), errorDuplicatePublishing)
}

func Test_handleFunctionPublishComplete(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{
			Id:   "id",
			Hash: "hash",
		},
		Logger: logrus.New(),
		Output: "overwriteThis",
	}
	var testParams = &PublishParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var expectedFunction = &Function{
		Oem:     "expectedOem",
		Handle:  "expectedHandle",
		Version: "expectedVersion",
	}
	var restoredFactory = clientFactory.Get

	defer func() { clientFactory.Get = restoredFactory }()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			publishFunction: expectedFunction,
		}, nil
	}

	assert.NoError(t, handleFunctionPublishComplete(testParams, testState))
	assert.Equal(t, 1, len(testState.Reports))
	assert.Contains(t, testState.Reports[0], expectedFunction.Oem)
	assert.Contains(t, testState.Reports[0], expectedFunction.Handle)
	assert.Contains(t, testState.Reports[0], expectedFunction.Version)
	assert.Equal(t, expectedFunction, testState.Internal)
	assert.Empty(t, testState.Output)
}

func Test_handleFunctionPublishComplete_Error(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Internal: &shared.Identity{
			Id:   "id",
			Hash: "hash",
		},
		Logger: logrus.New(),
		Output: "overwriteThis",
	}
	var testParams = &PublishParams{
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
		return &stubFunctionClient{
			publishError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleFunctionPublishComplete(testParams, testState), expectedError)
	assert.Empty(t, testState.Internal)
	assert.Empty(t, testState.Output)
}

func Test_handleFunctionPublishComplete_NoCurrent(t *testing.T) {
	assert.ErrorIs(t, handleFunctionPublishComplete(&PublishParams{}, &task.State{}), errorNoProvision)
}

func Test_handleFunctionPublishComplete_NoRepoIdentifiers(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{},
	}

	assert.ErrorIs(t, handleFunctionPublishComplete(&PublishParams{}, testState), errorNoProvision)
}

func Test_handleFunctionPublishComplete_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Internal: &shared.Identity{
			Id:   "id",
			Hash: "hash",
		},
	}
	var testParams = &PublishParams{
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
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleFunctionPublishComplete(testParams, testState))
}

func Test_handleFunctionPublishIncomplete(t *testing.T) {
	var expectedError = errors.New("notExpected")
	var testParams = &PublishParams{SkipUnknown: false}
	var testState = &task.State{
		Error:    expectedError,
		Internal: &shared.Identity{Id: "id"},
	}

	assert.ErrorIs(t, handleFunctionPublishIncomplete(testParams, testState), expectedError)
	assert.True(t, testState.Completed)
	assert.NotEmpty(t, testState.Internal)
}

func Test_handleFunctionPublishIncomplete_Duplicate(t *testing.T) {
	var expectedPath = "path"
	var expectedFunction = &Function{Id: 37}
	var testParams = &PublishParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testState = &task.State{
		Error: errorDuplicatePublishing,
		Internal: &shared.Identity{
			Path: expectedPath,
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() { clientFactory.Get = restoredFactory }()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			getFunction: expectedFunction,
		}, nil
	}

	assert.NoError(t, handleFunctionPublishIncomplete(testParams, testState), ErrorNoSession)
	assert.True(t, testState.Abort)
	assert.Equal(t, 1, len(testState.Reports))
	assert.Contains(t, testState.Reports[0], expectedPath)
	assert.True(t, testState.Completed)
	assert.Equal(t, expectedFunction, testState.Internal)
	assert.Nil(t, testState.Error)
}

func Test_handleFunctionPublishIncomplete_Duplicate_NoSession(t *testing.T) {
	var expectedPath = "path"
	var testParams = &PublishParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testState = &task.State{
		Error: errorDuplicatePublishing,
		Internal: &shared.Identity{
			Path: expectedPath,
		},
	}

	assert.ErrorIs(t, handleFunctionPublishIncomplete(testParams, testState), ErrorNoSession)
	assert.True(t, testState.Abort)
	assert.Equal(t, 1, len(testState.Reports))
	assert.Contains(t, testState.Reports[0], expectedPath)
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Internal)
}

func Test_handleFunctionPublishIncomplete_NoSkip(t *testing.T) {
	var expectedError = errors.New("notExpected")
	var testParams = &PublishParams{SkipUnknown: true}
	var testState = &task.State{
		Error:    expectedError,
		Internal: &shared.Identity{},
	}

	assert.NoError(t, handleFunctionPublishIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Internal)
}

func Test_handleFunctionPublishIncomplete_SkipNoRepoIdentity(t *testing.T) {
	var testParams = &PublishParams{SkipUnknown: true}
	var testState = &task.State{
		Error:    errorNoRepoIdentity,
		Internal: &shared.Identity{},
		Logger:   logrus.New(),
	}

	assert.NoError(t, handleFunctionPublishIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Internal)
}

func Test_handleFunctionPublishIncomplete_SkipNoRepoProvisioning(t *testing.T) {
	var testParams = &PublishParams{SkipUnknown: true}
	var testState = &task.State{
		Error:    errorNoRepoProvisioning,
		Internal: &shared.Identity{},
		Logger:   logrus.New(),
	}

	assert.NoError(t, handleFunctionPublishIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Empty(t, testState.Internal)
}

func Test_handleFunctionPublishPretend(t *testing.T) {
	var expectedId = "testId"
	var testState = &task.State{
		Logger:   logrus.New(),
		Internal: &shared.Identity{Id: expectedId},
	}
	var testParams = &PublishParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Handle: "testHandle",
		Oem:    "testOem",
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
		return &stubFunctionClient{}, nil
	}

	assert.NoError(t, handleFunctionPublishPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedId)
}

func Test_handleFunctionPublishPretend_NoCurrent(t *testing.T) {
	assert.ErrorIs(t, handleFunctionPublishPretend(&PublishParams{}, &task.State{}), errorNoProvision)
}

func Test_handleFunctionPublishPretend_NoRepoIdentifiers(t *testing.T) {
	var testState = &task.State{
		Internal: &shared.Identity{},
	}

	assert.ErrorIs(t, handleFunctionPublishPretend(&PublishParams{}, testState), errorNoProvision)
}

func Test_handleFunctionPublishPretend_NoSession(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Internal: &shared.Identity{
			Id:   "id",
			Hash: "hash",
		},
	}
	var testParams = &PublishParams{
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
		return nil, expectedError
	}

	assert.ErrorIs(t, expectedError, handleFunctionPublishPretend(testParams, testState))
}
