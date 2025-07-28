package cmd

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/cmd/feature"
	"genaiz.com/genaiz-it/cmd/registry"
	"genaiz.com/genaiz-it/cmd/wiremock"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "genaiz-it",
		Short:   "GenAIz Integration Test Toolkit",
		Version: "0.0.2",
	}

	cmd.AddCommand(
		feature.NewFeature(),
		registry.NewRegistry(),
		wiremock.NewWiremock())
	return cmd
}
