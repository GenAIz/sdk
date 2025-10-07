package sf

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/recipe"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type CreateTaskFactory func() *task.Task[layout.CreateParams]

type RecipeTaskFactory func(...string) *task.Task[layout.RecipeParams]

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions
	FolderPath string

	createTaskFactory CreateTaskFactory
	initTaskFactory   InitTaskFactory
	recipeTaskFactory RecipeTaskFactory
}

func (ce *CreateExecutor) Display() {
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.FolderPath,
	},
		&ce.optionArches.Option,
		&ce.optionConfigType.Option,
		&ce.optionHandle.Option,
		&ce.optionName.Option,
		&ce.optionType.Option,
		&ce.optionOem.Option,
		&ce.optionRecipe.Option,
		&ce.optionVersion.Option,
	)
}

func (ce *CreateExecutor) Pretend() {
	var createParams = ce.makeCreateParams()
	var initParams = makeInitParams(ce.Ledger, ce.InitOptions)
	var recipeName = ce.Ledger.GetString(ce.optionRecipe)
	var plan = task.NewPlan("Create", ce.Ledger.Logger)
	var builder = ce.makeCreateBuilder(ce.Ledger, ce.Cli)
	var workers []task.Worker

	ce.Ledger.DisplayChangeDir()
	workers = append(workers, task.NewPretender(createParams, ce.createTaskFactory()))

	if recipeName != "" {
		var recipeParams = ce.makeRecipeParams(recipeName, ce.FolderPath)

		workers = append(workers, task.NewPretender(recipeParams, ce.recipeTaskFactory(ce.Ledger.TemplatePaths...)))
	}

	workers = append(workers, task.NewPretender(initParams, ce.initTaskFactory(builder)))
	plan.Sequence(workers...)
}

func (ce *CreateExecutor) Proceed() {
	var builder = ce.makeCreateBuilder(ce.Ledger, ce.Cli)
	var createParams = ce.makeCreateParams()
	var initParams = makeInitParams(ce.Ledger, ce.InitOptions)
	var recipeName = ce.Ledger.GetString(ce.optionRecipe)
	var plan = task.NewPlan("Create", ce.Ledger.Logger)

	var workers = []task.Worker{
		task.NewWorker(createParams, ce.createTaskFactory()),
		task.NewWorker(initParams, ce.initTaskFactory(builder)),
	}

	if recipeName != "" {
		var recipeParams = ce.makeRecipeParams(recipeName, ce.FolderPath)

		workers = append(workers, task.NewWorker(recipeParams, ce.recipeTaskFactory(ce.Ledger.TemplatePaths...)))
	}

	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
}

func (ce *CreateExecutor) makeCreateBuilder(ledger *config.Ledger, sfCli *Cli) layout.ConfigWriter {
	var dockerTag = ledger.GetString(sfCli.optionDockerTag)
	var result = &InitWriter{
		PublishOptions:   NewPublishOptions(sfCli),
		RunOptions:       NewRunOptions(sfCli),
		buildTagKeys:     &schema.Genaiz.Function.Build.Tag,
		buildVersionKeys: &schema.Genaiz.Function.Build.Version,
		vp:               viper.New(),
	}

	return result.WithTag(dockerTag)
}

func (ce *CreateExecutor) makeCreateParams() *layout.CreateParams {
	var configType, err = ce.Ledger.GetConfigType(ce.optionConfigType)

	lang.HandleExit(err)
	return &layout.CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   ce.Ledger.ConfigName,
			ConfigType:   configType,
			ConfigFolder: ce.FolderPath,
		},
	}
}

func (ce *CreateExecutor) makeRecipeParams(recipeName string, folderPath string) *layout.RecipeParams {
	var handleName = ce.Ledger.GetString(ce.optionHandle)
	var recipeOptions = []*config.Option{
		&ce.optionArches.Option,
		&ce.optionHandle.Option,
		&ce.optionMountInput.Option,
		&ce.optionMountOutput.Option,
		&ce.optionName.Option,
		&ce.optionOem.Option,
		&ce.optionType.Option,
		&ce.optionVersion.Option,
	}

	return &layout.RecipeParams{
		Name:         recipeName,
		InstanceName: handleName,
		Type:         recipe.TypeFunction,
		Destination:  folderPath,
		Options:      config.MapOptionsByEnvKey(ce.Ledger, recipeOptions...),
		Parameters:   config.MapOptionsByParam(ce.Ledger, recipeOptions...),
	}
}

