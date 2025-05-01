package sf

import (
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

var (
	debugCmd = &cobra.Command{
		Use:     "debug",
		Short:   "Debugs a Smart Function image interactively",
		Long:    "Debugs a Smart Function image, building it first if necessary, opening a shell on a disposable container. If the image is not specified build parameters are used",
		Example: "genaiz sf debug --image genaiz.com/sf/smartfunc:latest",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryDebug) {
				ConfirmExecute(cmd, confirmDebug, execDebug)
			}
		},
	}
)

func init() {
	initRun(debugCmd)
}

func bindDebug() {
	config.BindCmd(debugCmd, runBindings)
}

func confirmDebug() {
	fmt.Println("Confirming genaiz sf debug on:")
	displayDebug()
}

func displayDebug() {
	config.Display(sfBindings, runBindings)
}

func dryDebug() {
	fmt.Println("Dry-run for genaiz sf debug:")
	displayDebug()
}

func execDebug() {
	// TODO
	fmt.Println("Executing genaiz sf debug... TODO")
}
