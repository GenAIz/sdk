package sf

import (
	"context"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/logz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

const (
	keyDockerTag       = "SmartFunction.Build.Tag"     // Configuration key used in config files for specifying the Docker image tag in part or in full
	keyDockerVersion   = "SmartFunction.Build.Version" // Configuration key used in config files for specifying the Docker image version
	paramDockerTag     = "tag"                         // Parameter label used from the shell for specifying the Docker image tag in part or in full
	paramDockerVersion = "version"                     // Parameter label used from the shell for specifying the Docker image version
)

var (
	// SmartFunction build command for ensuring a Docker image is built under the specified context using an existing Dockerfile
	buildCmd = &cobra.Command{
		Use:     "build",
		Short:   "Builds a Smart Function image",
		Long:    "Builds a Smart Function docker image and tags it with tag and version values",
		Example: "genaiz sf build --file Dockerfile2 --context ../smartfunc --tag genaiz.com/sf/smartfunc --version v1.0",
		Run: func(cmd *cobra.Command, args []string) {
			var displayBuild = func() {
				config.Display(sfMappings, buildMappings)
			}

			if !PretendExecute(cmd, func() { pretendBuild(cmd.Context()) }) {
				if !DryExecute(cmd, displayBuild) {
					ConfirmExecute(cmd,
						func() { execBuild(cmd.Context()) },
						displayBuild)
				}
			}
		},
	}

	// Default values for the buildCmd parameters
	buildDefaults = map[string]func() string{
		keyDockerTag: func() string {
			var pwd, err = os.Getwd()

			cobra.CheckErr(err)
			return filepath.Base(pwd)
		},
	}

	// buildCmd mappings between config files and shell
	buildMappings = map[string]string{
		keyDockerTag:     paramDockerTag,
		keyDockerVersion: paramDockerVersion,
	}
)

func init() {
	initBuild(buildCmd)
}

// bindBuild binds buildMappings between shell and config files to buildCmd with buildDefaults
func bindBuild() {
	config.BindCmd(buildCmd, buildMappings)
	config.BindDefaults(buildDefaults)
}

// initBuild initializes a cobra.Command with the parameter flags to satisfy a build call
func initBuild(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(paramDockerTag, "t", "", "tag the smart function image. By default the tag is the name of the context folder")
	cmd.PersistentFlags().StringP(paramDockerVersion, "v", "", "version of the smart image to add to the tag. By default this will be \"latest\" if no \":string\" is present in the tag.")
}

// execBuild executes the buildCmd in isolation. If the build exists this call should return successfully with a no-op
//
//   - If --tag is present with no --version, we'll assume we need to tag latest, this may be refused on publish
//   - If --tag contains a revision, we'll ignore whatever is in --version
func execBuild(ctx context.Context) {
	var params = makeBuildParams(ctx)
	var plan = &task.Plan[docker.BuildParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Build incomplete with error: %s", err)
		},
		OnSuccess: func(out string) {
			config.Logger.Println("Build completed")
			logz.InfoOutput(config.Logger, out)
		},
	}

	plan.Single(params, docker.BuildTask())
}

// makeBuildParams creates a docker.BuildParams from resolving parameters, configuration files and environment variables
func makeBuildParams(ctx context.Context) *docker.BuildParams {
	var cwd, _ = os.Getwd()
	var dockerContext = viper.GetString(keyDockerContext)
	var dockerFile = viper.GetString(keyDockerFile)

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
		Env:           task.Env{Context: ctx},
		Dockerfile:    dockerFile,
		DockerContext: dockerContext,
		DockerTag:     viper.GetString(keyDockerTag),
		DockerVersion: viper.GetString(keyDockerVersion),
		Prune:         true,
	}
}

func pretendBuild(ctx context.Context) {
	docker.BuildTask().Pretend(makeBuildParams(ctx), config.Logger)
}
