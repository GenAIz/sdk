package ac

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type inspectAuthTaskFactory func() *task.Task[broker.Broker]

type InspectExecutor struct {
	*InspectOptions
	ledger *config.Ledger

	inspectAuthTaskFactory inspectAuthTaskFactory
	printerParams          cli.PrinterParametric
}

func (ie InspectExecutor) Inspect(hostAddr string) error {
	var inspectParams = ie.newInspectParams(hostAddr)
	var printer = ie.printerParams.Printer()
	var session broker.Session
	var failure interface{}
	var plan = task.NewPlanBuilder(ie.ledger.Logger).
		WithReturn(func(i interface{}) { session = i.(broker.Session) }).
		WithFailures(func(i interface{}) { failure = i }).
		Build()

	plan.Sequence(task.NewWorker(inspectParams, ie.inspectAuthTaskFactory()))

	if failure == nil {
		var userSession = mgmt.ToUserSession(&session)

		if ie.printerParams.IsDefault() {
			// the default printer is expecting a list
			return printer.Print([]mgmt.UserSession{*userSession})
		}

		return printer.Print(*userSession)
	}

	return printer.Error(failure)
}

func (ie InspectExecutor) newInspectParams(hostAddr string) *broker.Broker {
	return &broker.Broker{
		AuthFile: ie.ledger.AuthFile,
		HostAddr: hostAddr,
	}
}

type InspectOptions struct {
	optionJsonPrinter *config.BoolOption
}

func (lo InspectOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionJsonPrinter,
	}
}

func NewInspect(ledger *config.Ledger) *cobra.Command {
	var inspectOptions = NewInspectOptions()
	var inspectCmd = &cobra.Command{
		Use:     "inspect [HOST_ADDR]",
		Aliases: []string{"ls"},
		Short:   "Inspects the currently active session",
		Long:    "Inspects the currently active session from the orchestrator, if any exists",
		Example: "genaiz account inspect",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewInspectExecutor(ledger, inspectOptions)
			var hostAddr = cli.ArgsOptionalSingle(args)
			var err = exec.Inspect(hostAddr)

			lang.HandleExit(err)
		},
	}

	ledger.Register(inspectCmd, inspectOptions.allDefiners()...)
	return inspectCmd
}

func NewInspectExecutor(ledger *config.Ledger, options *InspectOptions) *InspectExecutor {
	return &InspectExecutor{
		InspectOptions: options,

		ledger:                 ledger,
		printerParams:          cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		inspectAuthTaskFactory: broker.NewInspectTask,
	}
}

func NewInspectOptions() *InspectOptions {
	return &InspectOptions{
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Account.Inspect.Printer).
			BuildBoolOption(),
	}
}
