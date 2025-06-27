package function

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

type BaseOptions struct {
	optionRegistry *config.StringOption
	optionWiremock *config.StringOption
}

type BaseCli struct {
	options *BaseOptions
	ledger  *config.Ledger
}

func (bc BaseCli) GetRegistryUrl() string {
	return bc.ledger.GetString(bc.options.optionRegistry)
}

func (bc BaseCli) GetWiremockUrl() string {
	return bc.ledger.GetString(bc.options.optionWiremock)
}

func NewFunction() *cobra.Command {
	var function = &cobra.Command{
		Use:   "function",
		Short: "Invokes wiremock function tooling",
		Long:  "Invokes wiremock function tooling and helpers to setup mocked-broker/v1/sf/... responses",
	}
	var cli = NewFunctionCli(function)

	function.AddCommand(
		NewProvision(cli),
	)
	return function
}

func NewFunctionCli(cmd *cobra.Command) *BaseCli {
	var ledger = config.NewLedger()
	var options = &BaseOptions{
		optionRegistry: newOptionRegistry(),
		optionWiremock: newOptionWiremock(),
	}

	ledger.Register(cmd, options.optionRegistry, options.optionWiremock)
	ledger.InitDefaults()
	return &BaseCli{
		options: options,
		ledger:  ledger,
	}
}

func newOptionRegistry() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Param:        "registry",
			DefaultValue: "localhost:5000",
			Usage:        "the docker registry to login",
		},
	}
}

func newOptionWiremock() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Param:        "wiremock",
			DefaultValue: "localhost:8080",
			Usage:        "the wiremock host to manipulate",
		},
	}
}
