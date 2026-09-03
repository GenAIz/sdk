package lk

import (
	"context"
	"errors"
	"os"
	"strconv"
	"strings"

	"github.com/awnumar/memguard"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/cmd/lk/source"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/locker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorDataLinkOemRequired     = errors.New("data link oem is required")
	errorDataLinkSequenceInvalid = errors.New("data link sequence should be a number")
	errorDataLinkVersionRequired = errors.New("data link version is required")
)

type SourceAddTaskFactory func() *task.Task[locker.SourceAddParams]
type SourceFindTaskFactory func() *task.Task[locker.SourceFindParams]
type SourceUpdateTaskFactory func() *task.Task[locker.SourceUpdateParams]

type SourceExecutor struct {
	BaseExecutor
	*SourceOptions

	addHandle   string
	addOem      string
	addSequence *int
	addVersion  string
	handleArg   string
	keyArg      string
	secretArg   *memguard.Enclave
	valueArg    string

	accountParams config.AccountParametric

	dataLinksWriterFactory dk.DataLinksWriterFactory
	collectLinkTaskFactory dk.CollectLinkTaskFactory
	exportLinkTaskFactory  dk.ExportLinkTaskFactory

	sourceAddTaskFactory    SourceAddTaskFactory
	sourceFindTaskFactory   SourceFindTaskFactory
	sourceUpdateTaskFactory SourceUpdateTaskFactory
}

func (se *SourceExecutor) Add(handleArg string, dataLinkArg string) error {
	se.handleArg = handleArg
	se.addOem, se.addHandle, se.addVersion = dk.ParseDataLinkArgument(dataLinkArg)

	if se.addOem == "" {
		return errorDataLinkOemRequired
	}

	if se.addVersion == "" {
		return errorDataLinkVersionRequired
	}

	if versionParts := strings.Split(se.addVersion, "-rc-"); len(versionParts) == 2 {
		var seq int
		var err error

		if seq, err = strconv.Atoi(versionParts[1]); err != nil {
			return errorDataLinkSequenceInvalid
		}

		se.addSequence = &seq
	}

	se.Cli.Exec(se.Ledger, se)
	return nil
}

func (se *SourceExecutor) Display() {
	var argMap = map[string]string{
		"handle": se.handleArg,
	}

	if se.addHandle == "" {
		argMap["prop-key"] = se.keyArg

		if se.secretArg == nil {
			argMap["prop-value"] = se.valueArg
		} else {
			// Never display STDIN values, regardless of whether it is used for a secret or not.
			argMap["prop-value"] = "********"
		}
	} else {
		argMap["datalink-oem"] = se.addOem
		argMap["datalink-handle"] = se.addHandle
		argMap["datalink-version"] = se.addVersion
		argMap["datalink-seq"] = cast.ToString(se.addSequence)
	}

	se.Ledger.DisplayOptionsWithMap(&argMap,
		&se.optionAccount.Option,
		&se.optionLocker.Option)
}

