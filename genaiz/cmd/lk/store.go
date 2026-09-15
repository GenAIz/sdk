package lk

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/cmd/lk/store"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/locker"
	"genaiz.com/genaiz/task/shared"
)

type StoreAddTaskFactory func() *task.Task[locker.StoreAddParams]
type StoreFindTaskFactory func() *task.Task[locker.StoreFindParams]
type StorePublishTaskFactory func() *task.Task[locker.StorePublishParams]
type StoreSyncTaskFactory func() *task.Task[locker.StoreFindParams]
type StoreUpdateTaskFactory func() *task.Task[locker.StoreUpdateParams]

type StorePublishExecutor struct {
	BasePublishExecutor

	storeFindTaskFactory    StoreFindTaskFactory
	storePublishTaskFactory StorePublishTaskFactory
	storeSyncTaskFactory    StoreSyncTaskFactory
}

func (spe *StorePublishExecutor) Pretend() {
	var brokerParams = spe.accountParams.BrokerParams()
	var publishParams = spe.newPublishParams(brokerParams)
	var baseParams = spe.newBaseParams(spe.handleArg, spe.optionLocker)
	var findParams = spe.newFindParams(baseParams, brokerParams)
	var plan = task.NewPlan("publish-locker-source", spe.Ledger.Logger)
	var pretenders []task.Worker

	publishParams.BaseParams = *baseParams
	pretenders = append(pretenders, task.NewPretender(findParams, spe.storeFindTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(findParams.DataLinkParams, spe.dataLinkFindTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(findParams, spe.storeSyncTaskFactory()))
	pretenders = append(pretenders, task.NewPretender(publishParams, spe.storePublishTaskFactory()))
	plan.Sequence(pretenders...)
}

func (spe *StorePublishExecutor) Proceed() {
	var brokerParams = spe.accountParams.BrokerParams()
	var publishParams = spe.newPublishParams(brokerParams)
	var baseParams = spe.newBaseParams(spe.handleArg, spe.optionLocker)
	var findParams = spe.newFindParams(baseParams, brokerParams)
	var plan = task.NewPlan("publish-locker-store", spe.Ledger.Logger)
	var workers []task.Worker

	publishParams.BaseParams = *baseParams
	workers = append(workers, task.NewWorker(findParams, spe.storeFindTaskFactory()))
	workers = append(workers, task.NewWorker(findParams.DataLinkParams, spe.dataLinkFindTaskFactory()))
	workers = append(workers, task.NewWorker(findParams, spe.storeSyncTaskFactory()))
	workers = append(workers, task.NewWorker(publishParams, spe.storePublishTaskFactory()))
	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (spe *StorePublishExecutor) Publish(handleArg string) error {
	spe.handleArg = handleArg
	spe.Cli.Exec(spe.Ledger, spe)
	return nil
}

func (spe *StorePublishExecutor) newFindParams(baseParams *locker.BaseParams, brokerParams *broker.Broker) *locker.StoreFindParams {
	return &locker.StoreFindParams{
		BaseParams: *baseParams,
		DataLinkParams: &broker.DataLinkParams{
			Broker: *brokerParams,
		},
		StoreName: spe.Ledger.GetString(spe.optionName),
	}
}

func (spe *StorePublishExecutor) newPublishParams(brokerParams *broker.Broker) *locker.StorePublishParams {
	spe.Ledger.InitValue(spe.optionName, spe.handleArg)
	return &locker.StorePublishParams{
		Broker:      *brokerParams,
		Name:        spe.Ledger.GetString(spe.optionName),
		Description: spe.Ledger.GetString(spe.optionDescription),
		Visibility:  spe.Ledger.GetString(spe.optionVisibility),
	}
}

type StoreLockerExecutor struct {
	BaseLockerExecutor

	storeAddTaskFactory    StoreAddTaskFactory
	storeFindTaskFactory   StoreFindTaskFactory
	storeUpdateTaskFactory StoreUpdateTaskFactory
}

func (sle *StoreLockerExecutor) Add(handleArg string, dataLinkArg string) error {
	var err error

	if err = sle.BaseLockerExecutor.Add(handleArg, dataLinkArg); err == nil {
		sle.Cli.Exec(sle.Ledger, sle)
		return nil
	}

	return err
}

func (sle *StoreLockerExecutor) Pretend() {
	var brokerParams = sle.accountParams.BrokerParams()
	var configParams = sle.newConfigParams()
	var linkParams = sle.newDataLinkParams(brokerParams, configParams)
	var writer = sle.dataLinksWriterFactory(sle.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-store", sle.Ledger.Logger)
	var pretenders []task.Worker

	if sle.addHandle == "" {
		var updateParams = sle.newStoreUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		pretenders = append(pretenders, task.NewPretender(updateParams.StoreFindParams, sle.storeFindTaskFactory()))
		pretenders = append(pretenders, task.NewPretender(linkParams, sle.collectLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(updateParams, sle.storeUpdateTaskFactory()))
	} else {
		var addParams = sle.newStoreAddParams(brokerParams)

		pretenders = append(pretenders, task.NewPretender(linkParams, sle.exportLinkTaskFactory(writer)))
		pretenders = append(pretenders, task.NewPretender(addParams, sle.storeAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(pretenders...)
}

func (sle *StoreLockerExecutor) Proceed() {
	var brokerParams = sle.accountParams.BrokerParams()
	var configParams = sle.newConfigParams()
	var linkParams = sle.newDataLinkParams(brokerParams, configParams)
	var writer = sle.dataLinksWriterFactory(sle.Ledger, configParams.GetConfigPath())
	var plan = task.NewPlan("locker-source", sle.Ledger.Logger)
	var workers []task.Worker

	if sle.addHandle == "" {
		var updateParams = sle.newStoreUpdateParams(brokerParams, configParams)

		// Collecting propSpecs of a datalink requires oem/handle:version. We get this from the existing locker sources for the account
		workers = append(workers, task.NewWorker(updateParams.StoreFindParams, sle.storeFindTaskFactory()))
		workers = append(workers, task.NewWorker(updateParams.StoreFindParams.DataLinkParams, sle.collectLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(updateParams, sle.storeUpdateTaskFactory()))
	} else {
		var addParams = sle.newStoreAddParams(brokerParams)

		workers = append(workers, task.NewWorker(linkParams, sle.exportLinkTaskFactory(writer)))
		workers = append(workers, task.NewWorker(addParams, sle.storeAddTaskFactory()))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (sle *StoreLockerExecutor) Update(handleArg, keyArg, valueArg string) error {
	var err error

	if err = sle.BaseLockerExecutor.Update(handleArg, keyArg, valueArg); err == nil {
		sle.Cli.Exec(sle.Ledger, sle)
		return nil
	}

	return err
}

func (sle *StoreLockerExecutor) newStoreAddParams(brokerParams *broker.Broker) *locker.StoreAddParams {
	return &locker.StoreAddParams{
		BaseParams: *sle.newBaseParams(sle.handleArg, sle.optionLocker),
		Broker:     *brokerParams,
		LinkParams: locker.LinkParams{
			Oem:     sle.addOem,
			Handle:  sle.addHandle,
			Version: sle.addVersion,
		},
	}
}

func (sle *StoreLockerExecutor) newStoreUpdateParams(brokerParams *broker.Broker, configParams *shared.ConfigParams) *locker.StoreUpdateParams {
	var propertyParams = &locker.PropertyParams{
		Key:   sle.keyArg,
		Value: sle.valueArg,
	}

	if sle.secretArg != nil {
		propertyParams.Secret = sle.secretArg
	}

	return &locker.StoreUpdateParams{
		StoreFindParams: &locker.StoreFindParams{
			BaseParams: *sle.newBaseParams(sle.handleArg, sle.optionLocker),
			DataLinkParams: &broker.DataLinkParams{
				Broker:       *brokerParams,
				ConfigParams: *configParams,
			},
		},
		PropertyParams: *propertyParams,
	}
}

func NewStore(ledger *config.Ledger, lkCli *Cli) *cobra.Command {
	var strAddOptions = NewStoreAddOptions()
	var strPublishOptions = NewStorePublishOptions()
	var strUpdateOptions = NewStoreUpdateOptions()
	var strAddCmd = store.NewAddStore(newStoreAddExecutorFactory(ledger, lkCli, strAddOptions))
	var strPublishCmd = store.NewPublishStore(newStorePublishExecutorFactory(ledger, lkCli, strPublishOptions))
	var strUpdateCmd = store.NewUpdateStore(newStoreUpdateExecutorFactory(ledger, lkCli, strUpdateOptions))
	var strCmd = &cobra.Command{
		Use:     "store",
		Aliases: []string{"str"},
		Short:   "Manages data stores under a Locker",
	}

	strCmd.AddCommand(strAddCmd)
	strCmd.AddCommand(strUpdateCmd)
	strCmd.AddCommand(strPublishCmd)
	ledger.Register(strAddCmd, strAddOptions.allDefiners()...)
	ledger.Register(strPublishCmd, strPublishOptions.allDefiners()...)
	ledger.Register(strUpdateCmd, strUpdateOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(strAddCmd, ledger, strAddOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(strPublishCmd, ledger, strPublishOptions.optionAccount)
	cli.AutoBridge.Accounts().Option(strUpdateCmd, ledger, strUpdateOptions.optionAccount)
	return strCmd
}

func NewStoreLockerExecutor(ctx context.Context, ledger *config.Ledger, lkCli *Cli, options *LockerOptions) *StoreLockerExecutor {
	return &StoreLockerExecutor{
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

		storeAddTaskFactory:    locker.NewStoreAddTask,
		storeFindTaskFactory:   locker.NewStoreFindTask,
		storeUpdateTaskFactory: locker.NewStoreUpdateTask,
	}
}

func NewStorePublishExecutor(ctx context.Context, ledger *config.Ledger, lkCli *Cli, options *PublishOptions) *StorePublishExecutor {
	return &StorePublishExecutor{
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

		storeFindTaskFactory:    locker.NewStoreFindTask,
		storePublishTaskFactory: locker.NewStorePublishTask,
		storeSyncTaskFactory:    locker.NewStoreSyncTask,
	}
}

func NewStoreAddOptions() *LockerOptions {
	return &LockerOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Store.Add.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Store.Add.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func NewStorePublishOptions() *PublishOptions {
	return &PublishOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Store.Publish.Account).
			BuildStringOption(),
		optionDescription: cli.Options.Stores.Description().
			WithKeys(&schema.Genaiz.Store.Publish.Description).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Store.Publish.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
		optionName: cli.Options.Stores.Name().
			WithKeys(&schema.Genaiz.Store.Publish.Name).
			BuildStringOption(),
		optionVisibility: cli.Options.Stores.Visibility().
			WithKeys(&schema.Genaiz.Store.Publish.Visibility).
			WithDefaultValue(broker.VisibilityPrivate).
			BuildStringOption(),
	}
}

func NewStoreUpdateOptions() *LockerOptions {
	return &LockerOptions{
		optionAccount: cli.Options.Lockers.Account().
			WithKeys(&schema.Genaiz.Store.Update.Account).
			BuildStringOption(),
		optionLocker: cli.Options.Lockers.Path().
			WithKeys(&schema.Genaiz.Store.Update.Locker).
			WithParam("locker").
			WithShort("").
			BuildStringOption(),
	}
}

func newStoreAddExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *LockerOptions) store.AddExecutorFactory {
	return func(command *cobra.Command) store.AddExecutor {
		return NewStoreLockerExecutor(command.Context(), ledger, lkCli, options)
	}
}

func newStorePublishExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *PublishOptions) store.PublishExecutorFactory {
	return func(command *cobra.Command) store.PublishExecutor {
		return NewStorePublishExecutor(command.Context(), ledger, lkCli, options)
	}
}

func newStoreUpdateExecutorFactory(ledger *config.Ledger, lkCli *Cli, options *LockerOptions) store.UpdateExecutorFactory {
	return func(command *cobra.Command) store.UpdateExecutor {
		return NewStoreLockerExecutor(command.Context(), ledger, lkCli, options)
	}
}
