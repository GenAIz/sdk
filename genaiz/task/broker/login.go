package broker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
)

var (
	ErrorNoAuth    = errors.New("auth not established")
	ErrorNoLogin   = errors.New("not logged in")
	ErrorNoSession = errors.New("could not elect a session")
)

type AuthAccount struct {
	*AuthSession
	HostAddr string
}

type AuthData struct {
	Active   int
	Accounts []*AuthAccount
}

func (ad *AuthData) Find(hostAddr string) (*AuthAccount, error) {
	var key = sanitizeHostUrl(hostAddr)

	for _, account := range ad.Accounts {
		if strings.EqualFold(account.HostAddr, key) {
			return account, nil
		}
	}

	return nil, errors.New("no broker session found")
}

func (ad *AuthData) ForHostUser(host string, username string) (*AuthAccount, error) {
	var accounts []*AuthAccount

	if host == "" {
		accounts = ad.Accounts
	} else {
		for _, account := range ad.Accounts {
			if strings.EqualFold(account.HostAddr, host) {
				accounts = append(accounts, account)
			}
		}
	}

	if username != "" {
		for _, account := range accounts {
			if strings.EqualFold(account.Username, username) {
				return account, nil
			}
		}
	}

	return nil, ErrorNoSession
}

func (ad *AuthData) ForToken(token string) (*AuthAccount, error) {
	for _, account := range ad.Accounts {
		if account.Token == token {
			return account, nil
		}
	}

	return nil, ErrorNoSession
}

func (ad *AuthData) Push(hostAddr string, session *AuthSession) *AuthData {
	var key = sanitizeHostUrl(hostAddr)
	var accounts = []*AuthAccount{
		{
			AuthSession: &AuthSession{
				Expiry:    session.Expiry,
				SessionId: session.SessionId,
				Token:     session.Token,
				UserId:    session.UserId,
				Username:  session.Username,
			},
			HostAddr: key,
		},
	}

	for _, a := range ad.Accounts {
		if !strings.EqualFold(key, hostAddr) {
			accounts = append(accounts, a)
		}
	}

	return &AuthData{
		Active:   0,
		Accounts: accounts,
	}
}

func (ad *AuthData) Write(outFile string) error {
	var dir = filepath.Dir(outFile)
	var file = filepath.Base(outFile)
	var fw *os.File
	var err error

	if fw, err = filez.CreateRecursive(dir, file); err == nil {
		var buff = bufio.NewWriter(fw)
		var data []byte

		defer filez.CloseSilently(fw)
		if data, err = yaml.Marshal(ad); err == nil {
			if _, err = buff.Write(data); err == nil {
				err = buff.Flush()
			}
		}

		_ = os.Chmod(fw.Name(), 0600)
	}

	return err
}

type AuthSession struct {
	Expiry    int64
	SessionId int64
	Token     string
	UserId    int
	Username  string
}

func (s *AuthSession) IsExpired() bool {
	return s.Expiry != -1 && s.Expiry <= time.Now().UTC().UnixMilli()
}

type LoginParams struct {
	*Broker
	Password *[]byte
	Username string
}

func NewAuthData(authFile ...string) *AuthData {
	var bytes []byte
	var err error

	if len(authFile) > 0 {
		if bytes, err = os.ReadFile(authFile[0]); err == nil {
			var auth AuthData

			if err = yaml.Unmarshal(bytes, &auth); err == nil {
				return &auth
			}
		}
	}

	return &AuthData{}
}

func NewAuthSession(session *Session, username string, token string) *AuthSession {
	return &AuthSession{
		Expiry:    session.Expiry,
		Token:     token,
		SessionId: session.Id,
		UserId:    session.UserId,
		Username:  username,
	}
}

func NewLoginTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:       "broker-login",
		OnPrepare:  handleLoginContext,
		OnComplete: handleLoginCreate,
		OnPretend:  handleLoginPretend,
	}
}

func NewLogoutTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:       "broker-logout",
		OnPrepare:  handleLogoutContext,
		OnComplete: handleLoginDelete,
		OnPretend:  handleLogoutPretend,
	}
}

func NewSessionTask() *task.Task[Broker] {
	return &task.Task[Broker]{
		Name:      "broker-session",
		OnPrepare: handleSessionContext,
	}
}

