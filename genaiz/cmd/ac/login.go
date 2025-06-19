package ac

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LoginTaskFactory func() *task.Task[broker.LoginParams]

type SessionTaskFactory func() *task.Task[broker.Broker]

type LoginExecutor struct {
	Repo *config.Repo

	optionPassword *config.StringOption
	optionRefresh  *config.BoolOption
	optionUsername *config.StringOption

	loginTaskFactory   LoginTaskFactory
	sessionTaskFactory SessionTaskFactory
}

func (le *LoginExecutor) Login(brokerAddr string) {
	var currentSession string
	var refresh = le.Repo.GetBool(le.optionRefresh)
	var brokerParam = le.makeBrokerParams(brokerAddr)

	if !refresh {
		var sessionPlan = &task.Plan{
			Logger: le.Repo.Logger,
			OnFailure: func(msg interface{}) {
				refresh = true
			},
			OnSuccess: func(msg interface{}) {
				currentSession = cast.ToString(msg)
			},
		}

		task.Attempt(sessionPlan, brokerParam, le.sessionTaskFactory())

		if currentSession != "" {
			fmt.Printf("%s\n", currentSession)
		}
	}

	if refresh {
		var params = le.makeLoginParams(brokerParam)
		var loginPlan = task.NewPlan("Login", le.Repo.Logger)

		task.Single(loginPlan, params, le.loginTaskFactory())
	}

}

func (le *LoginExecutor) makeBrokerParams(brokerAddr string) *broker.Broker {
	return &broker.Broker{
		AuthFile: le.Repo.AuthFile,
		HostAddr: brokerAddr,
	}
}

func (le *LoginExecutor) makeLoginParams(brokerParam *broker.Broker) *broker.LoginParams {
	return &broker.LoginParams{
		Broker:   brokerParam,
		Username: le.queryUsername(),
		Password: le.queryPassword(),
	}
}

func (le *LoginExecutor) queryPassword() *[]byte {
	var password = le.Repo.GetString(le.optionPassword)
	var result *[]byte

	if password == "" {
		result = le.Repo.QuerySecret("password: ")
	} else {
		var bytes = []byte(password)

		result = &bytes
	}

	return result
}

func (le *LoginExecutor) queryUsername() string {
	var username = le.Repo.GetString(le.optionUsername)

	if username == "" {
		username = le.Repo.QueryMandatory("username: ")
	}

	return strings.TrimSpace(username)
}

func NewLogin(repo *config.Repo) *cobra.Command {
	var exec = NewLoginExecutor(repo)
	var login = &cobra.Command{
		Use:     "login HOST",
		Short:   "Authenticates an account with a Genaiz broker",
		Long:    "Authenticates a username and password with a Genaiz broker provided a url as argument",
		Example: "genaiz ac login www.genaiz.com",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			exec.Login(args[0])
		},
	}

	repo.Register(login, exec.optionUsername, exec.optionPassword, exec.optionRefresh)
	return login
}

func NewLoginExecutor(repo *config.Repo) *LoginExecutor {
	return &LoginExecutor{
		Repo: repo,

		optionPassword: newOptionPassword(),
		optionRefresh:  newOptionRefresh(),
		optionUsername: newOptionUsername(),

		loginTaskFactory:   broker.NewLoginTask,
		sessionTaskFactory: broker.NewSessionTask,
	}
}

func newOptionPassword() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key: "password",
			Env: "GENAIZ_PASSWORD",
		},
	}
}

func newOptionRefresh() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Key:          "AC.Refresh",
			Param:        "refresh",
			Short:        "r",
			DefaultValue: "false",
		},
	}
}

func newOptionUsername() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "AC.Username",
			Param: "username",
			Short: "u",
			Env:   "GENAIZ_USERNAME",
		},
	}
}
