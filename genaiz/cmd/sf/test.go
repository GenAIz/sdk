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

	te.Repo.DisplayChangeDir()
	docker.NewTestTask().Pretend(params, te.Repo.Logger)
}

func (te *TestExecutor) Proceed() {
	var testParams = te.makeTestParams()

	execRunParamsTask(te.BaseExecutor, te.RunOptions, testParams, docker.NewTestTask())
}

func (te *TestExecutor) makeTestParams() *docker.ContainerParams {
	return makeRunParams(te.BaseExecutor, te.RunOptions)
}

func NewTest(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewTestOptions(cli)
	var test = &cobra.Command{
		Use:     "test",
		Short:   "Runs the Smart Function attached for testing",
		Long:    "Runs the Smart Function, building it first if necessary",
		Example: "genaiz sf test --tag mytag/myfunction --version latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			repo.FromWorkDir(options.optionMountInput, cmd.Flags())
			repo.FromWorkDir(options.optionMountLog, cmd.Flags())
			repo.FromWorkDir(options.optionMountOutput, cmd.Flags())
			repo.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.optionRunImage)
			cli.Exec(repo, NewTestExecutor(cmd.Context(), repo, cli, options))
		},
	}

	repo.Register(test, options.allDefiners()...)
	return test
}

func NewTestExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *RunOptions) *TestExecutor {
	return &TestExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Repo:    repo,
		},
		RunOptions: options,
	}
}

func NewTestOptions(cli *Cli) *RunOptions {
	var optionOutput = newOptionMountOutput("Test")

	return &RunOptions{
		optionMountInput:  newOptionMountInput("Test"),
		optionMountLog:    newOptionMountLog("Test", optionOutput),
		optionMountOutput: optionOutput,
		optionMountVar:    newOptionMountVar("Test", optionOutput),
		optionRunImage:    newOptionCmdImage("Test"),
		optionRunPrefix:   NewOptionContainerPrefix("Test", cli),
	}
}
