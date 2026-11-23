package ac

import (
	"fmt"

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
	*LoginOptions

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
			le.Ledger.StdOut("Already logged in to %s\n", brokerParams.HostAddr)
			return
		}
	}

	if refresh {
		le.loginWithOidc(brokerParams)
	}
}

func (le *LoginExecutor) loginWithOidc(brokerParams *broker.Broker) {
	var oidcWorker = task.NewWorker(le.makeOidcParams(brokerParams), le.oidcTaskFactory())
	var oidcError error
	var oidcPlan = &task.Plan{
		Logger: le.Ledger.Logger,
		OnSuccess: func(i interface{}) {
			if oidcError == nil {
				le.printSuccessHandler(brokerParams)(i)
			}
		},
	}

	oidcPlan.Single(oidcWorker, lang.HandleExit)
}

func (le *LoginExecutor) makeBrokerParams(brokerAddr string) *broker.Broker {
	var prefix, version string

	if !le.Ledger.GetBool(le.optionLegacy) {
		prefix = le.Ledger.GetString(le.optionPrefix)

		if v := le.Ledger.GetString(le.optionVersion); v != "" {
			version = fmt.Sprintf("v%s", v)
		}
	}

	return &broker.Broker{
		AuthFile:   le.Ledger.AuthFile,
		ApiPrefix:  prefix,
		ApiVersion: version,
		HostAddr:   brokerAddr,
	}
}

func (le *LoginExecutor) makeOidcParams(brokerParams *broker.Broker) *broker.OidcParams {
	return &broker.OidcParams{
		Broker:          brokerParams,
		BrowserRedirect: !le.Ledger.GetBool(le.optionNoBrowser),
		LegacyApi:       le.Ledger.GetBool(le.optionLegacy),
	}
}

func (le *LoginExecutor) printSuccessHandler(brokerParams *broker.Broker) func(interface{}) {
	return func(msg interface{}) {
		fmt.Printf("Logged in to %s\n", brokerParams.HostAddr)
	}
}

type LoginOptions struct {
	optionLegacy    *config.BoolOption
	optionNoBrowser *config.BoolOption
	optionPrefix    *config.StringOption
	optionRefresh   *config.BoolOption
	optionVersion   *config.StringOption
}

func (le *LoginOptions) allDefiners() []config.Definer {
	return []config.Definer{
		le.optionLegacy,
		le.optionNoBrowser,
		le.optionPrefix,
		le.optionRefresh,
		le.optionVersion,
	}
}

func NewLogin(ledger *config.Ledger) *cobra.Command {
	var loginOptions = NewLoginOptions()
	var loginCmd = &cobra.Command{
		Use:     "login HOST",
		Short:   "Authenticates an account with a GenAIz Broker",
		Long:    "Authenticates an account with a GenAIz Broker provided an OIDC host url as argument",
		Example: "genaiz account login my-instance.genaiz.com",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewLoginExecutor(ledger, loginOptions)

			exec.Login(args[0])
		},
	}

	ledger.Register(loginCmd, loginOptions.allDefiners()...)
	return loginCmd
}

func NewLoginExecutor(ledger *config.Ledger, options *LoginOptions) *LoginExecutor {
	return &LoginExecutor{
		Ledger:       ledger,
		LoginOptions: options,

		oidcTaskFactory:    broker.NewOidcTask,
		sessionTaskFactory: broker.NewSessionTask,
	}
}

func NewLoginOptions() *LoginOptions {
	return &LoginOptions{
		optionLegacy: cli.Options.Accounts.Legacy().
			WithKeys(&schema.Genaiz.Account.Login.Legacy).
			BuildBoolOption(),
		optionNoBrowser: cli.Options.Accounts.NoBrowser().
			BuildBoolOption(),
		optionPrefix: cli.Options.Accounts.Prefix().
			WithKeys(&schema.Genaiz.Account.Login.Prefix).
			BuildStringOption(),
		optionRefresh: cli.Options.Accounts.Refresh().
			WithKeys(&schema.Genaiz.Account.Login.Refresh).
			BuildBoolOption(),
		optionVersion: cli.Options.Accounts.Version().
			WithKeys(&schema.Genaiz.Account.Login.Version).
			Optional(true).
			BuildStringOption(),
	}
}
