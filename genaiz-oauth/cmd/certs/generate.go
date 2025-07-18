package certs

import (
	"crypto/x509"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-oauth/cert"
)

const (
	bundleOption     = "bundleCrt"
	bundleUsage      = "path or name of the root bundle file to generate"
	caCertOption     = "caCert"
	caCertUsage      = "path or name of the ca cert file to read or generate"
	caCnOption       = "caCN"
	caCnUsage        = "common name of the ca cert"
	caKeyOption      = "caKey"
	caKeyUsage       = "path or name of the ca key file to read or generate"
	caLifetimeOption = "caLifetime"
	caLifetimeUsage  = "duration of the authority before expiry, in days"
	countryOption    = "country"
	countryUsage     = "country of the authority and issued cert"
	genCaOption      = "genCA"
	genCaUsage       = "generate a self-signed authority certificate"
	genCertOption    = "genCert"
	genCertUsage     = "generate a server certificate signed by an authority"
	genBundleOption  = "genBundle"
	genBundleUsage   = "generate a root certificate bundle, the public keys of the authority and the server"
	localityOption   = "locality"
	localityUsage    = "locality of the authority and issued cert"
	orgOption        = "organization"
	orgUsage         = "organization of the authority and issued cert"
	provinceOption   = "province"
	provinceUsage    = "province of the authority and issued cert"
	svCertOption     = "svCert"
	svCertUsage      = "path or name of the server cert file to generate"
	svCnOption       = "svCN"
	svCnUsage        = "common name of the server cert"
	svKeyOption      = "svKey"
	svKeyUsage       = "path or name of the server key file to generate"
	svLifetimeOption = "lifetime"
	svLifetimeUsage  = "duration of the server certificate before expiry, in hours"
)

type generateOptions struct {
	certsOptions

	genBundle bool
	genCA     bool
	genCert   bool

	caCommonName string
	caLifetime   int

	commonName   string
	organization string
	country      string
	province     string
	lifetime     int
	locality     string
}

func (g *generateOptions) init(cmd *cobra.Command) {
	var flags = cmd.PersistentFlags()

	flags.BoolVar(&g.genCA, genCaOption, true, genCaUsage)
	flags.BoolVar(&g.genCert, genCertOption, true, genCertUsage)
	flags.BoolVar(&g.genBundle, genBundleOption, false, genBundleUsage)
	flags.StringVar(&g.fileBundle, bundleOption, defaults.fileBundle, bundleUsage)
	flags.StringVar(&g.fileCaCert, caCertOption, defaults.fileCaCert, caCertUsage)
	flags.StringVar(&g.fileCaKey, caKeyOption, defaults.fileCaKey, caKeyUsage)
	flags.StringVar(&g.fileServerCert, svCertOption, defaults.fileServerCert, svCertUsage)
	flags.StringVar(&g.fileServerKey, svKeyOption, defaults.fileServerKey, svKeyUsage)
	flags.StringVar(&g.caCommonName, caCnOption, "iss.genaiz.com", caCnUsage)
	flags.IntVar(&g.caLifetime, caLifetimeOption, 5, caLifetimeUsage)
	flags.StringVar(&g.commonName, svCnOption, "dev.genaiz.com", svCnUsage)
	flags.IntVar(&g.lifetime, svLifetimeOption, 24, svLifetimeUsage)
	flags.StringVar(&g.organization, orgOption, "GenAIz", orgUsage)
	flags.StringVar(&g.country, countryOption, "CA", countryUsage)
	flags.StringVar(&g.province, provinceOption, "Quebec", provinceUsage)
	flags.StringVar(&g.locality, localityOption, "Montreal", localityUsage)
}

func newGenerate(handler Handler) *cobra.Command {
	var options = &generateOptions{}
	var generate = &cobra.Command{
		Use:   "generate [WORK_DIR]",
		Short: "Generates certificate PEM files",
		Long:  "Generates certificate PEM files under WORKDIR or the current directory",
		Run: func(cmd *cobra.Command, args []string) {
			var result *x509.Certificate
			var reset func()
			var err error

			if reset, err = dirz.CreateWorkingDir(args...); err == nil {
				defer errorz.DeferOnExit(&err, reset)()
				var arbiter = cert.NewArbiter().
					WithOrganization(options.organization).
					WithCountry(options.country).
					WithProvince(options.province).
					WithLocality(options.locality).
					WithCaCommonName(options.caCommonName).
					WithCaLifetime(options.caLifetime).
					WithCommonName(options.commonName).
					WithLifetime(options.lifetime)

				options.anchorFilePaths()
				arbiter.WithBundle(options.fileBundle).
					WithAuthority(options.fileCaCert, options.fileCaKey).
					WithServer(options.fileServerCert, options.fileServerKey)

				if options.genCert {
					if !options.genCA {
						if err = arbiter.ValidateAuthority(); err != nil {
							return
						}
					}

					if err = arbiter.BuildCert(); err == nil {
						result, err = arbiter.ParseCert()
					}
				} else if options.genCA {
					if err = arbiter.BuildAuthority(); err == nil {
						result, err = arbiter.ParseAuthority()
					}
				}

				if options.genBundle {
					err = arbiter.BuildRootBundle()
				}
			}

			if result != nil {
				handler(result)
			}
		},
	}

	options.init(generate)
	return generate
}

func NewGenerate() *cobra.Command {
	return newGenerate(defaultPrintHandler)
}
