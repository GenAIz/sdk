package jwks

import "github.com/spf13/cobra"

const (
	jwksOption = "jwks"
)

var (
	defaults = jwksDefaults{
		fileJwks: "keys.jwks",
	}
)

type jwksDefaults struct {
	fileJwks string
}

type jwksOptions struct {
	fileJwks string
}

func NewJwks() *cobra.Command {
	var jwks = &cobra.Command{
		Use:   "jwks",
		Short: "Manages operations on JWKS files",
		Long:  "Manages operations on JWKS files, creating, appending and deleting keys used to sign tokens",
	}

	jwks.AddCommand(
		NewAppend(),
		NewCreate(),
		NewDelete())
	return jwks
}
