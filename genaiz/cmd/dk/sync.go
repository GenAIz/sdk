package dk

import (
	"github.com/spf13/cast"
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

type ExportLinkTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type SyncExecutor struct {
	BaseExecutor
	*SyncOptions

	linkArgument string

	dataLinksWriterFactory DataLinksWriterFactory
	exportLinkTaskFactory  ExportLinkTaskFactory
}

func (se SyncExecutor) Display() {
	se.setOptions(se.Ledger, se.linkArgument)
	se.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": se.getConfigPath(se.optionUserDefined),
	},
		&se.optionOem.Option,
		&se.optionHandle.Option,
		&se.optionVersion.Option,
		&se.optionSequence.Option,
		&se.optionConfigType.Option,
	)
}

func (se SyncExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = se.makeConfigParams(se.optionConfigType, se.optionUserDefined); err == nil {
		var params = se.makeDataLinkParams(*configParams)
		var writer = se.dataLinksWriterFactory(se.Ledger, configParams.GetConfigPath())

		se.exportLinkTaskFactory(writer).Pretend(params, se.Ledger.Logger)
		return
	}

	lang.HandleExit(err)
}

func (se SyncExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = se.makeConfigParams(se.optionConfigType, se.optionUserDefined); err == nil {
		var params = se.makeDataLinkParams(*configParams)
		var writer = se.dataLinksWriterFactory(se.Ledger, configParams.GetConfigPath())
		var plan = task.NewPlan("DataLink", se.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, se.exportLinkTaskFactory(writer))
		return
	}

	lang.HandleExit(err)
}

func (se SyncExecutor) makeDataLinkParams(configParams shared.ConfigParams) *broker.DataLinkParams {
	var seqString = se.Ledger.GetString(se.optionSequence)
	var seq = new(cast.ToInt(seqString))

	se.setOptions(se.Ledger, se.linkArgument)
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: se.Ledger.AuthFile,
		},
		ConfigParams: configParams,
		DataLink: &broker.DataLink{
			Handle:  se.Ledger.GetString(se.optionHandle),
			Oem:     se.Ledger.GetString(se.optionOem),
			Version: se.Ledger.GetString(se.optionVersion),
			Seq:     seq,
		},
	}
}

type SyncOptions struct {
	BaseOptions
	optionSequence *config.StringOption
}

func (so SyncOptions) allDefiners() []config.Definer {
	var definers = so.BaseOptions.allDefiners()

	definers = append(definers, so.optionSequence)
	return definers
}

func NewSync(ledger *config.Ledger, dkCli *Cli) *cobra.Command {
	var syncOptions = NewSyncOptions()
	var syncCmd = &cobra.Command{
		Use:     "sync [OEM/]HANDLE[:VERSION] [CONFIG_FOLDER]",
		Short:   "Synchronizes a Data Link definition",
		Long:    "Synchronizes a Data Link definition from an active Orchestrator session if it exists",
		Example: "genaiz dk sync datalink-1:1.0.0 --oem=com.genaiz.dev",
		Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2),
			cli.ArgsOptionalFolder("function", 2, config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp func() (string, error)
			var err error

			if len(args) == 2 {
				ledger.OverrideBool(syncOptions.optionUserDefined, false)
				wdp = dirz.OptionalWorkingDir(args[1:]...)
			} else {
				wdp = dirz.OptionalWorkingDir()
			}

			if ledger.WorkDir, err = wdp(); err == nil {
				dkCli.Exec(ledger, NewSyncExecutor(cmd, ledger, dkCli, args[0], syncOptions))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(syncCmd, syncOptions.allDefiners()...)
	return syncCmd
}

func NewSyncExecutor(cmd *cobra.Command, ledger *config.Ledger, dkCli *Cli, linkArgument string, options *SyncOptions) *SyncExecutor {
	return &SyncExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: cmd.Context(),
			Ledger:  ledger,
		},
		SyncOptions: options,

		linkArgument: linkArgument,

		dataLinksWriterFactory: NewDataLinksWriter,
		exportLinkTaskFactory:  broker.NewDataLinkExportTask,
	}
}

func NewSyncOptions() *SyncOptions {
	return &SyncOptions{
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.Sync.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.Sync.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.Sync.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.Sync.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.Sync.Version).
				WithValidator(config.Validation.Version).
				WithDefaultValue("1.0.0").
				BuildStringOption(),
		},
		optionSequence: cli.Options.DataLinks.Sequence().
			WithKeys(&schema.Genaiz.DataLink.Sync.Sequence).
			Optional(true).
			BuildStringOption(),
	}
}
