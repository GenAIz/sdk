package broker

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	mio "genaiz.com/genaiz-lib/mock/io"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
)

type stubOidcClient struct {
	client

	deviceAuth              *DeviceAuth
	deviceAuthClient        *DeviceClient
	deviceAuthError         error
	deviceAuthUrl           string
	deviceUrl               string
	deviceUrlError          error
	tokenCreateDeviceCode   string
	tokenCreateDeviceClient *DeviceClient
	tokenCreateError        error
	tokenCreateToken        string
	tokenCreateUrl          string
	tokenSessionAuthSession *AuthSession
	tokenSessionAuthToken   string
	tokenSessionError       error
	tokenSessionUrl         string
	tokenUrl                string
	tokenUrlError           error
}

func (soc *stubOidcClient) OidcDeviceCode(deviceAuthUrl string, deviceAuthClient *DeviceClient) (*DeviceAuth, error) {
	soc.deviceAuthUrl = deviceAuthUrl
	soc.deviceAuthClient = deviceAuthClient
	return soc.deviceAuth, soc.deviceAuthError
}

func (soc *stubOidcClient) OidcDeviceUrl() (string, error) {
	return soc.deviceUrl, soc.deviceUrlError
}

func (soc *stubOidcClient) OidcTokenCreate(tokenCreateUrl, tokenCreateDeviceCode string, tokenCreateDeviceClient *DeviceClient) (string, error) {
	soc.tokenCreateUrl = tokenCreateUrl
	soc.tokenCreateDeviceCode = tokenCreateDeviceCode
	soc.tokenCreateDeviceClient = tokenCreateDeviceClient
	return soc.tokenCreateToken, soc.tokenCreateError
}

func (soc *stubOidcClient) OidcTokenSession(tokenSessionUrl, tokenSessionAuthToken string) (*AuthSession, error) {
	soc.tokenSessionUrl = tokenSessionUrl
	soc.tokenSessionAuthToken = tokenSessionAuthToken
	return soc.tokenSessionAuthSession, soc.tokenSessionError
}

func (soc *stubOidcClient) OidcTokenUrl() (string, error) {
	return soc.tokenUrl, soc.tokenUrlError
}

func TestOidcParams_GetInput(t *testing.T) {
	var testParams = &OidcParams{}
	var expectedIn = io.NopCloser(bufio.NewReader(&bytes.Buffer{}))

	assert.Same(t, os.Stdin, testParams.GetInput())
	testParams.Input = expectedIn
	assert.Equal(t, expectedIn, testParams.GetInput())
}

func TestOidcParams_GetOutput(t *testing.T) {
	var testParams = &OidcParams{}
	var expectedOut = bufio.NewWriter(&bytes.Buffer{})

	assert.Same(t, os.Stdout, testParams.GetOutput())
	testParams.Output = expectedOut
	assert.Same(t, expectedOut, testParams.GetOutput())
}

func TestOidcState_IsSupported(t *testing.T) {
	var testState = &task.State{}
	var testOidcState = NewOidcState(testState)

	assert.False(t, testOidcState.IsSupported())
	testOidcState.SetDeviceUrl("url")
	assert.False(t, testOidcState.IsSupported())
	testOidcState.SetTokenUrl("url")
	assert.True(t, testOidcState.IsSupported())
}

func TestNewOidcState(t *testing.T) {
	var expectedDeviceUrl = "expectedDeviceUrl"
	var expectedTokenUrl = "expectedTokenUrl"
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: expectedDeviceUrl,
			tokenUrl:  expectedTokenUrl,
		},
	}
	var testOidcState = NewOidcState(testState)

	assert.Equal(t, expectedDeviceUrl, testOidcState.deviceUrl)
	assert.Equal(t, expectedTokenUrl, testOidcState.tokenUrl)
	assert.Same(t, testState, testOidcState.state)
}

func TestNewOidcState_EmptyState(t *testing.T) {
	var testState = &task.State{}
	var testOidcState = NewOidcState(testState)

	assert.Empty(t, testOidcState.deviceUrl)
	assert.Empty(t, testOidcState.tokenUrl)
	assert.Same(t, testState, testOidcState.state)
}

