// Package sf provides commands for managing Genaiz Smart Functions.
// Smart Functions commands include build, create, debug, init, run, start, stop and test.
//
// See: genaiz sf --help
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

const (
	keyDockerFile      = "SmartFunction.Dockerfile" // Configuration key used in config files for specifying the Dockerfile path
	keyDockerContext   = "SmartFunction.Context"    // Configuration key used in config files for specifying the Docker build context
	paramConfirm       = "confirm"                  // Parameter label used from the shell for specifying the execution confirmation flag
	paramDry           = "dry"                      // Parameter label used from the shell for specifying a dry execution
	paramDockerFile    = "file"                     // Parameter label used from the shell for specifying the Dockerfile path
	paramDockerContext = "context"                  // Parameter label used from the shell for specifying the Docker build context
)

var (
	// SmartFunction command for initiating sub-commands, resolving the default context the default Dockerfile and the current work dir
	sfCmd = &cobra.Command{
		Use:   "sf",
		Short: "Genaiz Smart Function Utilities",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var context, err = os.Getwd()

			cobra.CheckErr(err)
			viper.SetDefault(keyDockerContext, context)
			viper.SetDefault(keyDockerFile, context+"/Dockerfile")

			if viperCtx := viper.GetString(keyDockerContext); !strings.EqualFold(context, viperCtx) {
				cobra.CheckErr(os.Chdir(context))
			}
		},
	}

	// sfCmd mappings between config files and shell
	sfMappings = map[string]string{
		keyDockerFile:    paramDockerFile,
		keyDockerContext: paramDockerContext,
	}
)

// Cmd returns a fully initialized sf command with all sub-commands initialized with shell, config and default configuration bindings
func Cmd() *cobra.Command {
	return sfCmd
}

// Confirm returns true if the user acknowledges the display func, false otherwise
//
// Note: if display is not used, the method will still ask to proceed or not
func Confirm(cmd *cobra.Command, display ...func()) bool {
	if confirm, err := cmd.Flags().GetBool(paramConfirm); err == nil && confirm {
		var r = bufio.NewReader(os.Stdin)
		var message = "Proceed?"

		if len(display) > 0 {
			for _, d := range display {
				d()
			}

			message = "Confirm all options?"
		}

		for {
			if _, err := fmt.Fprintf(os.Stdout, "%s (%s) ", message, "[y/n]"); err == nil {
				var s, _ = r.ReadString('\n')

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
			} else {
				return false
			}
		}
	}

	return true
}

// ConfirmExecute will call exec if the cmd requires confirmation through the display function
func ConfirmExecute(cmd *cobra.Command, exec func(), display ...func()) {
	if Confirm(cmd, display...) {
		exec()
	} else {
		fmt.Println("Cancelled, exiting")
		os.Exit(0)
	}
}

// DryExecute will invoke display if the cmd is set in dry-run mode
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
	sfCmd.PersistentFlags().StringP(paramDockerFile, "f", "", "path of the Dockerfile, [context]/Dockerfile by default")
	sfCmd.PersistentFlags().StringP(paramDockerContext, "c", "", "path of Docker context, PWD by default")
	sfCmd.PersistentFlags().Bool(paramConfirm, false, "confirm parameters before executing")
	sfCmd.PersistentFlags().Bool(paramDry, false, "dry-run only displays parameter resolution then exits")
	sfCmd.AddCommand(buildCmd, debugCmd, runCmd, startCmd, stopCmd, testCmd)
	cobra.OnInitialize(bindSf, bindBuild, bindDebug, bindRun, bindStart, bindStop, bindTest)
}

// bindSf binds sfMappings between shell and config files to sfCmd
func bindSf() {
	config.BindCmd(sfCmd, sfMappings)
}
