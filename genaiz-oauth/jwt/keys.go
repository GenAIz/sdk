package jwt

import (
	"encoding/json"
	"fmt"
	"os"

	"github.com/lestrrat-go/jwx/v3/jwk"
)

type KeyManager interface {
	MergeKeySet() error

	RemoveKeySet() error

	WithPemKeys(...string) KeyManager

	WithSetFile(string) KeyManager

	WriteKeySet() error
}

type keyManager struct {
	jwksFile string

	pemKeyFiles []string
}

func (co *keyManager) appendJwkSet(set jwk.Set) error {
	var keyBytes []byte
	var err error

	for _, kf := range co.pemKeyFiles {
		if keyBytes, err = os.ReadFile(kf); err == nil {
			var key jwk.Key

			if key, err = jwk.ParseKey(keyBytes, jwk.WithPEM(true)); err == nil {
				if err = jwk.AssignKeyID(key); err == nil {
					err = set.AddKey(key)
				}
			}
		}

		if err != nil {
			break
		}
	}

	return err
}

func (co *keyManager) createJwkSet() (jwk.Set, error) {
	var result = jwk.NewSet()

	return result, co.appendJwkSet(result)
}

func (co *keyManager) removeJwkSet(set jwk.Set) error {
	var keyBytes []byte
	var err error

	for _, kf := range co.pemKeyFiles {
		if keyBytes, err = os.ReadFile(kf); err == nil {
			var key jwk.Key

			if key, err = jwk.ParseKey(keyBytes, jwk.WithPEM(true)); err == nil {
				if err = jwk.AssignKeyID(key); err == nil {
					var ok bool
					var id string

					if id, ok = key.KeyID(); ok {
						if key, ok = set.LookupKeyID(id); ok {
							err = set.RemoveKey(key)
						}
					}

					if !ok || err != nil {
						return fmt.Errorf("could not remove key [%s]", id)
					}
				}
			}
		}
	}

	return err
}

func (co *keyManager) writeJwkSet(keySet jwk.Set) error {
	var fd *os.File
	var err error

	if fd, err = os.Create(co.jwksFile); err == nil {
		var encoder = json.NewEncoder(fd)

		encoder.SetIndent("", "  ")
		return encoder.Encode(keySet)
	}

	return nil
}

func (co *keyManager) MergeKeySet() error {
	var ks jwk.Set
	var setBytes []byte
	var err error

	if setBytes, err = os.ReadFile(co.jwksFile); err == nil {
		if ks, err = jwk.Parse(setBytes); err == nil {
			if err = co.appendJwkSet(ks); err == nil {
				err = co.writeJwkSet(ks)
			}
		}
	}

	return err
}

func (co *keyManager) RemoveKeySet() error {
	var ks jwk.Set
	var setBytes []byte
	var err error

	if setBytes, err = os.ReadFile(co.jwksFile); err == nil {
		if ks, err = jwk.Parse(setBytes); err == nil {
			if err = co.removeJwkSet(ks); err == nil {
				err = co.writeJwkSet(ks)
			}
		}
	}

	return err
}

func (co *keyManager) WithPemKeys(keyFiles ...string) KeyManager {
	co.pemKeyFiles = keyFiles
	return co
}

func (co *keyManager) WithSetFile(jwksFile string) KeyManager {
	co.jwksFile = jwksFile
	return co
}

func (co *keyManager) WriteKeySet() error {
	var ks jwk.Set
	var err error

	if ks, err = co.createJwkSet(); err == nil {
		err = co.writeJwkSet(ks)
	}

	return err
}

func NewKeyManager() KeyManager {
	return &keyManager{}
}
