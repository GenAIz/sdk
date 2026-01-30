package sf

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/cmd/sf/store"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type StoreExecutor struct {
	DataLinkExecutor

	addParams     *broker.DataLinkParams
	rmParams      *broker.DataLinkParams
	innerType     *config.StringOption
	innerStores   *config.ListOption
	updatedStores []string

	initTaskFactory        InitTaskFactory
	listLinksTaskFactory   ListLinksTaskFactory
	syncLinksTaskFactory   SyncLinksTaskFactory
	dataLinksWriterFactory dk.DataLinksWriterFactory
}

func (se *StoreExecutor) Add(fqdnv string) error {
	var err error

	if err = se.validateConnector(se.innerType); err == nil {
		var dataStores = se.Ledger.GetList(se.innerStores)
		var params = se.makeDataLinkParams(fqdnv)
		var value = params.ToString()

		if slices.Contains(dataStores, value) {
			return fmt.Errorf("data store [%s] is already configured", value)
		} else {
			dataStores = append(dataStores, value)
		}

		se.addParams = params
		se.updatedStores = dataStores
		se.Cli.Exec(se.Ledger, se)
	}

	return err
}

func (se *StoreExecutor) Display() {
	var params *broker.DataLinkParams

	if se.addParams != nil {
		_, _ = fmt.Printf("Adding the following data store:\n")
		params = se.addParams
	} else if se.rmParams != nil {
		_, _ = fmt.Printf("Removing the following data store:\n")
		params = se.rmParams
	}

	if params != nil {
		var detailsMap = map[string]string{}

		detailsMap["handle"] = params.Handle
		detailsMap["oem"] = params.Oem
		detailsMap["version"] = params.Version
		detailsMap["no-validation"] = cast.ToString(params.NoValidation)
		se.Ledger.DisplayOptionsWithMap(&detailsMap)
	}
}

func (se *StoreExecutor) Pretend() {
	var builder = makeInitBuilder(se.Ledger, se.Cli)
	var initParams = se.makeInitParams()
	var plan = task.NewPlan("DataStore", se.Ledger.Logger)
	var workers []task.Worker

	if se.addParams != nil {
		var noValidation = se.Ledger.GetBool(se.optionNoValidation)

		if !noValidation {
			var syncConfigParams = se.makeSyncParams(se.addParams)
			var dataLinkWriter = se.dataLinksWriterFactory(se.Ledger, syncConfigParams.GetConfigFile())

			workers = append(workers, task.NewPretender(syncConfigParams, se.syncLinksTaskFactory(dataLinkWriter)))
		}

		workers = append(workers, task.NewPretender(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(builder)))
	} else {
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(builder)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *StoreExecutor) Proceed() {
	var builder = makeInitBuilder(se.Ledger, se.Cli)
	var initParams = se.makeInitParams()
	var plan = task.NewPlan("DataStore", se.Ledger.Logger)
	var workers []task.Worker

	if se.addParams != nil {
		var noValidation = se.Ledger.GetBool(se.optionNoValidation)

		if !noValidation {
			var syncConfigParams = se.makeSyncParams(se.addParams)
			var dataLinkWriter = se.dataLinksWriterFactory(se.Ledger, syncConfigParams.GetConfigFile())

			workers = append(workers, task.NewWorker(syncConfigParams, se.syncLinksTaskFactory(dataLinkWriter)))
		}

		workers = append(workers, task.NewWorker(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(builder)))
	} else if se.rmParams != nil {
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(builder)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *StoreExecutor) Remove(fqdnv string) {
	var dataStores = se.Ledger.GetList(se.innerStores)
	var params = se.makeDataLinkParams(fqdnv)
	var value = params.ToString()

	se.rmParams = params
	se.updatedStores = slices.DeleteFunc(dataStores, func(s string) bool {
		return strings.EqualFold(value, s)
	})
	se.Cli.Exec(se.Ledger, se)
}

func (se *StoreExecutor) makeInitParams() *layout.InitParams {
	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: se.Ledger.ConfigName,
			},
		},
		DataStores: se.updatedStores,
	}
}

func NewStore(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var addOptions = NewStoreAddOptions(sfCli)
	var addCommand = store.NewAddStore(newStoreAddFactory(ledger, sfCli, addOptions))
	var removeOptions = NewStoreRemoveOptions(sfCli)
	var removeCommand = store.NewRemoveStore(newStoreRemoveFactory(ledger, sfCli, removeOptions))
	var storeCmd = &cobra.Command{
		Use:     "store",
		Aliases: []string{"str"},
		Short:   "Manages data Store configurations for Smart Functions",
	}

	ledger.Register(addCommand, addOptions.addDefiners()...)
	ledger.Register(removeCommand, removeOptions.removeDefiners()...)
	storeCmd.AddCommand(addCommand)
	storeCmd.AddCommand(removeCommand)
	return storeCmd
}

func NewStoreExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *DataLinkOptions) *StoreExecutor {
	return &StoreExecutor{
		DataLinkExecutor: DataLinkExecutor{
			BaseExecutor: BaseExecutor{
				Context: ctx,
				Cli:     sfCli,
				Ledger:  ledger,
			},
			DataLinkOptions: options,
		},

		innerType: cli.Options.Functions.Type().
			WithKeys(&schema.Genaiz.Function.Publish.Type).
			BuildStringOption(),
		innerStores: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataStores).
			BuildListOption(),

		initTaskFactory:        layout.NewInitTask,
		listLinksTaskFactory:   broker.NewDataLinkFindTask,
		syncLinksTaskFactory:   broker.NewDataLinkExportTask,
		dataLinksWriterFactory: dk.NewDataLinksWriter,
	}
}

func NewStoreAddOptions(sfCli *Cli) *DataLinkOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &DataLinkOptions{
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreAdd.Handle).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreAdd.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreAdd.Version).
			BuildStringOption(),
		optionNoValidation: cli.Options.DataLinks.NoValidation().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreAdd.NoValidation).
			BuildBoolOption(),
	}
}

func NewStoreRemoveOptions(sfCli *Cli) *DataLinkOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &DataLinkOptions{
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreRemove.Handle).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreRemove.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.Function.Publish.DataStoreRemove.Version).
			BuildStringOption(),
	}
}

func newStoreAddFactory(ledger *config.Ledger, sfCli *Cli, addOptions *DataLinkOptions) store.AddExecutorFactory {
	return func(cmd *cobra.Command) store.AddExecutor {
		return NewStoreExecutor(cmd.Context(), ledger, sfCli, addOptions)
	}
}

func newStoreRemoveFactory(ledger *config.Ledger, sfCli *Cli, removeOptions *DataLinkOptions) store.RemoveExecutorFactory {
	return func(cmd *cobra.Command) store.RemoveExecutor {
		return NewStoreExecutor(cmd.Context(), ledger, sfCli, removeOptions)
	}
}
