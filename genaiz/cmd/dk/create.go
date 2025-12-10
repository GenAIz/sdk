package dk

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type PublishLinkTaskFactory func() *task.Task[broker.DataLinkParams]

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions

	linkHandle             string
	publishLinkTaskFactory PublishLinkTaskFactory
}

func (ce CreateExecutor) Display() {
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.Ledger.WorkDir,
		"handle": ce.linkHandle,
	},
		&ce.optionDescription.Option,
		&ce.optionName.Option,
		&ce.optionOem.Option,
		&ce.optionVersion.Option,
	)
}

func (ce CreateExecutor) Pretend() {
	var params = ce.makeDataLinkParams()

	ce.publishLinkTaskFactory().Pretend(params, ce.Ledger.Logger)
}

func (ce CreateExecutor) Proceed() {
	var params = ce.makeDataLinkParams()
	var plan = task.NewPlan("CreateLink", ce.Ledger.Logger)

	plan.PrintReportsOnly = true
	task.Single(plan, params, ce.publishLinkTaskFactory())
}

func (ce CreateExecutor) makeDataLinkParams() *broker.DataLinkParams {
	ce.Ledger.InitValue(ce.optionName, ce.linkHandle)
	ce.Ledger.InitValue(ce.optionHandle, ce.linkHandle)
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: ce.Ledger.AuthFile,
		},
		Description: ce.Ledger.GetString(ce.optionDescription),
		Handle:      ce.Ledger.GetString(ce.optionHandle),
		Name:        ce.Ledger.GetString(ce.optionName),
		Oem:         ce.Ledger.GetString(ce.optionOem),
		Version:     ce.Ledger.GetString(ce.optionVersion),
	}
}

type CreateOptions struct {
	optionDescription *config.StringOption
	// omit it from the definers, it will be initialised from the args
	optionHandle  *config.StringOption
	optionOem     *config.StringOption
	optionName    *config.StringOption
	optionVersion *config.StringOption
}

func (co CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionDescription,
		co.optionOem,
		co.optionName,
		co.optionVersion,
	}
}

func NewCreate(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var createOptions = NewCreateOptions()
	var createCmd = &cobra.Command{
		Use:     "create HANDLE [FUNCTION_PATH]",
		Short:   "Creates a Data Link definition",
		Long:    "Creates a Data Link definition, possibly attached to a function, published to a broker",
		Example: "genaiz dk create datalink-1 function-1 --oem=com.genaiz.dev --name='Data Link One'",
		Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2),
			cli.ArgsOptionalFolder("function", 2, config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp func() (string, error)
			var err error

			if len(args) == 2 {
				wdp = dirz.OptionalWorkingDir(args[1:]...)
			} else {
				wdp = dirz.OptionalWorkingDir()
			}

			if ledger.WorkDir, err = wdp(); err == nil {
				dkCli.Exec(ledger, NewCreateExecutor(cmd, ledger, dkCli, args[0], createOptions))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(createCmd, createOptions.allDefiners()...)
	return createCmd
}

func NewCreateExecutor(cmd *cobra.Command, ledger *config.Ledger, dkCli *Cli, handle string, options *CreateOptions) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: cmd.Context(),
			Ledger:  ledger,
		},
		CreateOptions: options,

		linkHandle: handle,

		publishLinkTaskFactory: broker.NewDataLinkPublishTask,
	}
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		optionDescription: cli.Options.DataLinks.Description().
			WithKeys(&schema.Genaiz.DataLink.Create.Description).
			BuildStringOption(),
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.DataLink.Create.Handle).
			BuildStringOption(),
		optionName: cli.Options.DataLinks.Name().
			WithKeys(&schema.Genaiz.DataLink.Create.Name).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.DataLink.Create.Oem).
			BuildStringOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.DataLink.Create.Version).
			WithDefaultValue("1.0.0").
			BuildStringOption(),
	}
}
