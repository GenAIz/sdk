package sf

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

var (
	buildBindings = map[string]string{
		keyDockerTag:     paramDockerTag,
		keyDockerVersion: paramDockerVersion,
	}

	buildCmd = &cobra.Command{
		Use:     "build",
		Short:   "Builds a Smart Function image",
		Long:    "Builds a Smart Function docker image and tags it with tag and version values",
		Example: "genaiz sf build --file Dockerfile2 --context ../smartfunc --tag genaiz.com/sf/smartfunc --version v1.0",
		Run: func(cmd *cobra.Command, args []string) {
			if !DryExecute(cmd, dryBuild) {
				ConfirmExecute(cmd, confirmBuild, execBuild)
			}
		},
	}

	buildDefaults = map[string]func() string{
		keyDockerTag: resolveDefaultTag,
	}

	keyDockerTag       = "SmartFunction.Build.Tag"
	keyDockerVersion   = "SmartFunction.Build.Version"
	paramDockerTag     = "tag"
	paramDockerVersion = "version"
)

func init() {
	initBuild(buildCmd)
}

func bindBuild() {
	config.BindCmd(buildCmd, buildBindings)
	config.BindDefaults(buildDefaults)
}

func confirmBuild() {
	fmt.Println("Confirming genaiz sf build on:")
	displayBuild()
}

func displayBuild() {
	config.Display(sfBindings, buildBindings)
}

func dryBuild() {
	fmt.Println("Dry-run for genaiz sf build:")
	displayBuild()
}

func initBuild(cmd *cobra.Command) {
	cmd.PersistentFlags().StringP(paramDockerTag, "t", "", "tag the smart function image, by default the tag is the name of the folder")
	cmd.PersistentFlags().StringP(paramDockerVersion, "v", "", "version of the smart image as part of the tag")
}

func execBuild() {
	// TODO
	fmt.Println("Executing genaiz sf build... TODO")
}

func resolveDefaultTag() string {
	var pwd, err = os.Getwd()

	cobra.CheckErr(err)
	return filepath.Base(pwd)
}
