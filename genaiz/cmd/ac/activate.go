package ac

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type ActivateTaskFactory func() *task.Task[broker.Broker]

type ActivateExecutor struct {
	*ActivateOptions
	hostAddr            string
	ledger              *config.Ledger
	activateTaskFactory ActivateTaskFactory
}

func (ae ActivateExecutor) Activate() {
	var brokerParams = ae.newBrokerParams()
	var activatePlan = task.NewPlan("Activate", ae.ledger.Logger)

	activatePlan.OnSuccess = func(i interface{}) {
		if brokerParams.Username == "" {
			fmt.Printf("Session for host %s activated\n", brokerParams.HostAddr)
		} else {
			fmt.Printf("Session for host %s@%s activated\n", brokerParams.Username, brokerParams.HostAddr)
		}
	}
	task.Single(activatePlan, brokerParams, ae.activateTaskFactory())
}

func (ae ActivateExecutor) newBrokerParams() *broker.Broker {
	var username, hostAddr string

	if strings.Contains(ae.hostAddr, "@") {
		var parts = strings.Split(ae.hostAddr, "@")

		username = parts[0]
		hostAddr = parts[1]
	} else {
		username = ae.ledger.GetString(ae.optionUsername)
		hostAddr = ae.hostAddr
	}

	return &broker.Broker{
		AuthFile: ae.ledger.AuthFile,
		HostAddr: hostAddr,
		Username: username,
	}
}

type ActivateOptions struct {
	optionUsername *config.StringOption
}

func (ao ActivateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		ao.optionUsername,
	}
}

func NewActivate(ledger *config.Ledger) *cobra.Command {
	var options = NewActivateOptions()
	var cmdActivate = &cobra.Command{
		Use:     "activate HOST",
		Short:   "Activates any previously acquired session",
		Long:    "Activates any previously acquired session tokens held for the current user",
		Args:    cobra.ExactArgs(1),
		Example: "genaiz ac activate dev.genaiz.com",
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewActivateExecutor(ledger, options, args[0])

			exec.Activate()
		},
	}

	ledger.Register(cmdActivate, options.allDefiners()...)
	cli.AutoBridge.Accounts().Arguments(cmdActivate, ledger)
	return cmdActivate
}

func NewActivateExecutor(ledger *config.Ledger, options *ActivateOptions, hostAddr string) *ActivateExecutor {
	return &ActivateExecutor{
		ActivateOptions: options,

		hostAddr:            hostAddr,
		ledger:              ledger,
		activateTaskFactory: broker.NewActivateTask,
	}
}

func NewActivateOptions() *ActivateOptions {
	return &ActivateOptions{
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Activate.Username).
			BuildStringOption(),
	}
}
