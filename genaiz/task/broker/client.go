package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"maps"
	"net/http"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"time"

	"github.com/spf13/cast"
	"resty.dev/v3"

	"genaiz.com/genaiz-lib/lang/intz"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
	"genaiz.com/genaiz/version/env"
)

const (
	defaultCookieName     = "genaiz_token"
	defaultExpiryMinutes  = 5 * 24 * 60
	defaultTimeoutSeconds = 30

	apiVersion1     version = "v1"
	pathDataLink    path    = "datalink"
	pathFunction    path    = "sf"
	pathOemSolution path    = "oem/solution"
	pathOidcDevice  path    = "oidc/device"
	pathOidcToken   path    = "oidc/token"
	pathSession     path    = "user/session"
	pathSolution    path    = "solution"
	pathWorkspace   path    = "workspace"
)

var (
	errorBadRequest         = task.NewRequestError("bad broker request", 400)
	errorDatalinkUnknown    = task.NewRequestError("datalink is unknown to the broker", 404)
	errorDisallowedProtocol = task.NewRequestError("broker protocol is not allowed", 505)
	errorDisconnected       = task.NewRequestError("broker connection timed out", 0)
	errorForbidden          = task.NewRequestError("broker refused the request", 403)
	errorInternal           = task.NewRequestError("broker request crashed internally", 500)
	errorInvalidHost        = task.NewRequestError("invalid host address", 410)
	errorNoAuth             = task.NewRequestError("client not authenticated", 401)
	errorNotFound           = task.NewRequestError("broker did not find the entity", 404)
	errorNoResponse         = task.NewRequestError("client expected a response, but got none", 0)
	errorNoPath             = task.NewRequestError("broker path is not available", 501)
	errorNoToken            = task.NewRequestError("could not read token from cookie", 500)
	errorSessionExpired     = task.NewRequestError("broker session is expired", 401)
	errorUnauthorized       = task.NewRequestError("unauthorized, please login", 401)

	clientByHost = map[string]Client{}
	clientErrors = map[int]error{
		0:   errorDisconnected,
		400: errorBadRequest,
		401: errorUnauthorized,
		403: errorForbidden,
		404: errorNotFound,
		410: errorInvalidHost,
		500: errorInternal,
		501: errorNoPath,
		505: errorDisallowedProtocol,
	}
	clientFactory = NewClientFactory()
)

type path string

type version string

type Client interface {
	CreateWorkspace(*Workspace) (*Workspace, error)

	CreateWorkspaceUrl() string

	CreateWorkspaceFlow(int64, int64, string, string) (*WorkspaceFlow, error)

	CreateWorkspaceFlowUrl() string

	ExportDataLink(string, string, string, string) (*DataLink, error)

	FindDataLink(string, string, string) (*DataLink, error)

	FindSolution(string, string, string) (*Solution, error)

	FindSolutionUrl() string

	GetAuthToken() string

	GetExpiry() int

	GetFunction(int64) (*Function, error)

	GetFunctionUrl() string

	GetHostAddr() string

	GetTimeout() int

	GetUserId() int

	ListDataLinks(string, string, int) ([]DataLink, error)

	ListDataLinksUrl() string

	ListSolutions(string) ([]Solution, error)

	ListSolutionsUrl() string

	ListWorkspaces(int, int) ([]Workspace, error)

	ListWorkspacesUrl() string

	Login(string, []byte) (*AuthSession, error)

	LoginUrl() string

	Logout(string) error

	LogoutUrl() string

	OidcDeviceCode(string, *DeviceClient) (*DeviceAuth, error)

	OidcDeviceUrl() (string, error)

	OidcTokenCreate(string, string, *DeviceClient) (string, error)

	OidcTokenSession(string, string) (*AuthSession, error)

	OidcTokenUrl() (string, error)

	ProvisionFunction(*Function, map[string]any) (*shared.Identity, error)

	ProvisionFunctionUrl() string

	PublishDataLink(*DataLink) (*DataLink, error)

	PublishFunction(*shared.Identity) (*Function, error)

	PublishFunctionUrl() string

	PublishSolution(*Solution) (*Solution, error)

	PublishSolutionUrl() string

	Session() (*Session, error)

	SessionUrl() string

	SessionValid(*Session) bool

	WithAccount(*AuthAccount) (Client, error)
}

