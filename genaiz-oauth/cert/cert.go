package cert

import (
	"bytes"
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"fmt"
	"math/big"
	"net"
	"os"
	"time"
)

type Arbiter interface {
	BuildAuthority() error

	BuildCert() error

	BuildRootBundle() error

	BuildSigningKey() (string, string, []byte)

	GetRawCert() ([]byte, error)

	GetRawKey() ([]byte, error)

	ParseAuthority() (*x509.Certificate, error)

	ParseBundle() ([]*x509.Certificate, error)

	ParseCert() (*x509.Certificate, error)

	ValidateAuthority() error

	WithAuthority(string, string) Arbiter

	WithBundle(string) Arbiter

	WithCaCommonName(string) Arbiter

	WithCaLifetime(int) Arbiter

	WithCommonName(string) Arbiter

	WithCountry(string) Arbiter

	WithLifetime(int) Arbiter

	WithLocality(string) Arbiter

	WithOrganization(string) Arbiter

	WithProvince(string) Arbiter

	WithServer(string, string) Arbiter
}

type arbiter struct {
	bundleFile string
	caCertFile string
	caKeyFile  string
	certFile   string
	keyFile    string

	caCommonName   string
	caOrganization string
	caCountry      string
	caProvince     string
	caLocality     string
	caLifetime     time.Duration

	certCommonName   string
	certOrganization string
	certCountry      string
	certProvince     string
	certLocality     string
	certLifetime     time.Duration
	certLocal        bool
}

func (a *arbiter) buildAuthorityCert() *x509.Certificate {
	var now = time.Now()

	return &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixMilli()),
		Subject: pkix.Name{
			CommonName:   a.caCommonName,
			Organization: []string{a.caOrganization},
			Country:      []string{a.caCountry},
			Province:     []string{a.caProvince},
			Locality:     []string{a.caLocality},
		},

		NotBefore:             now,
		NotAfter:              now.Add(a.caLifetime),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
}

func (a *arbiter) buildAuthorityEntity() (*entity, error) {
	var ent *entity
	var err error

	if ent, err = a.loadEntity(a.caCertFile, a.caKeyFile); os.IsNotExist(err) {
		var cert = a.buildAuthorityCert()

		if ent, err = a.writeSelfSignedEntity(cert, cert, a.caCertFile, a.caKeyFile); err == nil {
			return ent, nil
		}
	} else if err = a.validateAuthority(ent); err == nil {
		return ent, nil
	}

	return nil, err
}

func (a *arbiter) buildServerCert() *x509.Certificate {
	var now = time.Now()

	var result = &x509.Certificate{
		SerialNumber: big.NewInt(time.Now().UnixMilli()),
		Subject: pkix.Name{
			CommonName:   a.certCommonName,
			Organization: []string{a.certOrganization},
			Country:      []string{a.certCountry},
			Province:     []string{a.certProvince},
			Locality:     []string{a.certLocality},
		},
		NotBefore:    now,
		NotAfter:     now.Add(a.certLifetime),
		SubjectKeyId: []byte{1, 2, 3, 4, 6},
		ExtKeyUsage:  []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:     x509.KeyUsageDigitalSignature,
	}

	if a.certLocal {
		result.IPAddresses = []net.IP{net.IPv4(127, 0, 0, 1), net.IPv6loopback}
	}

	return result
}

func (a *arbiter) loadEntity(certFile, keyFile string) (*entity, error) {
	var certBytes, keyBytes []byte
	var err error

	if certBytes, err = os.ReadFile(certFile); err == nil {
		var certBlock, _ = pem.Decode(certBytes)
		var cert *x509.Certificate
		var key *ecdsa.PrivateKey

		if certBlock == nil {
			return nil, fmt.Errorf("%s is not a valid pem file", certFile)
		}

		if cert, err = x509.ParseCertificate(certBlock.Bytes); err == nil {
			if keyBytes, err = os.ReadFile(keyFile); err == nil {
				var keyBlock, _ = pem.Decode(keyBytes)

				if key, err = x509.ParseECPrivateKey(keyBlock.Bytes); err == nil {
					return &entity{
						cert: cert,
						key:  key,
					}, nil
				}
			}
		}
	}

	return nil, err
}

