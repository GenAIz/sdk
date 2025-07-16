package certs

import (
	"crypto/x509"
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/cert"
	"genaiz.com/genaiz-oauth/lang"
)

const (
	parseCaOption     = "parseCA"
	parseCaUsage      = "parse a self-signed authority certificate"
	parseCertOption   = "parseCert"
	parseCertUsage    = "parse a server certificate signed by an authority"
	parseBundleOption = "parseBundle"
	parseBundleUsage  = "parse a root certificate bundle, the public keys of the authority and the server"
)

type parseOptions struct {
	certsOptions

	parseBundle bool
	parseCa     bool
	parseCert   bool
}

func (p *parseOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.BoolVar(&p.parseCa, parseCaOption, true, parseCaUsage)
	flags.BoolVar(&p.parseCert, parseCertOption, true, parseCertUsage)
	flags.BoolVar(&p.parseBundle, parseBundleOption, false, parseBundleUsage)
	flags.StringVar(&p.fileBundle, bundleOption, defaults.fileBundle, bundleUsage)
	flags.StringVar(&p.fileCaCert, caCertOption, defaults.fileCaCert, caCertUsage)
	flags.StringVar(&p.fileCaKey, caKeyOption, defaults.fileCaKey, caKeyUsage)
	flags.StringVar(&p.fileServerCert, svCertOption, defaults.fileServerCert, svCertUsage)
	flags.StringVar(&p.fileServerKey, svKeyOption, defaults.fileServerKey, svKeyUsage)
}

func newParse(handler Handler) *cobra.Command {
	var options = &parseOptions{}
	var parse = &cobra.Command{
		Use:   "parse PEM_FILE",
		Short: "Parses a PEM file",
		Long:  "Parses a PEM file according to the provided types and algorithms",
		Run: func(cmd *cobra.Command, args []string) {
			var crt *x509.Certificate
			var reset func()
			var err error

			if reset, err = lang.Chdir(args...); err == nil {
				defer reset()
				var arbiter = cert.NewArbiter().
					WithAuthority(options.fileCaCert, options.fileCaKey).
					WithBundle(options.fileBundle).
					WithServer(options.fileServerCert, options.fileServerKey)

				if options.parseCa {
					if crt, err = arbiter.ParseAuthority(); err == nil {
						handler(crt)
					}
				}

				if err == nil && options.parseCert {
					if crt, err = arbiter.ParseCert(); err == nil {
						handler(crt)
					}
				}

				if err == nil && options.parseBundle {
					var certs []*x509.Certificate

					if certs, err = arbiter.ParseBundle(); err == nil {
						for _, crt = range certs {
							handler(crt)
						}
					}
				}
			}

			cobra.CheckErr(err)
		},
	}

	options.init(parse)
	return parse
}

func NewParse() *cobra.Command {
	return newParse(func(crt *x509.Certificate) {
		fmt.Println("-----")
		defaultPrintHandler(crt)
		fmt.Println("-----")
	})
}