type client struct {
	AuthToken string
	Expiry    int
	HostAddr  string
	UserId    int

	requestBridge func() requestBridge
}

type clientPayload[P any] struct {
	Code   int
	Data   P
	Status string
}

type oauthResponse struct {
	AccessToken      string `json:"access_token"`
	ExpiresIn        int    `json:"expires_in"`
	IdToken          string `json:"id_token"`
	NotBeforePolicy  int    `json:"not_before_policy"`
	RefreshExpiresIn int    `json:"refresh_expires_in"`
	RefreshToken     string `json:"refresh_token"`
	Scope            string `json:"scope"`
	SessionState     string `json:"session_state"`
	TokenType        string `json:"token_type"`
}

type solutionSlices struct {
	Solution       Solution       `json:"solution"`
	Workflows      []Workflow     `json:"workflows"`
	WorkflowLinks  []WorkflowLink `json:"workflowLinks"`
	WorkflowNodes  []WorkflowNode `json:"workflowNodes"`
	SmartFunctions []Function     `json:"smartFunctions"`
}

func (ss solutionSlices) graph() *Solution {
	var workflowsById = mapz.Mapped(ss.Workflows, func(workflow Workflow) string {
		return cast.ToString(workflow.Id)
	})
	var functionsById = mapz.Mapped(ss.SmartFunctions, func(function Function) string {
		return cast.ToString(function.Id)
	})
	var nodesByWorkflow = make(map[int64][]WorkflowNode)
	var nodesById = make(map[int64]WorkflowNode)
	var linksByWorkflow = make(map[int64][]WorkflowLink)
	var graphedWorkflows []Workflow

	for _, node := range ss.WorkflowNodes {
		nodesById[*node.Id] = node

		if node.WorkflowId != nil {
			var wf Workflow
			var ok bool

			if wf, ok = workflowsById[cast.ToString(*node.WorkflowId)]; ok {
				var nodeFunction *WorkflowNodeFunction

				if node.SmartFunctionId != nil {
					var function Function

					if function, ok = functionsById[cast.ToString(*node.SmartFunctionId)]; ok {
						nodeFunction = &WorkflowNodeFunction{
							Oem:     function.Oem,
							Handle:  function.Handle,
							Version: function.Version,
							Seq:     intz.IntToDefault(function.Seq, 0),
						}
					}
				}

				nodesByWorkflow[*wf.Id] = append(nodesByWorkflow[*wf.Id], WorkflowNode{
					Id:          node.Id,
					Handle:      node.Handle,
					Name:        node.Name,
					Description: node.Description,
					Props:       node.Props,
					Sf:          nodeFunction,
				})
			}
		}
	}

	for _, link := range ss.WorkflowLinks {
		if link.WorkflowId != nil {
			if wf, ok := workflowsById[cast.ToString(link.WorkflowId)]; ok {
				var leftNodeHandle, rightNodeHandle string

				if link.LhsNodeId != nil {
					var leftNode WorkflowNode

					if leftNode, ok = nodesById[*link.LhsNodeId]; ok {
						leftNodeHandle = leftNode.Handle
					}
				}

				if link.RhsNodeId != nil {
					var rightNode WorkflowNode

					if rightNode, ok = nodesById[*link.RhsNodeId]; ok {
						rightNodeHandle = rightNode.Handle
					}
				}

				if leftNodeHandle != "" && rightNodeHandle != "" {
					linksByWorkflow[*wf.Id] = append(linksByWorkflow[*wf.Id], WorkflowLink{
						LhsNodeId:   link.LhsNodeId,
						LhsNode:     leftNodeHandle,
						LhsNodePort: link.LhsNodePort,
						RhsNodeId:   link.RhsNodeId,
						RhsNode:     rightNodeHandle,
						RhsNodePort: link.RhsNodePort,
					})
				}
			}
		}
	}

	for _, wf := range slices.Collect(maps.Values(workflowsById)) {
		graphedWorkflows = append(graphedWorkflows, Workflow{
			Id:          wf.Id,
			Created:     wf.Created,
			Modified:    wf.Modified,
			Flags:       wf.Flags,
			Handle:      wf.Handle,
			Name:        wf.Name,
			Description: wf.Description,
			Links:       linksByWorkflow[*wf.Id],
			Nodes:       nodesByWorkflow[*wf.Id],
		})
	}

	return &Solution{
		Id:          ss.Solution.Id,
		Created:     ss.Solution.Created,
		Modified:    ss.Solution.Modified,
		Flags:       ss.Solution.Flags,
		Seq:         ss.Solution.Seq,
		Name:        ss.Solution.Name,
		Description: ss.Solution.Description,
		Oem:         ss.Solution.Oem,
		Handle:      ss.Solution.Handle,
		Version:     ss.Solution.Version,
		Digest:      ss.Solution.Digest,
		Workflows:   graphedWorkflows,
	}
}

