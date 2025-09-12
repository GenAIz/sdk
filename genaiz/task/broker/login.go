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

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
)

var (
	ErrorNoAuth          = errors.New("auth not established")
	ErrorNoLogin         = errors.New("not logged in")
	ErrorNoSession       = errors.New("could not elect a session")
	ErrorSessionConflict = errors.New("could not choose a session to logout")
	ErrorSessionInvalid  = errors.New("session is not valid")
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

	return nil, ErrorNoSession
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
		if !strings.EqualFold(key, a.HostAddr) {
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
		defer filez.CloseSilently(fw)
		var buff = bufio.NewWriter(fw)
		var data, _ = yaml.Marshal(ad)

		_, err = buff.Write(data)
		panicz.PanicIfError(err)
		panicz.PanicIfError(buff.Flush())
		panicz.PanicIfError(os.Chmod(fw.Name(), 0600))
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
	Password []byte
	Username string
}

func NewAuthData(authFile ...string) *AuthData {
	if len(authFile) > 0 {
		var bytes []byte
		var err error

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
		Name:       "broker-session",
		OnPrepare:  handleSessionContext,
		OnComplete: handleSessionValidate,
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
		var brokerClient = clientFactory.New(params.HostAddr)
		var session *AuthSession

		state.Logger.Debugf("Creating session on url [%s]", brokerClient.LoginUrl())

		if session, err = brokerClient.Login(params.Username, params.Password); err == nil {
			var auth = NewAuthData(params.AuthFile)
			var newAuth = auth.Push(params.HostAddr, session)

			if err = newAuth.Write(params.AuthFile); err == nil {
				state.Output = session.Token
				return nil
			}
		}

		state.Logger.Errorf("Could not create session: [%s]", err)
		return err
	}

	return nil
}

func handleLoginDelete(params *LoginParams, state *task.State) error {
	if state.Output != "" {
		var auth = NewAuthData(params.AuthFile)
		var accounts []*AuthAccount
		var brokerClient Client
		var err error

		state.Logger.Debugf("Logging out session id [%s]", state.Output)

		if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
			if err = brokerClient.Logout(state.Output); err != nil {
				state.Logger.Warnf("Could not delete session for host [%s]: %s", params.HostAddr, err)
			}
		}

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

	return ErrorNoSession
}

func handleLoginPretend(params *LoginParams, state *task.State) error {
	if errors.Is(state.Error, ErrorNoAuth) {
		var brokerClient = clientFactory.New(params.HostAddr)

		state.Logger.Debugf("Pretending to login to [%s] with username [%s]", params.HostAddr, params.Username)
		fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("-d username=%s\\\n", params.Username)
		fmt.Printf("-d password=**********\\\n")
		fmt.Printf("-d expiry=%d\\\n", brokerClient.GetExpiry())
		fmt.Printf("%s\n", brokerClient.LoginUrl())
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

		if params.Username != "" {
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

			return ErrorSessionConflict
		}

		return err
	}

	return ErrorNoLogin
}

func handleLogoutPretend(params *LoginParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
			state.Logger.Debugf("Pretending to logout from session id [%s]", state.Output)
			state.Logger.Debugf("For host [%s]", params.HostAddr)
			state.Logger.Debugf("And username [%s]", params.Username)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("-d id=%s\\\n", state.Output)
			fmt.Printf("%s\n", brokerClient.LogoutUrl())
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

func handleSessionValidate(params *Broker, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
			var session *Session

			if session, err = brokerClient.Session(); err == nil {
				if brokerClient.SessionValid(session) {
					state.Logger.Debugf("Session [%d] validated for user [%d]", session.Id, session.UserId)
					return nil
				} else {
					err = ErrorSessionInvalid
				}
			}
		}

		state.Output = ""
		return err
	}

	return ErrorNoAuth
}
