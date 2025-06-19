package cmd

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/cmd/wiremock"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "genaiz-it",
		Short:   "Genaiz Integration Test Toolkit",
		Version: "0.0.1",
	}

	cmd.AddCommand(wiremock.NewWiremock())
	return cmd
}
