package jwks

import (
	"errors"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/jwt"
)

const (
	jwksCreateUsage = "path or name of the jwks file to create"
)

type createOptions struct {
	jwksOptions
}

func (co *createOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&co.fileJwks, jwksOption, defaults.fileJwks, jwksCreateUsage)
}

func NewCreate() *cobra.Command {
	var options = &createOptions{}
	var createCmd = &cobra.Command{
		Use:   "create CERT_FILES...",
		Short: "Creates a JWKS file from a set of PEM-based public keys",
		Long:  "Creates a JWKS file from a set of PEM-based public keys, the certificate parts of key pairs",
		Args:  cobra.MinimumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.fileJwks == "" {
				err = errors.New("missing jwks file parameter")
			} else {
				var manager = jwt.NewKeyManager().
					WithSetFile(options.fileJwks).
					WithPemKeys(args...)

				err = manager.WriteKeySet()
			}

			cobra.CheckErr(err)
		},
	}

	options.init(createCmd)
	return createCmd
}
