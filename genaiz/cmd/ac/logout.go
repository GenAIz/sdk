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
	Repo *config.Repo

	optionHost     *config.StringOption
	optionUsername *config.StringOption

	logoutTaskFactory LogoutTaskFactory
}

func (le *LogoutExecutor) Logout() {
	var logoutParams = le.makeLogoutParams()
	var logoutPlan = task.NewPlan("Logout", le.Repo.Logger)

	logoutPlan.OnSuccess = func(i interface{}) {
		fmt.Printf("%s logged out\n", i)
	}
	task.Single(logoutPlan, logoutParams, le.logoutTaskFactory())
}

func (le *LogoutExecutor) makeLogoutParams() *broker.LoginParams {
	return &broker.LoginParams{
		Broker: &broker.Broker{
			AuthFile: le.Repo.AuthFile,
			HostAddr: le.Repo.GetString(le.optionHost),
		},
		Username: le.Repo.GetString(le.optionUsername),
	}
}

func NewLogout(repo *config.Repo) *cobra.Command {
	var exec = NewLogoutExecutor(repo)
	var logout = &cobra.Command{
		Use:     "logout",
		Short:   "Removes any previously acquired session",
		Long:    "Removes any previously acquired session tokens held for the current user",
		Example: "genaiz ac logout",
		Run: func(cmd *cobra.Command, args []string) {
			exec.Logout()
		},
	}

	repo.Register(logout, exec.optionHost, exec.optionUsername)
	return logout
}

func NewLogoutExecutor(repo *config.Repo) *LogoutExecutor {
	return &LogoutExecutor{
		Repo: repo,

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
