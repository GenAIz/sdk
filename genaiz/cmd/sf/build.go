package sf

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type BuildTaskFactory func() *task.Task[docker.BuildParams]

type BuildExecutor struct {
	BaseExecutor

	buildTaskFactory BuildTaskFactory
}

func (be *BuildExecutor) Display() {
	be.Repo.DisplayOptions(be.Cli.SfOptions()...)
}

func (be *BuildExecutor) Pretend() {
	var params = be.makeBuildParams()

	be.Repo.DisplayChangeDir()
	be.buildTaskFactory().Pretend(params, be.Repo.Logger)
}

func (be *BuildExecutor) Proceed() {
	var params = be.makeBuildParams()
	var plan = task.NewPlan("Build", be.Repo.Logger)

	task.Single(plan, params, be.buildTaskFactory())
}

// makeBuildParams creates a docker.BuildParams from resolving parameters, configuration files and environment variables
func (be *BuildExecutor) makeBuildParams() *docker.BuildParams {
	return makeBuildParams(&be.BaseExecutor)
}

func NewBuild(repo *config.Repo, cli *Cli) *cobra.Command {
	var build = &cobra.Command{
		Use:     "build",
		Short:   "Builds a Smart Function",
		Long:    "Builds a Smart Function image tagging it with tag and version values",
		Example: "genaiz sf build --file Dockerfile2 --context ../smartfunc --tag genaiz.com/sf/smartfunc --version v1.0",
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(repo, NewBuildExecutor(cmd.Context(), repo, cli))
		},
	}

	return build
}

func NewBuildExecutor(ctx context.Context, repo *config.Repo, cli *Cli) *BuildExecutor {
	return &BuildExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Cli:     cli,
			Repo:    repo,
		},
		buildTaskFactory: docker.NewBuildTask,
	}
}

func makeBuildParams(base *BaseExecutor) *docker.BuildParams {
	var cwd, _ = os.Getwd()
	var dockerContext = base.Repo.GetString(base.Cli.optionDockerContext)
	var dockerFile = base.Repo.GetString(base.Cli.optionDockerFile)

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
		DockerTag:     base.Repo.GetString(base.Cli.optionDockerTag),
		DockerVersion: base.Repo.GetString(base.Cli.optionDockerVersion),
		Prune:         true,
	}
}