type workspaceList struct {
	Workspaces []Workspace `json:"workspaces"`
}

type workspaceSlices struct {
	Workspace *Workspace `json:"workspace"`
}

type workspaceFlowsSlices struct {
	WorkspaceFlow WorkspaceFlow `json:"workspaceFlow"`
}

func (c *client) CreateWorkspace(workspace *Workspace) (*Workspace, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathWorkspace, "create"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[workspaceSlices]{}).
				FormData(map[string]string{
					"name":        workspace.Name,
					"description": workspace.Description,
					"visibility":  strings.ToUpper(workspace.Visibility),
					"rcEnabled":   cast.ToString(workspace.IsRcEnabled()),
				}).
				Post(url)

			if err == nil {
				return resultOrError(resp, func(body any) *Workspace {
					var payload = resp.Result().(*clientPayload[workspaceSlices])

					return payload.Data.Workspace
				})
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) CreateWorkspaceUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathWorkspace, "create")
}

func (c *client) CreateWorkspaceFlow(workspaceId int64, workflowId int64, name string, description string) (*WorkspaceFlow, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathWorkspace, "flow", "create"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *WorkspaceFlow

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[workspaceFlowsSlices]{}).
				FormData(map[string]string{
					"workspaceId": cast.ToString(workspaceId),
					"workflowId":  cast.ToString(workflowId),
					"name":        name,
					"description": description,
				}).Post(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *WorkspaceFlow {
					var payload = resp.Result().(*clientPayload[workspaceFlowsSlices])

					return &payload.Data.WorkspaceFlow
				}); err == nil {
					return result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) CreateWorkspaceFlowUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathWorkspace, "flow", "create")
}

func (c *client) ExportDataLink(oem, handle, version, sequence string) (*DataLink, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathDataLink, "exportByFqdn"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *DataLink

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[DataLink]{}).
				QueryParams(map[string]string{
					"oem":      oem,
					"handle":   handle,
					"version":  version,
					"sequence": sequence,
				}).Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *DataLink {
					var payload = resp.Result().(*clientPayload[DataLink])

					return &payload.Data
				}); err == nil {
					// broker chose to answer with successful empty definition, which is an error
					if result.Handle == "" {
						return nil, errorDatalinkUnknown
					}

					return result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) FindDataLink(oem, handle, version string) (*DataLink, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathDataLink, "getByFqdn"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *DataLink

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[DataLink]{}).
				QueryParams(map[string]string{
					"oem":     oem,
					"handle":  handle,
					"version": version,
				}).Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *DataLink {
					var payload = resp.Result().(*clientPayload[DataLink])

					return &payload.Data
				}); err == nil {
					return result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) FindSolution(oem, handle, version string) (*Solution, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathSolution, "readByFqdn"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *Solution

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[solutionSlices]{}).
				QueryParams(map[string]string{
					"oem":     oem,
					"handle":  handle,
					"version": version,
				}).Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *Solution {
					var payload = resp.Result().(*clientPayload[solutionSlices])

					return payload.Data.graph()
				}); err == nil {
					return result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) FindSolutionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSolution, "readByFqdn")
}

