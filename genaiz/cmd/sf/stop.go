package sf

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type DisposeTaskFactory func() *task.Task[docker.ContainerParams]

type StopTaskFactory func() *task.Task[docker.ContainerParams]

type StopExecutor struct {
	BaseExecutor
	*StopOptions

	disposeTaskFactory DisposeTaskFactory
	stopTaskFactory    StopTaskFactory
}

func (se *StopExecutor) Display() {
	var options = []*config.Option{
		&se.optionRunImage.Option,
		&se.optionContainerName.Option,
		&se.optionContainerPrefix.Option,
		&se.optionContainerPreserve.Option,
	}

	options = append(options, se.Cli.SfOptions()...)
	se.Ledger.DisplayOptions(options...)
}

func (se *StopExecutor) Pretend() {
	var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
	var params = se.makeContainerParams()

	params.Force = !preserve

	if preserve {
		se.stopTaskFactory().Pretend(params, se.Ledger.Logger)
	} else {
		se.disposeTaskFactory().Pretend(params, se.Ledger.Logger)
	}
}

func (se *StopExecutor) Proceed() {
	var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
	var params = se.makeContainerParams()
	var plan = task.NewPlan("Stop", se.Ledger.Logger)

	params.Force = !preserve
	task.Conditional(plan, preserve, params, se.stopTaskFactory, se.disposeTaskFactory)
}

func (se *StopExecutor) makeContainerParams() *docker.ContainerParams {
	return makeContainerParams(se.BaseExecutor, se.StopOptions, se.RunOptions)
}

type StopOptions struct {
	*RunOptions
	optionContainerName     *config.StringOption
	optionContainerPrefix   *config.StringOption
	optionContainerPreserve *config.BoolOption
}

func (so *StopOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.RunOptions.optionRunImage,
		so.optionContainerName,
		so.optionContainerPrefix,
		so.optionContainerPreserve,
	}
}

func NewStop(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewStopOptions(cli)
	var stop = &cobra.Command{
		Use:     "stop",
		Short:   "Stops the container of a Smart Function",
		Long:    "Stops a Smart Function, removing its container by default",
		Example: "genaiz sf stop --name mycontainer-myfunc1 --preserve",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewStopExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(stop, options.allDefiners()...)
	return stop
}

func NewStopExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *StopOptions) *StopExecutor {
	return &StopExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		StopOptions: options,

		disposeTaskFactory: docker.NewDisposeTask,
		stopTaskFactory:    docker.NewStopTask,
	}
}

func NewStopOptions(cli *Cli) *StopOptions {
	var stopCmd = "Stop"
	var runOptions = &RunOptions{
		optionRunImage: newOptionCmdImage(stopCmd),
	}

	return newStopOptions(cli, runOptions, stopCmd, false)
}

func makeContainerParams(be BaseExecutor, so *StopOptions, ro *RunOptions) *docker.ContainerParams {
	return &docker.ContainerParams{
		RunParams: docker.RunParams{
			Env: task.Env{
				Context: be.Context,
			},
		},
		DockerImage: be.Ledger.GetString(ro.optionRunImage),
		Name:        be.Ledger.GetString(so.optionContainerName),
		Prefix:      be.Ledger.GetString(so.optionContainerPrefix),
	}
}

func newStopOptions(cli *Cli, runOptions *RunOptions, cmd string, preserve bool) *StopOptions {
	return &StopOptions{
		RunOptions: runOptions,

		optionContainerName:     newOptionContainerName(cmd),
		optionContainerPrefix:   newOptionContainerPrefix(cmd, cli),
		optionContainerPreserve: newOptionContainerPreserve(preserve),
	}
}

func newOptionContainerName(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Container." + cmd + "Name",
			Param: "name",
			Short: "n",
			Usage: "name of the container to start/stop",
		},
	}
}

func newOptionContainerPrefix(cmd string, cli *Cli) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Container." + cmd + "Prefix",
			Param: "prefix",
			Short: "p",
			Usage: "prefix to use for creating new containers",
			DefaultGetter: func(ledger *config.Ledger) any {
				var tag = strings.ReplaceAll(ledger.GetString(cli.optionDockerTag), "/", "-")
				var workspace = ledger.GetWorkspace()

				if workspace != "" {
					return workspace + "-" + tag
				}

				return tag
			},
		},
	}
}

func newOptionContainerPreserve(defaultValue bool) *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Key:          "SF.Container.Preserve",
			Param:        "preserve",
			Usage:        "preserves the container after it exits",
			DefaultValue: defaultValue,
		},
	}
}
