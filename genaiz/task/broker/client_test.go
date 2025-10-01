package broker

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"
	"resty.dev/v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/version/env"
)

type emptyClient struct {
	client
}

type stubBridge struct {
	jsonRequest bool
	cookie      *http.Cookie
	err         error
	params      map[string]string
	result      any
	response    responseBridge
	timeout     time.Duration
	url         string
}

func (s *stubBridge) Close() error {
	return nil
}

func (s *stubBridge) Cookie(cookie *http.Cookie) requestBridge {
	s.cookie = cookie
	return s
}

func (s *stubBridge) Get(url string) (responseBridge, error) {
	s.url = url
	return s.response, s.err
}

func (s *stubBridge) Json() requestBridge {
	s.jsonRequest = true
	return s
}

func (s *stubBridge) Params(params map[string]string) requestBridge {
	s.params = params
	return s
}

func (s *stubBridge) Resulting(result any) requestBridge {
	s.result = result
	return s
}

func (s *stubBridge) Post(url string) (responseBridge, error) {
	s.url = url
	return s.response, s.err
}

func (s *stubBridge) Timeout() time.Duration {
	return s.timeout
}

type stubResponse struct {
	cookies    []*http.Cookie
	result     any
	status     string
	statusCode int
	success    bool
}

func (s stubResponse) Bytes() []byte {
	return nil
}

func (s stubResponse) Cookies() []*http.Cookie {
	return s.cookies
}

func (s stubResponse) IsSuccess() bool {
	return s.success
}

func (s stubResponse) Result() any {
	return s.result
}

func (s stubResponse) Status() string {
	return s.status
}

func (s stubResponse) StatusCode() int {
	return s.statusCode
}

