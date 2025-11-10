package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
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
		&se.optionEnvFile.Option,
		&se.optionEnvVars.Option,
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
	var envMap map[string]string
	var err error

	if envMap, err = se.makeEnvMap(se.Ledger); err == nil {
		var replace = se.Ledger.GetBool(se.optionContainerReplace)
		var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
		var buildParams = makeBuildParams(&se.BaseExecutor)
		var params = se.makeStartParams(replace, envMap)
		var plan = task.NewPlan("Start", se.Ledger.Logger)
		var workers []task.Worker

		if se.rebuildImage {
			workers = append(workers, task.NewPretender(buildParams, se.buildTaskFactory()))
		}

		if replace {
			workers = append(workers, task.NewPretender(params, se.disposeTaskFactory()))
		}

		workers = append(workers,
			task.NewPretender(params, se.containerTaskFactory()),
			task.NewPretender(params, se.startTaskFactory()))

		if !preserve {
			var disposeParams = se.makeStartParams(false, envMap)

			workers = append(workers, task.NewPretender(disposeParams, se.stopTaskFactory()))
			workers = append(workers, task.NewPretender(disposeParams, se.disposeTaskFactory()))
		}

		plan.Sequence(workers...)
		return
	}

	lang.HandleExit(err)
}

func (se *StartExecutor) Proceed() {
	var envMap map[string]string
	var err error

	if envMap, err = se.makeEnvMap(se.Ledger); err == nil {
		var replace = se.Ledger.GetBool(se.optionContainerReplace)
		var params = se.makeStartParams(replace, envMap)
		var preserve = se.Ledger.GetBool(se.optionContainerPreserve)
		var buildParams = makeBuildParams(&se.BaseExecutor)
		var plan = task.NewPlan("Start", se.Ledger.Logger)
		var workers []task.Worker

		if se.rebuildImage {
			workers = append(workers, task.NewWorker(buildParams, se.buildTaskFactory()))
		}

		if replace {
			workers = append(workers, task.NewWorker(params, se.disposeTaskFactory()))
		}

		workers = append(workers,
			task.NewWorker(params, se.containerTaskFactory()),
			task.NewWorker(params, se.startTaskFactory()))

		if !preserve {
			var disposeParams = se.makeStartParams(false, envMap)

			workers = append(workers, task.NewWorker(disposeParams, se.stopTaskFactory()))
			workers = append(workers, task.NewWorker(disposeParams, se.disposeTaskFactory()))
		}

		plan.PrintReportsOnly = true
		plan.Sequence(workers...)
		return
	}

	lang.HandleExit(err)
}

func (se *StartExecutor) makeStartParams(force bool, envMap map[string]string) *docker.ContainerParams {
	var result = makeContainerParams(se.BaseExecutor, se.StartOptions.StopOptions, se.StartOptions.RunOptions)

	result.EnvVars = envMap
	result.MountInput = se.Ledger.GetPath(se.optionMountInput)
	result.MountOutput = se.Ledger.GetPath(se.optionMountOutput)
	result.MountLog = se.Ledger.GetPath(se.optionMountLog)
	result.MountVar = se.Ledger.GetPath(se.optionMountVar)
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
		so.optionEnvFile,
		so.optionEnvVars,
	}
}

func NewStart(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewStartOptions(cli)
	var start = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Smart Function, creating a container if necessary",
		Long:    "Starts the Smart Function, building it first if necessary, and creating a container matching the name and version of its context if it doesn't exist",
		Example: "genaiz sf start --image=my-project/my-func:latest --name=my-container-my-func --replace",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionEnvFile, cmd.Flags())
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			var imageFlag = cmd.Flags().Lookup(options.optionRunImage.Param)

			options.rebuildImage = imageFlag.Value.String() == ""
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

func NewStartOptions(sfCli *Cli) *StartOptions {
	var outputMountOption = cli.Options.Functions.MountOutput().
		WithKeys(&schema.Genaiz.Function.Start.MountOutput).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return ledger.GetString(cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Run.MountOutput).
				BuildStringOption())
		}).
		Optional(false).
		BuildStringOption()

	return &StartOptions{
		RunOptions: &RunOptions{
			EnvOptions: EnvOptions{
				optionEnvFile: cli.Options.Docker.EnvFile().
					WithKeys(&schema.Genaiz.Function.Start.EnvFile).
					BuildStringOption(),
				optionEnvVars: cli.Options.Docker.EnvVar().
					WithKeys(&schema.Genaiz.Function.Start.EnvVars).
					BuildListOption(),
			},
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Start.MountInput).
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return ledger.GetString(cli.Options.Functions.MountInput().
						WithKeys(&schema.Genaiz.Function.Run.MountInput).
						BuildStringOption())
				}).
				Optional(false).
				BuildStringOption(),
			optionMountLog: cli.Options.Functions.MountLog().
				WithKeys(&schema.Genaiz.Function.Start.MountLog).
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return ledger.GetString(outputMountOption)
				}).
				BuildStringOption(),
			optionMountOutput: outputMountOption,
			optionMountVar: cli.Options.Functions.MountVar().
				WithKeys(&schema.Genaiz.Function.Start.MountVar).
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return ledger.GetString(outputMountOption)
				}).
				BuildStringOption(),
			optionRunImage: cli.Options.Docker.Image().
				WithKeys(&schema.Genaiz.Function.Start.Image).
				WithDefaultGetter(sfCli.DefaultRunImage).
				BuildStringOption(),
		},
		StopOptions: &StopOptions{
			optionContainerName: cli.Options.Docker.ContainerName().
				WithKeys(&schema.Genaiz.Function.Start.Name).
				BuildStringOption(),
			optionContainerPrefix: cli.Options.Docker.ContainerPrefix().
				WithKeys(&schema.Genaiz.Function.Start.Prefix).
				WithDefaultGetter(sfCli.ContainerPrefix).
				BuildStringOption(),
			optionContainerPreserve: cli.Options.Docker.Preserve().
				WithKeys(&schema.Genaiz.Function.Start.Preserve).
				BuildBoolOption(),
		},

		optionContainerReplace: cli.Options.Docker.Replace().
			BuildBoolOption(),
	}
}
