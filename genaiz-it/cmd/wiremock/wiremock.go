package wiremock

import (
	"github.com/spf13/cobra"
)

func NewWiremock() *cobra.Command {
	var wiremock = &cobra.Command{
		Use:   "wiremock",
		Short: "Manages a standalone Wiremock Broker",
		Long:  "Manages a standalone Wiremock Broker with the provided internal resources as mocks",
	}

	wiremock.AddCommand(
		NewStart(),
		NewStop())
	return wiremock
}
