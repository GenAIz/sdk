package feature

import (
	"errors"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
)

const (
	prefixOption    = "prefix"
	prefixUsage     = "use the specified prefix for test work directories"
	timestampOption = "timestamp"
	timestampUsage  = "prefix feature work directories with a timestamp to avoid collisions across multiple runs"
)

type startOptions struct {
	baseOptions

	prefix    string
	timestamp bool
}

func (to *startOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&to.orchestrator, orchestratorOption, "wiremock", orchestratorUsage)
	flags.StringVar(&to.prefix, prefixOption, "", prefixUsage)
	flags.StringSliceVar(&to.tests, testOption, []string{}, testUsage)
	flags.BoolVar(&to.timestamp, timestampOption, false, timestampUsage)
	flags.StringVar(&to.version, versionOption, "latest", versionUsage)
}

func (to *startOptions) validate() error {
	if len(to.tests) == 0 {
		return errors.New("missing tests to start")
	}

	return nil
}

func NewStart() *cobra.Command {
	var options = &startOptions{}
	var startCmd = &cobra.Command{
		Use:     "start [WORKDIR]",
		Short:   "Starts test(s) using the genaiz docker image",
		Long:    "Starts test(s) using the genaiz docker images and bundled features and scenarios",
		Example: "genaiz-it feature start myOutput --test=account_login:create_session_ok",
		Args:    cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var services = compose.NewServices(cmd.Context())
			var err error

			if err = options.validate(); err == nil {
				var reset func()

				if reset, err = dirz.CreateWorkingDir(args...); err == nil {
					defer errorz.DeferOnExit(&err, reset)()
					var features = options.getFeatures()
					var service = services.Genaiz(options.version, features...)

					if err = service.Init(); err == nil {
						err = service.Start()
					}
				}
			}
		},
	}

	options.init(startCmd)
	return startCmd
}
