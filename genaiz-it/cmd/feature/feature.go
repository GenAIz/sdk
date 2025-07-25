package feature

import (
	"maps"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/cucumber"
)

const (
	orchestratorOption = "orchestrator"
	orchestratorUsage  = "the orchestrator to test, either an internal profile or a url starting with http(s)"
	testOption         = "test"
	testUsage          = "a feature with or without scenario to start"
	versionOption      = "version"
	versionUsage       = "Version of the genaiz image to test"
)

type baseOptions struct {
	orchestrator string
	tests        []string
	version      string
}

func (bo *baseOptions) getFeatures() []cucumber.Feature {
	var result = map[string]cucumber.Feature{}

	for _, test := range bo.tests {
		var parts = strings.Split(test, ":")
		var scenarios []string

		if len(parts) > 1 {
			scenarios = parts[1:]
		}

		if feat, ok := result[parts[0]]; ok {
			feat.Scenarios = append(feat.Scenarios, scenarios...)
		} else {
			result[parts[0]] = *cucumber.NewFeature(bo.orchestrator, parts[0], scenarios)
		}
	}

	return slices.Collect(maps.Values(result))
}

func NewFeature() *cobra.Command {
	var featureCmd = &cobra.Command{
		Use:   "feature",
		Short: "Manages bundled Gherkin features",
		Long:  "Manages bundled Gherkin features to execute genaiz-it tests",
	}

	featureCmd.AddCommand(
		NewReport(),
		NewStart(),
		NewStop())
	return featureCmd
}
