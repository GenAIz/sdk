package sf

import (
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
)

var (
	keyDockerImage        = "SmartFunction.Run.Image"
	keyDockerInputPath    = "SmartFunction.Run.InputPath"
	keyDockerOutputPath   = "SmartFunction.Run.OutputPath"
	paramDockerImage      = "image"
	paramDockerInputPath  = "in"
	paramDockerOutputPath = "out"

	runBindings = map[string]string{
		keyDockerImage:      paramDockerImage,
		keyDockerInputPath:  paramDockerInputPath,
		keyDockerOutputPath: paramDockerOutputPath,
	}

	runCmd = &cobra.Command{
		Use:     "run",
		Short:   "Runs a Smart Function in detach mode",
		Long:    "Runs a Smart Function image detached, building it first if necessary. If the image is not specified build parameters are used",
		Example: "genaiz sf run --image genaiz.com/sf/smartfunc:latest",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryRun) {
				ConfirmExecute(cmd, confirmRun, execRun)
			}
		},
	}

	runDefaults = map[string]func() string{
		keyDockerImage: resolveDefaultImage,
	}
)

func init() {
	initRun(runCmd)
}

func bindRun() {
	config.BindCmd(runCmd, runBindings)
	config.BindDefaults(runDefaults)
}

func confirmRun() {
	fmt.Println("Confirming genaiz sf run on:")
	displayRun()
}

func displayRun() {
	config.Display(sfBindings, runBindings)
}

func dryRun() {
	fmt.Println("Dry-run for genaiz sf run:")
	displayRun()
}

func execRun() {
	// TODO
	fmt.Println("Executing genaiz sf run... TODO")
}

func initRun(cmd *cobra.Command) {
	initBuild(cmd)
	cmd.PersistentFlags().String("image", "", "full name of an image with the version, defaults to build params if not specified")
	cmd.PersistentFlags().String("in", "", "full path of the input files folder. No input will be configured if not specified")
	cmd.PersistentFlags().String("out", "", "full path of the output files folder. No output will be configured if not specified")
}

func resolveDefaultImage() string {
	var tag = viper.GetString(keyDockerTag)
	var version = viper.GetString(keyDockerVersion)

	if version == "" || strings.Contains(tag, ":") {
		return tag
	}

	return tag + ":" + version
}
