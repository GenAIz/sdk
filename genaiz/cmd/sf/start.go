package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type ContainerTaskFactory func() *task.Task[docker.ContainerParams]

type StartTaskFactory func() *task.Task[docker.ContainerParams]

type StartExecutor struct {
	BaseExecutor
	*StartOptions

	buildTaskFactory     BuildTaskFactory
	containerTaskFactory ContainerTaskFactory
	disposeTaskFactory   DisposeTaskFactory
	startTaskFactory     StartTaskFactory
	stopTaskFactory      StopTaskFactory
}

func (se *StartExecutor) Display() {
	var options = []*config.Option{
		&se.optionRunImage.Option,
		&se.optionContainerName.Option,
		&se.optionContainerPrefix.Option,
		&se.optionContainerReplace.Option,
		&se.optionMountInput.Option,
		&se.optionMountOutput.Option,
		&se.optionMountVar.Option,
		&se.optionMountLog.Option,
	}

	options = append(options, se.Cli.SfOptions()...)
	se.Ledger.DisplayOptions(options...)
}

func (se *StartExecutor) Pretend() {
	var replace = se.Ledger.GetBool(se.optionContainerReplace)
	var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
	var buildParams = makeBuildParams(&se.BaseExecutor)
	var params = se.makeStartParams(replace)
	var plan = task.NewPlan("Start", se.Ledger.Logger)
	var workers = []task.Worker{
		task.NewPretender(buildParams, se.buildTaskFactory()),
	}

	if replace {
		workers = append(workers, task.NewPretender(params, se.disposeTaskFactory()))
	}

	workers = append(workers,
		task.NewPretender(params, se.containerTaskFactory()),
		task.NewPretender(params, se.startTaskFactory()))

	if !preserve {
		var disposeParams = se.makeStartParams(false)

		workers = append(workers, task.NewPretender(disposeParams, se.stopTaskFactory()))
		workers = append(workers, task.NewPretender(disposeParams, se.disposeTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (se *StartExecutor) Proceed() {
	var replace = se.Ledger.GetBool(se.optionContainerReplace)
	var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
	var buildParams = makeBuildParams(&se.BaseExecutor)
	var params = se.makeStartParams(replace)
	var plan = task.NewPlan("Start", se.Ledger.Logger)
	var workers = []task.Worker{
		task.NewWorker(buildParams, se.buildTaskFactory()),
	}

	if replace {
		workers = append(workers, task.NewWorker(params, se.disposeTaskFactory()))
	}

	workers = append(workers,
		task.NewWorker(params, se.containerTaskFactory()),
		task.NewWorker(params, se.startTaskFactory()))

	if !preserve {
		var disposeParams = se.makeStartParams(false)

		workers = append(workers, task.NewWorker(disposeParams, se.stopTaskFactory()))
		workers = append(workers, task.NewWorker(disposeParams, se.disposeTaskFactory()))
	}

	plan.Sequence(workers...)
}

func (se *StartExecutor) makeStartParams(force bool) *docker.ContainerParams {
	var result = makeContainerParams(se.BaseExecutor, se.StartOptions.StopOptions, se.StartOptions.RunOptions)

	result.MountOutput = se.Ledger.GetString(se.optionMountOutput)
	result.MountInput = se.Ledger.GetString(se.optionMountInput)
	result.Force = force
	return result
}

type StartOptions struct {
	*RunOptions
	*StopOptions

	optionContainerReplace *config.BoolOption
}

func (so *StartOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionContainerPrefix,
		so.optionContainerName,
		so.optionContainerReplace,
		so.optionContainerPreserve,
		so.optionMountInput,
		so.optionMountLog,
		so.optionMountOutput,
		so.optionMountVar,
		so.optionRunImage,
	}
}

func NewStart(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewStartOptions(cli)
	var start = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Smart Function, creating a container if necessary",
		Long:    "Starts the Smart Function, building it first if necessary, and creating a container matching the name and version of its context if it doesn't exist",
		Example: "genaiz sf start --image myproject/myfunc:latest --name mycontainer-myfunc --replace",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.RunOptions.optionRunImage)
			cli.Exec(ledger, NewStartExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(start, options.allDefiners()...)
	return start
}

func NewStartExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *StartOptions) *StartExecutor {
	return &StartExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Ledger:  ledger,
			Cli:     cli,
		},
		StartOptions: options,

		buildTaskFactory:     docker.NewBuildTask,
		containerTaskFactory: docker.NewCreateTask,
		disposeTaskFactory:   docker.NewDisposeTask,
		startTaskFactory:     docker.NewStartTask,
		stopTaskFactory:      docker.NewStopTask,
	}
}

func NewStartOptions(cli *Cli) *StartOptions {
	var startCmd = "Start"
	var runOptions = newRunOptions(cli, startCmd)
	var stopOptions = newStopOptions(cli, runOptions, startCmd, true)

	return &StartOptions{
		RunOptions:  runOptions,
		StopOptions: stopOptions,

		optionContainerReplace: newOptionContainerReplace(),
	}
}

func newOptionContainerReplace() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Key:          "SF.Start.Replace",
			Param:        "replace",
			Short:        "r",
			Usage:        "removes any previous containers before creating a new one",
			DefaultValue: "false",
		},
	}
}
