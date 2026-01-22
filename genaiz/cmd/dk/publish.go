package dk

import (
	"context"

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

type PublishLinkTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type PublishExecutor struct {
	BaseExecutor
	*PublishOptions

	linkArgument string

	publishLinkTaskFactory PublishLinkTaskFactory
	dataLinksWriterFactory dataLinksWriterFactory
}

func (pe PublishExecutor) Display() {
	pe.initDataLinkOptions()
	pe.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": pe.getConfigPath(pe.optionUserDefined),
	},
		&pe.optionOem.Option,
		&pe.optionHandle.Option,
		&pe.optionVersion.Option,
		&pe.optionConfigType.Option,
	)
}

func (pe PublishExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pe.makeConfigParams(pe.optionConfigType, pe.optionUserDefined); err == nil {
		var params = pe.makeDataLinkParams(*configParams)
		var writer = pe.dataLinksWriterFactory(pe.Ledger, configParams.GetConfigPath())

		pe.publishLinkTaskFactory(writer).Pretend(params, pe.Ledger.Logger)
		return
	}

	lang.HandleExit(err)
}

func (pe PublishExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pe.makeConfigParams(pe.optionConfigType, pe.optionUserDefined); err == nil {
		var params = pe.makeDataLinkParams(*configParams)
		var writer = pe.dataLinksWriterFactory(pe.Ledger, configParams.GetConfigPath())
		var plan = task.NewPlan("DataLink", pe.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, pe.publishLinkTaskFactory(writer))
		return
	}

	lang.HandleExit(err)
}

func (pe PublishExecutor) initDataLinkOptions() {
	var oem, handle, version = pe.parseDataLinkArgument(pe.linkArgument)

	pe.Ledger.InitValue(pe.optionHandle, handle)
	pe.Ledger.InitValue(pe.optionOem, oem)
	pe.Ledger.InitValue(pe.optionVersion, version)
}

func (pe PublishExecutor) makeDataLinkParams(configParams shared.ConfigParams) *broker.DataLinkParams {
	pe.initDataLinkOptions()
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: pe.Ledger.AuthFile,
		},
		ConfigParams: configParams,
		DataLink: &broker.DataLink{
			Handle:  pe.Ledger.GetString(pe.optionHandle),
			Oem:     pe.Ledger.GetString(pe.optionOem),
			Version: pe.Ledger.GetString(pe.optionVersion),
		},
	}
}

type PublishOptions struct {
	optionConfigType  *config.StringOption
	optionHandle      *config.StringOption
	optionOem         *config.StringOption
	optionUserDefined *config.BoolOption
	optionVersion     *config.StringOption
}

func (po PublishOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionConfigType,
		po.optionOem,
		po.optionVersion,
		po.optionUserDefined,
	}
}

func NewPublish(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var publishOptions = NewPublishOptions()
	var publishCmd = &cobra.Command{
		Use:     "publish [OEM/]HANDLE[:VERSION] [CONFIG_PATH]",
		Short:   "Publishes a Data Link definition",
		Long:    "Publishes a Data Link definition, possibly attached to a function, published to a broker",
		Example: "genaiz dk publish com.genaiz/datalink-1",
		Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2),
			cli.ArgsOptionalFolder("function", 2, config.Validation.DirExists)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp func() (string, error)
			var err error

			if len(args) == 2 {
				ledger.OverrideBool(publishOptions.optionUserDefined, false)
				wdp = dirz.OptionalWorkingDir(args[1:]...)
			} else {
				wdp = dirz.OptionalWorkingDir()
			}

			if ledger.WorkDir, err = wdp(); err == nil {
				dkCli.Exec(ledger, NewPublishExecutor(cmd.Context(), ledger, dkCli, args[0], publishOptions))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(publishCmd, publishOptions.allDefiners()...)
	return publishCmd
}

func NewPublishExecutor(ctx context.Context, ledger *config.Ledger, dkCli *Cli, handle string, options *PublishOptions) *PublishExecutor {
	return &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: ctx,
			Ledger:  ledger,
		},
		PublishOptions: options,

		linkArgument:           handle,
		publishLinkTaskFactory: broker.NewDataLinkPublishTask,
		dataLinksWriterFactory: newDataLinksWriter,
	}
}

func NewPublishOptions() *PublishOptions {
	return &PublishOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.DataLink.Publish.ConfigType).
			WithDefaultValue("yaml").
			BuildStringOption(),
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.DataLink.Publish.Handle).
			WithValidator(config.Validation.Handle).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.DataLink.Publish.Oem).
			WithValidator(config.Validation.Oem).
			BuildStringOption(),
		optionUserDefined: cli.Options.DataLinks.UserDefined().
			WithKeys(&schema.Genaiz.DataLink.Publish.UserDefined).
			WithDefaultValue("True").
			BuildBoolOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.DataLink.Publish.Version).
			WithValidator(config.Validation.Version).
			Optional(true).
			BuildStringOption(),
	}
}
