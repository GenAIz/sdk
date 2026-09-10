package locker

import (
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/chacha20poly1305"

	"genaiz.com/genaiz-lib/lang/filez"
	gio "genaiz.com/genaiz-lib/mock/io"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubClient struct {
	decoratedClient broker.Client
}

func (s stubClient) CreateDataSource(*broker.DataLinkInstance, map[string]string) (*broker.DataLinkInstance, error) {
	panic("unimplemented")
}

func (s stubClient) CreateDataSourceUrl() string {
	return s.decoratedClient.CreateDataSourceUrl()
}

func (s stubClient) CreateWorkspace(*broker.Workspace) (*broker.Workspace, error) {
	panic("unimplemented")
}

func (s stubClient) CreateWorkspaceUrl() string {
	panic("unimplemented")
}

func (s stubClient) CreateWorkspaceFlow(int64, int64, string, string) (*broker.WorkspaceFlow, error) {
	panic("unimplemented")
}

func (s stubClient) CreateWorkspaceFlowUrl() string {
	panic("unimplemented")
}

func (s stubClient) ExportDataLink(string, string, string, string) (*broker.DataLink, error) {
	panic("unimplemented")
}

func (s stubClient) FindDataLink(string, string, string) (*broker.DataLink, error) {
	panic("unimplemented")
}

func (s stubClient) FindSolution(string, string, string) (*broker.Solution, error) {
	panic("unimplemented")
}

func (s stubClient) FindSolutionUrl() string {
	panic("unimplemented")
}

func (s stubClient) GetAuthToken() string {
	return s.decoratedClient.GetAuthToken()
}

func (s stubClient) GetExpiry() int {
	panic("unimplemented")
}

func (s stubClient) GetFunction(int64) (*broker.Function, error) {
	panic("unimplemented")
}

func (s stubClient) GetFunctionUrl() string {
	panic("unimplemented")
}

func (s stubClient) GetHostAddr() string {
	return s.decoratedClient.GetHostAddr()
}

func (s stubClient) GetSolution(int64) (*broker.Solution, error) {
	panic("unimplemented")
}

func (s stubClient) GetSolutionUrl() string {
	panic("unimplemented")
}

func (s stubClient) GetTimeout() int {
	panic("unimplemented")
}

func (s stubClient) GetUserId() int {
	panic("unimplemented")
}

func (s stubClient) ListDataLinks(string, string, int) ([]broker.DataLink, error) {
	panic("unimplemented")
}

func (s stubClient) ListDataLinksUrl() string {
	panic("unimplemented")
}

func (s stubClient) ListDataSources() ([]broker.DataLinkInstance, error) {
	panic("unimplemented")
}

func (s stubClient) ListDataSourcesUrl() string {
	return s.decoratedClient.ListDataSourcesUrl()
}

func (s stubClient) ListSolutions(string) ([]broker.Solution, error) {
	panic("unimplemented")
}

func (s stubClient) ListSolutionsUrl() string {
	panic("unimplemented")
}

func (s stubClient) ListWorkspaceFlows(int64, int, int) ([]broker.WorkspaceFlow, error) {
	panic("unimplemented")
}

func (s stubClient) ListWorkspaceFlowsUrl() string {
	panic("unimplemented")
}

func (s stubClient) ListWorkspaceNodes(int64) ([]broker.WorkspaceNode, error) {
	panic("unimplemented")
}

func (s stubClient) ListWorkspaceNodesUrl() string {
	panic("unimplemented")
}

func (s stubClient) ListWorkspaces(int, int) ([]broker.Workspace, error) {
	panic("unimplemented")
}

func (s stubClient) ListWorkspacesUrl() string {
	panic("unimplemented")
}

func (s stubClient) Login(string, *memguard.Enclave) (*broker.AuthSession, error) {
	panic("unimplemented")
}

func (s stubClient) LoginUrl() string {
	panic("unimplemented")
}

func (s stubClient) Logout(string) error {
	panic("unimplemented")
}

func (s stubClient) LogoutUrl() string {
	panic("unimplemented")
}

func (s stubClient) OidcDeviceCode(string, *broker.DeviceClient) (*broker.DeviceAuth, error) {
	panic("unimplemented")
}

func (s stubClient) OidcDeviceUrl() (string, error) {
	panic("unimplemented")
}

func (s stubClient) OidcTokenCreate(string, string, *broker.DeviceClient) (string, error) {
	panic("unimplemented")
}

func (s stubClient) OidcTokenSession(string, string) (*broker.AuthSession, error) {
	panic("unimplemented")
}

func (s stubClient) OidcTokenUrl() (string, error) {
	panic("unimplemented")
}

func (s stubClient) ProvisionFunction(*broker.Function, map[string]any) (*shared.Identity, error) {
	panic("unimplemented")
}

func (s stubClient) ProvisionFunctionUrl() string {
	panic("unimplemented")
}

func (s stubClient) PublishDataLink(*broker.DataLink) (*broker.DataLink, error) {
	panic("unimplemented")
}

func (s stubClient) PublishFunction(*shared.Identity) (*broker.Function, error) {
	panic("unimplemented")
}

func (s stubClient) PublishFunctionUrl() string {
	panic("unimplemented")
}

func (s stubClient) PublishSolution(*broker.Solution) (*broker.Solution, error) {
	panic("unimplemented")
}

func (s stubClient) PublishSolutionUrl() string {
	panic("unimplemented")
}

func (s stubClient) Session() (*broker.Session, error) {
	panic("unimplemented")
}

func (s stubClient) SessionUrl() string {
	panic("unimplemented")
}

func (s stubClient) SessionValid(*broker.Session) bool {
	panic("unimplemented")
}

func (s stubClient) UpdateDataSource(*broker.DataLinkInstance, map[string]string) (*broker.DataLinkInstance, error) {
	panic("unimplemented")
}

func (s stubClient) UpdateDataSourceUrl() string {
	return s.decoratedClient.UpdateDataSourceUrl()
}

func (s stubClient) WithAccount(*broker.AuthAccount) (broker.Client, error) {
	panic("unimplemented")
}

type stubEnclave struct {
	openBuffer *memguard.LockedBuffer
	openError  error
	size       int
}

func (se stubEnclave) Open() (*memguard.LockedBuffer, error) {
	return se.openBuffer, se.openError
}

func (se stubEnclave) Size() int {
	return se.size
}

func Test_newEmptyEnclave(t *testing.T) {
	var actual = newEmptyEnclave()

	assert.Equal(t, 0, actual.Size())
	lb, err := actual.Open()
	assert.NoError(t, err)
	assert.Empty(t, lb.Bytes())
}

func TestLockerAccount_withSource(t *testing.T) {
	var expectedSource = &lockerLink{
		LockerHandle: "replaced",
		LinkOem:      "replacedOem",
		LinkHandle:   "replacedHandle",
		LinkVersion:  "replacedVersion",
		Properties:   "replacedProperties",
	}
	var testAccount = &lockerAccount{
		DataSources: []lockerLink{
			{
				LockerHandle: "keep",
			},
			{
				LockerHandle: "replaced",
				LinkOem:      "oldOem",
				LinkHandle:   "oldHandle",
				LinkVersion:  "oldVersion",
				Properties:   "oldProperties",
			},
		},
	}

	actual := testAccount.withSource(expectedSource)
	assert.NotNil(t, actual)
	assert.Contains(t, actual.DataSources, *expectedSource)
}

func TestLockerBody_withAccount(t *testing.T) {
	var expectedAccount = &lockerAccount{
		AccountUrl: "replaced",
		DataSources: []lockerLink{
			{
				LockerHandle: "replacedSource",
			},
		},
	}
	var testBody = &lockerBody{
		Accounts: []lockerAccount{
			{
				AccountUrl: "keep",
			},
			{
				AccountUrl: "replaced",
			},
		},
	}

	actualBody := testBody.withAccount(expectedAccount)
	assert.NotNil(t, actualBody)
	actual, err := actualBody.findAccount(expectedAccount.AccountUrl)
	assert.Equal(t, expectedAccount, actual)
	assert.NoError(t, err)
}

func TestLockerHeader_Decrypt_OpenFail(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = lockerHeader{}
	var testEnclave = &stubEnclave{
		openError: expectedError,
	}

	actual, err := testHeader.Decrypt([]byte{}, testEnclave)
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, actual)
}

