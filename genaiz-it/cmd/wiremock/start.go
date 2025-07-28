package wiremock

import (
	"errors"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
	"genaiz.com/genaiz-it/cucumber"
	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
)

const (
	featureOption  = "feature"
	featureUsage   = "The name of the feature to start, ex: \"account_login\""
	scenarioOption = "scenario"
	scenarioUsage  = "Scenarios within the feature to start, ex: \"session_create_ok\""
)

type startOptions struct {
	feature   string
	scenarios []string
}

func (so *startOptions) getFeature() cucumber.Feature {
	return *cucumber.NewFeature("wiremock", so.feature, so.scenarios)
}

func (so *startOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&so.feature, featureOption, "", featureUsage)
	flags.StringSliceVar(&so.scenarios, scenarioOption, []string{}, scenarioUsage)
}

func (so *startOptions) validate() error {
	if so.feature == "" {
		return errors.New("missing feature to start")
	}

	return nil
}

func NewStart() *cobra.Command {
	var options = &startOptions{}
	var startCmd = &cobra.Command{
		Use:   "start [WORKDIR]",
		Short: "Starts the Wiremock orchestration service",
		Long:  "Starts the Wiremock orchestration service for a single feature with a set of scenarios or all of them",
		Run: func(cmd *cobra.Command, args []string) {
			var services = compose.NewServices(cmd.Context())
			var err error

			if err = options.validate(); err == nil {
				var reset func()

				if reset, err = dirz.CreateWorkingDir(args...); err == nil {
					defer errorz.DeferOnExit(&err, reset)()
					var features = []cucumber.Feature{options.getFeature()}
					var service = services.Wiremock(features...)

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
