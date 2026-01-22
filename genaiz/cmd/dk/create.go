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
	"genaiz.com/genaiz/task/shared"
)

type CreateLinkTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions

	linkArgument           string
	createLinkTaskFactory  CreateLinkTaskFactory
	dataLinksWriterFactory dataLinksWriterFactory
}

func (ce CreateExecutor) Display() {
	ce.initDataLinkOptions()
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.getConfigPath(ce.optionUserDefined),
	},
		&ce.optionHandle.Option,
		&ce.optionOem.Option,
		&ce.optionVersion.Option,
		&ce.optionDescription.Option,
		&ce.optionName.Option,
		&ce.optionConfigType.Option,
	)
}

func (ce CreateExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = ce.makeConfigParams(ce.optionConfigType, ce.optionUserDefined); err == nil {
		var params = ce.makeDataLinkParams(*configParams)
		var writer = ce.dataLinksWriterFactory(ce.Ledger, configParams.GetConfigPath())

		ce.createLinkTaskFactory(writer).Pretend(params, ce.Ledger.Logger)
		return
	}

	lang.HandleExit(err)
}

func (ce CreateExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = ce.makeConfigParams(ce.optionConfigType, ce.optionUserDefined); err == nil {
		var params = ce.makeDataLinkParams(*configParams)
		var writer = ce.dataLinksWriterFactory(ce.Ledger, configParams.GetConfigPath())
		var plan = task.NewPlan("DataLink", ce.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, ce.createLinkTaskFactory(writer))
		return
	}

	lang.HandleExit(err)
}

func (ce CreateExecutor) initDataLinkOptions() {
	var oem, handle, version = ce.parseDataLinkArgument(ce.linkArgument)

	ce.Ledger.OverrideString(ce.optionHandle, handle)
	ce.Ledger.InitValue(ce.optionName, handle)
	ce.Ledger.OverrideString(ce.optionOem, oem)
	ce.Ledger.OverrideString(ce.optionVersion, version)
}

func (ce CreateExecutor) makeDataLinkParams(configParams shared.ConfigParams) *broker.DataLinkParams {
	ce.initDataLinkOptions()
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: ce.Ledger.AuthFile,
		},
		ConfigParams: configParams,
		DataLink: &broker.DataLink{
			Description: ce.Ledger.GetString(ce.optionDescription),
			Handle:      ce.Ledger.GetString(ce.optionHandle),
			Name:        ce.Ledger.GetString(ce.optionName),
			Oem:         ce.Ledger.GetString(ce.optionOem),
			Version:     ce.Ledger.GetString(ce.optionVersion),
		},
	}
}

type CreateOptions struct {
	optionConfigType  *config.StringOption
	optionDescription *config.StringOption
	optionHandle      *config.StringOption
	optionName        *config.StringOption
	optionOem         *config.StringOption
	optionUserDefined *config.BoolOption
	optionVersion     *config.StringOption
}

func (co CreateOptions) allDefiners() []config.Definer {
	// Handle is not registered because it's a mandatory positional argument
	return []config.Definer{
		co.optionConfigType,
		co.optionDescription,
		co.optionName,
		co.optionOem,
		co.optionUserDefined,
		co.optionVersion,
	}
}

func NewCreate(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var createOptions = NewCreateOptions()
	var createCmd = &cobra.Command{
		Use:     "create [OEM/]HANDLE[:VERSION] [CONFIG_FOLDER]",
		Short:   "Creates a Data Link definition",
		Long:    "Creates a Data Link definition, possibly attached to a function, published to a broker",
		Example: "genaiz dk create datalink-1 function-1 --oem=com.genaiz.dev --name='Data Link One'",
		Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2),
			cli.ArgsOptionalFolder("function", 2, config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp func() (string, error)
			var err error

			if len(args) == 2 {
				ledger.OverrideBool(createOptions.optionUserDefined, false)
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

func NewCreateExecutor(cmd *cobra.Command, ledger *config.Ledger, dkCli *Cli, arg string, options *CreateOptions) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: cmd.Context(),
			Ledger:  ledger,
		},
		CreateOptions: options,

		linkArgument:           arg,
		createLinkTaskFactory:  broker.NewDataLinkCreateTask,
		dataLinksWriterFactory: newDataLinksWriter,
	}
}

func NewCreateOptions() *CreateOptions {
	return &CreateOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.DataLink.Create.ConfigType).
			WithDefaultValue("yaml").
			BuildStringOption(),
		optionDescription: cli.Options.DataLinks.Description().
			WithKeys(&schema.Genaiz.DataLink.Create.Description).
			BuildStringOption(),
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.DataLink.Create.Handle).
			WithValidator(config.Validation.Handle).
			BuildStringOption(),
		optionName: cli.Options.DataLinks.Name().
			WithKeys(&schema.Genaiz.DataLink.Create.Name).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.DataLink.Create.Oem).
			WithValidator(config.Validation.Oem).
			BuildStringOption(),
		optionUserDefined: cli.Options.DataLinks.UserDefined().
			WithKeys(&schema.Genaiz.DataLink.Create.UserDefined).
			WithDefaultValue("True").
			BuildBoolOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.DataLink.Create.Version).
			WithValidator(config.Validation.Version).
			WithDefaultValue("1.0.0").
			BuildStringOption(),
	}
}
