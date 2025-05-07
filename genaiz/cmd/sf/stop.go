package sf

import (
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/logz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

const (
	keyDockerContainerPreserve   = "SmartFunction.Stop.Preserve" // Configuration key used in config files for specifying whether to preserve the container or not
	paramDockerContainerPreserve = "preserve"                    // Parameter label used from the shell for specifying  whether to preserve the container or not
)

var (
	// SmartFunction stop command for stopping a Docker container by name
	stopCmd = &cobra.Command{
		Use:     "stop",
		Short:   "Stops a Smart Function, removing its container",
		Long:    "Stops a Smart Function, removing its container by default, if no container can be resolved this is a no-op",
		Example: "genaiz sf stop --name mycontainer-myfunc1 --preserve",
		Run: func(cmd *cobra.Command, args []string) {
			var displayStop = func() {
				config.Display(sfMappings, runMappings, stopMappings)
			}

			if !DryExecute(cmd, displayStop) {
				ConfirmExecute(cmd, execStop, displayStop)
			}
		},
	}

	// stopCmd mappings between config files and shell
	stopMappings = map[string]string{
		keyDockerContainer:         paramDockerContainer,
		keyDockerContainerPreserve: paramDockerContainerPreserve,
	}
)

func init() {
	initStart(stopCmd)
	initStop(stopCmd)
}

// bindStop binds stopMappings between shell and config files to stopCmd
func bindStop() {
	config.BindCmd(stopCmd, stopMappings)
}

// execStop executes the stopCmd after ensuring the container specified exists
//
//   - If the --preserve flag is not specified the container will be disposed of after stoppage.
func execStop() {
	var preserve = viper.GetBool(keyDockerContainerPreserve)
	var params = makeStopParams()
	var plan = &task.Plan[docker.ContainerParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Could not stop container %s, error: %s", params.Name, err)
		},
		OnSuccess: func(out string) {
			logz.InfoOutput(config.Logger, out)
			config.Logger.Printf("Stopped container %s", params.Name)
		},
	}

	if preserve {
		plan.Single(params, docker.StopTask())
	} else {
		plan.Sequence(
			task.Execution(params, docker.StopTask()),
			task.Execution(params, docker.DisposeTask()),
		)
	}
}

// initStop initializes a cobra.Command with the parameter flags to satisfy a stop call
func initStop(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(paramDockerContainerPreserve, "p", false, "preserves the container after it exits, defaults to false")
}

// makeStopParams creates a docker.ContainerParams from resolving parameters, configuration files and environment variables
func makeStopParams() *docker.ContainerParams {
	return &docker.ContainerParams{
		RunParams: *makeRunParams(),
		Name:      viper.GetString(keyDockerContainer),
	}
}