func TestNewOidcTask(t *testing.T) {
	var testTask = NewOidcTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.Empty(t, testTask.OnPretend)
}

func TestOidcUrlHandler_Init(t *testing.T) {
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return nil
	})
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		BrowserRedirect: true,
		Output:          testOutput,
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)

	defer patch.Unpatch()
	assert.NoError(t, testHandler.Init())
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}

func TestOidcUrlHandler_Init_BrowserPrintError(t *testing.T) {
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return nil
	})
	var expectedError = errors.New("expected")
	var testParams = &OidcParams{
		BrowserRedirect: true,
		Output:          &mio.StubWriter{WriteError: expectedError},
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)

	defer patch.Unpatch()
	assert.ErrorIs(t, testHandler.Init(), expectedError)
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
}

func TestOidcUrlHandler_Init_BrowserUnsupported(t *testing.T) {
	var expectedError = errors.New("expected")
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return expectedError
	})
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		BrowserRedirect: true,
		Output:          testOutput,
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)

	defer patch.Unpatch()
	assert.NoError(t, testHandler.Init())
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}

func TestOidcUrlHandler_Init_CopyPaste(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		BrowserRedirect: false,
		Output:          testOutput,
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)

	assert.NoError(t, testHandler.Init())
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}

func TestOidcUrlHandler_PollForToken_OutError(t *testing.T) {
	var expectedErrors = errors.New("expectedError")
	var testParams = &OidcParams{
		BrowserRedirect: false,
		Input:           io.NopCloser(strings.NewReader("\n")),
		Output: &mio.StubWriter{
			WriteError: expectedErrors,
		},
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)

	actual, err := testHandler.PollForToken(nil, "token", nil)
	assert.ErrorIs(t, err, expectedErrors)
	assert.Empty(t, actual)
}

func TestOidcUrlHandler_PollForToken_PendingError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		BrowserRedirect: false,
		Input:           io.NopCloser(strings.NewReader("\n")),
		Output:          testOutput,
		Timeout:         lang.Ref(0 * time.Millisecond),
		Tries:           lang.Ref(1),
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)
	var testClient = &stubOidcClient{
		tokenCreateError: ErrorOidcPending,
	}

	actual, err := testHandler.PollForToken(testClient, "token", nil)
	assert.ErrorIs(t, err, ErrorOidcClientTimeout)
	assert.Empty(t, actual)
}

func TestOidcUrlHandler_PollForToken_SlowDownError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		BrowserRedirect: false,
		Input:           io.NopCloser(strings.NewReader("\n")),
		Output:          testOutput,
		Timeout:         lang.Ref(0 * time.Millisecond),
		Tries:           lang.Ref(1),
	}
	var expectedUrl = "expectedUrl"
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}
	var testHandler = NewOidcUrlHandler(testParams, testAuth)
	var testClient = &stubOidcClient{
		tokenCreateError: ErrorOidcSlowDown,
	}

	actual, err := testHandler.PollForToken(testClient, "token", nil)
	assert.ErrorIs(t, err, ErrorOidcClientTimeout)
	assert.Empty(t, actual)
}

func Test_handleOidcContext(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".auth")
	var expectedDeviceUrl = "deviceUrl"
	var expectedTokenUrl = "tokenUrl"
	var restoredFactory = clientFactory.New
	var testParams = &OidcParams{Broker: &Broker{AuthFile: testFile}}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceUrl: expectedDeviceUrl,
			tokenUrl:  expectedTokenUrl,
		}
	}

	assert.NoError(t, handleOidcContext(testParams, testState))
	actual := NewOidcState(testState)
	assert.Equal(t, expectedDeviceUrl, actual.deviceUrl)
	assert.Equal(t, expectedTokenUrl, actual.tokenUrl)
}

func Test_handleOidcContext_DeviceUrlError(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".auth")
	var restoredFactory = clientFactory.New
	var testParams = &OidcParams{Broker: &Broker{AuthFile: testFile}}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceUrlError: errors.New("urlError"),
		}
	}

	assert.ErrorIs(t, handleOidcContext(testParams, testState), ErrorOidcNotSupported)
}