func (se *SourceExecutor) Pretend() {
	var brokerParams = se.accountParams.BrokerParams()
	var configParams = se.newSourceConfigParams()
	var linkParams = se.newDataLinkSyncParams(brokerParams, configParams)
	var writer = se.dataLinksWriterFactory(se.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-source", se.Ledger.Logger)
	var pretenders []task.Worker

	if se.addHandle == "" {
		var updateParams = se.newSourceUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		pretenders = append(pretenders, task.NewPretender(updateParams.SourceFindParams, se.sourceFindTaskFactory()))
		pretenders = append(pretenders, task.NewPretender(linkParams, se.collectLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(updateParams, se.sourceUpdateTaskFactory()))
	} else {
		var addParams = se.newSourceAddParams(brokerParams)

		pretenders = append(pretenders, task.NewPretender(linkParams, se.exportLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(addParams, se.sourceAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(pretenders...)
}

func (se *SourceExecutor) Proceed() {
	var brokerParams = se.accountParams.BrokerParams()
	var configParams = se.newSourceConfigParams()
	var linkParams = se.newDataLinkSyncParams(brokerParams, configParams)
	var writer = se.dataLinksWriterFactory(se.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-source", se.Ledger.Logger)
	var workers []task.Worker

	if se.addHandle == "" {
		var updateParams = se.newSourceUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		workers = append(workers, task.NewWorker(updateParams.SourceFindParams, se.sourceFindTaskFactory()))
		workers = append(workers, task.NewWorker(updateParams.SourceFindParams.DataLinkParams, se.collectLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(updateParams, se.sourceUpdateTaskFactory()))
	} else {
		var addParams = se.newSourceAddParams(brokerParams)

		workers = append(workers, task.NewWorker(linkParams, se.exportLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(addParams, se.sourceAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *SourceExecutor) Update(handleArg string, keyArg string, valueArg string) error {
	var err error

	if se.secretArg, err = se.Ledger.QueryPipe(); err == nil {
		se.handleArg = handleArg
		se.keyArg = keyArg
		se.valueArg = valueArg
		se.Cli.Exec(se.Ledger, se)
		return nil
	}

	return err
}

func (se *SourceExecutor) newDataLinkSyncParams(brokerParams *broker.Broker, configParams *shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker:       *brokerParams,
		ConfigParams: *configParams,
		DataLink: &broker.DataLink{
			Oem:     se.addOem,
			Handle:  se.addHandle,
			Version: se.addVersion,
		},
	}
}

func (se *SourceExecutor) newSourceAddParams(brokerParams *broker.Broker) *locker.SourceAddParams {
	return &locker.SourceAddParams{
		BaseParams: *se.newSourceBaseParams(),
		Broker:     *brokerParams,
		LinkParams: locker.LinkParams{
			Oem:     se.addOem,
			Handle:  se.addHandle,
			Version: se.addVersion,
		},
		SourceHandle: se.handleArg,
	}
}

func (se *SourceExecutor) newSourceBaseParams() *locker.BaseParams {
	var baseParams = &locker.BaseParams{
		LockerPath: se.Ledger.GetString(se.optionLocker),
	}

	if envPwd := os.Getenv(passphraseEnvKey); envPwd != "" {
		baseParams.Passphrase = memguard.NewEnclave([]byte(envPwd))
	} else if pwdEnclave := se.Ledger.QuerySecret(passphrasePrompt); pwdEnclave != nil {
		baseParams.Passphrase = pwdEnclave
	}

	return baseParams
}

func (se *SourceExecutor) newSourceConfigParams() *shared.ConfigParams {
	return &shared.ConfigParams{
		ConfigName:   se.Ledger.ConfigName,
		ConfigFolder: se.Ledger.UserPath,
		ConfigType:   new(shared.ConfigTypeYaml),
	}
}

func (se *SourceExecutor) newSourceUpdateParams(brokerParams *broker.Broker, configParams *shared.ConfigParams) *locker.SourceUpdateParams {
	var propertyParams = &locker.PropertyParams{
		Key:   se.keyArg,
		Value: se.valueArg,
	}

	if se.secretArg != nil {
		propertyParams.Secret = se.secretArg
	}

	return &locker.SourceUpdateParams{
		SourceFindParams: &locker.SourceFindParams{
			BaseParams: *se.newSourceBaseParams(),
			DataLinkParams: &broker.DataLinkParams{
				Broker:       *brokerParams,
				ConfigParams: *configParams,
			},
			SourceHandle: se.handleArg,
		},
		PropertyParams: *propertyParams,
	}
}

type SourceOptions struct {
	optionAccount *config.StringOption
	optionLocker  *config.StringOption
}

func (so SourceOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionAccount,
		so.optionLocker,
	}
}

func NewSource(ledger *config.Ledger, lkCli *Cli) *cobra.Command {
	var srcAddOptions = NewSourceAddOptions()
	var srcUpdateOptions = NewSourceUpdateOptions()
	var srcAddCmd = source.NewAddSource(newSourceAddExecutorFactory(ledger, lkCli, srcAddOptions))
	var srcUpdateCmd = source.NewUpdateSource(newSourceUpdateExecutorFactory(ledger, lkCli, srcUpdateOptions))
	var srcCmd = &cobra.Command{
		Use:     "source",
		Aliases: []string{"src"},
		Short:   "Manages data sources under a Locker",
	}

	srcCmd.AddCommand(srcAddCmd)
	srcCmd.AddCommand(srcUpdateCmd)
	ledger.Register(srcAddCmd, srcAddOptions.allDefiners()...)
	ledger.Register(srcUpdateCmd, srcUpdateOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(srcAddCmd, ledger, srcAddOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(srcUpdateCmd, ledger, srcUpdateOptions.optionAccount)
	return srcCmd
}

func NewSourceExecutor(ctx context.Context, ledger *config.Ledger, lkCli *Cli, options *SourceOptions) *SourceExecutor {
	return &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     lkCli,
			Context: ctx,
			Ledger:  ledger,
		},
		SourceOptions: options,

		accountParams:           config.NewAccountParams(ledger, options.optionAccount),
		collectLinkTaskFactory:  broker.NewDataLinkCollectTask,
		dataLinksWriterFactory:  dk.NewDataLinksWriter,
		exportLinkTaskFactory:   broker.NewDataLinkExportTask,
		sourceAddTaskFactory:    locker.NewSourceAddTask,
		sourceFindTaskFactory:   locker.NewSourceFindTask,
		sourceUpdateTaskFactory: locker.NewSourceUpdateTask,
	}
}

func NewSourceAddOptions() *SourceOptions {
	return &SourceOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Locker.Add.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Locker.Add.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func NewSourceUpdateOptions() *SourceOptions {
	return &SourceOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Locker.Update.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Locker.Update.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func newSourceAddExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *SourceOptions) source.AddExecutorFactory {
	return func(command *cobra.Command) source.AddExecutor {
		return NewSourceExecutor(command.Context(), ledger, lkCli, options)
	}
}

func newSourceUpdateExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *SourceOptions) source.UpdateExecutorFactory {
	return func(command *cobra.Command) source.UpdateExecutor {
		return NewSourceExecutor(command.Context(), ledger, lkCli, options)
	}
}
