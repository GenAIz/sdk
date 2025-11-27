package ac

import (
	"fmt"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LoginTaskFactory func() *task.Task[broker.LoginParams]

type OidcTaskFactory func() *task.Task[broker.OidcParams]

type SessionTaskFactory func() *task.Task[broker.Broker]

type LoginExecutor struct {
	Ledger *config.Ledger

	optionNoBrowser *config.BoolOption
	optionPassword  *config.StringOption
	optionRefresh   *config.BoolOption
	optionUsername  *config.StringOption

	loginTaskFactory   LoginTaskFactory
	oidcTaskFactory    OidcTaskFactory
	sessionTaskFactory SessionTaskFactory
}

func (le *LoginExecutor) Login(brokerAddr string) {
	var brokerParams = le.makeBrokerParams(brokerAddr)
	var refresh = le.Ledger.GetBool(le.optionRefresh)

	if !refresh {
		var currentSession string
		var sessionPlan = &task.Plan{
			Logger:    le.Ledger.Logger,
			OnFailure: task.HandleFlag(&refresh, true),
			OnSuccess: task.HandleString(&currentSession),
		}

		task.Attempt(sessionPlan, brokerParams, le.sessionTaskFactory())

		if currentSession != "" {
			fmt.Printf("Already logged in to %s\n", brokerParams.HostAddr)
		}
	}

	if refresh {
		le.loginWithOidc(brokerParams)
	}
}

func (le *LoginExecutor) allDefiners() []config.Definer {
	return []config.Definer{
		le.optionNoBrowser,
		le.optionPassword,
		le.optionRefresh,
		le.optionUsername,
	}
}

func (le *LoginExecutor) loginWithOidc(brokerParams *broker.Broker) {
	var oidcWorker = task.NewWorker(le.makeOidcParams(brokerParams), le.oidcTaskFactory())
	var oidcPlan = &task.Plan{
		Logger:    le.Ledger.Logger,
		OnSuccess: le.printSuccessHandler(brokerParams),
	}

	oidcPlan.Single(oidcWorker, func(msg interface{}) {
		if strings.EqualFold(cast.ToString(msg), broker.ErrorOidcNotSupported.Error()) {
			le.loginWithUsername(brokerParams)
		} else {
			lang.HandleExit(msg)
		}
	})
}

func (le *LoginExecutor) loginWithUsername(brokerParams *broker.Broker) {
	var params = le.makeLoginParams(brokerParams)
	var loginPlan = task.NewPlan("Login", le.Ledger.Logger)

	loginPlan.OnSuccess = le.printSuccessHandler(brokerParams)
	task.Single(loginPlan, params, le.loginTaskFactory())
}

func (le *LoginExecutor) makeBrokerParams(brokerAddr string) *broker.Broker {
	return &broker.Broker{
		AuthFile: le.Ledger.AuthFile,
		HostAddr: brokerAddr,
	}
}

func (le *LoginExecutor) makeLoginParams(brokerParams *broker.Broker) *broker.LoginParams {
	return &broker.LoginParams{
		Broker:   brokerParams,
		Username: le.queryUsername(),
		Password: le.queryPassword(),
	}
}

func (le *LoginExecutor) makeOidcParams(brokerParams *broker.Broker) *broker.OidcParams {
	return &broker.OidcParams{
		Broker:          brokerParams,
		BrowserRedirect: !le.Ledger.GetBool(le.optionNoBrowser),
	}
}

func (le *LoginExecutor) printSuccessHandler(brokerParams *broker.Broker) func(interface{}) {
	return func(msg interface{}) {
		fmt.Printf("Logged in to %s\n", brokerParams.HostAddr)
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

	ledger.Register(login, exec.allDefiners()...)
	return login
}

func NewLoginExecutor(ledger *config.Ledger) *LoginExecutor {
	return &LoginExecutor{
		Ledger: ledger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionPassword:  cli.Options.Accounts.Password().BuildStringOption(),
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Login.Username).
			BuildStringOption(),

		loginTaskFactory:   broker.NewLoginTask,
		oidcTaskFactory:    broker.NewOidcTask,
		sessionTaskFactory: broker.NewSessionTask,
	}
}
