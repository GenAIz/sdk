package tokens

import (
	"os"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
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

func (co *tokenOptions) anchorFilePaths() {
	co.fileSigningCert = dirz.AnchorWorkingFile(co.fileSigningCert)
	co.fileSigningKey = dirz.AnchorWorkingFile(co.fileSigningKey)
}

func (to *tokenOptions) getOutput() (*os.File, error) {
	if to.output == "" {
		return os.Stdout, nil
	} else {
		var path = dirz.AnchorWorkingFile(to.output)

		return os.OpenFile(path, os.O_CREATE|os.O_WRONLY|os.O_TRUNC, 0600)
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
