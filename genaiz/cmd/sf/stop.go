package sf

import (
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

var (
	keyDockerContainerPreserve   = "SmartFunction.Stop.Preserve"
	paramDockerContainerPreserve = "preserve"

	stopBindings = map[string]string{
		keyDockerContainer:         paramDockerContainer,
		keyDockerContainerPreserve: paramDockerContainerPreserve,
	}

	stopCmd = &cobra.Command{
		Use:     "stop",
		Short:   "Stops a Smart Function, removing its container",
		Long:    "Stops a Smart Function, removing its container by default, if no container can be resolved this is a no-op",
		Example: "genaiz sf stop --name mycontainer-myfunc1 --preserve",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryStop) {
				ConfirmExecute(cmd, confirmStop, execStop)
			}
		},
	}
)

func init() {
	initStart(stopCmd)
	initStop(stopCmd)
}

func bindStop() {
	config.BindCmd(stopCmd, stopBindings)
}

func confirmStop() {
	fmt.Println("Confirming genaiz sf stop on:")
	displayStop()
}

func displayStop() {
	config.Display(sfBindings, runBindings, stopBindings)
}

func dryStop() {
	fmt.Println("Dry-run for genaiz sf stop:")
	displayStop()
}

func execStop() {
	// TODO
	fmt.Println("Executing genaiz sf stop... TODO")
}

func initStop(cmd *cobra.Command) {
	cmd.PersistentFlags().BoolP(paramDockerContainerPreserve, "p", false, "preserves the container after it exits, defaults to false")
}
