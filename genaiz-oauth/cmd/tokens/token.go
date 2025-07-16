package tokens

import (
	"os"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/lang"
)

const (
	outputOption = "out"
	outputUsage  = "path of a file to write the token to, if empty decode will output to STDOUT"

	signingCertOption = "signCert"
	signingKeyOption  = "signKey"
)

var (
	defaults = tokenDefaults{
		fileSigningCert: "server.cert",
		fileSigningKey:  "server.key",
	}
)

type tokenDefaults struct {
	fileSigningCert string
	fileSigningKey  string
}

type tokenOptions struct {
	fileSigningCert string
	fileSigningKey  string
	output          string
}

func (to *tokenOptions) getOutput() (*os.File, error) {
	if to.output == "" {
		return os.Stdout, nil
	} else {
		return os.OpenFile(lang.LocalDir(to.output), os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
	}
}

func NewToken() *cobra.Command {
	var token = &cobra.Command{
		Use:   "tokens",
		Short: "Reads and writes JWT encoded tokens",
		Long:  "Reads and writes JWT encoded tokens to use for oauth authentication",
	}

	token.AddCommand(
		NewGenerate(),
		NewDecode())
	return token
}
