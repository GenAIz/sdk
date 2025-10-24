package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
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
	var runParams = makeRunParams(re.BaseExecutor, re.RunOptions)
	var plan = task.NewPlan("Run", re.Ledger.Logger)
	var workers []task.Worker

	re.Ledger.DisplayChangeDir()

	if re.rebuildImage {
		workers = append(workers, task.NewPretender(makeBuildParams(&re.BaseExecutor), re.buildTaskFactory()))
	}

	workers = append(workers, task.NewPretender(runParams, re.runTaskFactory()))
	plan.Sequence(workers...)
}

func (re *RunExecutor) Proceed() {
	var runParams = makeRunParams(re.BaseExecutor, re.RunOptions)
	var plan = task.NewPlan("Run", re.Ledger.Logger)
	var workers []task.Worker

	if re.rebuildImage {
		workers = append(workers, task.NewWorker(makeBuildParams(&re.BaseExecutor), re.buildTaskFactory()))
	}

	workers = append(workers, task.NewWorker(runParams, re.runTaskFactory()))
	plan.PrintReportsOnly = true
	plan.Sequence(workers...)
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
		Example: "genaiz sf run --image=genaiz.com/sf/smartfunc:latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			var imageFlag = cmd.Flags().Lookup(options.optionRunImage.Param)

			options.rebuildImage = imageFlag.Value.String() == ""
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

func NewRunOptions(sfCli *Cli) *RunOptions {
	return &RunOptions{
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Run.MountLog).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Run.MountOutput).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Run.MountVar).
			BuildStringOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Run.Image).
			WithDefaultGetter(sfCli.DefaultRunImage).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Run.Prefix).
			WithDefaultGetter(sfCli.ContainerPrefix).
			BuildStringOption(),
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
	be.Ledger.DisplayOptions(options...)
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
		Prefix:      be.Ledger.GetString(ro.optionRunPrefix),
	}
}
