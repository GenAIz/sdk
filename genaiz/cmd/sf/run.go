package sf

import (
	"context"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/logz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

const (
	keyRunImage        = "SmartFunction.Run.Image"      // Configuration key used in config files for specifying the Docker image to run
	keyRunInputPath    = "SmartFunction.Run.InputPath"  // Configuration key used in config files for specifying the input mount path
	keyRunOutputPath   = "SmartFunction.Run.OutputPath" // Configuration key used in config files for specifying the output mount path
	paramRunImage      = "image"                        // Parameter label used from the shell for specifying the Docker image to run
	paramRunInputPath  = "in"                           // Parameter label used from the shell for specifying the input mount path
	paramRunOutputPath = "out"                          // Parameter label used from the shell for specifying the output mount path
)

var (
	dockerImage string // full name of the Docker image to run

	// SmartFunction run command for running a Docker image in detached mode, the command will build and/or use tag and version params to figure out what to run
	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Runs a Smart Function in detach mode",
		Long:    "Runs a Smart Function image detached, building it first if necessary. If the image is not specified build parameters are used",
		Example: "genaiz sf run --image genaiz.com/sf/smartfunc:latest",
		Run: func(cmd *cobra.Command, args []string) {
			var displayRun = func() {
				config.Display(sfMappings, runMappings)
			}

			if !DryExecute(cmd, displayRun) {
				ConfirmExecute(cmd,
					func() { execRun(cmd.Context()) },
					displayRun)
			}
		},
	}

	// Default values for the runCmd parameters
	runDefaults = map[string]func() string{
		keyRunImage: func() string {
			var tag = viper.GetString(keyDockerTag)
			var version = viper.GetString(keyDockerVersion)

			if version == "" || strings.Contains(tag, ":") {
				return tag
			}

			return tag + ":" + version
		},
	}

	// runCmd mappings between config files and shell
	runMappings = map[string]string{
		keyRunImage:      paramRunImage,
		keyRunInputPath:  paramRunInputPath,
		keyRunOutputPath: paramRunOutputPath,
	}
)

func init() {
	initRun(runCmd)
}

// bindRun binds runMappings between shell and config files to runCmd with runDefaults
func bindRun() {
	config.BindCmd(runCmd, runMappings)
	config.BindDefaults(runDefaults)
}

// execRun executes the runCmd after ensuring the image specified is available, otherwise it will attempt building an image and then run it.
//
//   - If the --image flag has a value we assume it means ignore the context, and we'll ignore --tag and --version.
//   - If --tag and/or --version are present without --image, we'll attempt finding or building a local image before running.
func execRun(ctx context.Context) {
	var runParams = makeRunParams()
	var plan = &task.Plan[docker.RunParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Could not run image %s, error: %s", dockerImage, err)
		},
		OnSuccess: func(out string) {
			var image = viper.GetString(keyRunImage)

			logz.InfoOutput(config.Logger, out)
			config.Logger.Printf("Running %s in detached mode", image)
		},
	}

	if dockerImage == "" {
		plan.Sequence(
			task.Execution(makeBuildParams(ctx), docker.BuildTask()),
			task.Execution(runParams, docker.RunTask()),
		)
	} else {
		plan.Single(runParams, docker.RunTask())
	}
}

// initRun initializes a cobra.Command with the parameter flags to satisfy a build and a run call
func initRun(cmd *cobra.Command) {
	initBuild(cmd)
	cmd.PersistentFlags().StringVar(&dockerImage, "image", "", "full name of an image with the version, defaults to build params if not specified")
	cmd.PersistentFlags().String("in", "", "full path of the input files folder. No input will be configured if not specified")
	cmd.PersistentFlags().String("out", "", "full path of the output files folder. No output will be configured if not specified")
}

// makeRunParams creates a docker.RunParams from resolving parameters, configuration files and environment variables
func makeRunParams() *docker.RunParams {
	return &docker.RunParams{
		DockerImage: viper.GetString(keyRunImage),
		MountInput:  viper.GetString(keyRunInputPath),
		MountOutput: viper.GetString(keyRunOutputPath),
	}
}
