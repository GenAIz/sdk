package lk

import (
	"context"
	"errors"

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
type SourcePublishTaskFactory func() *task.Task[locker.SourcePublishParams]
type SourceSyncTaskFactory func() *task.Task[locker.SourceFindParams]
type SourceUpdateTaskFactory func() *task.Task[locker.SourceUpdateParams]

type SourcePublishExecutor struct {
	BasePublishExecutor

	sourceFindTaskFactory    SourceFindTaskFactory
	sourcePublishTaskFactory SourcePublishTaskFactory
	sourceSyncTaskFactory    SourceSyncTaskFactory
}

func (pe *SourcePublishExecutor) Pretend() {
	var brokerParams = pe.accountParams.BrokerParams()
	var publishParams = pe.newPublishParams(brokerParams)
	var baseParams = pe.newBaseParams(pe.handleArg, pe.optionLocker)
	var findParams = pe.newFindParams(baseParams, brokerParams)
	var plan = task.NewPlan("publish-locker-source", pe.Ledger.Logger)
	var pretenders []task.Worker

	publishParams.BaseParams = *baseParams
	pretenders = append(pretenders, task.NewPretender(findParams, pe.sourceFindTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(findParams.DataLinkParams, pe.dataLinkFindTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(findParams, pe.sourceSyncTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(publishParams, pe.sourcePublishTaskFactory()))
	plan.Sequence(pretenders...)
}

func (pe *SourcePublishExecutor) Proceed() {
	var brokerParams = pe.accountParams.BrokerParams()
	var publishParams = pe.newPublishParams(brokerParams)
	var baseParams = pe.newBaseParams(pe.handleArg, pe.optionLocker)
	var findParams = pe.newFindParams(baseParams, brokerParams)
	var plan = task.NewPlan("publish-locker-source", pe.Ledger.Logger)
	var workers []task.Worker

	publishParams.BaseParams = *baseParams
	workers = append(workers, task.NewWorker(findParams, pe.sourceFindTaskFactory()))
	workers = append(workers, task.NewWorker(findParams.DataLinkParams, pe.dataLinkFindTaskFactory()))
	workers = append(workers, task.NewWorker(findParams, pe.sourceSyncTaskFactory()))
	workers = append(workers, task.NewWorker(publishParams, pe.sourcePublishTaskFactory()))
	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (pe *SourcePublishExecutor) Publish(handleArg string) error {
	pe.handleArg = handleArg
	pe.Cli.Exec(pe.Ledger, pe)
	return nil
}

func (pe *SourcePublishExecutor) newFindParams(baseParams *locker.BaseParams, brokerParams *broker.Broker) *locker.SourceFindParams {
	return &locker.SourceFindParams{
		BaseParams: *baseParams,
		DataLinkParams: &broker.DataLinkParams{
			Broker: *brokerParams,
		},
		SourceName: pe.Ledger.GetString(pe.optionName),
	}
}

func (pe *SourcePublishExecutor) newPublishParams(brokerParams *broker.Broker) *locker.SourcePublishParams {
	pe.Ledger.InitValue(pe.optionName, pe.handleArg)
	return &locker.SourcePublishParams{
		Broker:      *brokerParams,
		Name:        pe.Ledger.GetString(pe.optionName),
		Description: pe.Ledger.GetString(pe.optionDescription),
		Visibility:  pe.Ledger.GetString(pe.optionVisibility),
	}
}

type SourceLockerExecutor struct {
	BaseLockerExecutor

	sourceAddTaskFactory    SourceAddTaskFactory
	sourceFindTaskFactory   SourceFindTaskFactory
	sourceUpdateTaskFactory SourceUpdateTaskFactory
}

func (sle *SourceLockerExecutor) Add(handleArg string, dataLinkArg string) error {
	var err error

	if err = sle.BaseLockerExecutor.Add(handleArg, dataLinkArg); err == nil {
		sle.Cli.Exec(sle.Ledger, sle)
		return nil
	}

	return err
}

func (sle *SourceLockerExecutor) Pretend() {
	var brokerParams = sle.accountParams.BrokerParams()
	var configParams = sle.newConfigParams()
	var linkParams = sle.newDataLinkParams(brokerParams, configParams)
	var writer = sle.dataLinksWriterFactory(sle.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-source", sle.Ledger.Logger)
	var pretenders []task.Worker

	if sle.addHandle == "" {
		var updateParams = sle.newSourceUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		pretenders = append(pretenders, task.NewPretender(updateParams.SourceFindParams, sle.sourceFindTaskFactory()))
		pretenders = append(pretenders, task.NewPretender(linkParams, sle.collectLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(updateParams, sle.sourceUpdateTaskFactory()))
	} else {
		var addParams = sle.newSourceAddParams(brokerParams)

		pretenders = append(pretenders, task.NewPretender(linkParams, sle.exportLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(addParams, sle.sourceAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(pretenders...)
}

func (sle *SourceLockerExecutor) Proceed() {
	var brokerParams = sle.accountParams.BrokerParams()
	var configParams = sle.newConfigParams()
	var linkParams = sle.newDataLinkParams(brokerParams, configParams)
	var writer = sle.dataLinksWriterFactory(sle.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-source", sle.Ledger.Logger)
	var workers []task.Worker

	if sle.addHandle == "" {
		var updateParams = sle.newSourceUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		workers = append(workers, task.NewWorker(updateParams.SourceFindParams, sle.sourceFindTaskFactory()))
		workers = append(workers, task.NewWorker(updateParams.SourceFindParams.DataLinkParams, sle.collectLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(updateParams, sle.sourceUpdateTaskFactory()))
	} else {
		var addParams = sle.newSourceAddParams(brokerParams)

		workers = append(workers, task.NewWorker(linkParams, sle.exportLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(addParams, sle.sourceAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (sle *SourceLockerExecutor) Update(handleArg, keyArg, valueArg string) error {
	var err error

	if err = sle.BaseLockerExecutor.Update(handleArg, keyArg, valueArg); err == nil {
		sle.Cli.Exec(sle.Ledger, sle)
		return nil
	}

	return err
}

func (sle *SourceLockerExecutor) newSourceAddParams(brokerParams *broker.Broker) *locker.SourceAddParams {
	return &locker.SourceAddParams{
		BaseParams: *sle.newBaseParams(sle.handleArg, sle.optionLocker),
		Broker:     *brokerParams,
		LinkParams: locker.LinkParams{
			Oem:     sle.addOem,
			Handle:  sle.addHandle,
			Version: sle.addVersion,
		},
	}
}

func (sle *SourceLockerExecutor) newSourceUpdateParams(brokerParams *broker.Broker, configParams *shared.ConfigParams) *locker.SourceUpdateParams {
	var propertyParams = &locker.PropertyParams{
		Key:   sle.keyArg,
		Value: sle.valueArg,
	}

	if sle.secretArg != nil {
		propertyParams.Secret = sle.secretArg
	}

	return &locker.SourceUpdateParams{
		SourceFindParams: &locker.SourceFindParams{
			BaseParams: *sle.newBaseParams(sle.handleArg, sle.optionLocker),
			DataLinkParams: &broker.DataLinkParams{
				Broker:       *brokerParams,
				ConfigParams: *configParams,
			},
		},
		PropertyParams: *propertyParams,
	}
}

func NewSource(ledger *config.Ledger, lkCli *Cli) *cobra.Command {
	var srcAddOptions = NewSourceAddOptions()
	var srcPublishOptions = NewSourcePublishOptions()
	var srcUpdateOptions = NewSourceUpdateOptions()
	var srcAddCmd = source.NewAddSource(newSourceAddExecutorFactory(ledger, lkCli, srcAddOptions))
	var srcPublishCmd = source.NewPublishSource(newSourcePublishExecutorFactory(ledger, lkCli, srcPublishOptions))
	var srcUpdateCmd = source.NewUpdateSource(newSourceUpdateExecutorFactory(ledger, lkCli, srcUpdateOptions))
	var srcCmd = &cobra.Command{
		Use:     "source",
		Aliases: []string{"src"},
		Short:   "Manages data sources under a Locker",
	}

	srcCmd.AddCommand(srcAddCmd)
	srcCmd.AddCommand(srcUpdateCmd)
	srcCmd.AddCommand(srcPublishCmd)
	ledger.Register(srcAddCmd, srcAddOptions.allDefiners()...)
	ledger.Register(srcPublishCmd, srcPublishOptions.allDefiners()...)
	ledger.Register(srcUpdateCmd, srcUpdateOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(srcAddCmd, ledger, srcAddOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(srcPublishCmd, ledger, srcPublishOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(srcUpdateCmd, ledger, srcUpdateOptions.optionAccount)
	return srcCmd
}

func NewSourceLockerExecutor(ctx context.Context, ledger *config.Ledger, lkCli *Cli, options *LockerOptions) *SourceLockerExecutor {
	return &SourceLockerExecutor{
		BaseLockerExecutor: BaseLockerExecutor{
			BaseExecutor: BaseExecutor{
				Cli:     lkCli,
				Context: ctx,
				Ledger:  ledger,
			},
			LockerOptions: options,

			accountParams: config.NewAccountParams(ledger, options.optionAccount),

			collectLinkTaskFactory: broker.NewDataLinkCollectTask,
			dataLinksWriterFactory: dk.NewDataLinksWriter,
			exportLinkTaskFactory:  broker.NewDataLinkExportTask,
		},

		sourceAddTaskFactory:    locker.NewSourceAddTask,
		sourceFindTaskFactory:   locker.NewSourceFindTask,
		sourceUpdateTaskFactory: locker.NewSourceUpdateTask,
	}
}

func NewSourcePublishExecutor(ctx context.Context, ledger *config.Ledger, lkCli *Cli, options *PublishOptions) *SourcePublishExecutor {
	return &SourcePublishExecutor{
		BasePublishExecutor: BasePublishExecutor{
			BaseExecutor: BaseExecutor{
				Cli:     lkCli,
				Context: ctx,
				Ledger:  ledger,
			},
			PublishOptions: options,

			accountParams:           config.NewAccountParams(ledger, options.optionAccount),
			dataLinkFindTaskFactory: broker.NewDataLinkFindTask,
		},

		sourceFindTaskFactory:    locker.NewSourceFindTask,
		sourcePublishTaskFactory: locker.NewSourcePublishTask,
		sourceSyncTaskFactory:    locker.NewSourceSyncTask,
	}
}

func NewSourceAddOptions() *LockerOptions {
	return &LockerOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Source.Add.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Source.Add.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func NewSourcePublishOptions() *PublishOptions {
	return &PublishOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Source.Publish.Account).
			BuildStringOption(),
		optionDescription: cli.Options.Sources.Description().
			WithKeys(&schema.Genaiz.Source.Publish.Description).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Source.Publish.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
		optionName: cli.Options.Sources.Name().
			WithKeys(&schema.Genaiz.Source.Publish.Name).
			BuildStringOption(),
		optionVisibility: cli.Options.Sources.Visibility().
			WithKeys(&schema.Genaiz.Source.Publish.Visibility).
			WithDefaultValue(broker.VisibilityPrivate).
			BuildStringOption(),
	}
}

func NewSourceUpdateOptions() *LockerOptions {
	return &LockerOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Source.Update.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Source.Update.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func newSourceAddExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *LockerOptions) source.AddExecutorFactory {
	return func(command *cobra.Command) source.AddExecutor {
		return NewSourceLockerExecutor(command.Context(), ledger, lkCli, options)
	}
}

func newSourcePublishExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *PublishOptions) source.PublishExecutorFactory {
	return func(command *cobra.Command) source.PublishExecutor {
		return NewSourcePublishExecutor(command.Context(), ledger, lkCli, options)
	}
}

func newSourceUpdateExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *LockerOptions) source.UpdateExecutorFactory {
	return func(command *cobra.Command) source.UpdateExecutor {
		return NewSourceLockerExecutor(command.Context(), ledger, lkCli, options)
	}
}
