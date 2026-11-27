package sf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/docker/adapt"
)

type BuildExecutor struct {
	BaseExecutor
	*BuildOptions

	buildTaskFactory docker.BuildTaskFactory
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
	var result = makeBuildParams(&be.BaseExecutor, be.BuildOptions)

	result.Label = be.Ledger.GetBool(be.optionLabelling)
	result.Prune = be.Ledger.GetBool(be.optionPruning)
	return result
}

type BuildOptions struct {
	optionBuildArgs *config.ListOption
	optionLabelling *config.BoolOption
	optionLegacy    *config.BoolOption
	optionNoCache   *config.BoolOption
	optionPlatform  *config.StringOption
	optionPruning   *config.BoolOption
}

func (bo BuildOptions) allDefiners() []config.Definer {
	return []config.Definer{
		bo.optionBuildArgs,
		bo.optionLabelling,
		bo.optionLegacy,
		bo.optionNoCache,
		bo.optionPlatform,
		bo.optionPruning,
	}
}

func NewBuild(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var buildOptions = NewBuildOptions()
	var buildCmd = &cobra.Command{
		Use:     "build",
		Short:   "Builds a Smart Function",
		Long:    "Builds a Smart Function image tagging it with repository and version values",
		Example: "genaiz sf build --file=Dockerfile2 --context=../smartfunc --repository=genaiz.com/sf/smartfunc --version=v1.0",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewBuildExecutor(cmd.Context(), ledger, cli, buildOptions))
		},
	}

	ledger.Register(buildCmd, buildOptions.allDefiners()...)
	return buildCmd
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
				return docker.BuildTasks.LegacyTaskFactory()
			}

			return docker.BuildTasks.StandardTaskFactory()
		},
	}
}

func NewBuildOptions() *BuildOptions {
	return &BuildOptions{
		optionBuildArgs: cli.Options.Docker.BuildArg().BuildListOption(),
		optionLabelling: cli.Options.Docker.Label().BuildBoolOption(),
		optionLegacy:    cli.Options.Docker.Legacy().BuildBoolOption(),
		optionNoCache:   cli.Options.Docker.NoCache().BuildBoolOption(),
		optionPlatform:  cli.Options.Docker.Platform().BuildStringOption(),
		optionPruning:   cli.Options.Docker.Prune().BuildBoolOption(),
	}
}

func makeBuildParams(base *BaseExecutor, buildOptions *BuildOptions) *docker.BuildParams {
	var adaptedArgs = newBuildArguments(base)
	var buildArgs = base.Ledger.GetList(buildOptions.optionBuildArgs)
	var platform = base.Ledger.GetString(buildOptions.optionPlatform)

	if platform != "" {
		adaptedArgs.Platform = platform
	}

	for _, arg := range buildArgs {
		if strings.Contains(arg, "$") {
			var argTokens = strings.Split(arg, "=")

			if len(argTokens) == 2 {
				adaptedArgs.BuildArgs = append(adaptedArgs.BuildArgs, fmt.Sprintf("%s=%s", argTokens[0], os.ExpandEnv(argTokens[1])))
				continue
			}
		}

		adaptedArgs.BuildArgs = append(adaptedArgs.BuildArgs, arg)
	}

	return &docker.BuildParams{
		Env:            task.Env{Context: base.Context},
		BuildArguments: *adaptedArgs,
	}
}

func newBuildArguments(base *BaseExecutor) *adapt.BuildArguments {
	var cwd, _ = os.Getwd()
	var dockerContext = base.Ledger.GetString(base.Cli.optionDockerContext)
	var dockerFile = base.Ledger.GetString(base.Cli.optionDockerFile)
	var dockerRepo = base.Ledger.GetString(base.Cli.optionDockerRepo)
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

	return &adapt.BuildArguments{
		Dockerfile:       dockerFile,
		DockerContext:    dockerContext,
		DockerRepository: strings.ToLower(dockerRepo),
		DockerVersion:    strings.ToLower(dockerVersion),
	}
}
