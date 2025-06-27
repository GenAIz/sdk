package ac

import (
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type LogoutTaskFactory func() *task.Task[broker.LoginParams]

type LogoutExecutor struct {
	Ledger *config.Ledger

	optionHost     *config.StringOption
	optionUsername *config.StringOption

	logoutTaskFactory LogoutTaskFactory
}

func (le *LogoutExecutor) Logout() {
	var logoutParams = le.makeLogoutParams()
	var logoutPlan = task.NewPlan("Logout", le.Ledger.Logger)

	logoutPlan.OnSuccess = func(i interface{}) {
		fmt.Printf("%s logged out\n", i)
	}
	task.Single(logoutPlan, logoutParams, le.logoutTaskFactory())
}

func (le *LogoutExecutor) makeLogoutParams() *broker.LoginParams {
	return &broker.LoginParams{
		Broker: &broker.Broker{
			AuthFile: le.Ledger.AuthFile,
			HostAddr: le.Ledger.GetString(le.optionHost),
		},
		Username: le.Ledger.GetString(le.optionUsername),
	}
}

func NewLogout(ledger *config.Ledger) *cobra.Command {
	var exec = NewLogoutExecutor(ledger)
	var logout = &cobra.Command{
		Use:     "logout",
		Short:   "Removes any previously acquired session",
		Long:    "Removes any previously acquired session tokens held for the current user",
		Example: "genaiz ac logout",
		Run: func(cmd *cobra.Command, args []string) {
			exec.Logout()
		},
	}

	ledger.Register(logout, exec.optionHost, exec.optionUsername)
	return logout
}

func NewLogoutExecutor(ledger *config.Ledger) *LogoutExecutor {
	return &LogoutExecutor{
		Ledger: ledger,

		optionHost:     newOptionHost(),
		optionUsername: newOptionUsername("Logout"),

		logoutTaskFactory: broker.NewLogoutTask,
	}
}

func newOptionHost() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "AC.Host",
			Param: "host",
		},
	}
}
