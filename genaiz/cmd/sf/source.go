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
	"genaiz.com/genaiz/cmd/sf/source"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type ListLinksTaskFactory func() *task.Task[broker.DataLinkParams]

type SourceExecutor struct {
	BaseExecutor
	*SourceOptions

	addParams      *broker.DataLinkParams
	rmParams       *broker.DataLinkParams
	innerSources   *config.ListOption
	updatedSources []string

	initTaskFactory      InitTaskFactory
	listLinksTaskFactory ListLinksTaskFactory
}

func (se *SourceExecutor) Add(fqdnv string) error {
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
	return nil
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
	var builder = makeInitBuilder(se.Ledger, se.Cli)
	var initParams = se.makeInitParams()
	var plan = task.NewPlan("AddSource", se.Ledger.Logger)
	var workers []task.Worker

	if se.addParams != nil {
		workers = append(workers, task.NewPretender(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(builder)))
	} else {
		workers = append(workers, task.NewPretender(initParams, se.initTaskFactory(builder)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (se *SourceExecutor) Proceed() {
	var builder = makeInitBuilder(se.Ledger, se.Cli)
	var initParams = se.makeInitParams()
	var plan = task.NewPlan("AddSource", se.Ledger.Logger)
	var workers []task.Worker

	if se.addParams != nil {
		workers = append(workers, task.NewWorker(se.addParams, se.listLinksTaskFactory()))
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(builder)))
	} else if se.rmParams != nil {
		workers = append(workers, task.NewWorker(initParams, se.initTaskFactory(builder)))
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

func (se *SourceExecutor) makeDataLinkParams(fqdnv string) *broker.DataLinkParams {
	var handle, oem, version string
	var validate = false

	if se.optionNoValidation != nil {
		validate = se.Ledger.GetBool(se.optionNoValidation)
	}

	if fqdnv != "" {
		oem, handle, version = parseFqdnv(fqdnv)
		se.Ledger.InitValue(se.optionHandle, handle)
		se.Ledger.InitValue(se.optionOem, oem)
		se.Ledger.InitValue(se.optionVersion, version)
	}

	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: se.Ledger.AuthFile,
		},
		Handle:       se.Ledger.GetString(se.optionHandle),
		Oem:          se.Ledger.GetString(se.optionOem),
		Version:      se.Ledger.GetString(se.optionVersion),
		NoValidation: validate,
	}
}

func (se *SourceExecutor) makeInitParams() *layout.InitParams {
	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: se.Ledger.ConfigName,
			},
		},
		DataSources: se.updatedSources,
	}
}

type SourceOptions struct {
	optionHandle       *config.StringOption
	optionNoValidation *config.BoolOption
	optionOem          *config.StringOption
	optionVersion      *config.StringOption
}

func (so SourceOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionHandle,
		so.optionNoValidation,
		so.optionOem,
		so.optionVersion,
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

	ledger.Register(addCommand, addOptions.allDefiners()...)
	sourceCmd.AddCommand(addCommand)
	sourceCmd.AddCommand(removeCommand)
	return sourceCmd
}

func NewSourceExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *SourceOptions) *SourceExecutor {
	return &SourceExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Cli:     sfCli,
			Ledger:  ledger,
		},
		SourceOptions: options,

		innerSources: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.DataSources).
			BuildListOption(),

		initTaskFactory:      layout.NewInitTask,
		listLinksTaskFactory: broker.NewDataLinkFindTask,
	}
}

func NewSourceAddOptions(sfCli *Cli) *SourceOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &SourceOptions{
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

func NewSourceRemoveOptions(sfCli *Cli) *SourceOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Init.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirParent()
		}).BuildStringOption()

	return &SourceOptions{
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

func newSourceAddFactory(ledger *config.Ledger, sfCli *Cli, addOptions *SourceOptions) source.AddExecutorFactory {
	return func(cmd *cobra.Command) source.AddExecutor {
		return NewSourceExecutor(cmd.Context(), ledger, sfCli, addOptions)
	}
}

func newSourceRemoveFactory(ledger *config.Ledger, sfCli *Cli, removeOptions *SourceOptions) source.RemoveExecutorFactory {
	return func(cmd *cobra.Command) source.RemoveExecutor {
		return NewSourceExecutor(cmd.Context(), ledger, sfCli, removeOptions)
	}
}

func parseFqdnv(fqdnv string) (string, string, string) {
	var parts = strings.SplitN(fqdnv, "/", 2)
	var oem, handle, version string

	if len(parts) == 2 {
		oem, parts = parts[0], parts[1:]
	}

	if len(parts) == 1 {
		parts = strings.SplitN(parts[0], ":", 2)
		handle = parts[0]

		if len(parts) == 2 {
			version = parts[1]
		}
	}

	return oem, handle, version
}