func handleLoginContext(params *LoginParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)
	var accounts []*AuthAccount

	state.Logger.Debugf("Pruning sessions on host [%s]", params.HostAddr)

	for _, a := range auth.Accounts {
		if !strings.EqualFold(a.HostAddr, sanitizeHostUrl(params.HostAddr)) {
			accounts = append(accounts, a)
		}
	}

	auth.Accounts = accounts
	return auth.Write(params.AuthFile)
}

func handleLoginCreate(params *LoginParams, state *task.State) error {
	var err error

	if state.Output == "" {
		var client = NewClient(params.HostAddr)
		var session *AuthSession

		state.Logger.Debugf("Creating session on url [%s]", client.loginUrl())

		if session, err = client.Login(params.Username, params.Password); err == nil {
			var auth = NewAuthData(params.AuthFile)
			var newAuth = auth.Push(params.HostAddr, session)

			if err = newAuth.Write(params.AuthFile); err == nil {
				state.Output = session.Token
				return nil
			}
		}
	}

	state.Logger.Errorf("Could not create session: [%s]", err)
	return err
}

func handleLoginDelete(params *LoginParams, state *task.State) error {
	if state.Output != "" {
		var client *Client
		var err error

		state.Logger.Debugf("Logging out session id [%s]", state.Output)

		if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
			if err = client.Logout(state.Output); err == nil {
				var auth = NewAuthData(params.AuthFile)
				var accounts []*AuthAccount

				state.Logger.Debugf("Pruning session id [%s]", state.Output)

				for _, a := range auth.Accounts {
					if state.Output != fmt.Sprintf("%d", a.SessionId) {
						accounts = append(accounts, a)
					}
				}

				auth.Accounts = accounts
				state.Output = params.Username
				return auth.Write(params.AuthFile)
			}
		}

		return err
	}

	return ErrorNoSession
}

func handleLoginPretend(params *LoginParams, state *task.State) error {
	if errors.Is(state.Error, ErrorNoAuth) {
		var client = NewClient(params.HostAddr)

		state.Logger.Debugf("Pretending to login to [%s] with username [%s]", params.HostAddr, params.Username)
		fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("-d username=%s\\\n", params.Username)
		fmt.Printf("-d password=**********\\\n")
		fmt.Printf("-d expiry=%d\\\n", client.Expiry)
		fmt.Printf("%s\n", client.loginUrl())
		return nil
	}

	state.Logger.Debugf("Would not pretend login, since auth to [%s] is already established", params.HostAddr)
	return nil
}

func handleLogoutContext(params *LoginParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)
	var size = len(auth.Accounts)

	state.Logger.Debugf("Looking for sessions under [%s]", params.AuthFile)

	if size != 0 {
		var account *AuthAccount
		var err error

		if params.HostAddr != "" || params.Username != "" {
			account, err = auth.ForHostUser(params.HostAddr, params.Username)
		} else if size == 1 {
			account = auth.Accounts[0]
		}

		if err == nil {
			if account != nil {
				params.Username = account.Username
				params.HostAddr = account.HostAddr
				state.Output = fmt.Sprintf("%d", account.SessionId)
				return nil
			}

			return ErrorNoSession
		}
	}

	return ErrorNoLogin
}

func handleLogoutPretend(params *LoginParams, state *task.State) error {
	if state.Output != "" {
		var client *Client
		var err error

		if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
			state.Logger.Debugf("Pretending to logout from session id [%s]", state.Output)

			if params.HostAddr != "" {
				state.Logger.Debugf("For host [%s]", params.HostAddr)
			}

			if params.Username != "" {
				state.Logger.Debugf("And username [%s]", params.Username)
			}

			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("-d id=%s\\\n", state.Output)
			fmt.Printf("%s\n", client.logoutUrl())
			return nil
		}

		return err
	}

	return state.Error
}

func handleSessionContext(params *Broker, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)

	state.Logger.Debugf("Looking for accounts under [%s]", params.AuthFile)

	if params.HostAddr != "" {
		state.Logger.Debugf("Finding session for broker address [%s]", params.HostAddr)

		if account, err := auth.Find(params.HostAddr); err == nil {
			state.Output = account.Token
			return nil
		} else {
			return err
		}
	} else if auth.Active <= len(auth.Accounts)-1 && !auth.Accounts[auth.Active].IsExpired() {
		var account = auth.Accounts[auth.Active]

		state.Logger.Debugf("Found session for active broker address [%s]", account.HostAddr)
		state.Output = account.Token
		return nil
	}

	return ErrorNoAuth
}
