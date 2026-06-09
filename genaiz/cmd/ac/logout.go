package ac

import (
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LogoutTaskFactory func() *task.Task[broker.LoginParams]

type LogoutExecutor struct {
	*LogoutOptions

	host              string
	ledger            *config.Ledger
	logoutTaskFactory LogoutTaskFactory
}

type LogoutOptions struct {
	optionUsername *config.StringOption
}

func (lo LogoutOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionUsername,
	}
}

func (le *LogoutExecutor) Logout() {
	var logoutParams = le.makeLogoutParams()
	var logoutPlan = task.NewPlan("Logout", le.ledger.Logger)

	logoutPlan.OnSuccess = func(i interface{}) {
		fmt.Printf("%s logged out\n", i)
	}
	task.Single(logoutPlan, logoutParams, le.logoutTaskFactory())
}

func (le *LogoutExecutor) makeLogoutParams() *broker.LoginParams {
	return &broker.LoginParams{
		Broker: &broker.Broker{
			AuthFile: le.ledger.AuthFile,
			HostAddr: le.host,
		},
		Username: le.ledger.GetString(le.optionUsername),
	}
}

func NewLogout(ledger *config.Ledger) *cobra.Command {
	var options = NewLogoutOptions()
	var logoutCmd = &cobra.Command{
		Use:     "logout [[USER_STRING@]HOST]",
		Short:   "Removes any previously acquired session",
		Long:    "Removes any previously acquired session tokens held for the current user",
		Args:    cobra.MaximumNArgs(1),
		Example: "genaiz ac logout dev.genaiz.com",
		Run: func(cmd *cobra.Command, args []string) {
			var host = cli.ArgsOptionalSingle(args)
			var exec = NewLogoutExecutor(ledger, options, host)

			exec.Logout()
		},
	}

	ledger.Register(logoutCmd, options.allDefiners()...)
	cli.AutoBridge.Accounts().Arguments(logoutCmd, ledger)
	return logoutCmd
}

func NewLogoutExecutor(ledger *config.Ledger, options *LogoutOptions, host string) *LogoutExecutor {
	return &LogoutExecutor{
		LogoutOptions: options,

		host:              host,
		ledger:            ledger,
		logoutTaskFactory: broker.NewLogoutTask,
	}
}

func NewLogoutOptions() *LogoutOptions {
	return &LogoutOptions{
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Logout.Username).
			BuildStringOption(),
	}
}
