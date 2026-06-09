package broker

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"
	"time"

	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
)

var (
	ErrorNoAuth          = task.NewError("auth not established")
	ErrorNoHostAddr      = task.NewError("host address not specified")
	ErrorNoLogin         = task.NewError("not logged in")
	ErrorNoSession       = task.NewError("could not elect a session")
	ErrorSessionConflict = task.NewError("could not choose a session to logout")
	ErrorSessionExists   = task.NewError("session exists")
	ErrorSessionExpired  = task.NewError("session is expired")
	ErrorSessionInvalid  = task.NewError("session is not valid")
	ErrorSessionsEmpty   = task.NewError("no sessions found")
)

type AuthAccount struct {
	*AuthSession
	HostAddr string
}

type AuthData struct {
	Active   int
	Accounts []*AuthAccount
}

func (ad *AuthData) AccountIndexForToken(token string) (int, *AuthAccount) {
	if i := slices.IndexFunc(ad.Accounts, func(account *AuthAccount) bool {
		return account.Token == token
	}); i >= 0 {
		return i, ad.Accounts[i]
	}

	return -1, nil
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
				Created:   session.Created,
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

type AuthParams struct {
	*Broker
	Expired bool
}

type AuthSession struct {
	Created   int64
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
		Created:   time.Now().UnixMilli(),
		Expiry:    session.Expiry,
		Token:     token,
		SessionId: session.Id,
		UserId:    session.UserId,
		Username:  username,
	}
}

func NewActivateTask() *task.Task[Broker] {
	return &task.Task[Broker]{
		Name:       "activate-auth",
		OnPrepare:  handleActivateContext,
		OnComplete: handleSessionActivate,
	}
}

func NewAuthTask() *task.Task[AuthParams] {
	return &task.Task[AuthParams]{
		Name:       "auth-data",
		OnPrepare:  handleAuthContext,
		OnComplete: handleAuthList,
	}
}

func NewLoginTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:       "broker-login",
		OnPrepare:  handleLoginContext,
		OnComplete: handleLoginCreate,
	}
}

func NewLogoutTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:       "broker-logout",
		OnPrepare:  handleLogoutContext,
		OnComplete: handleLoginDelete,
	}
}

func NewSessionTask() *task.Task[Broker] {
	return &task.Task[Broker]{
		Name:         "broker-session",
		OnPrepare:    handleSessionContext,
		OnIncomplete: handleSessionValidate,
		OnComplete:   handleSessionActivate,
	}
}

func handleActivateContext(params *Broker, state *task.State) error {
	if state.Output == "" {
		if params.HostAddr != "" {
			var auth = NewAuthData(params.AuthFile)

			state.Logger.Debugf("Looking for a session on [%s] under [%s]", params.HostAddr, params.AuthFile)

			if params.Username != "" {
				state.Logger.Debugf("Username [%s]", params.Username)
			}

			for _, acc := range auth.Accounts {
				if strings.EqualFold(acc.HostAddr, params.HostAddr) {
					if params.Username == "" ||
						(acc.Username == "token" && acc.Token[:10] == params.Username) ||
						strings.HasPrefix(acc.Username, params.Username) {
						state.Output = acc.Token
						return nil
					}
				}
			}

			return ErrorNoSession
		}

		return ErrorNoHostAddr
	}

	return nil
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
		var deletedIndex *int
		var err error

		state.Logger.Debugf("Logging out session id [%s]", state.Output)

		if brokerClient, err = params.GetClient(); err == nil {
			if err = brokerClient.Logout(state.Output); err != nil {
				state.Logger.Warnf("Could not delete session for host [%s]: %s", params.HostAddr, err)
			}
		}

		state.Logger.Debugf("Pruning session id [%s]", state.Output)

		for i, a := range auth.Accounts {
			if state.Output != fmt.Sprintf("%d", a.SessionId) {
				accounts = append(accounts, a)
			} else {
				state.Output = a.HostAddr
				deletedIndex = &i
			}
		}

		if deletedIndex != nil && *deletedIndex <= auth.Active && auth.Active > 0 {
			auth.Active -= 1
		}

		auth.Accounts = accounts
		return auth.Write(params.AuthFile)
	}

	return ErrorNoSession
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
		} else if auth.Active < size {
			account = auth.Accounts[auth.Active]
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

func handleSessionActivate(params *Broker, state *task.State) error {
	if state.Output != "" {
		var auth = NewAuthData(params.AuthFile)
		var err error

		if i, account := auth.AccountIndexForToken(state.Output); i >= 0 {
			if account.IsExpired() {
				return ErrorSessionExpired
			}

			state.Logger.Debugf("Activating session on host [%s] for user [%d]", params.HostAddr, account.UserId)
			auth.Active = i

			if err = auth.Write(params.AuthFile); err == nil {
				state.Output = account.Token
				return nil
			}

			return err
		}
	}

	return ErrorNoSession
}

func handleSessionContext(params *Broker, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)

	state.Logger.Debugf("Looking for accounts under [%s]", params.AuthFile)

	if params.HostAddr != "" {
		state.Logger.Debugf("Finding session for broker address [%s]", params.HostAddr)

		if account, err := auth.Find(params.HostAddr); err == nil {
			state.Output = account.Token
			return ErrorSessionExists
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

func handleAuthContext(params *AuthParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)

	if len(auth.Accounts) > 0 {
		state.Logger.Debug("Accounts found")
		return nil
	}

	return ErrorSessionsEmpty
}

func handleAuthList(params *AuthParams, state *task.State) error {
	var auth = NewAuthData(params.AuthFile)
	var effectiveAuth *AuthData

	if auth.Active < 0 || auth.Active >= len(auth.Accounts) {
		effectiveAuth = &AuthData{
			Active:   0,
			Accounts: auth.Accounts,
		}
		state.Logger.Debug("Invalid activated index, defaulting to 0")
	} else {
		effectiveAuth = auth
	}

	if !params.Expired {
		var accountsFiltered []*AuthAccount
		var activeFiltered = effectiveAuth.Active

		state.Logger.Debug("Filtering expired accounts")

		for i, acc := range effectiveAuth.Accounts {
			if !acc.IsExpired() {
				accountsFiltered = append(accountsFiltered, acc)

				if i == activeFiltered {
					activeFiltered = len(accountsFiltered) - 1
				}
			} else if i == activeFiltered {
				activeFiltered = 0
			}
		}

		state.Internal = &AuthData{
			Active:   activeFiltered,
			Accounts: accountsFiltered,
		}
		return nil
	}

	state.Internal = effectiveAuth
	return nil
}

func handleSessionValidate(params *Broker, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var session *Session

			if session, err = brokerClient.Session(); err == nil {
				if brokerClient.SessionValid(session) {
					state.Logger.Debugf("Session [%d] validated for user [%d]", session.Id, session.UserId)
					return nil
				}

				err = ErrorSessionInvalid
			}
		}

		state.Output = ""
		return err
	}

	return ErrorNoAuth
}
