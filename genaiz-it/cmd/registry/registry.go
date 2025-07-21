package registry

import "github.com/spf13/cobra"

func NewRegistry() *cobra.Command {
	var cert = &cobra.Command{
		Use:   "registry",
		Short: "Manages the CNCF Distribution registry",
		Long:  "Manages the CNCF Distribution registry for genaiz-it tests",
	}

	cert.AddCommand(
		NewStart(),
		NewStop())
	return cert
}