func (a *arbiter) validateAuthority(ent *entity) error {
	if !ent.cert.IsCA {
		return fmt.Errorf("%s is not a certificate authority", a.caCertFile)
	} else {
		var verifyOptions = x509.VerifyOptions{
			KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		}

		if ent.cert.Issuer.CommonName == ent.cert.Subject.CommonName {
			verifyOptions.Roots = x509.NewCertPool()
			verifyOptions.Roots.AddCert(ent.cert)
		}

		if _, err := ent.cert.Verify(verifyOptions); err != nil {
			return fmt.Errorf("%s is an invalid authority: %s", a.caCertFile, err)
		}
	}

	return nil
}

func (a *arbiter) writeCaSignedEntity(cert *x509.Certificate, parent *x509.Certificate, parentKey *ecdsa.PrivateKey) (*entity, error) {
	var key *ecdsa.PrivateKey
	var err error

	if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err == nil {
		return a.writeSignedEntity(cert, key, parent, parentKey, a.certFile, a.keyFile)
	}

	return nil, err
}

func (a *arbiter) writeSelfSignedEntity(cert *x509.Certificate, parent *x509.Certificate, certFile, keyFile string) (*entity, error) {
	var key *ecdsa.PrivateKey
	var err error

	if key, err = ecdsa.GenerateKey(elliptic.P256(), rand.Reader); err == nil {
		return a.writeSignedEntity(cert, key, parent, key, certFile, keyFile)
	}

	return nil, err
}

func (a *arbiter) writeSignedEntity(
	cert *x509.Certificate, certKey *ecdsa.PrivateKey,
	parent *x509.Certificate, parentKey *ecdsa.PrivateKey,
	certFile, keyFile string) (*entity, error) {
	var certBytes, keyBytes []byte
	var err error

	if certBytes, err = x509.CreateCertificate(rand.Reader, cert, parent, &certKey.PublicKey, parentKey); err == nil {
		var buffer = new(bytes.Buffer)

		if keyBytes, err = x509.MarshalECPrivateKey(certKey); err == nil {
			var pemKeyBlock = &pem.Block{Type: "EC PRIVATE KEY", Bytes: keyBytes}
			var pemCertBlock = &pem.Block{Type: "CERTIFICATE", Bytes: certBytes}

			if err = pem.Encode(buffer, pemKeyBlock); err == nil {
				if err = os.WriteFile(keyFile, buffer.Bytes(), 0600); err != nil {
					return nil, err
				}
			}

			buffer.Reset()

			if err = pem.Encode(buffer, pemCertBlock); err == nil {
				if err = os.WriteFile(certFile, buffer.Bytes(), 0600); err != nil {
					return nil, err
				}
			}

			return &entity{
				key:  certKey,
				cert: cert,
			}, nil
		}
	}

	return nil, err
}

func (a *arbiter) BuildAuthority() error {
	var _, err = a.buildAuthorityEntity()

	return err
}

func (a *arbiter) BuildCert() error {
	var auth, ent *entity
	var err error

	if ent, err = a.loadEntity(a.certFile, a.keyFile); os.IsNotExist(err) {
		if auth, err = a.buildAuthorityEntity(); err == nil {
			var cert = a.buildServerCert()

			_, err = a.writeCaSignedEntity(cert, auth.cert, auth.key)
		}
	} else if ent.cert.NotAfter.Before(time.Now()) {
		err = fmt.Errorf("%s is an expired cert", a.certFile)
	}

	return err
}

func (a *arbiter) BuildRootBundle() error {
	var ent *entity
	var err error

	if ent, err = a.loadEntity(a.certFile, a.keyFile); err == nil {
		var auth *entity

		if auth, err = a.loadEntity(a.caCertFile, a.caKeyFile); err == nil {
			var buffer = new(bytes.Buffer)
			var pemCertBlock = &pem.Block{Type: "CERTIFICATE", Bytes: ent.cert.Raw}
			var pemAuthBlock = &pem.Block{Type: "CERTIFICATE", Bytes: auth.cert.Raw}

			if err = pem.Encode(buffer, pemAuthBlock); err == nil {
				if err = pem.Encode(buffer, pemCertBlock); err == nil {
					err = os.WriteFile(a.bundleFile, buffer.Bytes(), 0600)
				}
			}
		}
	}

	return err
}

