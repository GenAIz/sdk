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
	// SmartFunction test command for running a Docker image in attached mode, the command will build and/or use tag and version params to figure out what to test
	testCmd = &cobra.Command{
		Use:     "test",
		Short:   "Runs the Smart Function attached for testing",
		Long:    "Runs the Smart Function, building it first if necessary, in a disposable attached container for testing",
		Example: "genaiz sf test --tag mytag/myfunction --version latest",
		Run: func(cmd *cobra.Command, args []string) {
			var displayTest = func() {
				config.Display(sfMappings, runMappings)
			}

			if !DryExecute(cmd, displayTest) {
				ConfirmExecute(cmd,
					func() { execTest(cmd.Context()) },
					displayTest)
			}
		},
	}
)

func init() {
	initRun(testCmd)
}

// bindTest binds runMappings between shell and config files to testCmd with runDefaults
func bindTest() {
	config.BindCmd(testCmd, runMappings)
	config.BindDefaults(runDefaults)
}

// execTest executes the testCmd after ensuring the image specified is available, otherwise it will attempt building an image and then run it.
//
// Note: it does the same logic as execRun but with a docker.TestTask instead
func execTest(ctx context.Context) {
	var runParams = makeRunParams()
	var plan = &task.Plan[docker.RunParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Could not debug image %s, error: %s", dockerImage, err)
		},
		OnSuccess: func(out string) {
			var image = viper.GetString(keyRunImage)

			logz.InfoOutput(config.Logger, out)
			config.Logger.Printf("Testing %s in attached mode", image)
		},
	}

	if dockerImage == "" {
		plan.Sequence(
			task.Execution(makeBuildParams(ctx), docker.BuildTask()),
			task.Execution(runParams, docker.TestTask()),
		)
	} else {
		plan.Single(runParams, docker.TestTask())
	}
}