func TestClient_Login(t *testing.T) {
	var expectedUser = "user"
	var expectedPassword = "password"
	var expectedToken = "token"
	var expectedSessionId = int64(37)
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			cookies: []*http.Cookie{{Name: "s", Value: expectedToken}},
			result: &clientPayload[Session]{
				Data: Session{
					Id: expectedSessionId,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge)
	var actual, err = testClient.Login(expectedUser, []byte(expectedPassword))

	assert.NoError(t, err)
	assert.Equal(t, expectedSessionId, actual.SessionId)
	assert.Equal(t, expectedToken, actual.Token)
	assert.Equal(t, expectedUser, actual.Username)
}

func TestClient_Login_InvalidUrl(t *testing.T) {
	var testClient = &client{}
	var actual, err = testClient.Login("", []byte(""))

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_Login_NoCookie(t *testing.T) {
	var testBridge = &stubBridge{
		response: newTestResponse(200, ""),
	}
	var testClient = newTestClient(testBridge)
	var actual, err = testClient.Login("", []byte(""))

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoToken)
}

func TestClient_Login_NoResponse(t *testing.T) {
	var expectedErr = errors.New("expected")
	var testBridge = &stubBridge{
		err: expectedErr,
	}
	var testClient = newTestClient(testBridge)
	var actual, err = testClient.Login("", []byte(""))

	assert.Empty(t, actual)
	assert.ErrorIs(t, err, expectedErr)
}

func TestClient_Login_NoResponseCustomStatus(t *testing.T) {
	var expectedStatus = "status"
	var testBridge = &stubBridge{
		response: newTestResponse(0, expectedStatus),
	}
	var testClient = newTestClient(testBridge)
	var actual, err = testClient.Login("", []byte(""))

	assert.Empty(t, actual)
	assert.Error(t, err)
	assert.Equal(t, err.Error(), expectedStatus)
}

func TestClient_Login_NoResponseStatus(t *testing.T) {
	var testBridge = &stubBridge{
		response: newTestResponse(0, ""),
	}
	var testClient = newTestClient(testBridge)
	var actual, err = testClient.Login("", []byte(""))

	assert.Empty(t, actual)
	assert.Error(t, err)
	assert.Contains(t, err.Error(), testClient.HostAddr)
}

func TestClient_LoginUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.LoginUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_Logout(t *testing.T) {
	var expectedSessionId = "sessionId"
	var expectedSessionToken = "token"
	var testBridge = &stubBridge{
		response: newTestResponse(200, ""),
	}
	var testClient = newTestClient(testBridge, expectedSessionToken)
	var err = testClient.Logout(expectedSessionId)

	assert.NoError(t, err)
	assert.Equal(t, expectedSessionId, testBridge.params["id"])
	assert.Equal(t, expectedSessionToken, testBridge.cookie.Value)
}

func TestClient_Logout_InvalidUrl(t *testing.T) {
	var testClient = &client{}

	assert.NoError(t, testClient.Logout(""))
}

func TestClient_Logout_NoResponse(t *testing.T) {
	var expectedErr = errors.New("expected")
	var testBridge = &stubBridge{
		err: expectedErr,
	}
	var testClient = newTestClient(testBridge)
	var err = testClient.Logout("")

	assert.ErrorIs(t, err, expectedErr)
}

func TestClient_Logout_NoResponseCustomStatus(t *testing.T) {
	var expectedStatus = "status"
	var testBridge = &stubBridge{
		response: newTestResponse(0, expectedStatus),
	}
	var testClient = newTestClient(testBridge)
	var err = testClient.Logout("")

	assert.Error(t, err)
	assert.Equal(t, err.Error(), expectedStatus)
}

func TestClient_Logout_NoResponseStatus(t *testing.T) {
	var testBridge = &stubBridge{
		response: newTestResponse(0, ""),
	}
	var testClient = newTestClient(testBridge)
	var err = testClient.Logout("")

	assert.Error(t, err)
	assert.Contains(t, err.Error(), testClient.HostAddr)
}

func TestClient_LogoutUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.LogoutUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_ProvisionFunction(t *testing.T) {
	var expectedId = 37
	var expectedAuth = "auth"
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			result: &clientPayload[provisionData]{
				Data: provisionData{
					Sf: Function{
						Id: expectedId,
					},
					Auth: expectedAuth,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)
	var testFunction = &Function{
		Name:        "name",
		Description: "description",
		Handle:      "handle",
		Oem:         "oem",
		Type:        "type",
		Arches:      []string{"arch1", "arch2"},
	}
	var actual, err = testClient.ProvisionFunction(testFunction)
	var actualModel *functionModel

	assert.NoError(t, err)
	assert.Equal(t, cast.ToString(expectedId), actual.Id)
	assert.Equal(t, expectedAuth, actual.Auth)
	assert.NotEmpty(t, testBridge.cookie)
	assert.Equal(t, expectedToken, testBridge.cookie.Value)
	assert.NoError(t, json.Unmarshal([]byte(testBridge.params["model"]), &actualModel))
	assert.Equal(t, testFunction.Name, actualModel.Name)
	assert.Equal(t, testFunction.Description, actualModel.Description)
	assert.Equal(t, testFunction.Oem, actualModel.Oem)
	assert.Equal(t, testFunction.Handle, actualModel.Handle)
	assert.Equal(t, testFunction.Type, actualModel.Type)
	assert.Equal(t, testFunction.Version, actualModel.Version)
}

func TestClient_ProvisionFunction_InvalidUrl(t *testing.T) {
	var testClient = &client{AuthToken: "token"}
	var testFunction = &Function{}
	var _, err = testClient.ProvisionFunction(testFunction)

	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ProvisionFunction_NoAuth(t *testing.T) {
	var testClient = &client{}
	var testFunction = &Function{}
	var _, err = testClient.ProvisionFunction(testFunction)

	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ProvisionFunctionUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.ProvisionFunctionUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_PublishFunction(t *testing.T) {
	var expectedToken = "token"
	var testFunction = &Function{
		Id:          37,
		Name:        "name",
		Description: "description",
		Digest:      "digest",
		Handle:      "handle",
		Oem:         "oem",
		Type:        "type",
		Arches:      []string{"arch1", "arch2"},
	}
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			result: &clientPayload[publishingData]{
				Data: publishingData{
					Sf: *testFunction,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)
	var actual, err = testClient.PublishFunction(testFunction.asIdentity())

	assert.NoError(t, err)
	assert.NotEmpty(t, testBridge.cookie)
	assert.Equal(t, expectedToken, testBridge.cookie.Value)
	assert.Equal(t, cast.ToString(testFunction.Id), testBridge.params["id"])
	assert.Equal(t, testFunction.Digest, testBridge.params["digest"])
	assert.Equal(t, testFunction.Id, actual.Id)
	assert.Equal(t, testFunction.Name, actual.Name)
	assert.Equal(t, testFunction.Description, actual.Description)
	assert.Equal(t, testFunction.Digest, actual.Digest)
	assert.Equal(t, testFunction.Oem, actual.Oem)
	assert.Equal(t, testFunction.Handle, actual.Handle)
	assert.Equal(t, testFunction.Type, actual.Type)
	assert.Equal(t, testFunction.Version, actual.Version)
	assert.Equal(t, testFunction.Arches, actual.Arches)
}

func TestClient_PublishFunction_InvalidUrl(t *testing.T) {
	var testClient = &client{AuthToken: "token"}
	var testFunction = &Function{}
	var _, err = testClient.PublishFunction(testFunction.asIdentity())

	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_PublishFunction_NoAuth(t *testing.T) {
	var testClient = &client{}
	var testFunction = &Function{}
	var _, err = testClient.PublishFunction(testFunction.asIdentity())

	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_PublishFunctionUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.PublishFunctionUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_PublishSolution(t *testing.T) {
	var expectedDigest = "digest"
	var expectedId = int64(37)
	var expectedPath = "path"
	var expectedToken = "token"
	var expectedWorkflowHandle = "WorkflowHandle"
	var testSolution = &Solution{
		Name:        "SolutionName",
		Description: "SolutionDesc",
		Handle:      "SolutionHandle",
		Oem:         "SolutionOEM",
		Version:     "SolutionVersion",
		Workflows: []Workflow{
			{
				Handle: expectedWorkflowHandle,
			},
		},
	}
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			result: &clientPayload[solutionSlices]{
				Data: solutionSlices{
					Solution: SolutionRemote{
						Solution: Solution{
							Version: testSolution.Version,
						},
						Id:     expectedId,
						Digest: expectedDigest,
						Fqdn:   expectedPath,
					},
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)
	var actual, err = testClient.PublishSolution(testSolution)
	var serialized *Solution

	assert.NoError(t, err)
	assert.NotEmpty(t, testBridge.cookie)
	assert.Equal(t, expectedToken, testBridge.cookie.Value)
	assert.NoError(t, json.Unmarshal([]byte(testBridge.params["solution"]), &serialized))
	assert.Equal(t, testSolution.Description, serialized.Description)
	assert.Equal(t, testSolution.Handle, serialized.Handle)
	assert.Equal(t, testSolution.Name, serialized.Name)
	assert.Equal(t, testSolution.Oem, serialized.Oem)
	assert.Equal(t, testSolution.Version, serialized.Version)
	assert.Equal(t, expectedWorkflowHandle, serialized.Workflows[0].Handle)
	assert.Equal(t, expectedId, cast.ToInt64(actual.Id))
	assert.Equal(t, expectedDigest, actual.Hash)
	assert.Equal(t, expectedPath, actual.Path)
	assert.Equal(t, testSolution.Version, actual.Version)
}

func TestClient_PublishSolution_InvalidUrl(t *testing.T) {
	var testClient = &client{AuthToken: "token"}
	var testSolution = &Solution{}
	var _, err = testClient.PublishSolution(testSolution)

	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_PublishSolution_NoAuth(t *testing.T) {
	var testClient = &client{}
	var testSolution = &Solution{}
	var _, err = testClient.PublishSolution(testSolution)

	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_PublishSolutionUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.PublishSolutionUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_Session(t *testing.T) {
	var expectedToken = "token"
	var expectedId = int64(37)
	var expectedUserId = 31
	var expectedExpiry = int64(1337)
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			result: &clientPayload[Session]{
				Data: Session{
					Id:     expectedId,
					UserId: expectedUserId,
					Expiry: expectedExpiry,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)
	var actual, err = testClient.Session()

	assert.NoError(t, err)
	assert.NotEmpty(t, testBridge.cookie)
	assert.Equal(t, expectedToken, testBridge.cookie.Value)
	assert.Equal(t, expectedId, actual.Id)
	assert.Equal(t, expectedUserId, actual.UserId)
	assert.Equal(t, expectedExpiry, actual.Expiry)
}

func TestClient_Session_InvalidUrl(t *testing.T) {
	var testClient = &client{AuthToken: "token"}
	var _, err = testClient.Session()

	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_Session_NoAuth(t *testing.T) {
	var testClient = &client{}
	var _, err = testClient.Session()

	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_SessionUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.SessionUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestActiveClient(t *testing.T) {

	var testFile = filepath.Join(t.TempDir(), ".authFile")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var expectedHost = "host"
		var expectedToken = "token"
		var testAuth = &AuthData{
			Accounts: []*AuthAccount{
				{
					AuthSession: &AuthSession{
						Expiry: -1,
						Token:  expectedToken,
					},
					HostAddr: expectedHost,
				},
			},
		}

		defer filez.CloseSilently(fd)

		if err = testAuth.Write(fd.Name()); err == nil {
			var testClient Client

			testClient, err = ActiveClient(testFile)
			assert.Empty(t, err)
			assert.Equal(t, expectedToken, testClient.GetAuthToken())
			return
		}
	}

	assert.NoError(t, err)
}

func TestActiveClient_Cached(t *testing.T) {
	var expectedHost = "host"
	var testFile = filepath.Join(t.TempDir(), ".authFile")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var testAuth = &AuthData{
			Accounts: []*AuthAccount{
				{
					HostAddr: expectedHost,
				},
			},
		}

		defer filez.CloseSilently(fd)

		if err = testAuth.Write(fd.Name()); err == nil {
			var expectedClient = NewClient(expectedHost)
			var actualClient Client

			clientByHost[expectedHost] = expectedClient
			defer func() { delete(clientByHost, expectedHost) }()

			if actualClient, err = ActiveClient(testFile); err == nil {
				assert.Same(t, expectedClient, actualClient)
				return
			}
		}
	}

	assert.NoError(t, err)
}

func TestActiveClient_Expired(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), ".authFile")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var expectedHost = "host"
		var testAuth = &AuthData{
			Accounts: []*AuthAccount{
				{
					AuthSession: &AuthSession{
						Expiry: 37,
					},
					HostAddr: expectedHost,
				},
			},
		}

		defer filez.CloseSilently(fd)

		if err = testAuth.Write(fd.Name()); err == nil {
			var testClient Client

			testClient, err = ActiveClient(testFile)
			assert.Empty(t, testClient)
			assert.ErrorIs(t, err, errorSessionExpired)
			return
		}
	}

	assert.NoError(t, err)
}

func TestActiveClient_NoLogin(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), ".authFile")
	var testClient, err = ActiveClient(testFile)
	var fd *os.File

	assert.Empty(t, testClient)
	assert.ErrorIs(t, err, ErrorNoLogin)

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		var testAuth = &AuthData{Active: -1}

		assert.NoError(t, testAuth.Write(fd.Name()))
		testClient, err = ActiveClient(testFile)
		assert.Empty(t, testClient)
		assert.ErrorIs(t, err, ErrorNoLogin)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestGetClient(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := filez.CreateRecursive(testDir, "authFile"); err == nil {
		defer func() { clientByHost = make(map[string]Client) }()
		var expectedAddr = "addr"
		var expectedToken = "token"
		var auth = NewAuthData(fd.Name())
		var actual Client

		auth = auth.Push(expectedAddr, &AuthSession{Token: expectedToken, Expiry: -1})
		assert.NoError(t, auth.Write(fd.Name()))
		actual, err = GetClient(fd.Name(), expectedAddr)
		assert.NoError(t, err)
		assert.Equal(t, expectedToken, actual.GetAuthToken())
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestGetClient_Existing(t *testing.T) {
	var expectedAddr = "http://test"
	var actual Client
	var err error

	clientByHost[expectedAddr] = &emptyClient{}
	actual, err = GetClient("", expectedAddr)
	assert.NoError(t, err)
	assert.Same(t, clientByHost[expectedAddr], actual)
}

func TestGetClient_ExpiredAccount(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := filez.CreateRecursive(testDir, "authFile"); err == nil {
		defer func() { clientByHost = make(map[string]Client) }()
		var expectedAddr = "addr"
		var auth = NewAuthData(fd.Name())
		var actual Client

		auth = auth.Push(expectedAddr, &AuthSession{Token: "token", Expiry: 0})
		assert.NoError(t, auth.Write(fd.Name()))
		actual, err = GetClient(fd.Name(), expectedAddr)
		assert.Empty(t, actual)
		assert.ErrorIs(t, err, errorSessionExpired)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestGetClient_NoAccount(t *testing.T) {
	var testDir = t.TempDir()

	if fd, err := filez.CreateRecursive(testDir, "authNoAccountFile"); err == nil {
		var actual Client

		actual, err = GetClient(fd.Name(), "notExisting")
		assert.Empty(t, actual)
		assert.Error(t, err)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewClientFactory(t *testing.T) {
	var expectedAddr = "expectedAddr"
	var testClientFactory = NewClientFactory()
	var testClient = testClientFactory.New(expectedAddr)

	assert.Equal(t, defaultTimeoutSeconds, testClient.GetTimeout())
	assert.Equal(t, expectedAddr, testClient.GetHostAddr())
}

func Test_makeHostUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var expectedAddr = expectedPrefix + "/" + expectedHost
	var expectedVerb = "verb"

	assert.Equal(t, fmt.Sprintf("%s/%s/%s/%s", expectedPrefix, expectedHost, apiVersion1, pathFunction),
		makeHostUrl(expectedHost, apiVersion1, pathFunction))
	assert.Equal(t, fmt.Sprintf("%s/%s/%s/%s/%s", expectedPrefix, expectedHost, apiVersion1, pathFunction, expectedVerb),
		makeHostUrl(expectedHost, apiVersion1, pathFunction, expectedVerb))
	assert.Equal(t, fmt.Sprintf("%s/%s/%s", expectedAddr, apiVersion1, pathFunction),
		makeHostUrl(expectedAddr, apiVersion1, pathFunction))
	assert.Equal(t, fmt.Sprintf("%s/%s/%s/%s", expectedAddr, apiVersion1, pathFunction, expectedVerb),
		makeHostUrl(expectedAddr, apiVersion1, pathFunction, expectedVerb))
}

func Test_resultOrError_BadRequest(t *testing.T) {
	var emptyProvider = func(body any) *provisionData { return nil }
	var payload, err = resultOrError(&resty.Response{
		Request: &resty.Request{
			DoNotParseResponse: false,
		},
		RawResponse: &http.Response{
			StatusCode: 400,
		},
	}, emptyProvider)

	assert.Empty(t, payload)
	assert.ErrorIs(t, errorBadRequest, err)
}

func Test_resultOrError_BadRequestString(t *testing.T) {
	var emptyProvider = func(body any) *provisionData { return nil }
	var expectedResponse = "response"
	var payload, err = resultOrError(&resty.Response{
		Body: io.NopCloser(strings.NewReader(expectedResponse)),
		Request: &resty.Request{
			DoNotParseResponse: false,
		},
		RawResponse: &http.Response{
			StatusCode: 400,
		},
	}, emptyProvider)

	assert.Empty(t, payload)
	assert.Error(t, err)
	assert.Equal(t, expectedResponse, err.Error())
}

func Test_resultOrError_BadRequestJson(t *testing.T) {
	var emptyProvider = func(body any) *provisionData { return nil }
	var expectedResponse = Error{Code: 400, Status: "status", Message: "message"}
	var data, _ = json.Marshal(expectedResponse)
	var payload, err = resultOrError(&resty.Response{
		Body: io.NopCloser(bytes.NewReader(data)),
		Request: &resty.Request{
			DoNotParseResponse: false,
		},
		RawResponse: &http.Response{
			StatusCode: 400,
		},
	}, emptyProvider)

	assert.Empty(t, payload)
	assert.Error(t, err)
	assert.Equal(t, expectedResponse.Message, err.Error())
}

func Test_resultOrError_CustomStatus(t *testing.T) {
	var emptyProvider = func(body any) *provisionData { return nil }
	var expectedCode = 37
	var expectedStatus = "status"
	var payload, err = resultOrError(&resty.Response{
		RawResponse: &http.Response{
			StatusCode: expectedCode,
			Status:     expectedStatus,
		},
	}, emptyProvider)
	assert.Empty(t, payload)
	assert.Error(t, err)
	assert.Equal(t, fmt.Sprintf("%d %s", expectedCode, expectedStatus), err.Error())
}

func Test_resultOrError(t *testing.T) {
	var emptyProvider = func(body any) *provisionData { return nil }
	var payload, err = resultOrError(nil, emptyProvider)
	var expectedAuth = "auth"

	assert.Empty(t, payload)
	assert.ErrorIs(t, err, errorNoResponse)
	payload, err = resultOrError(&resty.Response{
		RawResponse: &http.Response{
			StatusCode: 200,
		},
		Request: &resty.Request{
			Result: "anything",
		},
	}, func(body any) *provisionData {
		return &provisionData{Auth: expectedAuth}
	})
	assert.NoError(t, err)
	assert.Equal(t, expectedAuth, payload.Auth)
}

func Test_sanitizeHostUrl(t *testing.T) {
	var expectedHost = "host"

	assert.Equal(t, expectedHost, sanitizeHostUrl("host"))
	assert.Equal(t, expectedHost, sanitizeHostUrl("host/"))
}

func newTestClient(testBridge *stubBridge, values ...string) *client {
	var token string

	if len(values) > 0 {
		token = values[0]
	}

	return &client{
		AuthToken: token,
		HostAddr:  "host",

		requestBridge: func() requestBridge {
			return testBridge
		},
	}
}

func newTestResponse(statusCode int, status string) responseBridge {
	return &stubResponse{
		success:    statusCode == 200 || statusCode == 202,
		status:     status,
		statusCode: statusCode,
	}
}
