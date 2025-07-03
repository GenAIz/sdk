package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type RunTaskFactory func() *task.Task[docker.ContainerParams]

type RunExecutor struct {
	BaseExecutor
	*RunOptions

	buildTaskFactory BuildTaskFactory
	runTaskFactory   RunTaskFactory
}

func (re *RunExecutor) Display() {
	displayRunOptions(re.BaseExecutor, re.RunOptions)
}

func (re *RunExecutor) Pretend() {
	var runParams = re.makeRunParams()

	re.Ledger.DisplayChangeDir()
	pretendRunParamsTask(re.BaseExecutor, re.RunOptions, runParams, re.buildTaskFactory, re.runTaskFactory)
}

func (re *RunExecutor) Proceed() {
	var runParams = re.makeRunParams()

	execRunParamsTask(re.BaseExecutor, re.RunOptions, runParams, re.buildTaskFactory, re.runTaskFactory)
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

func NewRun(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewRunOptions(cli)
	var run = &cobra.Command{
		Use:     "run",
		Short:   "Runs a Smart Function detached from the current shell",
		Long:    "Runs a Smart Function image detached, building it first if necessary, assigning it a disposable container",
		Example: "genaiz sf run --image genaiz.com/sf/smartfunc:latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.optionRunImage)
			cli.Exec(ledger, NewRunExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(run, options.allDefiners()...)
	return run
}

func NewRunExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *RunOptions) *RunExecutor {
	return &RunExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		RunOptions: options,

		buildTaskFactory: docker.NewBuildTask,
		runTaskFactory:   docker.NewRunTask,
	}
}

func NewRunOptions(cli *Cli) *RunOptions {
	return newRunOptions(cli, "Run")
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
	be.Ledger.DisplayOptions(options...)
}

func execRunParamsTask(be BaseExecutor, ro *RunOptions, params *docker.ContainerParams,
	buildTaskFactory BuildTaskFactory, runTaskFactory RunTaskFactory) {
	var plan = task.NewPlan("Run", be.Ledger.Logger)
	var workers []task.Worker

	if ro.rebuildImage {
		workers = append(workers, task.NewWorker(makeBuildParams(&be), buildTaskFactory()))
	}

	workers = append(workers, task.NewWorker(params, runTaskFactory()))
	plan.Sequence(workers...)
}

func makeRunParams(be BaseExecutor, ro *RunOptions) *docker.ContainerParams {
	return &docker.ContainerParams{
		RunParams: docker.RunParams{
			Env:      task.Env{Context: be.Context},
			Attached: false,
			Dispose:  true,
		},
		DockerImage: be.Ledger.GetString(ro.optionRunImage),
		MountInput:  be.Ledger.GetString(ro.optionMountInput),
		MountLog:    be.Ledger.GetString(ro.optionMountLog),
		MountOutput: be.Ledger.GetString(ro.optionMountOutput),
		MountVar:    be.Ledger.GetString(ro.optionMountVar),
		Prefix:      be.Ledger.GetString(be.Cli.optionDockerTag),
	}
}

func needsRebuildingImage(cmd *cobra.Command, option *config.StringOption) bool {
	var imageFlag = cmd.Flags().Lookup(option.Param)

	return imageFlag.Value.String() == ""
}

func newRunOptions(cli *Cli, cmd string) *RunOptions {
	var defaultOption = newOptionMountOutput(cmd, true)

	return &RunOptions{
		optionMountInput:  newOptionMountInput(cmd, true),
		optionMountLog:    newOptionMountLog(cmd, defaultOption),
		optionMountOutput: defaultOption,
		optionMountVar:    newOptionMountVar(cmd, defaultOption),
		optionRunImage:    newOptionCmdImage(cmd),
		optionRunPrefix:   newOptionContainerPrefix(cmd, cli),
	}
}

func newOptionCmdImage(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Image",
			Param: "image",
			Usage: "reference to an image with or without the version",
			DefaultGetter: func(ledger *config.Ledger) any {
				if cmd != "Run" {
					return ledger.GetValue("SF.Run.Image")
				}

				return ""
			},
		},
	}
}

func newOptionMountInput(cmd string, validated bool) *config.StringOption {
	var validator func(any) bool

	if validated {
		validator = config.Optionally(config.Validation.DirExists)
	}

	return &config.StringOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".Input",
			Param:     "in",
			Usage:     "path of the input files folder, read-only, if any",
			Validator: validator,
		},
	}
}

func newOptionMountLog(cmd string, defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Log",
			Param: "log",
			Usage: "path of the log files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.Get(&defaultOption.Option)
			},
			Validator: config.Optionally(config.Validation.DirCreated),
		},
	}
}

func newOptionMountOutput(cmd string, validated bool) *config.StringOption {
	var validator func(any) bool

	if validated {
		validator = config.Optionally(config.Validation.DirCreated)
	}

	return &config.StringOption{
		Option: config.Option{
			Key:       "SF." + cmd + ".Output",
			Param:     "out",
			Usage:     "path of the output files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			Validator: validator,
		},
	}
}

func newOptionMountVar(cmd string, defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF." + cmd + ".Var",
			Param: "var",
			Usage: "path of a solution state files folder, if any. " + cmd + " will attempt creating the path if does not exist",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.Get(&defaultOption.Option)
			},
			Validator: config.Optionally(config.Validation.DirCreated),
		},
	}
}

func pretendRunParamsTask(be BaseExecutor, ro *RunOptions, params *docker.ContainerParams,
	buildTaskFactory BuildTaskFactory, runTaskFactory RunTaskFactory) {
	var plan = task.NewPlan("Run", be.Ledger.Logger)
	var workers []task.Worker

	if ro.rebuildImage {
		workers = append(workers, task.NewPretender(makeBuildParams(&be), buildTaskFactory()))
	}

	workers = append(workers, task.NewPretender(params, runTaskFactory()))
	plan.Sequence(workers...)
}
