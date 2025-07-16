package cmd

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/cmd/certs"
	"genaiz.com/genaiz-oauth/cmd/jwks"
	"genaiz.com/genaiz-oauth/cmd/tokens"
)

func New() *cobra.Command {
	var cmd = &cobra.Command{
		Use:     "genaiz-oauth",
		Short:   "GenAIz OAuth Utility Kit",
		Version: "0.0.1",
	}

	cmd.AddCommand(
		certs.NewCert(),
		jwks.NewJwks(),
		tokens.NewToken())
	return cmd
}
