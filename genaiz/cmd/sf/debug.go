package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/docker"
)

type DebugExecutor struct {
	BaseExecutor
	*RunOptions
}

func (de *DebugExecutor) Display() {
	displayRunOptions(de.BaseExecutor, de.RunOptions)
}

func (de *DebugExecutor) Pretend() {
	var params = de.makeDebugParams()

	de.Repo.DisplayChangeDir()
	docker.NewDebugTask().Pretend(params, de.Repo.Logger)
}

func (de *DebugExecutor) Proceed() {
	var debugParams = de.makeDebugParams()

	execRunParamsTask(de.BaseExecutor, de.RunOptions, debugParams, docker.NewDebugTask())
}

func (de *DebugExecutor) makeDebugParams() *docker.ContainerParams {
	return makeRunParams(de.BaseExecutor, de.RunOptions)
}

func NewDebug(repo *config.Repo, cli *Cli) *cobra.Command {
	var options = NewDebugOptions(cli)
	var debug = &cobra.Command{
		Use:     "debug",
		Short:   "Debugs a Smart Function interactively",
		Long:    "Debugs a Smart Function image, building it first if necessary",
		Example: "genaiz sf debug --image genaiz.com/sf/smartfunc:latest",
		PreRun: func(cmd *cobra.Command, args []string) {
			repo.FromWorkDir(options.optionMountInput, cmd.Flags())
			repo.FromWorkDir(options.optionMountLog, cmd.Flags())
			repo.FromWorkDir(options.optionMountOutput, cmd.Flags())
			repo.FromWorkDir(options.optionMountVar, cmd.Flags())
		},
		Run: func(cmd *cobra.Command, args []string) {
			options.rebuildImage = needsRebuildingImage(cmd, options.optionRunImage)
			cli.Exec(repo, NewDebugExecutor(cmd.Context(), repo, cli, options))
		},
	}

	repo.Register(debug, options.allDefiners()...)
	return debug
}

func NewDebugExecutor(ctx context.Context, repo *config.Repo, cli *Cli, options *RunOptions) *DebugExecutor {
	return &DebugExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Repo:    repo,
		},
		RunOptions: options,
	}
}

func NewDebugOptions(cli *Cli) *RunOptions {
	return newRunOptions(cli, "Debug")
}
