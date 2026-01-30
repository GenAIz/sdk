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
	"genaiz.com/genaiz/task/layout"
)

type TestExecutor struct {
	BaseExecutor
	SyncExecutor
	*RunOptions

	buildTaskFactory BuildTaskFactory
	testTaskFactory  RunTaskFactory
}

func (te *TestExecutor) Display() {
	displayRunOptions(te.BaseExecutor, te.RunOptions)
}

func (te *TestExecutor) Pretend() {
	var testParams *docker.ContainerParams
	var err error

	if testParams, err = newRunParams(te.BaseExecutor, te.RunOptions); err == nil {
		var plan = task.NewPlan("Test", te.Ledger.Logger)
		var datalinkWorkers []task.Worker
		var workers []task.Worker

		te.Ledger.DisplayChangeDir()

		if te.rebuildImage {
			workers = append(workers, task.NewPretender(makeBuildParams(&te.BaseExecutor), te.buildTaskFactory()))
		}

		if datalinkWorkers, err = makeSyncPretenders(te.Ledger, te.SyncExecutor, te.optionNoPropSync); err == nil {
			workers = append(workers, datalinkWorkers...)
			workers = append(workers, task.NewPretender(testParams, te.testTaskFactory()))
			plan.Sequence(workers...)
			return
		}
	}

	lang.HandleExit(err)
}

func (te *TestExecutor) Proceed() {
	var testParams *docker.ContainerParams
	var err error

	if testParams, err = newRunParams(te.BaseExecutor, te.RunOptions); err == nil {
		var plan = task.NewPlan("Run", te.Ledger.Logger)
		var datalinkWorkers []task.Worker
		var workers []task.Worker

		if te.RunOptions.rebuildImage {
			workers = append(workers, task.NewWorker(makeBuildParams(&te.BaseExecutor), te.buildTaskFactory()))
		}

		if datalinkWorkers, err = makeSyncWorkers(te.Ledger, te.SyncExecutor, te.optionNoPropSync); err == nil {
			workers = append(workers, datalinkWorkers...)
			workers = append(workers, task.NewWorker(testParams, te.testTaskFactory()))
			plan.Sequence(workers...)
			return
		}
	}

	lang.HandleExit(err)
}

func NewTest(ledger *config.Ledger, sfCLi *Cli) *cobra.Command {
	var options = NewTestOptions(sfCLi)
	var test = &cobra.Command{
		Use:     "test",
		Short:   "Runs a Smart Function attached to the current shell for testing",
		Long:    "Runs a Smart Function attached, building it first if necessary, assigning it a disposable container",
		Example: "genaiz sf test --tag my-tag/my-function --version latest",
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
			sfCLi.Exec(ledger, NewTestExecutor(cmd.Context(), ledger, sfCLi, options))
		},
	}

	ledger.Register(test, options.allDefiners()...)
	return test
}

func NewTestExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *RunOptions) *TestExecutor {
	return &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		SyncExecutor: makeSyncExecutor(),
		RunOptions:   options,

		buildTaskFactory: docker.NewBuildTask,
		testTaskFactory:  docker.NewTestTask,
	}
}

func NewTestOptions(sfCli *Cli) *RunOptions {
	var runLayout = layout.NewRunLayout()
	var outputMountOption = cli.Options.Functions.MountOutput().
		WithKeys(&schema.Genaiz.Function.Test.MountOutput).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return ledger.GetPath(cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Run.MountOutput).
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return runLayout.DirOutput
				}).
				BuildStringOption())
		}).
		Optional(false).
		BuildStringOption()

	return &RunOptions{
		EnvOptions: EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Genaiz.Function.Test.EnvFile).
				BuildStringOption(),
			optionEnvVars: cli.Options.Docker.EnvVar().
				WithKeys(&schema.Genaiz.Function.Test.EnvVars).
				BuildListOption(),
		},
		optionMountInput: cli.Options.Functions.MountInput().
			WithKeys(&schema.Genaiz.Function.Test.MountInput).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetPath(cli.Options.Functions.MountInput().
					WithKeys(&schema.Genaiz.Function.Run.MountInput).
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return runLayout.DirInput
					}).
					BuildStringOption())
			}).
			Optional(false).
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Test.MountLog).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetPath(cli.Options.Functions.MountLog().
					WithKeys(&schema.Genaiz.Function.Run.MountLog).
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return runLayout.DirLog
					}).
					BuildStringOption())
			}).
			BuildStringOption(),
		optionMountOutput: outputMountOption,
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Test.MountVar).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetPath(cli.Options.Functions.MountVar().
					WithKeys(&schema.Genaiz.Function.Run.MountVar).
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return runLayout.DirVar
					}).
					BuildStringOption())
			}).
			BuildStringOption(),
		optionNoPropSync: cli.Options.Functions.NoPropSync().
			WithKeys(&schema.Genaiz.Function.Run.NoPropSync).
			BuildBoolOption(),
		optionRunImage: cli.Options.Docker.Image().
			WithKeys(&schema.Genaiz.Function.Test.Image).
			WithDefaultGetter(sfCli.DefaultRunImage).
			BuildStringOption(),
		optionRunPrefix: cli.Options.Docker.ContainerPrefix().
			WithKeys(&schema.Genaiz.Function.Test.Prefix).
			WithDefaultGetter(sfCli.ContainerPrefix).
			BuildStringOption(),
	}
}
