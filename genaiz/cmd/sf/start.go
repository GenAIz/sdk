package sf

import (
	"context"
	"fmt"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type StartExecutor struct {
	BaseExecutor
	*StartOptions
}

func (se *StartExecutor) Display() {
	var options = []*config.Option{
		&se.optionRunImage.Option,
		&se.optionContainerName.Option,
		&se.optionContainerPrefix.Option,
		&se.optionContainerReplace.Option,
		&se.optionMountInput.Option,
		&se.optionMountOutput.Option,
	}

	options = append(options, se.Cli.SfOptions()...)
	se.Ledger.DisplayOptions(options...)
}

func (se *StartExecutor) Pretend() {
	var force = se.Ledger.GetBool(se.optionContainerReplace)
	var params = se.makeStartParams(force)

	docker.NewStartTask().Pretend(params, se.Ledger.Logger)
}

func (se *StartExecutor) Proceed() {
	var replace = se.Ledger.GetBool(se.optionContainerReplace)
	var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
	var buildParams = makeBuildParams(&se.BaseExecutor)
	var params = se.makeStartParams(replace)
	var plan = task.Plan{
		Logger: se.Ledger.Logger,
		OnFailure: func(msg interface{}) {
			se.Ledger.Logger.Errorf("Could not start container %s, error: %s", params.Name, msg)
		},
		OnSuccess: func(msg interface{}) {
			var out = cast.ToString(msg)

			if out != "" {
				se.Ledger.Logger.Infof("Started container [%s]", out)
				fmt.Printf("%s\n", out)
			}
		},
	}
	var workers = []task.Worker{
		task.NewWorker(buildParams, docker.NewBuildTask()),
	}

	if replace {
		workers = append(workers, task.NewWorker(params, docker.NewDisposeTask()))
	}

	workers = append(workers,
		task.NewWorker(params, docker.NewCreateTask()),
		task.NewWorker(params, docker.NewStartTask()))

	if !preserve {
		var disposeParams = se.makeStartParams(false)

		workers = append(workers, task.NewWorker(disposeParams, docker.NewDisposeTask()))
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
	}
}

func NewStartOptions(cli *Cli) *StartOptions {
	var startCmd = "Start"
	var runOptions = newRunOptions(cli, startCmd)
	var stopOptions = newStopOptions(cli, runOptions, startCmd)

	return &StartOptions{
		RunOptions:             runOptions,
		StopOptions:            stopOptions,
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
