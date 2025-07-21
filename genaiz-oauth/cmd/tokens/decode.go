package tokens

import (
	"encoding/json"
	"io"
	"os"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-oauth/jwt"
)

const (
	inputOption            = "in"
	inputUsage             = "path of a file to read a token from, if both token and input are specified decode prioritizes token"
	signingCertDecodeUsage = "path or name of the PEM-based cert for signing tokens"
	signingKeyDecodeUsage  = "path of the PEM-based private key file for signing tokens"
	tokenOption            = "token"
	tokenUsage             = "base64 string value of the token to decode"
)

type decodeOptions struct {
	tokenOptions

	input string
	token string
}

func (do *decodeOptions) getInput() ([]byte, error) {
	if do.token == "" {
		return os.ReadFile(dirz.AnchorWorkingFile(do.input))
	} else {
		return []byte(do.token), nil
	}
}

func (do *decodeOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.StringVar(&do.fileSigningCert, signingCertOption, defaults.fileSigningCert, signingCertDecodeUsage)
	flags.StringVar(&do.fileSigningKey, signingKeyOption, defaults.fileSigningKey, signingKeyDecodeUsage)
	flags.StringVar(&do.input, inputOption, "", inputUsage)
	flags.StringVar(&do.output, outputOption, "", outputUsage)
	flags.StringVar(&do.token, tokenOption, "", tokenUsage)
}

func (do *decodeOptions) validate() error {
	return filez.IsReadable(do.fileSigningCert)
}

func newDecode(handler func(any, io.Writer) error) *cobra.Command {
	var options = &decodeOptions{}
	var decodeCmd = &cobra.Command{
		Use:   "decode [WORKDIR]",
		Short: "Decodes a JWT",
		Long:  "Decodes a JWT, potentially validating its signature",
		Run: func(cmd *cobra.Command, args []string) {
			var reset func()
			var err error

			if reset, err = dirz.CreateWorkingDir(args...); err == nil {
				defer errorz.DeferOnExit(&err, reset)
				options.anchorFilePaths()

				if err = options.validate(); err == nil {
					var tokenBytes []byte
					var builder = jwt.NewBuilder().
						WithSigner(options.fileSigningCert, options.fileSigningKey)

					if tokenBytes, err = options.getInput(); err == nil {
						var token any

						if token, err = builder.Decode(tokenBytes); err == nil {
							var out io.Writer

							if out, err = options.getOutput(); err == nil {
								err = handler(token, out)
							}
						}
					}
				}
			}
		},
	}

	options.init(decodeCmd)
	return decodeCmd
}

func NewDecode() *cobra.Command {
	return newDecode(func(token any, writer io.Writer) error {
		var encoder = json.NewEncoder(writer)

		encoder.SetIndent("", "  ")
		return encoder.Encode(token)
	})
}