func (c *client) GetAuthToken() string {
	return c.AuthToken
}

func (c *client) GetExpiry() int {
	return c.Expiry
}

func (c *client) GetFunction(id int64) (*Function, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathFunction, "get"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *Function

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[Function]{}).
				QueryParams(map[string]string{
					"id": cast.ToString(id),
				}).
				Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *Function {
					var payload = resp.Result().(*clientPayload[Function])

					return &payload.Data
				}); err == nil {
					return result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) GetFunctionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathFunction, "get")
}

func (c *client) GetHostAddr() string {
	return c.HostAddr
}

func (c *client) GetTimeout() int {
	var bridge = c.requestBridge()

	return int(bridge.Timeout() / time.Second)
}

func (c *client) GetUserId() int {
	return c.UserId
}

func (c *client) ListDataLinks(oem, handle string, flags int) ([]DataLink, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathDataLink, "list"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *[]DataLink

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[[]DataLink]{}).
				QueryParams(map[string]string{
					"oem":    oem,
					"handle": handle,
					"flags":  cast.ToString(flags),
				}).
				Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *[]DataLink {
					var payload = resp.Result().(*clientPayload[[]DataLink])

					return &payload.Data
				}); err == nil {
					return *result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) ListDataLinksUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathDataLink, "list")
}

func (c *client) ListSolutions(oem string) ([]Solution, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathSolution, "list"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *[]Solution

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[[]Solution]{}).
				QueryParams(map[string]string{
					"oem":   oem,
					"mask":  cast.ToString(SolutionFlags.Released | SolutionFlags.Active),
					"flags": cast.ToString(SolutionFlags.Active),
				}).
				Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *[]Solution {
					var payload = resp.Result().(*clientPayload[[]Solution])

					return &payload.Data
				}); err == nil {
					return *result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) ListSolutionsUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSolution, "list")
}

func (c *client) ListWorkspaces(mask, flags int) ([]Workspace, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathWorkspace, "list"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge
			var result *[]Workspace

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[*workspaceList]{}).
				QueryParams(map[string]string{
					"mask":  cast.ToString(mask),
					"flags": cast.ToString(flags),
				}).
				Get(url)

			if err == nil {
				if result, err = resultOrError(resp, func(body any) *[]Workspace {
					var payload = resp.Result().(*clientPayload[*workspaceList])

					return &payload.Data.Workspaces
				}); err == nil {
					return *result, nil
				}
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) ListWorkspacesUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathWorkspace, "list")
}

func (c *client) Login(username string, password []byte) (*AuthSession, error) {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathSession, "create"); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge

		defer c.closeSilently(rb)
		resp, err = rb.Json().
			Resulting(&clientPayload[Session]{}).
			FormData(map[string]string{
				"email":    username,
				"password": string(password),
				"expiry":   strconv.FormatInt(int64(c.Expiry*60), 10),
			}).
			Post(url)

		if err == nil {
			return c.sessionOrError(resp, username)
		}
	}

	return nil, err
}

func (c *client) LoginUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSession, "create")
}

func (c *client) Logout(sessionId string) error {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathSession, "delete"); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge

		defer c.closeSilently(rb)
		resp, err = rb.Json().
			Cookie(c.makeCookie()).
			Resulting(&clientPayload[Session]{}).
			FormData(map[string]string{
				"id": sessionId,
			}).
			Post(url)

		if resp != nil && !resp.IsStatusSuccess() {
			if resp.Status() != "" {
				err = errors.New(resp.Status())
			} else {
				err = fmt.Errorf("could not reach %s", c.HostAddr)
			}
		}

		return err
	}

	return nil
}

func (c *client) LogoutUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSession, "delete")
}

