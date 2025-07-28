package feature

import (
	"os/user"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
)

type stopOptions struct {
	baseOptions
}

func (so *stopOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&so.orchestrator, orchestratorOption, "wiremock", orchestratorUsage)
	flags.StringSliceVar(&so.tests, testOption, []string{}, testUsage)
	flags.StringVar(&so.version, versionOption, "latest", versionUsage)
}

func NewStop() *cobra.Command {
	var options = &stopOptions{}
	var stopCmd = &cobra.Command{
		Use:     "stop",
		Short:   "Stops test(s) and services used to for a list of features",
		Long:    "Stops test(s) and services containers used to execute a list of features",
		Example: "genaiz-it feature stop workDir --test=account_login:create_session_ok",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var currentUser, _ = user.Current()
			var features = options.getFeatures()
			var reset func()
			var err error

			if reset, err = dirz.ChangeWorkingDir(args...); err == nil {
				defer errorz.DeferOnExit(&err, reset)()
				var service = compose.NewGenaizService(cmd.Context(), currentUser, options.version, features...)

				err = service.Stop()
			}

			cobra.CheckErr(err)
		},
	}

	options.init(stopCmd)
	return stopCmd
}
