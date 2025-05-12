package sf

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type RunExecutor struct {
	BaseExecutor
	*RunOptions
}

func (re *RunExecutor) Display() {
	displayRunOptions(re.BaseExecutor, re.RunOptions)
}

func (re *RunExecutor) Pretend() {
	var params = re.makeRunParams()

	re.Repo.DisplayChangeDir()
	docker.NewRunTask().Pretend(params, re.Repo.Logger)
}

func (re *RunExecutor) Proceed() {
	var runParams = re.makeRunParams()

	execRunParamsTask(re.BaseExecutor, re.RunOptions, runParams, docker.NewRunTask())
}

func (re *RunExecutor) makeRunParams() *docker.ContainerParams {
	return makeRunParams(re.BaseExecutor, re.RunOptions)
}

type RunOptions struct {
	optionMountInput  *config.StringOption
	optionMountLog    *config.StringOption
	optionMountOutput *config.StringOption
	optionMountVar    *config.StringOption
	optionRunImage    *config.StringOption
	optionRunPrefix   *config.StringOption
	rebuildImage      bool
}

func (ro *RunOptions) allDefiners() []config.Definer {
	return []config.Definer{
		ro.optionRunImage,
		ro.optionRunPrefix,
		ro.optionMountInput,
		ro.optionMountLog,
		ro.optionMountOutput,
		ro.optionMountVar,
	}
}

func NewRun(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewRunOptions(cli)
	var run = &cobra.Command{
		Use:     "run",
		Short:   "Runs a Smart Function detached from the current shell",
		Long:    "Runs a Smart Function image detached, building it first if necessary, assigning it a disposable container",
		Example: "genaiz sf run --image genaiz.com/sf/smartfunc:latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			repo.FromWorkDir(options.optionMountInput, cmd.Flags())
			repo.FromWorkDir(options.optionMountLog, cmd.Flags())
			repo.FromWorkDir(options.optionMountOutput, cmd.Flags())
			repo.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.optionRunImage)
			cli.Exec(repo, NewRunExecutor(cmd.Context(), repo, cli, options))
		},
	}

	repo.Register(run, options.allDefiners()...)
	return run
}

func NewRunExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *RunOptions) *RunExecutor {
	return &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Repo:    repo,
		},
		RunOptions: options,
	}
}

func NewRunOptions(cli *Cli) *RunOptions {
	var outputOption = NewOptionMountOutput()

	return &RunOptions{
		optionMountInput:  NewOptionMountInput(),
		optionMountLog:    NewOptionMountLog(outputOption),
		optionMountOutput: outputOption,
		optionMountVar:    NewOptionMountVar(outputOption),
		optionRunImage:    NewOptionRunImage(cli),
		optionRunPrefix:   NewOptionContainerPrefix("Run", cli),
	}
}

func NewOptionMountInput() *config.StringOption {
	return newOptionMountInput("Run")
}

func NewOptionMountLog(defaultOption *config.StringOption) *config.StringOption {
	return newOptionMountLog("Run", defaultOption)
}

func NewOptionMountOutput() *config.StringOption {
	return newOptionMountOutput("Run")
}

func NewOptionMountVar(defaultOption *config.StringOption) *config.StringOption {
	return newOptionMountVar("Run", defaultOption)
}

func NewOptionRunImage(cli *Cli) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Run.Image",
			Param: "image",
			Usage: "reference to an image with or without the version",
			DefaultSetter: func(repo *config.Repo) any {
				var tag = repo.GetString(cli.optionDockerTag)
				var version = repo.GetString(cli.optionDockerVersion)

				if version == "" || strings.Contains(tag, ":") {
					return tag
				}

				return tag + ":" + version
			},
		},
	}
}

func displayRunOptions(be BaseExecutor, ro *RunOptions) {
	var options = []*config.Option{
		&ro.optionRunImage.Option,
		&ro.optionRunPrefix.Option,
		&ro.optionMountInput.Option,
		&ro.optionMountLog.Option,
		&ro.optionMountOutput.Option,
		&ro.optionMountVar.Option,
	}

	options = append(options, be.Cli.SfOptions()...)
	be.Repo.DisplayOptions(options...)
}

func execRunParamsTask(be BaseExecutor, ro *RunOptions, params *docker.ContainerParams, runTask *task.Task[docker.ContainerParams]) {
	var plan = &task.Plan[docker.ContainerParams]{
		Logger: be.Repo.Logger,
		OnError: func(err error) {
			be.Repo.Logger.Errorf("Could not %s on image %s, error: %s", runTask.Name, params.DockerImage, err)
			cobra.CheckErr(err)
		},
		OnSuccess: func(out string) {
			if out != "" {
				fmt.Printf("%s\n", out)
			}
		},
	}
	var executions []func(*task.State) error

	if ro.rebuildImage {
		executions = append(executions, task.Execution(makeBuildParams(be), docker.NewBuildTask()))
	}

	executions = append(executions, task.Execution(params, runTask))
	plan.Sequence(executions...)
}

func makeRunParams(be BaseExecutor, ro *RunOptions) *docker.ContainerParams {
	return &docker.ContainerParams{
		RunParams: docker.RunParams{
			Env:      task.Env{Context: be.Context},
			Attached: false,
			Dispose:  true,
		},
		DockerImage: be.Repo.GetString(ro.optionRunImage),
		MountInput:  be.Repo.GetString(ro.optionMountInput),
		MountLog:    be.Repo.GetString(ro.optionMountLog),
		MountOutput: be.Repo.GetString(ro.optionMountOutput),
		MountVar:    be.Repo.GetString(ro.optionMountVar),
		Prefix:      be.Repo.GetString(be.Cli.optionDockerTag),
	}
}

func needsRebuildingImage(cmd *cobra.Command, option *config.StringOption) bool {
	var imageFlag = cmd.Flags().Lookup(option.Param)

	return imageFlag.Value.String() == ""
}

func newOptionCmdImage(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Image",
			Param: "image",
			Usage: "reference to an image with or without the version",
			DefaultGetter: func(repo *config.Repo) any {
				return repo.GetValue("SF.Run.Image")
			},
		},
	}
}

func newOptionMountInput(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".InputPath",
			Param: "in",
			Usage: "path of the input files folder, read-only, if any",
			Validator: func(value any) bool {
				return config.ValidateOptional(value, config.ValidateDir)
			},
		},
	}
}

func newOptionMountLog(cmd string, defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".LogPath",
			Param: "log",
			Usage: "path of the log files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			DefaultGetter: func(repo *config.Repo) any {
				return repo.Get(&defaultOption.Option)
			},
			Validator: func(value any) bool {
				return config.ValidateOptional(value, config.ValidateDirCreation)
			},
		},
	}
}

func newOptionMountOutput(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".OutputPath",
			Param: "out",
			Usage: "path of the output files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			Validator: func(value any) bool {
				return config.ValidateOptional(value, config.ValidateDirCreation)
			},
		},
	}
}

func newOptionMountVar(cmd string, defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".VarPath",
			Param: "var",
			Usage: "path of a solution state files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			DefaultGetter: func(repo *config.Repo) any {
				return repo.Get(&defaultOption.Option)
			},
			Validator: func(value any) bool {
				return config.ValidateOptional(value, config.ValidateDirCreation)
			},
		},
	}
}