func (c *client) OidcDeviceCode(deviceUrl string, deviceClient *DeviceClient) (*DeviceAuth, error) {
	if deviceUrl != "" {
		var rb = c.requestBridge()
		var resp responseBridge
		var err error

		defer c.closeSilently(rb)
		resp, err = rb.Form().
			Resulting(&DeviceAuth{}).
			FormData(map[string]string{
				"client_id": deviceClient.ClientId,
				"scope":     deviceClient.ClientScope,
			}).
			Post(deviceUrl)

		if err != nil {
			return nil, err
		}

		return resultOrError(resp, func(body any) *DeviceAuth {
			return resp.Result().(*DeviceAuth)
		})
	}

	return nil, errors.New("missing device url")
}

func (c *client) OidcDeviceUrl() (string, error) {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathOidcDevice); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge
		var result *string

		defer c.closeSilently(rb)
		resp, err = rb.Json().
			Cookie(c.makeCookie()).
			Resulting(&clientPayload[string]{}).
			Get(url)

		if err == nil {
			if result, err = resultOrError(resp, func(body any) *string {
				var payload = resp.Result().(*clientPayload[string])

				return &payload.Data
			}); err == nil {
				return *result, nil
			}
		}
	}

	return "", err
}

func (c *client) OidcTokenCreate(tokenUrl string, deviceCode string, deviceClient *DeviceClient) (string, error) {
	if tokenUrl != "" {
		var rb = c.requestBridge()
		var resp responseBridge
		var result *oauthResponse
		var err error

		defer c.closeSilently(rb)
		resp, err = rb.Form().
			Resulting(&oauthResponse{}).
			FormData(map[string]string{
				"client_id":   deviceClient.ClientId,
				"device_code": deviceCode,
				"grant_type":  deviceClient.GrantType,
			}).
			Post(tokenUrl)

		if err == nil {
			if result, err = resultOrError(resp, func(body any) *oauthResponse {
				return resp.Result().(*oauthResponse)
			}); err == nil {
				return result.AccessToken, nil
			}
		}

		return "", err
	}

	return "", errors.New("missing token url")
}

func (c *client) OidcTokenSession(tokenUrl, token string) (*AuthSession, error) {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathSession, "oidc"); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge

		defer c.closeSilently(rb)
		resp, _ = rb.Json().
			Cookie(c.makeCookie()).
			Resulting(&clientPayload[Session]{}).
			FormData(map[string]string{
				"accessToken": token,
			}).
			Post(url)

		return c.sessionOrError(resp, filepath.Base(tokenUrl))
	}

	return nil, err
}

func (c *client) OidcTokenUrl() (string, error) {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathOidcToken); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge
		var result *string

		defer c.closeSilently(rb)
		resp, err = rb.Json().
			Cookie(c.makeCookie()).
			Resulting(&clientPayload[string]{}).
			Get(url)

		if err == nil {
			if result, err = resultOrError(resp, func(body any) *string {
				var payload = resp.Result().(*clientPayload[string])

				return &payload.Data
			}); err == nil {
				return *result, nil
			}
		}
	}

	return "", err
}

func (c *client) ProvisionFunction(function *Function, extras map[string]any) (*shared.Identity, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathFunction, "provision"); err == nil {
			var functionBytes, _ = json.Marshal(function.toModel())
			var expandedMap map[string]any

			if expandedMap, err = jsonExpand(functionBytes, extras); err == nil {
				var modelBytes, _ = json.Marshal(expandedMap)
				var rb = c.requestBridge()
				var resp responseBridge

				defer c.closeSilently(rb)
				resp, err = rb.Json().
					Cookie(c.makeCookie()).
					Resulting(&clientPayload[provisionData]{}).
					FormData(map[string]string{
						"model": string(modelBytes),
					}).
					Post(url)

				return resultOrError(resp, func(body any) *shared.Identity {
					var payload = resp.Result().(*clientPayload[provisionData])
					var result = payload.Data.Sf.asIdentity()

					result.Auth = payload.Data.Auth
					return result
				})
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) ProvisionFunctionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathFunction, "provision")
}

