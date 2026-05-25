package broker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"genaiz.com/genaiz-lib/browser"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/stdz"
	"genaiz.com/genaiz/task"
)

const (
	oidcClientId       = "com.genaiz.sdk"
	oidcClientScope    = "openid profile email"
	oidcDefaultTimeout = 3500 * time.Millisecond
	oidcDefaultTries   = 4
	oidcGrantType      = "urn:ietf:params:oauth:grant-type:device_code"
)

var (
	oidcDeviceClient = NewDeviceClient(oidcClientId, oidcClientScope, oidcGrantType)

	ErrorOidcNotSupported  = errors.New("broker does not support OIDC logins")
	ErrorOidcPending       = errors.New("authorization_pending")
	ErrorOidcSlowDown      = errors.New("slow_down")
	ErrorOidcClientTimeout = errors.New("client timeout, too many attempts")
)

type OidcParams struct {
	*Broker
	Input           io.ReadCloser
	Output          io.Writer
	BrowserRedirect bool
	Tries           *int
	Timeout         *time.Duration
}

func (op OidcParams) GetInput() io.ReadCloser {
	if op.Input == nil {
		return os.Stdin
	}

	return op.Input
}

func (op OidcParams) GetOutput() io.Writer {
	if op.Output == nil {
		return os.Stdout
	}

	return op.Output
}

func (op OidcParams) GetTimeout() time.Duration {
	if op.Timeout != nil {
		return *op.Timeout
	}

	return oidcDefaultTimeout
}

func (op OidcParams) GetTries() int {
	if op.Tries != nil {
		return *op.Tries
	}

	return oidcDefaultTries
}

type oidcTracking struct {
	deviceUrl string
	tokenUrl  string
}

type OidcState struct {
	oidcTracking
	state *task.State
}

func (os *OidcState) IsSupported() bool {
	return os.deviceUrl != "" && os.tokenUrl != ""
}

func (os *OidcState) SetDeviceUrl(deviceUrl string) {
	os.deviceUrl = deviceUrl
	os.state.Internal = os.oidcTracking
}

func (os *OidcState) SetTokenUrl(tokenUrl string) {
	os.tokenUrl = tokenUrl
	os.state.Internal = os.oidcTracking
}

type OidcUrlHandler struct {
	auth      *DeviceAuth
	params    *OidcParams
	pollInput stdz.Input
	once      sync.Once
	timeout   time.Duration
	tries     int
}

func (ou *OidcUrlHandler) getInitHandler() func() error {
	if ou.params.BrowserRedirect {
		return ou.initBrowserRedirectUrl
	}

	return ou.initCopyPasteUrl
}

func (ou *OidcUrlHandler) initBrowserRedirectUrl() error {
	var errOut = bufio.NewWriter(new(bytes.Buffer))
	var out = ou.params.GetOutput()
	var err error

	if err = browser.OpenUrl(ou.auth.VerificationUriComplete, errOut, out); err == nil {
		_, err = fmt.Fprintf(out, "Redirecting authorization to URL: %s\n", ou.auth.VerificationUriComplete)
	} else {
		return ou.initCopyPasteUrl()
	}

	return err
}

func (ou *OidcUrlHandler) initCopyPasteUrl() error {
	var out = ou.params.GetOutput()

	_, err := fmt.Fprintf(out, "Authorize with the following URL: %s\n", ou.auth.VerificationUriComplete)
	return err
}

func (ou *OidcUrlHandler) Init() error {
	return ou.getInitHandler()()
}

