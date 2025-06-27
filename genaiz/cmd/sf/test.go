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
}

func (te *TestExecutor) Display() {
	displayRunOptions(te.BaseExecutor, te.RunOptions)
}

func (te *TestExecutor) Pretend() {
	var params = te.makeTestParams()

	te.Ledger.DisplayChangeDir()
	docker.NewTestTask().Pretend(params, te.Ledger.Logger)
}

func (te *TestExecutor) Proceed() {
	var testParams = te.makeTestParams()

	execRunParamsTask(te.BaseExecutor, te.RunOptions, testParams, docker.NewTestTask())
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
	}
}

func NewTestOptions(cli *Cli) *RunOptions {
	return newRunOptions(cli, "Test")
}
