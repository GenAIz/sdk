package cmd

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/cmd/registry"
	"genaiz.com/genaiz-it/cmd/wiremock"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "genaiz-it",
		Short:   "GenAIz Integration Test Toolkit",
		Version: "0.0.1",
	}

	cmd.AddCommand(
		registry.NewRegistry(),
		wiremock.NewWiremock())
	return cmd
}
