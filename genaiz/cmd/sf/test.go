package sf

import (
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

var (
	testCmd = &cobra.Command{
		Use:     "test",
		Short:   "Runs the Smart Function attached for testing",
		Long:    "Runs the Smart Function, building it first if necessary, in a disposable attached container for testing",
		Example: "genaiz sf test --tag mytag/myfunction --version latest",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryTest) {
				ConfirmExecute(cmd, confirmTest, execTest)
			}
		},
	}
)

func init() {
	initRun(testCmd)
}

func bindTest() {
	config.BindCmd(debugCmd, runBindings)
}

func confirmTest() {
	fmt.Println("Confirming genaiz sf test on:")
	displayTest()
}

func displayTest() {
	config.Display(sfBindings, runBindings)
}

func dryTest() {
	fmt.Println("Dry-run for genaiz sf test:")
	displayTest()
}

func execTest() {
	// TODO
	fmt.Println("Executing genaiz sf test... TODO")
}
