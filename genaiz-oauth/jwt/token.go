package jwt

import (
	"crypto/x509"
	"encoding/pem"
	"os"
	"time"

	"github.com/google/uuid"
	"github.com/lestrrat-go/jwx/v3/jwa"
	"github.com/lestrrat-go/jwx/v3/jwk"
	"github.com/lestrrat-go/jwx/v3/jws"
	jwt3 "github.com/lestrrat-go/jwx/v3/jwt"
)

type Builder interface {
	Build() ([]byte, error)

	Decode([]byte) (any, error)

	WithAudience(string) Builder

	WithExpiry(int) Builder

	WithAccess(string, []string) Builder

	WithSigner(string, string) Builder
}

type access struct {
	Type    string   `json:"type"`
	Name    string   `json:"name"`
	Actions []string `json:"actions"`
}

type builder struct {
	audience string
	expiry   time.Duration
	access   []*access

	signingCertFile string
	signingKeyFile  string
}

func (b *builder) getSigningCert() (*x509.Certificate, error) {
	var certBytes []byte
	var err error

	if certBytes, err = os.ReadFile(b.signingCertFile); err == nil {
		var block *pem.Block

		if block, _ = pem.Decode(certBytes); block != nil {
			return x509.ParseCertificate(block.Bytes)
		}
	}

	return nil, err
}

func (b *builder) getSigningKey() (jwt3.SignEncryptParseOption, error) {
	var keyBytes []byte
	var err error

	if keyBytes, err = os.ReadFile(b.signingKeyFile); err == nil {
		var key jwk.Key

		if key, err = jwk.ParseKey(keyBytes, jwk.WithPEM(true)); err == nil {
			if err = jwk.AssignKeyID(key); err == nil {
				var alg, _ = jwa.KeyAlgorithmFrom("ES256")

				return jwt3.WithKey(alg, key), nil
			}
		}
	}

	return nil, err
}

func (b *builder) Build() ([]byte, error) {
	var cert *x509.Certificate
	var err error

	if cert, err = b.getSigningCert(); err == nil {
		var now = time.Now()
		var token jwt3.Token

		var jwtBuilder = jwt3.NewBuilder().
			Audience([]string{b.audience}).
			Claim("access", b.access).
			Expiration(now.Add(b.expiry)).
			Issuer(cert.Issuer.CommonName).
			JwtID(uuid.NewString()).
			NotBefore(now)

		if token, err = jwtBuilder.Build(); err == nil {
			var signingKeyOption jwt3.SignEncryptParseOption
			var result []byte

			if signingKeyOption, err = b.getSigningKey(); err == nil {
				if result, err = jwt3.Sign(token, signingKeyOption); err == nil {
					return result, nil
				}
			}
		}
	}

	return nil, err
}

func (b *builder) Decode(token []byte) (any, error) {
	var bytes []byte
	var err error

	if bytes, err = os.ReadFile(b.signingCertFile); err == nil {
		var key jwk.Key

		if key, err = jwk.ParseKey(bytes, jwk.WithPEM(true)); err == nil {
			var alg, _ = jwa.KeyAlgorithmFrom("ES256")

			if bytes, err = jws.Verify(token, jws.WithKey(alg, key)); err == nil {
				return jwt3.Parse(bytes, jwt3.WithVerify(false))
			}
		}
	}

	return nil, err
}
func (b *builder) WithAccess(repository string, operations []string) Builder {
	b.access = append(b.access, &access{
		Type:    "repository",
		Name:    repository,
		Actions: operations,
	})
	return b
}

func (b *builder) WithAudience(audience string) Builder {
	b.audience = audience
	return b
}

func (b *builder) WithExpiry(minutes int) Builder {
	b.expiry = time.Minute * time.Duration(minutes)
	return b
}

func (b *builder) WithSigner(signingCertFile, signingKeyFile string) Builder {
	b.signingCertFile = signingCertFile
	b.signingKeyFile = signingKeyFile
	return b
}

func NewBuilder() Builder {
	return &builder{}
}
