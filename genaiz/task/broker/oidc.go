package broker

import (
	"bufio"
	"bytes"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	"genaiz.com/genaiz-lib/browser"
	"genaiz.com/genaiz/task"
)

const (
	oidcClientId    = "com.genaiz.sdk"
	oidcClientScope = "openid profile email"
	oidcGrantType   = "urn:ietf:params:oauth:grant-type:device_code"
)

var (
	oidcDeviceClient      = NewDeviceClient(oidcClientId, oidcClientScope, oidcGrantType)
	oidcDefaultUrlHandler = handleOidcCopyPasteUrl

	ErrorOidcNotSupported = errors.New("broker does not support OIDC logins")
	ErrorOidcPending      = errors.New("authorization_pending")
)

type OidcParams struct {
	*Broker
	Input           io.Reader
	Output          io.Writer
	BrowserRedirect bool
}

func (op OidcParams) GetInput() io.Reader {
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

func (op OidcParams) GetUrlHandler() func(*OidcParams, *DeviceAuth) error {
	if op.BrowserRedirect {
		return handleOidcBrowserUrl
	}

	return oidcDefaultUrlHandler
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
		var urlHandler = params.GetUrlHandler()
		var deviceAuth *DeviceAuth
		var err error

		state.Logger.Debugf("Obtaining oidc device code from [%s]", oidcState.deviceUrl)

		if deviceAuth, err = brokerClient.OidcDeviceCode(oidcState.deviceUrl, oidcDeviceClient); err == nil {
			state.Logger.Debugf("Authorizing oidc device at [%s]", deviceAuth.VerificationUriComplete)

			if err = urlHandler(params, deviceAuth); err == nil {
				var authToken string

				for {
					if authToken, err = brokerClient.OidcTokenCreate(oidcState.tokenUrl, deviceAuth.DeviceCode, oidcDeviceClient); err == nil ||
						!strings.EqualFold(err.Error(), ErrorOidcPending.Error()) {
						break
					}

					if err = handleOidcPendingUrl(params, deviceAuth); err != nil {
						break
					}
				}

				if err == nil {
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

func handleOidcBrowserUrl(params *OidcParams, deviceAuth *DeviceAuth) error {
	var errOut = bufio.NewWriter(new(bytes.Buffer))
	var out = params.GetOutput()
	var err error

	if err = browser.OpenUrl(deviceAuth.VerificationUriComplete, errOut, out); err == nil {
		if _, err = fmt.Fprintf(out, "Redirecting authorization to URL: %s\n", deviceAuth.VerificationUriComplete); err == nil {
			handleOidcShellQuery(params)
		}
	} else {
		return handleOidcCopyPasteUrl(params, deviceAuth)
	}

	return err
}

func handleOidcCopyPasteUrl(params *OidcParams, deviceAuth *DeviceAuth) error {
	var out = params.GetOutput()
	var err error

	if _, err = fmt.Fprintf(out, "Authorize with the following URL: %s\n", deviceAuth.VerificationUriComplete); err == nil {
		handleOidcShellQuery(params)
	}

	return err
}

func handleOidcPendingUrl(params *OidcParams, deviceAuth *DeviceAuth) error {
	var out = params.GetOutput()
	var err error

	if _, err = fmt.Fprintf(out, "Waiting for authorization on URL: %s\n", deviceAuth.VerificationUriComplete); err == nil {
		handleOidcShellQuery(params)
	}

	return err
}

func handleOidcShellQuery(params *OidcParams) {
	var out = params.GetOutput()
	var buff = bufio.NewReader(params.GetInput())
	var message = "Press enter after login...\n"
	var keyPressed rune

	for {
		_, _ = fmt.Fprint(out, message)
		keyPressed, _, _ = buff.ReadRune()

		if keyPressed == '\n' {
			return
		}
	}
}
