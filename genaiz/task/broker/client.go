package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strconv"
	"strings"
	"time"

	"resty.dev/v3"

	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task/shared"
)

const (
	defaultExpiryMinutes  = 5 * 24 * 60
	defaultTimeoutSeconds = 15

	apiVersion1  version = "v1"
	pathFunction path    = "sf"
	pathSession  path    = "user/session"
	pathSolution path    = "oem/solution"
)

var (
	errorBadRequest     = errors.New("broker refused the request")
	errorForbidden      = errors.New("broker denied access")
	errorInternal       = errors.New("broker request crashed")
	errorInvalidHost    = errors.New("invalid host address")
	errorNoAuth         = errors.New("client not authenticated")
	errorNotFound       = errors.New("broker did not find the entity")
	errorNoResponse     = errors.New("client did not get a response")
	errorNoPath         = errors.New("broker path is not available")
	errorNoToken        = errors.New("could not read token from cookie")
	errorSessionExpired = errors.New("broker session is expired")
	errorUnauthorized   = errors.New("unauthorized, please login")

	clientByHost = map[string]Client{}
	clientErrors = map[int]error{
		400: errorBadRequest,
		401: errorUnauthorized,
		403: errorForbidden,
		404: errorNotFound,
		500: errorInternal,
		501: errorNoPath,
	}
	clientFactory = NewClientFactory()
)

type path string

type version string

type Client interface {
	GetAuthToken() string

	GetExpiry() int

	GetTimeout() int

	Login(string, []byte) (*AuthSession, error)

	LoginUrl() string

	Logout(string) error

	LogoutUrl() string

	ProvisionFunction(*Function) (*shared.Identity, error)

	ProvisionFunctionUrl() string

	PublishFunction(*shared.Identity) (*Function, error)

	PublishFunctionUrl() string

	PublishSolution(*Solution) (*shared.Identity, error)

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

type solutionSlices struct {
	Solution       SolutionRemote `json:"solution"`
	Workflows      []Workflow     `json:"workflows"`
	WorkflowLinks  []WorkflowLink `json:"workflowLinks"`
	WorkflowNodes  []WorkflowNode `json:"workflowNodes"`
	SmartFunctions []Function     `json:"smartFunctions"`
}

func (c *client) GetAuthToken() string {
	return c.AuthToken
}

func (c *client) GetExpiry() int {
	return c.Expiry
}

func (c *client) GetTimeout() int {
	var bridge = c.requestBridge()

	return int(bridge.Timeout() / time.Second)
}

func (c *client) Login(username string, password []byte) (*AuthSession, error) {
	var url string
	var err error

	if url, err = c.makeUrl(apiVersion1, pathSession, "create"); err == nil {
		var rb = c.requestBridge()
		var resp responseBridge
		var result *AuthSession

		defer c.closeSilently(rb)
		resp, err = rb.Json().
			Resulting(&clientPayload[Session]{}).
			Params(map[string]string{
				"email":    username,
				"password": string(password),
				"expiry":   strconv.FormatInt(int64(c.Expiry*60), 10),
			}).
			Post(url)

		if resp != nil {
			if resp.IsSuccess() {
				var cookieValue = c.authFromCookie("s", resp.Cookies())

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
		}

		return result, err
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
			Cookie(&http.Cookie{Name: "s", Value: c.AuthToken}).
			Resulting(&clientPayload[Session]{}).
			Params(map[string]string{
				"id": sessionId,
			}).
			Post(url)

		if resp != nil && !resp.IsSuccess() {
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

func (c *client) ProvisionFunction(function *Function) (*shared.Identity, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathFunction, "provision"); err == nil {
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(&http.Cookie{Name: "s", Value: c.AuthToken}).
				Resulting(&clientPayload[Provision]{}).
				Params(map[string]string{
					"name":        function.Name,
					"description": function.Description,
					"oem":         function.Oem,
					"handle":      function.Handle,
					"type":        function.Type,
					"version":     function.Version,
					"arches":      strings.Join(function.Arches, ","),
				}).
				Post(url)

			return resultOrError(resp, func(body any) *shared.Identity {
				var payload = resp.Result().(*clientPayload[Provision])
				var result = payload.Data.Sf.asIdentity()

				result.Auth = payload.Data.Auth
				return result
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) ProvisionFunctionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathFunction, "provision")
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
				Cookie(&http.Cookie{Name: "s", Value: c.AuthToken}).
				Resulting(&clientPayload[Function]{}).
				Params(map[string]string{
					"id":     identity.Id,
					"digest": identity.Hash,
				}).
				Post(url)

			return resultOrError(resp, func(body any) *Function {
				var payload = resp.Result().(*clientPayload[Function])

				return &payload.Data
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) PublishFunctionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathFunction, "publish")
}

func (c *client) PublishSolution(solution *Solution) (*shared.Identity, error) {
	if c.AuthToken != "" {
		var url string
		var err error

		if url, err = c.makeUrl(apiVersion1, pathSolution, "publish"); err == nil {
			var solutionBytes, _ = json.Marshal(solution)
			var rb = c.requestBridge()
			var resp responseBridge

			defer c.closeSilently(rb)
			resp, err = rb.Json().
				Cookie(&http.Cookie{Name: "s", Value: c.AuthToken}).
				Resulting(&clientPayload[solutionSlices]{}).
				Params(map[string]string{
					"solution": string(solutionBytes),
				}).
				Post(url)

			return resultOrError(resp, func(body any) *shared.Identity {
				var payload = resp.Result().(*clientPayload[solutionSlices])
				var graph = &payload.Data

				return graph.Solution.asIdentity()
			})
		}

		return nil, err
	}

	return nil, errorNoAuth
}

func (c *client) PublishSolutionUrl() string {
	return makeHostUrl(c.HostAddr, apiVersion1, pathSolution, "publish")
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
				Cookie(&http.Cookie{Name: "s", Value: c.AuthToken}).
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

func (c *client) makeUrl(version version, path path, rpc ...string) (string, error) {
	if c.HostAddr == "" {
		return "", errorInvalidHost
	}

	return makeHostUrl(c.HostAddr, version, path, rpc...), nil
}

type ClientFactory struct {
	New func(string) Client
	Get func(string, string) (Client, error)
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
		var cl = resty.New().SetTimeout(defaultTimeoutSeconds * time.Second)

		return &restyBridge{
			client:  cl,
			request: cl.R(),
		}
	})
}

func NewClientFactory() *ClientFactory {
	return &ClientFactory{
		New: NewClient,
		Get: GetClient,
	}
}

func makeHostUrl(host string, version version, path path, rpc ...string) string {
	var result []string

	if !strings.HasPrefix(host, "http") {
		result = append(result, "http:/")
	}

	result = append(result, sanitizeHostUrl(host), string(version), string(path))

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
		if response.IsSuccess() {
			return transformer(response.Result()), nil
		} else if response.StatusCode() == 400 {
			var bytes = response.Bytes()

			if bytes != nil {
				var respError Error

				if err := json.Unmarshal(bytes, &respError); err == nil {
					return nil, errors.New(respError.Message)
				} else {
					return nil, errors.New(string(bytes))
				}
			}

			return nil, clientErrors[400]
		} else {
			return nil, mapz.GetOrDefault(clientErrors, response.StatusCode(), func() error {
				return fmt.Errorf("%d %s", response.StatusCode(), response.Status())
			})
		}
	}

	return nil, errorNoResponse
}

func sanitizeHostUrl(hostUrl string) string {
	if strings.HasSuffix(hostUrl, "/") {
		return hostUrl[:len(hostUrl)-1]
	}

	return hostUrl
}
