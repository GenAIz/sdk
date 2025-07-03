package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
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
	var testParams = te.makeTestParams()

	te.Ledger.DisplayChangeDir()
	pretendRunParamsTask(te.BaseExecutor, te.RunOptions, testParams, te.buildTaskFactory, te.testTaskFactory)
}

func (te *TestExecutor) Proceed() {
	var testParams = te.makeTestParams()

	execRunParamsTask(te.BaseExecutor, te.RunOptions, testParams, te.buildTaskFactory, te.testTaskFactory)
}

func (te *TestExecutor) makeTestParams() *docker.ContainerParams {
	return makeRunParams(te.BaseExecutor, te.RunOptions)
}

func NewTest(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewTestOptions(cli)
	var test = &cobra.Command{
		Use:     "test",
		Short:   "Runs the Smart Function attached for testing",
		Long:    "Runs the Smart Function, building it first if necessary",
		Example: "genaiz sf test --tag my-tag/my-function --version latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			ledger.FromWorkDir(options.optionMountInput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountLog, cmd.Flags())
			ledger.FromWorkDir(options.optionMountOutput, cmd.Flags())
			ledger.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.optionRunImage)
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

func NewTestOptions(cli *Cli) *RunOptions {
	return newRunOptions(cli, "Test")
}
