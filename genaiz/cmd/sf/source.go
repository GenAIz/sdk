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
	"genaiz.com/genaiz/cmd/sf/source"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type SourceExecutor struct {
	DataLinkExecutor

	addParams      *broker.DataLinkParams
	rmParams       *broker.DataLinkParams
	innerType      *config.StringOption
	innerSources   *config.ListOption
	updatedSources []string

	initTaskFactory        InitTaskFactory
	listLinksTaskFactory   ListLinksTaskFactory
	syncLinksTaskFactory   SyncLinksTaskFactory
	dataLinksWriterFactory dk.DataLinksWriterFactory
}

func (se *SourceExecutor) Add(fqdnv string) error {
	var err error

	if err = se.validateConnector(se.innerType); err == nil {
		var dataSources = se.Ledger.GetList(se.innerSources)
		var params = se.makeDataLinkParams(fqdnv)
		var value = params.ToString()

		if slices.Contains(dataSources, value) {
			return fmt.Errorf("data source [%s] is already configured", value)
		} else {
			dataSources = append(dataSources, value)
		}

		se.addParams = params
		se.updatedSources = dataSources
		se.Cli.Exec(se.Ledger, se)
	}

	return err
}

func (se *SourceExecutor) Display() {
	var params *broker.DataLinkParams

	if se.addParams != nil {
		_, _ = fmt.Printf("Adding the following data source:\n")
		params = se.addParams
	} else if se.rmParams != nil {
		_, _ = fmt.Printf("Removing the following data source:\n")
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

func (se *SourceExecutor) Pretend() {
	var initParams = se.makeInitParams()
	var writer = newInitWriter(se.Cli)
	var plan = task.NewPlan("DataSource", se.Ledger.Logger)
	var workers []task.Worker

	if se.addParams != nil {
		var noValidation = se.Ledger.GetBool(se.optionNoValidation)

		if !noValidation {
			var syncConfigParams = se.makeSyncParams(se.addParams)
			var dataLinkWriter = se.dataLinksWriterFactory(se.Ledger, syncConfigParams.GetConfigFile())

			workers = append(workers, task.NewPretender(syncConfigParams, se.syncLinksTaskFactory(dataLinkWriter)))
		}

		workers = append(workers, task.NewPretender(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(writer)))
	} else {
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(writer)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *SourceExecutor) Proceed() {
	var plan = task.NewPlan("DataSource", se.Ledger.Logger)
	var initParams = se.makeInitParams()
	var writer = newInitWriter(se.Cli)
	var workers []task.Worker

	if se.addParams != nil {
		var noValidation = se.Ledger.GetBool(se.optionNoValidation)

		if !noValidation {
			var syncConfigParams = se.makeSyncParams(se.addParams)
			var dataLinkWriter = se.dataLinksWriterFactory(se.Ledger, syncConfigParams.GetConfigFile())

			workers = append(workers, task.NewWorker(syncConfigParams, se.syncLinksTaskFactory(dataLinkWriter)))
		}

		workers = append(workers, task.NewWorker(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(writer)))
	} else if se.rmParams != nil {
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(writer)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *SourceExecutor) Remove(fqdnv string) {
	var dataSources = se.Ledger.GetList(se.innerSources)
	var params = se.makeDataLinkParams(fqdnv)
	var value = params.ToString()

	se.rmParams = params
	se.updatedSources = slices.DeleteFunc(dataSources, func(s string) bool {
		return strings.EqualFold(value, s)
	})
	se.Cli.Exec(se.Ledger, se)
}

func (se *SourceExecutor) makeInitParams() *layout.InitParams {
	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName:   se.Ledger.ConfigName,
				ConfigFolder: se.Ledger.WorkDir,
			},
		},
		DataSources: se.updatedSources,
	}
}

func NewSource(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var addOptions = NewSourceAddOptions(sfCli)
	var addCommand = source.NewAddSource(newSourceAddFactory(ledger, sfCli, addOptions))
	var removeOptions = NewSourceRemoveOptions(sfCli)
	var removeCommand = source.NewRemoveSource(newSourceRemoveFactory(ledger, sfCli, removeOptions))
	var sourceCmd = &cobra.Command{
		Use:     "source",
		Aliases: []string{"src"},
		Short:   "Manages data source configurations for Smart Functions",
	}

	ledger.Register(addCommand, addOptions.addDefiners()...)
	ledger.Register(removeCommand, removeOptions.removeDefiners()...)
	sourceCmd.AddCommand(addCommand)
	sourceCmd.AddCommand(removeCommand)
	return sourceCmd
}

func NewSourceExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *DataLinkOptions) *SourceExecutor {
	return &SourceExecutor{
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
		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),

		initTaskFactory:        layout.NewInitTask,
		listLinksTaskFactory:   broker.NewDataLinkFindTask,
		syncLinksTaskFactory:   broker.NewDataLinkExportTask,
		dataLinksWriterFactory: dk.NewDataLinksWriter,
	}
}

func NewSourceAddOptions(sfCli *Cli) *DataLinkOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &DataLinkOptions{
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceAdd.Handle).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceAdd.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceAdd.Version).
			BuildStringOption(),
		optionNoValidation: cli.Options.DataLinks.NoValidation().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceAdd.NoValidation).
			BuildBoolOption(),
	}
}

func NewSourceRemoveOptions(sfCli *Cli) *DataLinkOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &DataLinkOptions{
		optionHandle: cli.Options.DataLinks.Handle().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceRemove.Handle).
			BuildStringOption(),
		optionOem: cli.Options.DataLinks.Oem().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceRemove.Oem).
			WithDefaultGetter(sfCli.ParentOem(parentOpt)).
			BuildStringOption(),
		optionVersion: cli.Options.DataLinks.Version().
			WithKeys(&schema.Genaiz.Function.Publish.DataSourceRemove.Version).
			BuildStringOption(),
	}
}

func newSourceAddFactory(ledger *config.Ledger, sfCli *Cli, addOptions *DataLinkOptions) source.AddExecutorFactory {
	return func(cmd *cobra.Command) source.AddExecutor {
		return NewSourceExecutor(cmd.Context(), ledger, sfCli, addOptions)
	}
}

func newSourceRemoveFactory(ledger *config.Ledger, sfCli *Cli, removeOptions *DataLinkOptions) source.RemoveExecutorFactory {
	return func(cmd *cobra.Command) source.RemoveExecutor {
		return NewSourceExecutor(cmd.Context(), ledger, sfCli, removeOptions)
	}
}
