package jwks

import (
	"errors"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/jwt"
)

const (
	jwksAppendUsage = "path or name of the jwks file to append to"
)

type appendOptions struct {
	jwksOptions
}

func (ao *appendOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&ao.fileJwks, jwksOption, defaults.fileJwks, jwksAppendUsage)
}

func NewAppend() *cobra.Command {
	var options = &appendOptions{}
	var appendCmd = &cobra.Command{
		Use:   "append CERT_FILES...",
		Short: "Appends PEM-based public keys to an existing JWKS file",
		Long:  "Appends PEM-based public keys, certificate parts of key pairs, to a an existing JWKS file",
		Run: func(cmd *cobra.Command, args []string) {
			var err error

			if options.fileJwks == "" {
				err = errors.New("missing jwks file parameter")
			} else {
				var manager = jwt.NewKeyManager().
					WithSetFile(options.fileJwks).
					WithPemKeys(args...)

				err = manager.MergeKeySet()
			}

			cobra.CheckErr(err)
		},
	}

	options.init(appendCmd)
	return appendCmd
}