func Test_handleOidcContext_LoginError(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, "notWriteable", "notExisting")
	var testParams = &OidcParams{Broker: &Broker{AuthFile: testFile}}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	if err := os.MkdirAll(filepath.Dir(testFile), 0222); err == nil {
		assert.Error(t, handleOidcContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleOidcContext_TokenUrlError(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".auth")
	var restoredFactory = clientFactory.New
	var testParams = &OidcParams{Broker: &Broker{AuthFile: testFile}}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceUrl:     "url",
			tokenUrlError: errors.New("urlError"),
		}
	}

	assert.ErrorIs(t, handleOidcContext(testParams, testState), ErrorOidcNotSupported)
}

func Test_handleOidcCreate(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), ".auth")
	var expectedSessionToken = "sessionToken"
	var restoredFactory = clientFactory.New
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Broker: &Broker{
			AuthFile: testFile,
			HostAddr: "host",
		},
		Output:  testOutput,
		Timeout: lang.Ref(0 * time.Millisecond),
		Tries:   lang.Ref(1),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{
				VerificationUriComplete: "uri",
			},
			tokenCreateToken: "token",
			tokenSessionAuthSession: &AuthSession{
				Token: expectedSessionToken,
			},
		}
	}

	assert.NoError(t, handleOidcCreate(testParams, testState))
	assert.Equal(t, expectedSessionToken, testState.Output)
}

func Test_handleOidcCreate_AuthWriteError(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "notWriteable", ".auth")
	var restoredFactory = clientFactory.New
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Broker:  &Broker{HostAddr: "host"},
		Output:  testOutput,
		Timeout: lang.Ref(0 * time.Millisecond),
		Tries:   lang.Ref(1),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{
				VerificationUriComplete: "uri",
			},
			tokenCreateToken: "token",
			tokenSessionAuthSession: &AuthSession{
				Token: "sessionToken",
			},
		}
	}

	if err := os.MkdirAll(filepath.Dir(testFile), 0222); err == nil {
		testParams.AuthFile = testFile
		assert.Error(t, handleOidcCreate(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleOidcCreate_DeviceCodeError(t *testing.T) {
	var expectedError = errors.New("expectedErrors")
	var restoredFactory = clientFactory.New
	var testParams = &OidcParams{Broker: &Broker{HostAddr: "host"}}
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuthError: expectedError,
		}
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedError)
}

func Test_handleOidcCreate_StateNotSupported(t *testing.T) {
	var testParams = &OidcParams{}
	var testState = &task.State{}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), ErrorOidcNotSupported)
}

func Test_handleOidcCreate_PollTokenError(t *testing.T) {
	var expectedErr = errors.New("expected")
	var restoredFactory = clientFactory.New
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Broker:  &Broker{HostAddr: "host"},
		Output:  testOutput,
		Timeout: lang.Ref(0 * time.Millisecond),
		Tries:   lang.Ref(1),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{
				VerificationUriComplete: "uri",
			},
			tokenCreateError: expectedErr,
		}
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedErr)
}

func Test_handleOidcCreate_TokenSessionError(t *testing.T) {
	var expectedErr = errors.New("expected")
	var restoredFactory = clientFactory.New
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Broker:  &Broker{HostAddr: "host"},
		Output:  testOutput,
		Timeout: lang.Ref(0 * time.Millisecond),
		Tries:   lang.Ref(1),
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{
				VerificationUriComplete: "uri",
			},
			tokenCreateToken:  "token",
			tokenSessionError: expectedErr,
		}
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedErr)
}

func Test_handleOidcCreate_UrlInitError(t *testing.T) {
	var expectedErr = errors.New("expected")
	var restoredFactory = clientFactory.New
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}
	var testParams = &OidcParams{
		Broker: &Broker{HostAddr: "host"},
		Output: &mio.StubWriter{
			WriteError: expectedErr,
		},
	}

	defer func() {
		clientFactory.New = restoredFactory
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{
				VerificationUriComplete: "uri",
			},
		}
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedErr)
}
