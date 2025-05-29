package ac

import (
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LoginExecutor struct {
	repo *config.Repo

	optionPassword *config.StringOption
	optionUsername *config.StringOption
}

func (le *LoginExecutor) Login(brokerAddr string) {
	var params = le.makeLoginParams(le.repo, brokerAddr)
	var plan = &task.Plan[broker.LoginParams]{
		Logger: le.repo.Logger,
		OnError: func(err error) {
			le.repo.Logger.Errorf("Could not authenticate with broker: %s", err)
		},
	}

	plan.Single(params, broker.NewLoginTask())
}

func (le *LoginExecutor) makeLoginParams(repo *config.Repo, brokerAddr string) *broker.LoginParams {
	return &broker.LoginParams{
		AuthFile: repo.AuthFile,
		HostAddr: brokerAddr,
		Username: le.queryUsername(),
		Password: le.queryPassword(),
		Refresh:  true,
	}
}

func (le *LoginExecutor) queryPassword() *[]byte {
	var password = le.repo.GetString(le.optionPassword)
	var result *[]byte

	if password == "" {
		result = le.repo.QuerySecret("password: ")
	} else {
		var bytes = []byte(password)

		result = &bytes
	}

	return result
}

func (le *LoginExecutor) queryUsername() string {
	var username = le.repo.GetString(le.optionUsername)

	if username == "" {
		username = le.repo.QueryMandatory("username: ")
	}

	return strings.TrimSpace(username)
}

func NewLogin(repo *config.Repo) *cobra.Command {
	var exec = NewLoginExecutor(repo)
	var login = &cobra.Command{
		Use:     "login host",
		Short:   "Authenticates an account with a Genaiz broker",
		Long:    "Authenticates a username and password with a Genaiz broker provided a url as argument",
		Example: "genaiz ac login www.genaiz.com",
		Args:    cobra.MatchAll(cobra.ExactArgs(1)),
		Run: func(cmd *cobra.Command, args []string) {
			exec.Login(args[0])
		},
	}

	repo.Register(login, exec.optionUsername, exec.optionPassword)
	return login
}

func NewLoginExecutor(repo *config.Repo) *LoginExecutor {
	return &LoginExecutor{
		optionPassword: NewOptionPassword(),
		optionUsername: NewOptionUsername(),
		repo:           repo,
	}
}

func NewOptionPassword() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key: "password",
		},
	}
}

func NewOptionUsername() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "AC.Username",
			Param: "username",
			Short: "u",
		},
	}
}