func (c *client) PublishDataLink(link *DataLink) (*DataLink, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathDataLink, "publish"); err == nil {
			var modelBytes, _ = json.Marshal(link)
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[DataLink]{}).
				FormData(map[string]string{
					"model": string(modelBytes),
				}).
				Post(url)

			if err == nil {
				return resultOrError(resp, func(body any) *DataLink {
					var payload = resp.Result().(*clientPayload[DataLink])

					return &payload.Data
				})
			}
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) PublishDataLinkUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathDataLink, "publish")
}

func (c *client) PublishFunction(identity *shared.Identity) (*Function, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathFunction, "publish"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[publishingData]{}).
				FormData(map[string]string{
					"id":     identity.Id,
					"digest": identity.Hash,
				}).
				Post(url)

			return resultOrError(resp, func(body any) *Function {
				var payload = resp.Result().(*clientPayload[publishingData])

				return &payload.Data.Sf
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) PublishFunctionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathFunction, "publish")
}

func (c *client) PublishSolution(solution *Solution) (*Solution, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathOemSolution, "publish"); err == nil {
			var solutionBytes, _ = json.Marshal(solution)
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[solutionSlices]{}).
				FormData(map[string]string{
					"solution": string(solutionBytes),
				}).
				Post(url)

			return resultOrError(resp, func(body any) *Solution {
				var payload = resp.Result().(*clientPayload[solutionSlices])
				var graph = &payload.Data

				return &graph.Solution
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) PublishSolutionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathOemSolution, "publish")
}

func (c *client) Session() (*Session, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathSession, "get"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(c.makeCookie()).
				Resulting(&clientPayload[Session]{}).
				Get(url)

			return resultOrError(resp, func(body any) *Session {
				var payload = resp.Result().(*clientPayload[Session])

				return &payload.Data
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) SessionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSession, "get")
}

func (c *client) SessionValid(session *Session) bool {
	return session.Id > 0 && session.UserId == c.UserId
}

func (c *client) WithAccount(account *AuthAccount) (Client, error) {
	panicz.RequiresNotNil("account", account)

	if !account.IsExpired() {
		c.AuthToken = account.Token
		c.UserId = account.UserId
		return c, nil
	}

	return nil, errorSessionExpired
}

func (c *client) authFromCookie(name string, cookies []*http.Cookie) string {
	for _, cookie := range cookies {
		if cookie.Name == name {
			return cookie.Value
		}
	}

	return ""
}

func (c *client) closeSilently(client io.Closer) {
	_ = client.Close()
}

func (c *client) makeCookie() *http.Cookie {
	return &http.Cookie{Name: defaultCookieName, Value: c.AuthToken}
}

func (c *client) makeUrl(version version, path path, rpc ...string) (string, error) {
	if c.HostAddr == "" {
		return "", errorInvalidHost
	}

	return makeHostUrl(c.HostAddr, version, path, rpc...), nil
}

func (c *client) sessionOrError(resp responseBridge, username string) (*AuthSession, error) {
	var result *AuthSession
	var err error

	if resp.IsStatusSuccess() {
		var cookieValue = c.authFromCookie(defaultCookieName, resp.Cookies())

		if cookieValue == "" {
			err = errorNoToken
		} else {
			var payload = resp.Result().(*clientPayload[Session])

			c.AuthToken = cookieValue
			c.UserId = payload.Data.UserId
			result = NewAuthSession(&payload.Data, username, cookieValue)
		}
	} else if resp.Status() != "" {
		err = errors.New(resp.Status())
	} else {
		err = fmt.Errorf("could not reach %s", c.HostAddr)
	}

	return result, err
}

type ClientFactory struct {
	Active func(string) (Client, error)
	Get    func(string, string) (Client, error)
	New    func(string) Client
	Seed   func(string, string) (Client, error)
}

