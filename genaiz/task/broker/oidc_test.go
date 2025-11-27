package broker

import (
	"bufio"
	"bytes"
	"errors"
	"io"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	gio "genaiz.com/genaiz-lib/mock/io"
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

func TestOidcParams_GetInput(t *testing.T) {
	var testParams = &OidcParams{}
	var expectedIn = bufio.NewReader(&bytes.Buffer{})

	assert.Same(t, os.Stdin, testParams.GetInput())
	testParams.Input = expectedIn
	assert.Same(t, expectedIn, testParams.GetInput())
}

func TestOidcParams_GetOutput(t *testing.T) {
	var testParams = &OidcParams{}
	var expectedOut = bufio.NewWriter(&bytes.Buffer{})

	assert.Same(t, os.Stdout, testParams.GetOutput())
	testParams.Output = expectedOut
	assert.Same(t, expectedOut, testParams.GetOutput())
}

func TestOidcParams_GetUrlHandler(t *testing.T) {
	var testParams = &OidcParams{}

	assert.Equal(t, reflect.ValueOf(handleOidcCopyPasteUrl).Pointer(),
		reflect.ValueOf(testParams.GetUrlHandler()).Pointer())
	testParams.BrowserRedirect = true
	assert.Equal(t, reflect.ValueOf(handleOidcBrowserUrl).Pointer(),
		reflect.ValueOf(testParams.GetUrlHandler()).Pointer())
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
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".auth")
	var expectedToken = "sessionToken"
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
	var testParams = &OidcParams{
		Broker: &Broker{
			AuthFile: testFile,
			HostAddr: "host",
		},
	}
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth:       &DeviceAuth{DeviceCode: "code"},
			tokenCreateToken: "deviceToken",
			tokenSessionAuthSession: &AuthSession{
				Token: expectedToken,
			},
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return nil
	}

	assert.NoError(t, handleOidcCreate(testParams, testState))
	assert.Equal(t, expectedToken, testState.Output)
}

func Test_handleOidcCreate_AuthFileError(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, "notWriteable", ".auth")
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
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
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth:              &DeviceAuth{DeviceCode: "code"},
			tokenCreateToken:        "deviceToken",
			tokenSessionAuthSession: &AuthSession{},
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return nil
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

func Test_handleOidcCreate_TokenCreateError(t *testing.T) {
	var expectedError = errors.New("expectedError")
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
	var testParams = &OidcParams{
		Broker: &Broker{HostAddr: "host"},
		Output: &gio.StubWriter{WriteError: expectedError},
	}
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth:       &DeviceAuth{DeviceCode: "code"},
			tokenCreateError: expectedError,
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return nil
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedError)
}

func Test_handleOidcCreate_TokenSessionError(t *testing.T) {
	var expectedError = errors.New("expectedSessionError")
	var expectedToken = "token"
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
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
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth:        &DeviceAuth{DeviceCode: "code"},
			tokenCreateToken:  expectedToken,
			tokenSessionError: expectedError,
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return nil
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedError)
}

func Test_handleOidcCreate_UrlHandlerError(t *testing.T) {
	var expectedError = errors.New("expectedErrors")
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
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
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth: &DeviceAuth{},
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return expectedError
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedError)
}

func Test_handleOidcCreate_UrlPendingError(t *testing.T) {
	var expectedError = errors.New("expectedError")
	var restoredFactory = clientFactory.New
	var restoredUrlHandler = oidcDefaultUrlHandler
	var testParams = &OidcParams{
		Broker: &Broker{HostAddr: "host"},
		Output: &gio.StubWriter{WriteError: expectedError},
	}
	var testState = &task.State{
		Internal: oidcTracking{
			deviceUrl: "deviceUrl",
			tokenUrl:  "tokenUrl",
		},
		Logger: logrus.New(),
	}

	defer func() {
		clientFactory.New = restoredFactory
		oidcDefaultUrlHandler = restoredUrlHandler
	}()
	clientFactory.New = func(hostAddr string) Client {
		return &stubOidcClient{
			deviceAuth:       &DeviceAuth{DeviceCode: "code"},
			tokenCreateError: ErrorOidcPending,
		}
	}
	oidcDefaultUrlHandler = func(params *OidcParams, deviceAuth *DeviceAuth) error {
		return nil
	}

	assert.ErrorIs(t, handleOidcCreate(testParams, testState), expectedError)
}

func Test_handleOidcBrowserUrl(t *testing.T) {
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return nil
	})
	var expectedUrl = "expectedUrl"
	var expectedError = errors.New("expectedError")
	var testInput = bufio.NewReader(strings.NewReader("\n"))
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Input:  testInput,
		Output: testOutput,
	}
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}

	defer patch.Unpatch()
	assert.NoError(t, handleOidcBrowserUrl(testParams, testAuth), expectedError)
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}

func Test_handleOidcBrowserUrl_BrowserError(t *testing.T) {
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return errors.New("not supported")
	})
	var expectedUrl = "expectedUrl"
	var testInput = bufio.NewReader(strings.NewReader("\n"))
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Input:  testInput,
		Output: testOutput,
	}
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}

	defer patch.Unpatch()
	assert.NoError(t, handleOidcBrowserUrl(testParams, testAuth))
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}

func Test_handleOidcBrowserUrl_FprintfError(t *testing.T) {
	var patch = mock.Patches{T: t}.BrowserOpenUrl(func(string, io.Writer, io.Writer) error {
		return nil
	})
	var expectedUrl = "expectedUrl"
	var expectedError = errors.New("expectedError")
	var testInput = bufio.NewReader(strings.NewReader("\n"))
	var testParams = &OidcParams{
		Input: testInput,
		Output: &gio.StubWriter{
			WriteError: expectedError,
		},
	}
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}

	defer patch.Unpatch()
	assert.ErrorIs(t, handleOidcBrowserUrl(testParams, testAuth), expectedError)
	assert.Equal(t, expectedUrl, cast.ToStringSlice(patch.CalledWith)[0])
}

func Test_handleOidcPendingUrl(t *testing.T) {
	var expectedUrl = "expectedUrl"
	var testInput = bufio.NewReader(strings.NewReader("NOT!\n"))
	var testOutput = new(bytes.Buffer)
	var testParams = &OidcParams{
		Input:  testInput,
		Output: testOutput,
	}
	var testAuth = &DeviceAuth{
		VerificationUriComplete: expectedUrl,
	}

	assert.NoError(t, handleOidcPendingUrl(testParams, testAuth))
	output := testOutput.String()
	assert.NotEmpty(t, output)
	assert.Contains(t, output, expectedUrl)
}
