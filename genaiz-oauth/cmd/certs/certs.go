package certs

import (
	"crypto/x509"
	"fmt"
	"strings"
	"time"

	"github.com/spf13/cobra"
)

var (
	defaults = certsDefaults{
		fileBundle:     "root.crt",
		fileCaCert:     "ca.cert",
		fileCaKey:      "ca.key",
		fileServerCert: "server.cert",
		fileServerKey:  "server.key",
	}
)

type certsDefaults struct {
	fileBundle     string
	fileCaCert     string
	fileCaKey      string
	fileServerCert string
	fileServerKey  string
}

type certsOptions struct {
	fileBundle     string
	fileCaCert     string
	fileCaKey      string
	fileServerCert string
	fileServerKey  string
}

type Handler func(certificate *x509.Certificate)

func NewCert() *cobra.Command {
	var cert = &cobra.Command{
		Use:   "certs",
		Short: "Reads and writes PEM-based certificates",
		Long:  "Reads and writes PEM-based certificates to use for oauth provisioning",
	}

	cert.AddCommand(NewGenerate(), NewParse())
	return cert
}

func defaultPrintHandler(cert *x509.Certificate) {
	fmt.Println("Issued To")
	fmt.Printf("  Common Name (CN): %s\n", cert.Subject.CommonName)
	fmt.Printf("  Organization (O): %s\n", strings.Join(cert.Subject.Organization, ","))
	fmt.Printf("  Organizational Unit (OU): %s\n", strings.Join(cert.Subject.OrganizationalUnit, ","))
	fmt.Printf("  Locality (L): %s\n", strings.Join(cert.Subject.Locality, ","))
	fmt.Printf("  State or Province (ST): %s\n", strings.Join(cert.Subject.Province, ","))
	fmt.Printf("  Country (C): %s\n", strings.Join(cert.Subject.Country, ","))
	fmt.Println("Expires")
	fmt.Printf("  Not After: %s\n", cert.NotAfter.Format(time.DateTime))

	if !cert.IsCA {
		fmt.Println("\nIssuer")
		fmt.Printf("  Common Name (CN): %s\n", cert.Issuer.CommonName)
		fmt.Printf("  Organization (O): %s\n", strings.Join(cert.Issuer.Organization, ","))
		fmt.Printf("  Organizational Unit (OU): %s\n", strings.Join(cert.Issuer.OrganizationalUnit, ","))
	}
}
