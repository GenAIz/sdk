package sf

import (
	"context"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/logz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

var (
	// SmartFunction debug command for running a Docker image in interactive mode, the command will build and/or use tag and version params to figure out what to debug
	debugCmd = &cobra.Command{
		Use:     "debug",
		Short:   "Debugs a Smart Function image interactively",
		Long:    "Debugs a Smart Function image, building it first if necessary, opening a shell on a disposable container. If the image is not specified build parameters are used",
		Example: "genaiz sf debug --image genaiz.com/sf/smartfunc:latest",
		Run: func(cmd *cobra.Command, args []string) {
			var displayDebug = func() {
				config.Display(sfMappings, runMappings)
			}

			if !DryExecute(cmd, displayDebug) {
				ConfirmExecute(cmd,
					func() { execDebug(cmd.Context()) },
					displayDebug)
			}
		},
	}
)

func init() {
	initRun(debugCmd)
}

// bindDebug binds runMappings between shell and config files to debugCmd with runDefaults
func bindDebug() {
	config.BindCmd(debugCmd, runMappings)
	config.BindDefaults(runDefaults)
}

// execDebug executes the debugCmd after ensuring the image specified is available, otherwise it will attempt building an image and then run it.
//
// Note: it does the same logic as execRun but with a docker.DebugTask instead
func execDebug(ctx context.Context) {
	var runParams = makeRunParams()
	var plan = &task.Plan[docker.RunParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Could not debug image %s, error: %s", dockerImage, err)
		},
		OnSuccess: func(out string) {
			var image = viper.GetString(keyRunImage)

			logz.InfoOutput(config.Logger, out)
			config.Logger.Printf("Running %s in interactive mode", image)
		},
	}

	if dockerImage == "" {
		plan.Sequence(
			task.Execution(makeBuildParams(ctx), docker.BuildTask()),
			task.Execution(runParams, docker.DebugTask()),
		)
	} else {
		plan.Single(runParams, docker.DebugTask())
	}
}
