package broker

import (
	"bufio"
	"os"
	"strings"

	"github.com/pkg/errors"

	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz/task"
)

type AuthAccount struct {
	*Session
	HostAddr string
}

type AuthData struct {
	Active   int
	Accounts []*AuthAccount
}

func (ad *AuthData) Push(hostAddr string, session *Session) *AuthData {
	var accounts = []*AuthAccount{
		{
			Session: &Session{
				Expiry: session.Expiry,
				Token:  session.Token,
			},
			HostAddr: hostAddr,
		},
	}

	for _, a := range ad.Accounts {
		if !strings.EqualFold(a.HostAddr, hostAddr) {
			accounts = append(accounts, a)
		}
	}

	return &AuthData{
		Active:   0,
		Accounts: accounts,
	}
}

func (ad *AuthData) Write(outFile string) error {
	var fw *os.File
	var err error

	if fw, err = os.Create(outFile); err == nil {
		var buff = bufio.NewWriter(fw)
		var data []byte

		if data, err = yaml.Marshal(ad); err == nil {
			if _, err = buff.Write(data); err == nil {
				err = buff.Flush()
			}
		}
	}

	return err
}

type LoginParams struct {
	AuthFile string
	HostAddr string
	Password *[]byte
	Refresh  bool
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

func NewLoginTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:         "broker-login",
		OnPrepare:    handleLoginContext,
		OnIncomplete: handleLoginCreate,
		OnPretend:    handleLoginPretend,
	}
}

func handleLoginContext(params *LoginParams, state *task.State) error {
	if !params.Refresh {
		var auth = NewAuthData(params.AuthFile)
		var size = len(auth.Accounts)

		if size > 0 {
			if params.HostAddr != "" {
				for _, a := range auth.Accounts {
					if strings.EqualFold(a.HostAddr, params.HostAddr) && !a.IsExpired() {
						state.Output = a.Token
						return nil
					}
				}
			} else if auth.Active <= size-1 && !auth.Accounts[auth.Active].IsExpired() {
				state.Output = auth.Accounts[auth.Active].Token
				return nil
			}
		}
	}

	return errors.New("auth not established")
}

func handleLoginCreate(params *LoginParams, state *task.State) error {
	var err error

	if params.Refresh || state.Output == "" {
		var session *Session
		var client = NewClient(params.HostAddr)

		if session, err = client.Login(params.Username, params.Password); err == nil {
			var auth = NewAuthData(params.AuthFile)
			var newAuth = auth.Push(params.HostAddr, session)

			if err = newAuth.Write(params.AuthFile); err == nil {
				state.Output = session.Token
				return nil
			}
		}
	}

	return err
}

func handleLoginPretend(params *LoginParams, state *task.State) error {
	return nil
}
