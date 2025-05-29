package sf

import (
	"context"
	"fmt"

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
	se.Repo.DisplayOptions(options...)
}

func (se *StartExecutor) Pretend() {
	var force = se.Repo.GetBool(se.optionContainerReplace)
	var params = se.makeStartParams(force)

	docker.NewStartTask().Pretend(params, se.Repo.Logger)
}

func (se *StartExecutor) Proceed() {
	var replace = se.Repo.GetBool(se.optionContainerReplace)
	var preserve = se.Repo.GetBool(se.optionContainerPreserve)
	var buildParams = makeBuildParams(se.BaseExecutor)
	var params = se.makeStartParams(replace)
	var plan = task.Plan[docker.ContainerParams]{
		Logger: se.Repo.Logger,
		OnError: func(err error) {
			se.Repo.Logger.Errorf("Could not start container %s, error: %s", params.Name, err)
		},
		OnSuccess: func(out string) {
			if out != "" {
				se.Repo.Logger.Infof("Started container [%s]", out)
				fmt.Printf("%s\n", out)
			}
		},
	}
	var executions = []func(*task.State) error{
		task.Execution(buildParams, docker.NewBuildTask()),
	}

	if replace {
		executions = append(executions, task.Execution(params, docker.NewDisposeTask()))
	}

	executions = append(executions,
		task.Execution(params, docker.NewCreateTask()),
		task.Execution(params, docker.NewStartTask()))

	if !preserve {
		var disposeParams = se.makeStartParams(false)

		executions = append(executions, task.Execution(disposeParams, docker.NewDisposeTask()))
	}

	plan.Sequence(executions...)
}

func (se *StartExecutor) makeStartParams(force bool) *docker.ContainerParams {
	var result = makeContainerParams(se.BaseExecutor, se.StartOptions.StopOptions, se.StartOptions.RunOptions)

	result.MountOutput = se.Repo.GetString(se.optionMountOutput)
	result.MountInput = se.Repo.GetString(se.optionMountInput)
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

func NewStart(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewStartOptions(cli)
	var start = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Smart Function, creating a container if necessary",
		Long:    "Starts the Smart Function, building it first if necessary, and creating a container matching the name and version of its context if it doesn't exist",
		Example: "genaiz sf start --image myproject/myfunc:latest --name mycontainer-myfunc --replace",
		PreRun: func(cmd *cobra.Command, args []string) {
			repo.FromWorkDir(options.optionMountInput, cmd.Flags())
			repo.FromWorkDir(options.optionMountLog, cmd.Flags())
			repo.FromWorkDir(options.optionMountOutput, cmd.Flags())
			repo.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.RunOptions.optionRunImage)
			cli.Exec(repo, NewStartExecutor(cmd.Context(), repo, cli, options))
		},
	}

	repo.Register(start, options.allDefiners()...)
	return start
}

func NewStartExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *StartOptions) *StartExecutor {
	return &StartExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Repo:    repo,
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
