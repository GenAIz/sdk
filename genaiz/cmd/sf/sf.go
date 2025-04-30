package sf

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/config"
)

var (
	keyDockerFile      = "SmartFunction.Dockerfile"
	keyDockerContext   = "SmartFunction.Context"
	paramConfirm       = "confirm"
	paramDry           = "dry"
	paramDockerFile    = "file"
	paramDockerContext = "context"

	sfBindings = map[string]string{
		keyDockerFile:    paramDockerFile,
		keyDockerContext: paramDockerContext,
	}

	sfCmd = &cobra.Command{
		Use:   "sf",
		Short: "Genaiz Smart Function Utilities",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			viper.SetDefault(keyDockerContext, resolveDefaultContext())
			viper.SetDefault(keyDockerFile, resolveDefaultFile())
			changeContext()
		},
	}
)

func Cmd() *cobra.Command {
	return sfCmd
}

func Confirm(cmd *cobra.Command, display func()) bool {
	confirm, err := cmd.Flags().GetBool(paramConfirm)

	if err == nil && confirm {
		var s string

		display()
		r := bufio.NewReader(os.Stdin)

		for {
			_, err := fmt.Fprintf(os.Stdout, "%s (%s) ", "Confirm all options?", "[y/n]")

			if err != nil {
				return false
			}

			s, _ = r.ReadString('\n')
			s = strings.TrimSpace(s)
			s = strings.ToLower(s)

			if s != "" {
				if s == "y" || s == "yes" {
					return true
				}

				if s == "n" || s == "no" {
					return false
				}
			}
		}
	}

	return true
}

func ConfirmExecute(cmd *cobra.Command, display func(), exec func()) {
	if Confirm(cmd, display) {
		exec()
	} else {
		fmt.Println("Cancelled, exiting")
		os.Exit(0)
	}
}

func DryExecute(cmd *cobra.Command, display func()) bool {
	var dry, err = cmd.Flags().GetBool(paramDry)

	cobra.CheckErr(err)

	if dry {
		display()
		return true
	}

	return false
}

func init() {
	cobra.OnInitialize(bindSf, bindBuild, bindRun, bindDebug, bindTest, bindStart, bindStop)

	sfCmd.PersistentFlags().StringP(paramDockerFile, "f", "", "path of the Dockerfile, [context]/Dockerfile by default")
	sfCmd.PersistentFlags().StringP(paramDockerContext, "c", "", "path of Docker context, PWD by default")
	sfCmd.PersistentFlags().Bool(paramConfirm, false, "confirm parameters before executing")
	sfCmd.PersistentFlags().Bool(paramDry, false, "dry-run only displays parameter resolution then exist")
	sfCmd.AddCommand(buildCmd)
	sfCmd.AddCommand(debugCmd)
	sfCmd.AddCommand(runCmd)
	sfCmd.AddCommand(startCmd)
	sfCmd.AddCommand(stopCmd)
	sfCmd.AddCommand(testCmd)
}

func bindSf() {
	config.BindCmd(sfCmd, sfBindings)
}

func changeContext() {
	var context = viper.GetString(keyDockerContext)
	var wd, err = os.Getwd()

	cobra.CheckErr(err)

	if !strings.EqualFold(context, wd) {
		cobra.CheckErr(os.Chdir(context))
	}
}

func resolveDefaultFile() string {
	var context = viper.GetString(keyDockerContext)

	return context + "/Dockerfile"
}

func resolveDefaultContext() string {
	var wd, err = os.Getwd()

	cobra.CheckErr(err)
	return wd
}
