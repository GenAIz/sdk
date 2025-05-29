package sf

import (
	"context"

	"github.com/pkg/errors"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
)

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions
	FolderPath string
}

func (ce *CreateExecutor) Display() {
	ce.Repo.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.FolderPath,
	},
		&ce.optionArchTypes.Option,
		&ce.optionConfigType.Option,
		&ce.optionFqdn.Option,
		&ce.optionFunctionName.Option,
		&ce.optionFunctionType.Option,
		&ce.optionOem.Option,
		&ce.optionVersion.Option,
	)
}

func (ce *CreateExecutor) Pretend() {
	var createParams = ce.makeCreateParams()
	var initParams = makeInitParams(ce.Repo, ce.InitOptions)
	var writer = makeInitBuilder(ce.Cli)

	ce.Repo.DisplayChangeDir()
	layout.NewCreateTask().Pretend(createParams, ce.Repo.Logger)
	layout.NewInitTask(writer).Pretend(initParams, ce.Repo.Logger)
}

func (ce *CreateExecutor) Proceed() {
	var builder = makeInitBuilder(ce.Cli)
	var createParams = ce.makeCreateParams()
	var initParams = makeInitParams(ce.Repo, ce.InitOptions)
	var plan = &task.Plan[layout.InitParams]{
		Logger: ce.Repo.Logger,
		OnError: func(err error) {
			ce.Repo.Logger.Errorf("Could not run create, error: %s", err)
			cobra.CheckErr(err)
		},
	}

	plan.Sequence(
		task.Execution(createParams, layout.NewCreateTask()),
		task.Execution(initParams, layout.NewInitTask(builder)),
	)
}

func (ce *CreateExecutor) makeCreateParams() *layout.CreateParams {
	return &layout.CreateParams{
		ConfigName: ce.Repo.ConfigName,
		ConfigType: toConfigType(ce.Repo, ce.optionConfigType),
		FolderPath: ce.FolderPath,
	}
}

type CreateOptions struct {
	*InitOptions
}

func (co *CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionArchTypes,
		co.optionConfigType,
		co.optionFqdn,
		co.optionFunctionName,
		co.optionFunctionType,
		co.optionInteractive,
		co.optionMountInput,
		co.optionMountOutput,
		co.optionOem,
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
			if !options.optionFunctionName.Validator(args[0]) {
				return errors.New("invalid folder name, only alphanumeric characters, dots, dashes and underscores are allowed")
			}

			return nil
		}),
		Run: func(cmd *cobra.Command, args []string) {
			repo.InitValue(options.optionFunctionName, args[0])
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
	}
}

func NewCreateOptions() *CreateOptions {
	var createCmd = "Create"

	return &CreateOptions{
		InitOptions: &InitOptions{
			PublishOptions:    newPublishOptions(createCmd),
			optionConfigType:  newOptionConfigType(createCmd),
			optionInteractive: newOptionInteractive(),
			optionMountInput:  newOptionMountInput(createCmd, false),
			optionMountOutput: newOptionMountOutput(createCmd, false),
		},
	}
}

func toConfigType(repo *config.Repo, option *config.StringOption) *layout.ConfigType {
	var configTypeString = repo.GetString(option)
	var result *layout.ConfigType
	var err error

	if configTypeString != "" {
		result, err = layout.ConfigTypes.FromString(configTypeString)
		cobra.CheckErr(err)
	}

	return result
}
