package sf

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type BuildTaskFactory func() *task.Task[docker.BuildParams]

type BuildExecutor struct {
	BaseExecutor
	*BuildOptions

	buildTaskFactory BuildTaskFactory
}

func (be *BuildExecutor) Display() {
	be.Ledger.DisplayOptions(be.Cli.SfOptions()...)
}

func (be *BuildExecutor) Pretend() {
	var params = be.makeBuildParams()

	be.Ledger.DisplayChangeDir()
	be.buildTaskFactory().Pretend(params, be.Ledger.Logger)
}

func (be *BuildExecutor) Proceed() {
	var params = be.makeBuildParams()
	var plan = task.NewPlan("Build", be.Ledger.Logger)

	plan.PrintReportsOnly = true
	task.Single(plan, params, be.buildTaskFactory())
}

// makeBuildParams creates a docker.BuildParams from resolving parameters, configuration files and environment variables
func (be *BuildExecutor) makeBuildParams() *docker.BuildParams {
	var result = makeBuildParams(&be.BaseExecutor)

	result.Label = be.Ledger.GetBool(be.optionLabelling)
	result.Prune = be.Ledger.GetBool(be.optionPruning)
	return result
}

type BuildOptions struct {
	optionLabelling *config.BoolOption
	optionLegacy    *config.BoolOption
	optionNoCache   *config.BoolOption
	optionPruning   *config.BoolOption
}

func (bo BuildOptions) allDefiners() []config.Definer {
	return []config.Definer{
		bo.optionLabelling,
		bo.optionLegacy,
		bo.optionNoCache,
		bo.optionPruning,
	}
}

func NewBuild(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var options = NewBuildOptions()
	var build = &cobra.Command{
		Use:     "build",
		Short:   "Builds a Smart Function",
		Long:    "Builds a Smart Function image tagging it with tag and version values",
		Example: "genaiz sf build --file=Dockerfile2 --context=../smartfunc --tag=genaiz.com/sf/smartfunc --version=v1.0",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewBuildExecutor(cmd.Context(), ledger, cli, options))
		},
	}

	ledger.Register(build, options.allDefiners()...)
	return build
}

func NewBuildExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *BuildOptions) *BuildExecutor {
	return &BuildExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Cli:     cli,
			Ledger:  ledger,
		},
		BuildOptions: options,
		buildTaskFactory: func() *task.Task[docker.BuildParams] {
			var useLegacy = ledger.GetBool(options.optionLegacy)

			if useLegacy {
				return docker.NewBuildLegacyTask()
			} else {
				return docker.NewBuildTask()
			}
		},
	}
}

func NewBuildOptions() *BuildOptions {
	return &BuildOptions{
		optionLabelling: cli.Options.Docker.Label().BuildBoolOption(),
		optionLegacy:    cli.Options.Docker.Legacy().BuildBoolOption(),
		optionNoCache:   cli.Options.Docker.NoCache().BuildBoolOption(),
		optionPruning:   cli.Options.Docker.Prune().BuildBoolOption(),
	}
}

func makeBuildParams(base *BaseExecutor) *docker.BuildParams {
	var cwd, _ = os.Getwd()
	var dockerContext = base.Ledger.GetString(base.Cli.optionDockerContext)
	var dockerFile = base.Ledger.GetString(base.Cli.optionDockerFile)
	var dockerTag = base.Ledger.GetString(base.Cli.optionDockerTag)
	var dockerVersion = base.Ledger.GetString(base.Cli.optionDockerVersion)

	if cwd == dockerContext {
		dockerContext = "."
	}

	if dir := filepath.Dir(dockerFile); dir != cwd && strings.Contains(dir, cwd) {
		dockerFile = filepath.Join(dir[len(cwd)+1:], filepath.Base(dockerFile))
	} else if dir == cwd {
		var baseFile = filepath.Base(dockerFile)

		if baseFile == "Dockerfile" {
			dockerFile = ""
		} else {
			dockerFile = filepath.Base(dockerFile)
		}
	}

	return &docker.BuildParams{
		Env:           task.Env{Context: base.Context},
		Dockerfile:    dockerFile,
		DockerContext: dockerContext,
		DockerTag:     strings.ToLower(dockerTag),
		DockerVersion: strings.ToLower(dockerVersion),
	}
}
