package jwks

import (
	"errors"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/jwt"
)

const (
	jwksDeleteUsage = "path or name of the jwks file to delete from"
)

type deleteOptions struct {
	jwksOptions
}

func (do *deleteOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&do.fileJwks, jwksOption, defaults.fileJwks, jwksDeleteUsage)
}

func NewDelete() *cobra.Command {
	var options = &deleteOptions{}
	var deleteCmd = &cobra.Command{
		Use:   "delete CERT_FILES...",
		Short: "Deletes PEM-based public keys from an existing JWKS file",
		Long:  "Deletes PEM-based public keys, certificate parts of key pairs, from an existing JWKS file",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.fileJwks == "" {
				err = errors.New("missing jwks file parameter")
			} else {
				var manager = jwt.NewKeyManager().
					WithSetFile(options.fileJwks).
					WithPemKeys(args...)

				err = manager.RemoveKeySet()
			}

			cobra.CheckErr(err)
		},
	}

	options.init(deleteCmd)
	return deleteCmd
}