func (ou *OidcUrlHandler) PollForToken(brokerClient Client, tokenUrl string, deviceClient *DeviceClient) (string, error) {
	var token string
	var err error

	ou.once.Do(func() {
		ou.pollInput = stdz.NewInput(ou.params.GetInput())
	})

	defer filez.CloseSilently(ou.pollInput)

	if _, err = fmt.Fprintln(ou.params.GetOutput(), "Press enter after login..."); err != nil {
		return token, err
	}

	for {
		if ou.tries > 0 {
			ou.pollInput.Poll(ou.timeout, "\n", func() {
				if token, err = brokerClient.OidcTokenCreate(tokenUrl, ou.auth.DeviceCode, deviceClient); err != nil {
					if strings.EqualFold(err.Error(), ErrorOidcPending.Error()) {
						err = nil
					} else if strings.EqualFold(err.Error(), ErrorOidcSlowDown.Error()) {
						ou.timeout += 1500 * time.Millisecond
						err = nil
					}
				}
			})

			if err != nil || token != "" {
				break
			}

			ou.tries -= 1
		} else {
			err = ErrorOidcClientTimeout
			break
		}
	}

	return token, err
}

func NewOidcState(state *task.State) *OidcState {
	var current, ok = state.Internal.(oidcTracking)
	var deviceUrl, tokenUrl string

	if ok {
		deviceUrl = current.deviceUrl
		tokenUrl = current.tokenUrl
	}

	return &OidcState{
		oidcTracking: oidcTracking{
			deviceUrl: deviceUrl,
			tokenUrl:  tokenUrl,
		},
		state: state,
	}
}

func NewOidcTask() *task.Task[OidcParams] {
	return &task.Task[OidcParams]{
		Name:       "broker-oidc",
		OnPrepare:  handleOidcContext,
		OnComplete: handleOidcCreate,
	}
}

func NewOidcUrlHandler(params *OidcParams, auth *DeviceAuth) *OidcUrlHandler {
	return &OidcUrlHandler{
		auth:    auth,
		once:    sync.Once{},
		params:  params,
		timeout: params.GetTimeout(),
		tries:   params.GetTries(),
	}
}

func handleOidcContext(params *OidcParams, state *task.State) error {
	var err error

	if err = handleLoginContext(&LoginParams{Broker: params.Broker}, state); err == nil {
		var oidcState = NewOidcState(state)
		var brokerClient = clientFactory.New(params.HostAddr)
		var deviceUrl string

		state.Logger.Debugf("Querying oidc coordinates on broker [%s]", params.HostAddr)

		if deviceUrl, err = brokerClient.OidcDeviceUrl(); err == nil {
			var tokenUrl string

			if tokenUrl, err = brokerClient.OidcTokenUrl(); err == nil {
				oidcState.SetDeviceUrl(deviceUrl)
				oidcState.SetTokenUrl(tokenUrl)
				return nil
			}
		}

		return ErrorOidcNotSupported
	}

	return err
}

func handleOidcCreate(params *OidcParams, state *task.State) error {
	var oidcState = NewOidcState(state)

	if oidcState.IsSupported() {
		var brokerClient = clientFactory.New(params.HostAddr)
		var deviceAuth *DeviceAuth
		var err error

		state.Logger.Debugf("Obtaining oidc device code from [%s]", oidcState.deviceUrl)

		if deviceAuth, err = brokerClient.OidcDeviceCode(oidcState.deviceUrl, oidcDeviceClient); err == nil {
			var urlHandler = NewOidcUrlHandler(params, deviceAuth)

			state.Logger.Debugf("Authorizing oidc device at [%s]", deviceAuth.VerificationUriComplete)

			if err = urlHandler.Init(); err == nil {
				var authToken string

				if authToken, err = urlHandler.PollForToken(brokerClient, oidcState.tokenUrl, oidcDeviceClient); err == nil {
					var session *AuthSession

					state.Logger.Debugf("Creating session token for broker [%s]", params.HostAddr)

					if session, err = brokerClient.OidcTokenSession(oidcState.tokenUrl, authToken); err == nil {
						var auth = NewAuthData(params.AuthFile)
						var newAuth = auth.Push(params.HostAddr, session)

						if err = newAuth.Write(params.AuthFile); err == nil {
							state.Output = session.Token
							return nil
						}
					}
				}
			}
		}

		return err
	}

	return ErrorOidcNotSupported
}