func ActiveClient(authFile string) (Client, error) {
	var auth = NewAuthData(authFile)

	if auth.Active >= 0 && len(auth.Accounts) > 0 {
		var account = auth.Accounts[auth.Active]
		var result Client

		if result = clientByHost[account.HostAddr]; result == nil {
			var err error

			if result, err = NewClient(account.HostAddr).WithAccount(account); err == nil {
				clientByHost[account.HostAddr] = result
				return result, nil
			}

			return nil, err
		}

		return result, nil
	}

	return nil, ErrorNoLogin
}

func GetClient(authFile string, addr string) (Client, error) {
	var key = sanitizeHostUrl(addr)
	var result Client

	if result = clientByHost[key]; result == nil {
		var auth = NewAuthData(authFile)
		var account *AuthAccount
		var err error

		if account, err = auth.Find(key); err == nil {
			if result, err = NewClient(key).WithAccount(account); err == nil {
				clientByHost[key] = result
				return result, nil
			}
		}

		return nil, err
	}

	return result, nil
}

func NewClient(addr string) Client {
	return newClientWithBridge(addr, defaultExpiryMinutes, func() requestBridge {
		var cl = resty.New().
			AddRequestMiddleware(protocolGateChecker).
			SetTimeout(defaultTimeoutSeconds * time.Second)

		return &restyBridge{
			client:  cl,
			request: cl.R(),
		}
	})
}

func SeedClient(addr string, authToken string) (Client, error) {
	var key = sanitizeHostUrl(addr)
	var account = &AuthAccount{
		HostAddr: addr,
		AuthSession: &AuthSession{
			Created:   -1,
			Expiry:    -1,
			SessionId: -1,
			Token:     authToken,
			UserId:    -1,
			Username:  "token",
		},
	}

	clientByHost[key], _ = NewClient(key).WithAccount(account)
	return clientByHost[key], nil
}

func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		Active: ActiveClient,
		Get:    GetClient,
		New:    NewClient,
		Seed:   SeedClient,
	}
}

func jsonExpand(bytes []byte, extras map[string]any) (map[string]any, error) {
	var expandedMap = map[string]any{}
	var err error

	if err = json.Unmarshal(bytes, &expandedMap); err == nil {
		var keys = slices.Collect(maps.Keys(expandedMap))

		for k, v := range extras {
			if !slices.ContainsFunc(keys, func(s string) bool {
				return strings.EqualFold(s, k)
			}) {
				expandedMap[k] = v
			}
		}

		return expandedMap, err
	}

	return nil, err
}

func makeHostUrl(host string, apiVersion version, path path, rpc ...string) string {
	var result []string

	if !strings.HasPrefix(host, "http") {
		result = append(result, env.DefaultProtocolPrefix(host))
	}

	result = append(result, sanitizeHostUrl(host), string(apiVersion), string(path))

	if len(rpc) > 0 {
		result = append(result, rpc...)
	}

	return strings.Join(result, "/")
}

func newClientWithBridge(addr string, expiry int, bridge func() requestBridge) Client {
	return &client{
		Expiry:        expiry,
		HostAddr:      addr,
		requestBridge: bridge,
	}
}

func resultOrError[T any](response responseBridge, transformer func(body any) *T) (*T, error) {
	if response != nil {
		if response.IsStatusSuccess() {
			return transformer(response.Result()), nil
		} else if response.StatusCode() == 400 {
			var bytes = response.Bytes()

			if bytes != nil {
				var respError Error

				if err := json.Unmarshal(bytes, &respError); err == nil {
					return nil, errors.New(respError.Message)
				}

				return nil, errors.New(string(bytes))
			}

			return nil, clientErrors[400]
		}

		return nil, mapz.GetOrDefault(clientErrors, response.StatusCode(), func() error {
			return fmt.Errorf("%d %s", response.StatusCode(), response.Status())
		})
	}

	return nil, errorNoResponse
}

func sanitizeHostUrl(hostUrl string) string {
	if strings.HasSuffix(hostUrl, "/") {
		return hostUrl[:len(hostUrl)-1]
	}

	return hostUrl
}
