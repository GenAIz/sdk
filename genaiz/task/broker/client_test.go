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
	formRequest bool
	jsonRequest bool
	cookie      *http.Cookie
	err         error
	formData    map[string]string
	queryParams map[string]string
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

func (s *stubBridge) Form() requestBridge {
	s.formRequest = true
	return s
}

func (s *stubBridge) FormData(formData map[string]string) requestBridge {
	s.formData = formData
	return s
}

func (s *stubBridge) Json() requestBridge {
	s.jsonRequest = true
	return s
}

func (s *stubBridge) Post(url string) (responseBridge, error) {
	s.url = url
	return s.response, s.err
}

func (s *stubBridge) QueryParams(queryParams map[string]string) requestBridge {
	s.queryParams = queryParams
	return s
}

func (s *stubBridge) Resulting(result any) requestBridge {
	s.result = result
	return s
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

func (s stubResponse) IsStatusSuccess() bool {
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

type testStruct struct {
	field1 string
	field2 int
}

func TestClient_CreateWorkspace(t *testing.T) {
	var expectedToken = "token"
	var expectedWorkspace = &Workspace{Name: "name"}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[workspaceSlices]{
				Data: workspaceSlices{
					Workspace: expectedWorkspace,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspace(expectedWorkspace)
	assert.NoError(t, err)
	assert.Equal(t, expectedWorkspace, actual)
}

func TestClient_CreateWorkspace_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.CreateWorkspace(&Workspace{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_CreateWorkspace_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspace(&Workspace{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_CreateWorkspace_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.CreateWorkspace(&Workspace{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_CreateWorkspace_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspace(&Workspace{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_CreateWorkspaceFlow(t *testing.T) {
	var expectedToken = "token"
	var expectedFlow = &WorkspaceFlow{
		Id:          int64(37),
		WorkflowId:  int64(42),
		Name:        "expectedName",
		Description: "expectedDescription",
	}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[workspaceFlowsSlices]{
				Data: workspaceFlowsSlices{
					WorkspaceFlow: *expectedFlow,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspaceFlow(expectedFlow.Id, expectedFlow.WorkflowId, expectedFlow.Name, expectedFlow.Description)
	assert.NoError(t, err)
	assert.Equal(t, expectedFlow, actual)
}

func TestClient_CreateWorkspaceFlow_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.CreateWorkspaceFlow(int64(37), int64(42), "name", "description")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_CreateWorkspaceFlow_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspaceFlow(int64(37), int64(42), "name", "description")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_CreateWorkspaceFlow_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.CreateWorkspaceFlow(int64(37), int64(42), "name", "description")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_CreateWorkspaceFlow_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.CreateWorkspaceFlow(int64(37), int64(42), "name", "description")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_ExportDataLink(t *testing.T) {
	var expectedToken = "token"
	var expectedDataLink = &DataLink{Handle: "handle"}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[DataLink]{
				Data: *expectedDataLink,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.NoError(t, err)
	assert.Equal(t, expectedDataLink, actual)
}

func TestClient_ExportDataLink_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ExportDataLink_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_ExportDataLink_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ExportDataLink_UnknownLink(t *testing.T) {
	var expectedToken = "token"
	var expectedDataLink = &DataLink{Id: new(int64(0))}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[DataLink]{
				Data: *expectedDataLink,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.ErrorIs(t, err, errorDatalinkUnknown)
	assert.Nil(t, actual)
}

func TestClient_ExportDataLink_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ExportDataLink("oem", "handle", "version", "sequence")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_FindDataLink(t *testing.T) {
	var expectedToken = "token"
	var expectedDataLink = &DataLink{Id: new(int64(37))}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[DataLink]{
				Data: *expectedDataLink,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindDataLink("oem", "handle", "version")
	assert.NoError(t, err)
	assert.Equal(t, expectedDataLink, actual)
}

func TestClient_FindDataLink_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.FindDataLink("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_FindDataLink_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindDataLink("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_FindDataLink_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.FindDataLink("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_FindDataLink_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindDataLink("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_FindSolution(t *testing.T) {
	var expectedToken = "token"
	var expectedFunctions = []Function{
		{
			Id:      1337,
			Oem:     "expectedFunctionOem",
			Handle:  "expectedFunctionHandle",
			Version: "expectedFunctionVersion",
			Seq:     new(1),
		},
	}
	var expectedSolution = &Solution{Id: new(int64(37))}
	var expectedWorkflows = []Workflow{
		{
			Id: new(int64(42)),
		},
	}
	var expectedWorkflowNodes = []WorkflowNode{
		{
			Id:     new(int64(1)),
			Handle: "ignored",
		},
		{
			Id:         new(int64(2)),
			WorkflowId: expectedWorkflows[0].Id,
			Handle:     "expectedNode",
		},
		{
			Id:         new(int64(3)),
			WorkflowId: new(int64(67)),
			Handle:     "unknown",
		},
		{
			Id:              new(int64(4)),
			WorkflowId:      expectedWorkflows[0].Id,
			SmartFunctionId: new(int64(expectedFunctions[0].Id)),
			Handle:          "expectedSfNode",
		},
		{
			Id:              new(int64(5)),
			WorkflowId:      expectedWorkflows[0].Id,
			SmartFunctionId: new(int64(99)),
			Handle:          "expectedUnknownSf",
		},
	}
	var expectedLinks = []WorkflowLink{
		{
			// This will be ignored because of no workflowId found
			LhsNodeId: expectedWorkflowNodes[3].Id,
			RhsNodeId: expectedWorkflowNodes[4].Id,
		},
		{
			// This will be ignored because of no workflow 73 in the solution
			WorkflowId: new(int64(73)),
			LhsNodeId:  expectedWorkflowNodes[1].Id,
			RhsNodeId:  expectedWorkflowNodes[4].Id,
		},
		{
			// This will be ignored because links always require a left node
			WorkflowId: expectedWorkflows[0].Id,
		},
		{
			// This will be ignored because links always require a right node
			WorkflowId: expectedWorkflows[0].Id,
			LhsNodeId:  expectedWorkflowNodes[1].Id,
		},
		{
			WorkflowId: expectedWorkflows[0].Id,
			LhsNodeId:  expectedWorkflowNodes[1].Id,
			RhsNodeId:  expectedWorkflowNodes[3].Id,
		},
	}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[solutionSlices]{
				Data: solutionSlices{
					Solution:       *expectedSolution,
					Workflows:      expectedWorkflows,
					WorkflowLinks:  expectedLinks,
					WorkflowNodes:  expectedWorkflowNodes,
					SmartFunctions: expectedFunctions,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindSolution("oem", "handle", "version")
	assert.NoError(t, err)
	assert.Equal(t, expectedSolution.Id, actual.Id)
	assert.Equal(t, expectedWorkflows[0].Id, actual.Workflows[0].Id)
	assert.Equal(t, expectedWorkflowNodes[1].Id, actual.Workflows[0].Nodes[0].Id)
	assert.Equal(t, expectedWorkflowNodes[3].Id, actual.Workflows[0].Nodes[1].Id)
	assert.Equal(t, expectedWorkflowNodes[4].Handle, actual.Workflows[0].Nodes[2].Handle)
	assert.Equal(t, expectedWorkflowNodes[1].Handle, actual.Workflows[0].Links[0].LhsNode)
	assert.Equal(t, expectedWorkflowNodes[3].Handle, actual.Workflows[0].Links[0].RhsNode)
}

func TestClient_FindSolution_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.FindSolution("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_FindSolution_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindSolution("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_FindSolution_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.FindSolution("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_FindSolution_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.FindSolution("oem", "handle", "version")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_GetFunction(t *testing.T) {
	var expectedToken = "token"
	var expectedFunction = &Function{Id: 37}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[Function]{
				Data: *expectedFunction,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.GetFunction(int64(expectedFunction.Id))
	assert.NoError(t, err)
	assert.Equal(t, expectedFunction, actual)
}

func TestClient_GetFunction_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.GetFunction(int64(37))
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_GetFunction_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.GetFunction(int64(37))
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_GetFunction_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.GetFunction(int64(37))
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_GetFunction_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.GetFunction(int64(37))
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_ListDataLinks(t *testing.T) {
	var expectedToken = "token"
	var expectedDataLinks = []DataLink{
		{
			Id: new(int64(37)),
		},
	}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[[]DataLink]{
				Data: expectedDataLinks,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListDataLinks("oem", "handle", DataLinkFlags.Active)
	assert.NoError(t, err)
	assert.Equal(t, expectedDataLinks, actual)
}

func TestClient_ListDataLinks_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.ListDataLinks("oem", "handle", DataLinkFlags.Active)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ListDataLinks_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListDataLinks("oem", "handle", DataLinkFlags.Active)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_ListDataLinks_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.ListDataLinks("oem", "handle", DataLinkFlags.Active)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ListDataLinks_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListDataLinks("oem", "handle", DataLinkFlags.Active)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_ListDataLinksUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.ListDataLinksUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_ListSolutions(t *testing.T) {
	var expectedToken = "token"
	var expectedSolution = &Solution{
		Id:   new(int64(37)),
		Name: "name",
	}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[[]Solution]{
				Data: []Solution{*expectedSolution},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListSolutions("oem")
	assert.NoError(t, err)
	assert.Equal(t, []Solution{*expectedSolution}, actual)
}

func TestClient_ListSolutions_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.ListSolutions("oem")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ListSolutions_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListSolutions("oem")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_ListSolutions_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.ListSolutions("oem")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ListSolutions_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListSolutions("oem")
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_ListSolutionsUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.ListSolutionsUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_ListWorkspaces(t *testing.T) {
	var expectedToken = "token"
	var expectedWorkspace = &Workspace{Name: "name"}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[*workspaceList]{
				Data: &workspaceList{
					Workspaces: []Workspace{*expectedWorkspace},
				},
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListWorkspaces(1, 1)
	assert.NoError(t, err)
	assert.Equal(t, []Workspace{*expectedWorkspace}, actual)
}

func TestClient_ListWorkspaces_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.ListWorkspaces(1, 1)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ListWorkspaces_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListWorkspaces(1, 1)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_ListWorkspaces_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.ListWorkspaces(1, 1)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ListWorkspaces_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.ListWorkspaces(1, 1)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_Login(t *testing.T) {
	var expectedUser = "user"
	var expectedPassword = "password"
	var expectedToken = "token"
	var expectedSessionId = int64(37)
	var testBridge = &stubBridge{
		response: &stubResponse{
			success: true,
			cookies: []*http.Cookie{{Name: defaultCookieName, Value: expectedToken}},
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
	assert.Equal(t, expectedSessionId, testBridge.formData["id"])
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

	if err == nil {
		assert.Fail(t, "no error returned")
	} else {
		assert.Equal(t, err.Error(), expectedStatus)
	}
}

func TestClient_Logout_NoResponseStatus(t *testing.T) {
	var testBridge = &stubBridge{
		response: newTestResponse(0, ""),
	}
	var testClient = newTestClient(testBridge)
	var err = testClient.Logout("")

	if err == nil {
		assert.Fail(t, "no error returned")
	} else {
		assert.Contains(t, err.Error(), testClient.HostAddr)
	}
}

func TestClient_LogoutUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.LogoutUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_OidcDeviceCode(t *testing.T) {
	var expectedCodeUrl = "codeUrl"
	var expectedAuth = &DeviceAuth{
		DeviceCode:              "expectedCode",
		VerificationUriComplete: "expectedUri",
	}
	var testBridge = &stubBridge{
		response: &stubResponse{
			success:    true,
			statusCode: 200,
			result:     expectedAuth,
		},
	}
	var testClient = newTestClient(testBridge)

	if actual, err := testClient.OidcDeviceCode(expectedCodeUrl, oidcDeviceClient); actual != nil {
		assert.Equal(t, expectedAuth, actual)
		assert.Equal(t, expectedCodeUrl, testBridge.url)
		assert.True(t, testBridge.formRequest)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestClient_OidcDeviceCode_EmptyUrl(t *testing.T) {
	var testClient = &client{}

	actual, err := testClient.OidcDeviceCode("", nil)
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestClient_OidcDeviceCode_UrlError(t *testing.T) {
	var expectedCodeUrl = "codeUrl"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcDeviceCode(expectedCodeUrl, oidcDeviceClient)
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_OidcDeviceUrl(t *testing.T) {
	var expectedUrl = "expectedUrl"
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[string]{
				Data: expectedUrl,
			},
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcDeviceUrl()
	assert.NoError(t, err)
	assert.Equal(t, expectedUrl, actual)
}

func TestClient_OidcDeviceUrl_RequestError(t *testing.T) {
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcDeviceUrl()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_OidcDeviceUrl_UnknownHost(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.OidcDeviceUrl()
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestClient_OidcDeviceUrl_UrlError(t *testing.T) {
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcDeviceUrl()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_OidcTokenCreate(t *testing.T) {
	var expectedToken = "expectedToken"
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &oauthResponse{
				AccessToken: expectedToken,
			},
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenCreate("url", "testCode", &DeviceClient{})
	assert.NoError(t, err)
	assert.Equal(t, expectedToken, actual)
}

func TestClient_OidcTokenCreate_EmptyUrl(t *testing.T) {
	var testClient = &client{}
	var actual, err = testClient.OidcTokenCreate("", "testCode", &DeviceClient{})

	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestClient_OidcTokenCreate_RequestError(t *testing.T) {
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenCreate("url", "testCode", &DeviceClient{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_OidcTokenCreate_UrlError(t *testing.T) {
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenCreate("url", "testCode", &DeviceClient{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_OidcTokenSession(t *testing.T) {
	var expectedId = int64(37)
	var expectedUserId = 42
	var expectedExpiry = int64(1)
	var expectedToken = "sessionToken"
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			cookies: []*http.Cookie{{Name: defaultCookieName, Value: expectedToken}},
			result: &clientPayload[Session]{
				Data: Session{
					Id:     expectedId,
					UserId: expectedUserId,
					Expiry: expectedExpiry,
				},
			},
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenSession("url", "testCode")
	assert.NoError(t, err)
	assert.Equal(t, expectedId, actual.SessionId)
	assert.Equal(t, expectedUserId, actual.UserId)
	assert.Equal(t, expectedExpiry, actual.Expiry)
	assert.Equal(t, expectedToken, actual.Token)
}

func TestClient_OidcTokenSession_UnknownHost(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.OidcTokenSession("url", "token")
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestClient_OidcTokenUrl(t *testing.T) {
	var expectedUrl = "expectedUrl"
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[string]{
				Data: expectedUrl,
			},
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenUrl()
	assert.NoError(t, err)
	assert.Equal(t, expectedUrl, actual)
}

func TestClient_OidcTokenUrl_RequestError(t *testing.T) {
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenUrl()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_OidcTokenUrl_UnknownHost(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.OidcTokenUrl()
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func TestClient_OidcTokenUrl_UrlError(t *testing.T) {
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge)

	actual, err := testClient.OidcTokenUrl()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
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
	var actual, err = testClient.ProvisionFunction(testFunction, nil)
	var actualModel *functionModel

	assert.NoError(t, err)
	assert.Equal(t, cast.ToString(expectedId), actual.Id)
	assert.Equal(t, expectedAuth, actual.Auth)
	assert.NotEmpty(t, testBridge.cookie)
	assert.Equal(t, expectedToken, testBridge.cookie.Value)
	assert.NoError(t, json.Unmarshal([]byte(testBridge.formData["model"]), &actualModel))
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
	var _, err = testClient.ProvisionFunction(testFunction, nil)

	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_ProvisionFunction_NoAuth(t *testing.T) {
	var testClient = &client{}
	var testFunction = &Function{}
	var _, err = testClient.ProvisionFunction(testFunction, nil)

	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_ProvisionFunctionUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.ProvisionFunctionUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
}

func TestClient_PublishDataLink(t *testing.T) {
	var expectedToken = "token"
	var expectedDataLink = DataLink{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var testBridge = &stubBridge{
		response: stubResponse{
			success: true,
			result: &clientPayload[DataLink]{
				Data: expectedDataLink,
			},
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.PublishDataLink(&expectedDataLink)
	assert.NoError(t, err)
	assert.Equal(t, &expectedDataLink, actual)
}

func TestClient_PublishDataLink_NoAuth(t *testing.T) {
	var testClient = &client{HostAddr: ""}

	actual, err := testClient.PublishDataLink(&DataLink{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorNoAuth)
}

func TestClient_PublishDataLink_RequestError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		response: stubResponse{
			success:    false,
			statusCode: 400,
		},
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.PublishDataLink(&DataLink{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorBadRequest)
}

func TestClient_PublishDataLink_UnknownHost(t *testing.T) {
	var expectedToken = "token"
	var testClient = &client{HostAddr: "", AuthToken: expectedToken}

	actual, err := testClient.PublishDataLink(&DataLink{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, errorInvalidHost)
}

func TestClient_PublishDataLink_UrlError(t *testing.T) {
	var expectedToken = "token"
	var testBridge = &stubBridge{
		err: errors.New("expected error"),
	}
	var testClient = newTestClient(testBridge, expectedToken)

	actual, err := testClient.PublishDataLink(&DataLink{})
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, testBridge.err)
}

func TestClient_PublishDataLinkUrl(t *testing.T) {
	var expectedHost = "host"
	var expectedPrefix = env.DefaultProtocolPrefix(expectedHost)
	var testClient = &client{HostAddr: expectedHost}

	assert.Contains(t, testClient.PublishDataLinkUrl(), fmt.Sprintf("%s/%s", expectedPrefix, expectedHost))
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
	assert.Equal(t, cast.ToString(testFunction.Id), testBridge.formData["id"])
	assert.Equal(t, testFunction.Digest, testBridge.formData["digest"])
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
	var expectedFqdn = "fqdn"
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
					Solution: Solution{
						Version: testSolution.Version,
						Id:      new(expectedId),
						Digest:  new(expectedDigest),
						Fqdn:    new(expectedFqdn),
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
	assert.NoError(t, json.Unmarshal([]byte(testBridge.formData["solution"]), &serialized))
	assert.Equal(t, testSolution.Description, serialized.Description)
	assert.Equal(t, testSolution.Handle, serialized.Handle)
	assert.Equal(t, testSolution.Name, serialized.Name)
	assert.Equal(t, testSolution.Oem, serialized.Oem)
	assert.Equal(t, testSolution.Version, serialized.Version)
	assert.Equal(t, expectedWorkflowHandle, serialized.Workflows[0].Handle)
	assert.Equal(t, expectedId, cast.ToInt64(actual.Id))
	assert.Equal(t, expectedDigest, *actual.Digest)
	assert.Equal(t, expectedFqdn, *actual.Fqdn)
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

func TestSeedClient(t *testing.T) {
	var expectedAddr = "tehAddr"
	var expectedToken = "leToken"

	if actual, err := SeedClient(expectedAddr, expectedToken); err == nil {
		assert.Equal(t, expectedAddr, actual.GetHostAddr())
		assert.Equal(t, expectedToken, actual.GetAuthToken())
		return
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewClientFactory(t *testing.T) {
	var expectedAddr = "expectedAddr"
	var testClientFactory = NewClientFactory()
	var testClient = testClientFactory.New(expectedAddr)

	assert.Equal(t, defaultExpiryMinutes, testClient.GetExpiry())
	assert.Equal(t, defaultTimeoutSeconds, testClient.GetTimeout())
	assert.Equal(t, expectedAddr, testClient.GetHostAddr())
}

func Test_jsonExpand(t *testing.T) {
	var testFunction = &Function{Name: "the_right_name"}
	var testExtras = map[string]any{
		"Name":            "not_the_right_name",
		"supplement_int":  89,
		"supplement_list": []string{"item1", "item2"},
		"supplement_obj":  testStruct{field1: "field1", field2: 37},
		"supplement_ref":  &testStruct{field1: "field2", field2: 42},
	}
	var tBytes []byte
	var err error

	if tBytes, err = json.Marshal(testFunction); err == nil {
		var actual map[string]any

		if actual, err = jsonExpand(tBytes, testExtras); err == nil {
			assert.Equal(t, testFunction.Name, actual["name"])
			assert.Equal(t, testExtras["supplement_int"], actual["supplement_int"])
			assert.Equal(t, testExtras["supplement_list"], actual["supplement_list"])
			assert.Equal(t, testExtras["supplement_obj"], actual["supplement_obj"])
			assert.Equal(t, testExtras["supplement_ref"], actual["supplement_ref"])
		}
	}

	assert.NoError(t, err)
	_, err = jsonExpand([]byte("notValidJson"), nil)
	assert.Error(t, err)
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
			IsResponseDoNotParse: false,
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
			IsResponseDoNotParse: false,
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
			IsResponseDoNotParse: false,
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
