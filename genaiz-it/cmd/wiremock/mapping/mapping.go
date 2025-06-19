package mapping

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/wiremock"
)

type MappingExecutor struct {
	Client *wiremock.Client
}

func (m *MappingExecutor) Print(name string) error {
	var mapping *wiremock.AdminMapping
	var err error

	if mapping, err = m.Client.GetStub(name); err == nil {
		var bytes []byte

		bytes, err = json.MarshalIndent(mapping, "", "  ")
		fmt.Printf("%s\n", string(bytes))
	}

	return err
}

func (m *MappingExecutor) Reset() error {
	return m.Client.Reset()
}

func NewMapping() *cobra.Command {
	var wiremockValue string
	var provision = &cobra.Command{
		Use:     "mapping MAPPING",
		Short:   "Invokes wiremock mapping querying",
		Long:    "Invokes wiremock mapping querying on wiremock/__admin/mappings",
		Example: "genaiz-it wiremock mapping function_mock_publish_simple_provision_ok",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewMappingExecutor(wiremockValue)

			cobra.CheckErr(exec.Print(args[0]))
		},
	}

	provision.AddCommand(NewReset(&wiremockValue))
	provision.PersistentFlags().StringVar(&wiremockValue, "wiremock", "localhost:8080", "the wiremock host to query")
	return provision
}

func NewMappingExecutor(wiremockUrl string) *MappingExecutor {
	return &MappingExecutor{
		Client: wiremock.NewWiremockClient(wiremockUrl),
	}
}
