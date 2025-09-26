package ac

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LoginTaskFactory func() *task.Task[broker.LoginParams]

type SessionTaskFactory func() *task.Task[broker.Broker]

type LoginExecutor struct {
	Ledger *config.Ledger

	optionPassword *config.StringOption
	optionRefresh  *config.BoolOption
	optionUsername *config.StringOption

	loginTaskFactory   LoginTaskFactory
	sessionTaskFactory SessionTaskFactory
}

func (le *LoginExecutor) Login(brokerAddr string) {
	var currentSession string
	var refresh = le.Ledger.GetBool(le.optionRefresh)
	var brokerParam = le.makeBrokerParams(brokerAddr)

	if !refresh {
		var sessionPlan = &task.Plan{
			Logger: le.Ledger.Logger,
			OnFailure: func(msg interface{}) {
				refresh = true
			},
			OnSuccess: func(msg interface{}) {
				currentSession = cast.ToString(msg)
			},
		}

		task.Attempt(sessionPlan, brokerParam, le.sessionTaskFactory())

		if currentSession != "" {
			fmt.Printf("Already logged in to %s\n", brokerParam.HostAddr)
		}
	}

	if refresh {
		var params = le.makeLoginParams(brokerParam)
		var loginPlan = task.NewPlan("Login", le.Ledger.Logger)

		loginPlan.OnSuccess = func(i interface{}) {
			fmt.Printf("Logged in to %s\n", brokerParam.HostAddr)
		}

		task.Single(loginPlan, params, le.loginTaskFactory())
	}
}

func (le *LoginExecutor) makeBrokerParams(brokerAddr string) *broker.Broker {
	return &broker.Broker{
		AuthFile: le.Ledger.AuthFile,
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

func (le *LoginExecutor) queryPassword() []byte {
	var password = le.Ledger.GetString(le.optionPassword)
	var result []byte

	if password == "" {
		result = *le.Ledger.QuerySecret("password: ")
	} else {
		result = []byte(password)
	}

	return result
}

func (le *LoginExecutor) queryUsername() string {
	var username = le.Ledger.GetString(le.optionUsername)

	if username == "" {
		username = le.Ledger.QueryMandatory("username: ")
	}

	return strings.TrimSpace(username)
}

func NewLogin(ledger *config.Ledger) *cobra.Command {
	var exec = NewLoginExecutor(ledger)
	var login = &cobra.Command{
		Use:     "login HOST",
		Short:   "Authenticates an account with a GenAIz Broker",
		Long:    "Authenticates an account with a GenAIz Broker provided a host url as argument",
		Example: "genaiz account login broker.genaiz.com --username=myUser --refresh",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			exec.Login(args[0])
		},
	}

	ledger.Register(login, exec.optionUsername, exec.optionPassword, exec.optionRefresh)
	return login
}

func NewLoginExecutor(ledger *config.Ledger) *LoginExecutor {
	return &LoginExecutor{
		Ledger: ledger,

		optionPassword: cli.Options.Accounts.Password().BuildStringOption(),
		optionRefresh:  cli.Options.Accounts.Refresh().BuildBoolOption(),
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Login.Username).
			BuildStringOption(),

		loginTaskFactory:   broker.NewLoginTask,
		sessionTaskFactory: broker.NewSessionTask,
	}
}
