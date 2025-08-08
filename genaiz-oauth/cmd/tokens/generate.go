package tokens

import (
	"errors"
	"fmt"
	"io"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-oauth/jwt"
)

const (
	audienceOption   = "aud"
	audienceUsage    = "audience to match for token verification"
	expiryOption     = "exp"
	expiryUsage      = "expiry of the generated tokens in minutes"
	operationsOption = "op"
	operationsUsage  = "list of operations to support (pull, push)"
	repositoryOption = "repo"
	repositoryUsage  = "scope of the token, limited to a repository with or without tag"

	signingCertGenerateUsage = "path or name of the PEM-based cert for signing tokens"
	signingKeyGenerateUsage  = "path of the PEM-based private key file for signing tokens"
)

type generateOptions struct {
	tokenOptions

	audience   string
	expiry     int
	operations []string
	repository string
}

func (co *generateOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&co.audience, audienceOption, "dev.genaiz.com", audienceUsage)
	flags.IntVar(&co.expiry, expiryOption, 10, expiryUsage)
	flags.StringSliceVar(&co.operations, operationsOption, []string{}, operationsUsage)
	flags.StringVar(&co.output, outputOption, "", outputUsage)
	flags.StringVar(&co.repository, repositoryOption, "", repositoryUsage)
	flags.StringVar(&co.fileSigningCert, signingCertOption, defaults.fileSigningCert, signingCertGenerateUsage)
	flags.StringVar(&co.fileSigningKey, signingKeyOption, defaults.fileSigningKey, signingKeyGenerateUsage)
}

func (co *generateOptions) validate() error {
	var err error

	if co.repository != "" {
		if err = filez.IsReadable(co.fileSigningCert); err == nil {
			err = filez.IsReadable(co.fileSigningKey)
		}
	} else {
		err = errors.New("required repository is missing")
	}

	return nil
}

func newGenerate(handler func([]byte, io.Writer) error) *cobra.Command {
	var options = &generateOptions{}
	var generateCmd = &cobra.Command{
		Use:   "generate [WORKDIR]",
		Short: "Generates JWT encoded and signed tokens",
		Long:  "Generates JWT encoded and signed tokens using certs under WORKDIR",
		Run: func(cmd *cobra.Command, args []string) {
			var reset func()
			var err error

			if reset, err = dirz.CreateWorkingDir(args...); err == nil {
				defer errorz.DeferOnExit(&err, reset)()

				if err = options.validate(); err == nil {
					var tokenBytes []byte
					var builder = jwt.NewBuilder().
						WithAudience(options.audience).
						WithExpiry(options.expiry).
						WithAccess(options.repository, options.operations)

					options.anchorFilePaths()
					builder.WithSigner(options.fileSigningCert, options.fileSigningKey)

					if tokenBytes, err = builder.Build(); err == nil {
						var out io.Writer

						if out, err = options.getOutput(); err == nil {
							err = handler(tokenBytes, out)
						}
					}
				}
			}
		},
	}

	options.init(generateCmd)
	return generateCmd
}

func NewGenerate() *cobra.Command {
	return newGenerate(func(tokenBytes []byte, writer io.Writer) error {
		_, err := fmt.Fprintln(writer, string(tokenBytes))
		return err
	})
}
