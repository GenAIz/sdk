package sf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

type BuildExecutor struct {
	BaseExecutor
}

func (be *BuildExecutor) Display() {
	be.Repo.DisplayOptions(be.Cli.SfOptions()...)
}

func (be *BuildExecutor) Pretend() {
	var params = be.makeBuildParams()

	be.Repo.DisplayChangeDir()
	docker.NewBuildTask().Pretend(params, be.Repo.Logger)
}

func (be *BuildExecutor) Proceed() {
	var params = be.makeBuildParams()
	var plan = &task.Plan[docker.BuildParams]{
		Logger: be.Repo.Logger,
		OnError: func(err error) {
			be.Repo.Logger.Errorf("Build incomplete with error: %s", err)
		},
		OnSuccess: func(out string) {
			if out != "" {
				fmt.Printf("%s\n", out)
			}
		},
	}

	plan.Single(params, docker.NewBuildTask())
}

// makeBuildParams creates a docker.BuildParams from resolving parameters, configuration files and environment variables
func (be *BuildExecutor) makeBuildParams() *docker.BuildParams {
	return makeBuildParams(be.BaseExecutor)
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
		BaseExecutor{
			Context: ctx,
			Cli:     cli,
			Repo:    repo,
		},
	}
}

func makeBuildParams(base BaseExecutor) *docker.BuildParams {
	var cwd, _ = os.Getwd()
	var dockerContext = base.Repo.GetString(base.Cli.optionDockerContext)
	var dockerFile = base.Repo.GetString(base.Cli.optionDockerFile)

	if cwd == dockerContext {
		dockerContext = "."
	}

	if dir := filepath.Dir(dockerFile); dir != cwd && strings.Contains(dir, cwd) {
		dockerFile = filepath.Join(dir[len(cwd)+1:], filepath.Base(dockerFile))
	} else if dir == cwd {
		var base = filepath.Base(dockerFile)

		if base != "Dockerfile" {
			dockerFile = filepath.Base(dockerFile)
		} else {
			dockerFile = ""
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