func (a *arbiter) BuildSigningKey() (string, string, []byte) {
	return "", "HS256", []byte{}
}

func (a *arbiter) GetCert() (*x509.Certificate, error) {
	var certBytes []byte
	var err error

	if certBytes, err = a.GetRawKey(); err == nil {
		return x509.ParseCertificate(certBytes)
	}

	return nil, err
}

func (a *arbiter) GetKey() (*ecdsa.PrivateKey, error) {
	var keyBytes []byte
	var err error

	if keyBytes, err = a.GetRawKey(); err == nil {
		return x509.ParseECPrivateKey(keyBytes)
	}

	return nil, err
}

func (a *arbiter) GetRawCert() ([]byte, error) {
	return os.ReadFile(a.certFile)
}

func (a *arbiter) GetRawKey() ([]byte, error) {
	return os.ReadFile(a.certFile)
}

func (a *arbiter) ParseAuthority() (*x509.Certificate, error) {
	var ent *entity
	var err error

	if ent, err = a.loadEntity(a.caCertFile, a.caKeyFile); err == nil {
		return ent.cert, nil
	}

	return nil, err
}

func (a *arbiter) ParseBundle() ([]*x509.Certificate, error) {
	var result []*x509.Certificate
	var certBytes []byte
	var err error

	if certBytes, err = os.ReadFile(a.bundleFile); err == nil {
		for block, rest := pem.Decode(certBytes); block != nil; block, rest = pem.Decode(rest) {
			if c, e := x509.ParseCertificate(block.Bytes); e == nil {
				result = append(result, c)
			} else {
				return nil, e
			}
		}
	}

	return result, err
}

func (a *arbiter) ParseCert() (*x509.Certificate, error) {
	var ent *entity
	var err error

	if ent, err = a.loadEntity(a.certFile, a.keyFile); err == nil {
		return ent.cert, nil
	}

	return nil, err
}

func (a *arbiter) ValidateAuthority() error {
	var ent *entity
	var err error

	if ent, err = a.loadEntity(a.caCertFile, a.caKeyFile); err == nil {
		err = a.validateAuthority(ent)
	}

	return err
}

func (a *arbiter) WithAuthority(certFile, keyFile string) Arbiter {
	a.caCertFile = certFile
	a.caKeyFile = keyFile
	return a
}

func (a *arbiter) WithBundle(bundleFile string) Arbiter {
	a.bundleFile = bundleFile
	return a
}

func (a *arbiter) WithCaCommonName(commonName string) Arbiter {
	a.caCommonName = commonName
	return a
}

func (a *arbiter) WithCaLifetime(days int) Arbiter {
	a.caLifetime = time.Hour * 24 * time.Duration(days)
	return a
}

func (a *arbiter) WithCommonName(commonName string) Arbiter {
	a.certCommonName = commonName
	return a
}

func (a *arbiter) WithCountry(country string) Arbiter {
	a.caCountry = country
	a.certCountry = country
	return a
}

func (a *arbiter) WithLifetime(hours int) Arbiter {
	a.certLifetime = time.Hour * time.Duration(hours)
	return a
}

func (a *arbiter) WithLocality(locality string) Arbiter {
	a.caLocality = locality
	a.certLocality = locality
	return a
}

func (a *arbiter) WithOrganization(organization string) Arbiter {
	a.caOrganization = organization
	a.certOrganization = organization
	return a
}

func (a *arbiter) WithProvince(province string) Arbiter {
	a.caProvince = province
	a.certProvince = province
	return a
}

func (a *arbiter) WithServer(certFile, keyFile string) Arbiter {
	a.certFile = certFile
	a.keyFile = keyFile
	return a
}

func NewArbiter() Arbiter {
	return &arbiter{}
}

type entity struct {
	key  *ecdsa.PrivateKey
	cert *x509.Certificate
}