type CreateOptions struct {
	*InitOptions

	optionRecipe *config.StringOption
}

func (co *CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionArches,
		co.optionConfigType,
		co.optionHandle,
		co.optionName,
		co.optionType,
		co.optionInteractive,
		co.optionMountInput,
		co.optionMountOutput,
		co.optionOem,
		co.optionRecipe,
		co.optionVersion,
	}
}

func NewCreate(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewCreateOptions(cli)
	var create = &cobra.Command{
		Use:     "create FOLDER_NAME",
		Short:   "Creates a Smart Function from scratch",
		Long:    "Creates a Smart Function from scratch, optionally using a recipe",
		Example: "genaiz sf create smart-function-1",
		Args: cobra.MatchAll(cobra.ExactArgs(1), func(cmd *cobra.Command, args []string) error {
			if !config.Validation.FolderName(args[0]) {
				return errors.New("invalid folder name, only alphanumeric characters, dots, dashes and underscores are allowed")
			}

			return nil
		}),
		Run: func(cmd *cobra.Command, args []string) {
			ledger.InitValue(options.optionHandle, filepath.Base(args[0]))
			cli.Exec(ledger, NewCreateExecutor(cmd.Context(), ledger, cli, options, args[0]))
		},
	}

	ledger.Register(create, options.allDefiners()...)
	return create
}

func NewCreateExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *CreateOptions, folderPath string) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		CreateOptions: options,
		FolderPath:    folderPath,

		createTaskFactory: layout.NewCreateTask,
		initTaskFactory:   layout.NewInitTask,
		recipeTaskFactory: layout.NewRecipeTask,
	}
}

func NewCreateOptions(sfCli *Cli) *CreateOptions {
	var parentOpt = cli.Options.Configs.SolutionPath().
		WithKeys(&schema.Genaiz.Function.Create.SolutionPath).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return dirz.WorkingDirOrPanic()
		}).BuildStringOption()
	var handleOpt = cli.Options.Functions.Handle().
		WithKeys(&schema.Genaiz.Function.Create.Handle).
		BuildStringOption()

	return &CreateOptions{
		InitOptions: &InitOptions{
			optionArches: cli.Options.Functions.Arches().
				WithKeys(&schema.Genaiz.Function.Create.Arches).
				BuildListOption(),
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.Function.Create.ConfigType).
				WithDefaultGetter(sfCli.ParentConfigType(parentOpt)).
				BuildStringOption(),
			optionHandle: handleOpt,
			optionInteractive: cli.Options.Modes.Interactive().
				BuildBoolOption(),
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Create.MountInput).
				Validated(false).
				BuildStringOption(),
			optionMountOutput: cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Create.MountOutput).
				Validated(false).
				BuildStringOption(),
			optionName: cli.Options.Functions.Name().
				WithKeys(&schema.Genaiz.Function.Create.Name).
				WithUsage("defaults to the handle value if not provided").
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return ledger.GetString(handleOpt)
				}).BuildStringOption(),
			optionOem: cli.Options.Functions.Oem().
				WithKeys(&schema.Genaiz.Function.Create.Oem).
				WithDefaultGetter(sfCli.ParentOem(parentOpt)).
				BuildStringOption(),
			optionType: cli.Options.Functions.Type().
				WithKeys(&schema.Genaiz.Function.Create.Type).
				BuildStringOption(),
			optionVersion: cli.Options.Functions.Version().
				WithKeys(&schema.Genaiz.Function.Create.Version).
				WithDefaultGetter(sfCli.ParentVersion(parentOpt)).
				BuildStringOption(),
		},
		optionRecipe: cli.Options.Functions.Recipe().
			WithKeys(&schema.Genaiz.Function.Create.Recipe).
			BuildStringOption(),
	}
}
