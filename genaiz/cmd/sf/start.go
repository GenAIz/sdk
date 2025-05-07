package sf

import (
	"os"
	"path/filepath"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/logz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/docker"
)

const (
	keyDockerContainer          = "SmartFunction.Start.Container" // Configuration key used in config files for specifying the Docker container to start/stop/dispose
	keyDockerContainerReplace   = "SmartFunction.Start.Replace"   // Configuration key used in config files for replacing any containers on creation
	paramDockerContainer        = "name"                          // Parameter label used from the shell for specifying  the Docker container to start/stop/dispose
	paramDockerContainerReplace = "replace"                       // Parameter label used from the shell for replacing any containers on creation
)

var (
	// SmartFunction start command for starting a Docker container by name, potentially creating it and building an image beforehand
	startCmd = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Smart Function, creating a container if necessary",
		Long:    "Starts the Smart Function, building it first if necessary, and creating a container matching the name and version of its context if it doesn't exist",
		Example: "genaiz sf start --image myproject/myfunc:latest --name mycontainer-myfunc --replace",
		Run: func(cmd *cobra.Command, args []string) {
			var displayStart = func() {
				config.Display(sfMappings, runMappings, startMappings)
			}

			if !DryExecute(cmd, displayStart) {
				ConfirmExecute(cmd, execStart, displayStart)
			}
		},
	}

	// Default values for the startCmd parameters
	startDefaults = map[string]func() string{
		keyDockerContainer: func() string {
			var wd, err = os.Getwd()

			cobra.CheckErr(err)
			return filepath.Base(filepath.Dir(wd)) + "-" + filepath.Base(wd)
		},
	}

	// startCmd mappings between config files and shell
	startMappings = map[string]string{
		keyDockerContainer:         paramDockerContainer,
		keyDockerContainerPreserve: paramDockerContainerPreserve,
		keyDockerContainerReplace:  paramDockerContainerReplace,
	}
)

func init() {
	initRun(startCmd)
	initStop(startCmd)
	initStart(startCmd)
	startCmd.PersistentFlags().BoolP(paramDockerContainerReplace, "r", false, "removes any previous containers before creating a new one. By default, the command will use an incremental naming scheme of [name]-i")
}

// bindStart binds startMappings between shell and config files to startCmd with startDefaults
func bindStart() {
	config.BindCmd(startCmd, startMappings)
	config.BindDefaults(startDefaults)
}

// execStart executes the startCmd after ensuring the container specified exists, potentially replacing existing one.
//
//   - If the --replace flag is specified, the current container will be forced stopped and disposed of before creating a new one
//   - If the --preserve flag is not specified the container will be disposed of after stoppage.
//   - With both flags set to false, the command will attempt creating a newly named container with an increment and start it
func execStart() {
	var replace = viper.GetBool(keyDockerContainerReplace)
	var preserve = viper.GetBool(keyDockerContainerPreserve)
	var params = makeStartParams(false)
	var plan = task.Plan[docker.ContainerParams]{
		Logger: config.Logger,
		OnError: func(err error) {
			config.Logger.Errorf("Could not start container %s, error: %s", params.Name, err)
		},
		OnSuccess: func(out string) {
			logz.InfoOutput(config.Logger, out)
			config.Logger.Printf("Started container %s", params.Name)
		},
	}

	if replace && preserve {
		plan.Sequence(
			task.Execution(params, docker.DisposeTask()),
			task.Execution(params, docker.CreateTask()),
			task.Execution(params, docker.StartTask()))
	} else if replace {
		plan.Sequence(
			task.Execution(makeStartParams(true), docker.DisposeTask()),
			task.Execution(params, docker.CreateTask()),
			task.Execution(params, docker.StartTask()),
			task.Execution(params, docker.DisposeTask()))
	} else if preserve {
		plan.Sequence(
			task.Execution(params, docker.CreateTask()),
			task.Execution(params, docker.StartTask()))
	} else {
		plan.Sequence(
			task.Execution(params, docker.CreateTask()),
			task.Execution(params, docker.StartTask()),
			task.Execution(params, docker.DisposeTask()))
	}
}

func initStart(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(paramDockerContainer, "n", "", "name of the container to start/stop, defaults to the [parentDir-currentDir] if none specified")
}

// makeStartParams creates a docker.ContainerParams from resolving parameters, configuration files and environment variables
func makeStartParams(force bool) *docker.ContainerParams {
	// It's the same as stop for now, but creating containers can get very hairy, very quickly, so leave it be.
	return &docker.ContainerParams{
		RunParams: *makeRunParams(),
		Name:      viper.GetString(keyDockerContainer),
		Force:     force,
	}
}
