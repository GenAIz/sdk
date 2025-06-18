package wiremock

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/cmd/wiremock/function"
	"genaiz.com/genaiz-it/cmd/wiremock/mapping"
)

func NewWiremock() *cobra.Command {
	var wiremock = &cobra.Command{
		Use:   "wiremock",
		Short: "Invokes wiremock tooling",
		Long:  "Invokes wiremock tooling and helpers to setup mocked tests",
	}

	wiremock.AddCommand(function.NewFunction(), mapping.NewMapping())
	return wiremock
}
