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

type TestExecutor struct {
	BaseExecutor
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

	if testParams, err = makeRunParams(te.BaseExecutor, te.RunOptions); err == nil {
		var plan = task.NewPlan("Test", te.Ledger.Logger)
		var workers []task.Worker

		te.Ledger.DisplayChangeDir()

		if te.rebuildImage {
			workers = append(workers, task.NewPretender(makeBuildParams(&te.BaseExecutor), te.buildTaskFactory()))
		}

		workers = append(workers, task.NewPretender(testParams, te.testTaskFactory()))
		plan.Sequence(workers...)
		return
	}

	lang.HandleExit(err)
}

func (te *TestExecutor) Proceed() {
	var testParams *docker.ContainerParams
	var err error

	if testParams, err = makeRunParams(te.BaseExecutor, te.RunOptions); err == nil {
		var plan = task.NewPlan("Run", te.Ledger.Logger)
		var workers []task.Worker

		if te.RunOptions.rebuildImage {
			workers = append(workers, task.NewWorker(makeBuildParams(&te.BaseExecutor), te.buildTaskFactory()))
		}

		workers = append(workers, task.NewWorker(testParams, te.testTaskFactory()))
		plan.Sequence(workers...)
		return
	}

	lang.HandleExit(err)
}

func NewTest(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewTestOptions(cli)
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
			cli.Exec(ledger, NewTestExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(test, options.allDefiners()...)
	return test
}

func NewTestExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *RunOptions) *TestExecutor {
	return &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		RunOptions: options,

		buildTaskFactory: docker.NewBuildTask,
		testTaskFactory:  docker.NewTestTask,
	}
}

func NewTestOptions(sfCli *Cli) *RunOptions {
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
			BuildStringOption(),
		optionMountLog: cli.Options.Functions.MountLog().
			WithKeys(&schema.Genaiz.Function.Test.MountLog).
			BuildStringOption(),
		optionMountOutput: cli.Options.Functions.MountOutput().
			WithKeys(&schema.Genaiz.Function.Test.MountOutput).
			BuildStringOption(),
		optionMountVar: cli.Options.Functions.MountVar().
			WithKeys(&schema.Genaiz.Function.Test.MountVar).
			BuildStringOption(),
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