func TestLockerHeader_Encrypt_OpenFail(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = lockerHeader{}
	var testEnclave = &stubEnclave{
		openError: expectedError,
	}

	actual, err := testHeader.Encrypt(memguard.NewEnclave([]byte("hello")), testEnclave)
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, actual)
}

func TestLockerLink_encodeProperties(t *testing.T) {
	var expectedKey = "expectedKey"
	var expectedValue = "expectedValue"
	var testLink = &lockerLink{
		LockerHandle: "testHandle",
		LinkOem:      "testOem",
		LinkHandle:   "testHandle",
		LinkVersion:  "testVersion",
	}
	var testProperties = map[string]string{
		expectedKey: expectedValue,
	}
	var testPassphrase = memguard.NewEnclave([]byte("test"))
	var err error

	if testLink.Properties, err = testLink.encodeProperties(testProperties, testPassphrase); err == nil {
		var actual map[string]string

		assert.NotEmpty(t, testLink.Properties)
		assert.NoError(t, err)

		if actual, err = testLink.decodeProperties(testPassphrase); err == nil {
			assert.Equal(t, expectedValue, actual[expectedKey])
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestLockerLink_encodeProperties_Empty(t *testing.T) {
	var testLink = &lockerLink{
		LockerHandle: "testHandle",
		LinkOem:      "testOem",
		LinkHandle:   "testHandle",
		LinkVersion:  "testVersion",
	}
	var testProperties = map[string]string{}
	var testPassphrase = memguard.NewEnclave([]byte("test"))
	var err error

	if testLink.Properties, err = testLink.encodeProperties(testProperties, testPassphrase); err == nil {
		var actual map[string]string

		assert.Empty(t, testLink.Properties)
		assert.NoError(t, err)

		if actual, err = testLink.decodeProperties(testPassphrase); err == nil {
			assert.Empty(t, actual)
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestLockerLink_encodeProperties_EncryptError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLink = &lockerLink{
		LockerHandle: "testHandle",
		LinkOem:      "testOem",
		LinkHandle:   "testHandle",
		LinkVersion:  "testVersion",
	}
	var testProperties = map[string]string{
		"akey": "aValue",
	}
	var testPassphrase = stubEnclave{
		openError: expectedError,
	}
	var err error

	testLink.Properties, err = testLink.encodeProperties(testProperties, testPassphrase)
	assert.Empty(t, testLink.Properties)
	assert.ErrorIs(t, err, expectedError)
}

func TestLockerLink_refreshSources(t *testing.T) {
	var testOldPass = memguard.NewEnclave([]byte("old"))
	var testPass = memguard.NewEnclave([]byte("new"))
	var testLink = &lockerLink{}
	var testProps = map[string]string{"key": "value"}
	var err error

	if testLink.Properties, err = testLink.encodeProperties(testProps, testOldPass); err == nil {
		var testAccount = &lockerAccount{
			DataSources: []lockerLink{
				*testLink,
			},
		}
		var links []lockerLink

		links, err = testAccount.refreshSources(testOldPass, testPass)
		assert.NoError(t, err)
		assert.Equal(t, 1, len(links))
		return
	}

	assert.Fail(t, err.Error())
}

func TestLockerLink_refreshSources_decodeError(t *testing.T) {
	var testAccount = &lockerAccount{
		DataSources: []lockerLink{
			{
				Properties: "$",
			},
		},
	}

	actual, err := testAccount.refreshSources(nil, nil)
	assert.ErrorAs(t, err, new(base64.CorruptInputError))
	assert.Empty(t, actual)
}

func TestLockerLink_refreshSources_encodeError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testOldPass = memguard.NewEnclave([]byte("old"))
	var testPass = &stubEnclave{openError: expectedError}
	var testLink = &lockerLink{}
	var testProps = map[string]string{"key": "value"}
	var err error

	if testLink.Properties, err = testLink.encodeProperties(testProps, testOldPass); err == nil {
		var testAccount = &lockerAccount{
			DataSources: []lockerLink{
				*testLink,
			},
		}
		var links []lockerLink

		links, err = testAccount.refreshSources(testOldPass, testPass)
		assert.ErrorIs(t, err, expectedError)
		assert.Empty(t, links)
		return
	}

	assert.Fail(t, err.Error())
}

func TestNewSecuredLockerState(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("expected"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.True(t, actual.IsOpened())
	assert.Equal(t, actual.current, expectedBuffer)
	assert.Equal(t, actual.currentPath, expectedPath)
}

func TestNewSecuredLockerState_Close(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.NoError(t, actual.Close(expectedPath))
}

func TestNewSecuredLockerState_Destroy(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.NotEqual(t, 0, actual.current.Size())
	actual.Destroy()
	assert.Equal(t, 0, actual.current.Size())
}

func TestNewSecuredLockerState_isOpen(t *testing.T) {
	var actual = NewSecuredLockerState(&task.State{})

	assert.False(t, actual.IsOpened())
}

func TestNewSecuredLockerState_unfold(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var testLocker = NewSecuredLockerState(testState)

	actual, err := testLocker.unfold()
	assert.NoError(t, err)
	assert.Empty(t, actual.Accounts)
}

func TestNewSecuredLockerState_unfold_JsonError(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("expected"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var testLocker = NewSecuredLockerState(testState)

	actual, err := testLocker.unfold()
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestSecuredLockerTracking_Update(t *testing.T) {
	var testDir = t.TempDir()
	var testLockerPath = filepath.Join(testDir, "locker.bin")
	var testLink = &lockerLink{
		LockerHandle: "testHandle",
		LinkOem:      "expectedOem",
		LinkHandle:   "expectedHandle",
		LinkVersion:  "expectedVersion",
	}
	var testUrl = "testUrl"
	var testPassPhrase Enclave
	var err error

	if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, nil); err == nil {
		var testState = NewSecuredLockerState(&task.State{})

		if err = testState.Read(testLockerPath, testPassPhrase); err == nil {
			var newPassPhrase = memguard.NewEnclave([]byte("passphrase"))

			assert.NoError(t, testState.Update(newPassPhrase, testPassPhrase))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestSecuredLockerTracking_Update_NoAccounts(t *testing.T) {
	var testDir = t.TempDir()
	var testLockerPath = filepath.Join(testDir, "locker.bin")
	var testUrl = "testUrl"
	var testPassPhrase Enclave
	var err error

	if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, nil, nil); err == nil {
		var testState = NewSecuredLockerState(&task.State{})

		if err = testState.Read(testLockerPath, testPassPhrase); err == nil {
			var newPassPhrase = memguard.NewEnclave([]byte("wut"))

			assert.NoError(t, testState.Update(newPassPhrase, testPassPhrase))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestSecuredLockerTracking_Update_RefreshError(t *testing.T) {
	var testDir = t.TempDir()
	var testLockerPath = filepath.Join(testDir, "locker.bin")
	var testUrl = "testUrl"
	var testLink = &lockerLink{
		LockerHandle: "testHandle",
		LinkOem:      "expectedOem",
		LinkHandle:   "expectedHandle",
		LinkVersion:  "expectedVersion",
	}
	var testProps = map[string]string{
		"key": "value",
	}
	var testPassPhrase Enclave
	var err error

	if testPassPhrase, err = writeTestLocker(testLockerPath, testUrl, testLink, testProps); err == nil {
		var testState = NewSecuredLockerState(&task.State{})

		if err = testState.Read(testLockerPath, testPassPhrase); err == nil {
			var newPassPhrase = memguard.NewEnclave([]byte("wut"))

			// inverted params
			assert.Error(t, testState.Update(testPassPhrase, newPassPhrase))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestSecuredLockerTracking_Update_UnfoldError(t *testing.T) {
	var testDir = t.TempDir()
	var testLockerPath = filepath.Join(testDir, "locker.bin")
	var testPassPhrase Enclave
	var err error

	if testPassPhrase, err = writeTestJsonError(testLockerPath); err == nil {
		var testState = NewSecuredLockerState(&task.State{})

		if err = testState.Read(testLockerPath, testPassPhrase); err == nil {
			var newPassPhrase = memguard.NewEnclave([]byte("passphrase"))

			assert.Error(t, testState.Update(newPassPhrase, testPassPhrase))
			return
		}
	}

	assert.Fail(t, err.Error())
}

func Test_readLockerData_ErrorCopyEmpty(t *testing.T) {
	var testNonce = make([]byte, chacha20poly1305.NonceSizeX)
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	if _, err = rand.Read(testNonce); err == nil {
		testBytes = append(testBytes, []byte{37}...)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
		testBytes = append(testBytes, []byte{3}...)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
		testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
		testBytes = append(testBytes, testSalt...)
		testBytes = append(testBytes, testNonce...)
		testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
		assert.Nil(t, testHeader)
		assert.Empty(t, actualBytes)
		assert.ErrorIs(t, err, errorLockerContentEmpty)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_readLockerData_ErrorIterations(t *testing.T) {
	var testReader = strings.NewReader(string([]byte{27}))
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testHeader, actualBytes, err = readLockerData(testReader)
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorKeyLength(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorMemoryKb(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorNonce(t *testing.T) {
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
	testBytes = append(testBytes, testSalt...)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorSalt(t *testing.T) {
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorSaltLength(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorThreads(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorVersion(t *testing.T) {
	var testReader = strings.NewReader(string([]byte{}))
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testHeader, actualBytes, err = readLockerData(testReader)
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_writeLockerData_ErrorIterations(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		iterations: 2,
	}
	var testWriter = &gio.StubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.iterations)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorKeyLength(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		keyLength: 37,
	}
	var testWriter = &gio.StubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.keyLength)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorMemoryKb(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		memoryKb: 1024,
	}
	var testWriter = &gio.StubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.memoryKb)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorNonce(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		nonce: []byte("N once init"),
	}
	var testWriter = &gio.StubWriter{
		MatchError: testHeader.nonce,
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorSaltLength(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		salt: []byte("salt is good"),
	}
	var testWriter = &gio.StubWriter{
		MatchError: make([]byte, 2),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint16(testWriter.MatchError, uint16(len(testHeader.salt)))
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorSalt(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		salt: []byte("salt is good"),
	}
	var testWriter = &gio.StubWriter{
		MatchError: testHeader.salt,
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorThreads(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		threads: 37,
	}
	var testWriter = &gio.StubWriter{
		MatchError: []byte{testHeader.threads},
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorVersion(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		version: 37,
	}
	var testWriter = &gio.StubWriter{
		MatchError: []byte{testHeader.version},
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func writeTestJsonError(lockerPath string) (Enclave, error) {
	var fd *os.File
	var err error

	if fd, err = os.OpenFile(lockerPath, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0660); err == nil {
		defer filez.CloseSilently(fd)
		var header = newLockerHeader()
		var bodyEnclave = memguard.NewEnclave([]byte("- not json ever"))
		var passphrase = memguard.NewEnclave([]byte("test"))
		var encrypted []byte

		if encrypted, err = header.Encrypt(bodyEnclave, passphrase); err == nil {
			if err = writeLockerData(header, fd, encrypted); err == nil {
				return passphrase, nil
			}
		}
	}

	return nil, err
}

func writeTestLocker(lockerPath, testUrl string, link *lockerLink, properties map[string]string) (Enclave, error) {
	var lockerState = NewSecuredLockerState(&task.State{})
	var testPassphrase = memguard.NewEnclave([]byte("test"))
	var err error

	if testUrl != "" && link != nil {
		if err = lockerState.addSource(testUrl, link); err == nil {
			if len(properties) > 0 {
				for k, v := range properties {
					var valueEnclave = memguard.NewEnclave([]byte(v))

					if err = lockerState.updateSource(testUrl, link.LockerHandle, k, valueEnclave, testPassphrase); err != nil {
						break
					}
				}
			}
		}
	}

	if err == nil {
		if err = lockerState.Write(lockerPath, testPassphrase); err == nil {
			return testPassphrase, nil
		}
	}

	return nil, err
}
