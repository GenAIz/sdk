package sf

import (
	"context"
	"path/filepath"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/recipe"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
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
	ce.Repo.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.FolderPath,
	},
		&ce.optionArches.Option,
		&ce.optionConfigType.Option,
		&ce.optionFqdn.Option,
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
	var initParams = makeInitParams(ce.Repo, ce.InitOptions)
	var recipeName = ce.Repo.GetString(ce.optionRecipe)
	var plan = task.NewPlan("Create", ce.Repo.Logger)
	var builder = makeInitBuilder(ce.Cli)
	var workers []task.Worker

	ce.Repo.DisplayChangeDir()
	workers = append(workers, task.NewPretender(createParams, ce.createTaskFactory()))

	if recipeName != "" {
		var recipeParams = ce.makeRecipeParams(recipeName, ce.FolderPath)

		workers = append(workers, task.NewPretender(recipeParams, ce.recipeTaskFactory(ce.Repo.TemplatePaths...)))
	}

	workers = append(workers, task.NewPretender(initParams, ce.initTaskFactory(builder)))
	plan.Sequence(workers...)
}

func (ce *CreateExecutor) Proceed() {
	var builder = makeInitBuilder(ce.Cli)
	var createParams = ce.makeCreateParams()
	var initParams = makeInitParams(ce.Repo, ce.InitOptions)
	var recipeName = ce.Repo.GetString(ce.optionRecipe)
	var plan = task.NewPlan("Create", ce.Repo.Logger)

	var workers = []task.Worker{
		task.NewWorker(createParams, ce.createTaskFactory()),
		task.NewWorker(initParams, ce.initTaskFactory(builder)),
	}

	if recipeName != "" {
		var recipeParams = ce.makeRecipeParams(recipeName, ce.FolderPath)

		workers = append(workers, task.NewWorker(recipeParams, ce.recipeTaskFactory(ce.Repo.TemplatePaths...)))
	}

	plan.Sequence(workers...)
}

func (ce *CreateExecutor) makeCreateParams() *layout.CreateParams {
	return &layout.CreateParams{
		ConfigName: ce.Repo.ConfigName,
		ConfigType: toConfigType(ce.Repo, ce.optionConfigType),
		FolderPath: ce.FolderPath,
	}
}

func (ce *CreateExecutor) makeRecipeParams(recipeName string, folderPath string) *layout.RecipeParams {
	var handleName = ce.Repo.GetString(ce.optionHandle)
	var recipeOptions = []*config.Option{
		&ce.optionArches.Option,
		&ce.optionFqdn.Option,
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
		Options:      config.MapOptionsByEnvKey(ce.Repo, recipeOptions...),
		Parameters:   config.MapOptionsByParam(ce.Repo, recipeOptions...),
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
		co.optionFqdn,
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

func NewCreate(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewCreateOptions()
	var create = &cobra.Command{
		Use:     "create FOLDER_NAME",
		Short:   "Creates a Smart Function from scratch",
		Long:    "Creates a Smart Function from scratch, interactively by default, optionally using a selected template",
		Example: "genaiz sf create smart-function-1",
		Args: cobra.MatchAll(cobra.ExactArgs(1), func(cmd *cobra.Command, args []string) error {
			if !options.optionHandle.Validator(args[0]) {
				return errors.New("invalid folder name, only alphanumeric characters, dots, dashes and underscores are allowed")
			}

			return nil
		}),
		Run: func(cmd *cobra.Command, args []string) {
			repo.InitValue(options.optionHandle, filepath.Base(args[0]))
			cli.Exec(repo, NewCreateExecutor(cmd.Context(), repo, cli, options, args[0]))
		},
	}

	repo.Register(create, options.allDefiners()...)
	return create
}

func NewCreateExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *CreateOptions, folderPath string) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Repo:    repo,
		},
		CreateOptions: options,
		FolderPath:    folderPath,

		createTaskFactory: layout.NewCreateTask,
		initTaskFactory:   layout.NewInitTask,
		recipeTaskFactory: layout.NewRecipeTask,
	}
}

func NewCreateOptions() *CreateOptions {
	var createCmd = "Create"
	var publishOptions = newPublishOptions(createCmd)

	// Disable defaultSetter for create, since it doesn't exist yet
	publishOptions.optionHandle.DefaultSetter = nil

	return &CreateOptions{
		InitOptions: &InitOptions{
			PublishOptions:    publishOptions,
			optionConfigType:  newOptionConfigType(createCmd),
			optionInteractive: newOptionInteractive(),
			optionMountInput:  newOptionMountInput(createCmd, false),
			optionMountOutput: newOptionMountOutput(createCmd, false),
		},
		optionRecipe: newOptionRecipe(),
	}
}

func newOptionRecipe() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Create.Recipe",
			Param: "recipe",
			Short: "r",
			Usage: "name of a recipe to use as template for the Smart Function",
		},
	}
}

func toConfigType(repo *config.Repo, option *config.StringOption) *layout.ConfigType {
	var configTypeString = repo.GetString(option)
	var result *layout.ConfigType
	var err error

	if configTypeString != "" {
		result, err = layout.ConfigTypes.FromString(configTypeString)
		lang.HandleExit(err)
	}

	return result
}
