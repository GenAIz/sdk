package sf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

var (
	keyDockerContainer          = "SmartFunction.Start.Container"
	keyDockerContainerReplace   = "SmartFunction.Start.Replace"
	paramDockerContainer        = "name"
	paramDockerContainerReplace = "replace"

	startBindings = map[string]string{
		keyDockerContainer:         paramDockerContainer,
		keyDockerContainerPreserve: paramDockerContainerPreserve,
		keyDockerContainerReplace:  paramDockerContainerReplace,
	}

	startCmd = &cobra.Command{
		Use:     "start",
		Short:   "Starts the Smart Function, creating a container if necessary",
		Long:    "Starts the Smart Function, building it first if necessary, and creating a container matching the name and version of its context if it doesn't exist",
		Example: "genaiz sf start --image myproject/myfunc:latest --name mycontainer-myfunc --replace",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryStart) {
				ConfirmExecute(cmd, confirmStart, execStart)
			}
		},
	}

	startDefaults = map[string]func() string{
		keyDockerContainer: resolveDefaultContainer,
	}
)

func init() {
	initRun(startCmd)
	initStop(startCmd)
	initStart(startCmd)
	startCmd.PersistentFlags().BoolP(paramDockerContainerReplace, "r", false, "removes any previous containers before creating a new one. By default, the command will use an incremental naming scheme of [name]-i")
}

func bindStart() {
	config.BindCmd(startCmd, startBindings)
	config.BindDefaults(startDefaults)
}

func confirmStart() {
	fmt.Println("Confirming genaiz sf start on:")
	displayStart()
}

func displayStart() {
	config.Display(sfBindings, runBindings, startBindings)
}

func dryStart() {
	fmt.Println("Dry-run for genaiz sf start:")
	displayStart()
}

func execStart() {
	// TODO
	fmt.Println("Executing genaiz sf start... TODO")
}

func initStart(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(paramDockerContainer, "n", "", "name of the container to start/stop, defaults to the [parentDir-currentDir] if none specified")
}

func resolveDefaultContainer() string {
	var wd, err = os.Getwd()

	cobra.CheckErr(err)
	return filepath.Base(filepath.Dir(wd)) + "-" + filepath.Base(wd)
}
