// Package locker contains functionality to decrypt/encrypt locker files and write sensitive properties to local storage
package locker

import (
	"bytes"
	"crypto/cipher"
	"crypto/rand"
	"encoding/base64"
	"encoding/binary"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"slices"
	"strings"

	"github.com/awnumar/memguard"
	"golang.org/x/crypto/argon2"
	"golang.org/x/crypto/chacha20poly1305"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task"
)

const (
	// Recommendation #1 from https://www.rfc-editor.org/rfc/rfc9106.html#section-4

	DefaultVersion    = 1
	DefaultIterations = 1
	DefaultMemoryKb   = 2 * 1024 * 1024
	DefaultThreads    = 4
	DefaultKeyLength  = 32
)

var (
	errorLockerAccountNotFound = task.NewError("locker account not found")
	errorLockerContentEmpty    = task.NewError("locker is empty")
	errorLockerDataLinkFound   = task.NewError("locker data link known")
	errorLockerDataLinkInvalid = task.NewError("locker data link invalid")
	errorLockerSourceNotFound  = task.NewError("locker source not found")
)

// Enclave is an adapter to memguard.Enclave
type Enclave interface {
	Open() (*memguard.LockedBuffer, error)

	Size() int
}

type emptyEnclave struct {
}

func (ee emptyEnclave) Open() (*memguard.LockedBuffer, error) {
	return memguard.NewBuffer(0), nil
}

func (ee emptyEnclave) Size() int {
	return 0
}

type RemoteLink interface {
	GetPublishing() (string, string, string)
}

type SecuredCipher interface {
	Decrypt([]byte, *memguard.Enclave) (*memguard.LockedBuffer, error)

	Encrypt(*memguard.Enclave, *memguard.Enclave) ([]byte, error)
}

type BaseParams struct {
	LockerPath string
	Passphrase Enclave
}

type LinkParams struct {
	Oem     string
	Handle  string
	Version string
}

type PropertyParams struct {
	Key    string
	Value  string
	Secret Enclave
}

type lockerAccount struct {
	AccountUrl  string       `json:"url,omitempty"`
	DataSources []lockerLink `json:"dataSources,omitempty"`
	DataStores  []lockerLink `json:"dataStores,omitempty"`
}

func (la lockerAccount) findSource(handle string) (*lockerLink, error) {
	if i := slices.IndexFunc(la.DataSources, func(link lockerLink) bool {
		return handle == link.LockerHandle
	}); i >= 0 {
		return &la.DataSources[i], nil
	}

	return nil, errorLockerSourceNotFound
}

func (la lockerAccount) withSource(source *lockerLink) *lockerAccount {
	var sources []lockerLink

	for _, s := range la.DataSources {
		if s.LockerHandle == source.LockerHandle {
			sources = append(sources, *source)
		} else {
			sources = append(sources, s)
		}
	}

	return &lockerAccount{
		AccountUrl:  la.AccountUrl,
		DataSources: sources,
		DataStores:  la.DataStores,
	}
}

type lockerBody struct {
	Accounts []lockerAccount `json:"accounts"`
}

func (lb lockerBody) findAccount(accountUrl string) (*lockerAccount, error) {
	if i := slices.IndexFunc(lb.Accounts, func(account lockerAccount) bool {
		return strings.EqualFold(account.AccountUrl, accountUrl)
	}); i >= 0 {
		return &lb.Accounts[i], nil
	}

	return nil, errorLockerAccountNotFound
}

func (lb lockerBody) withAccount(account *lockerAccount) *lockerBody {
	var accounts []lockerAccount

	for _, a := range lb.Accounts {
		if a.AccountUrl == account.AccountUrl {
			accounts = append(accounts, *account)
		} else {
			accounts = append(accounts, a)
		}
	}

	return &lockerBody{
		Accounts: accounts,
	}
}

type lockerHeader struct {
	version    uint8
	iterations uint32
	memoryKb   uint32
	threads    uint8
	keyLength  uint32
	nonce      []byte // nonce is an initialization vector used by chacha20poly1305
	salt       []byte // salt is used for creating the argon2 hashed key from the passphrase supplied
}

func (lh lockerHeader) Decrypt(encrypted []byte, passphrase Enclave) (*memguard.LockedBuffer, error) {
	var keyBuffer *memguard.LockedBuffer
	var err error

	if keyBuffer, err = lh.makeArgon2idKey(passphrase); err == nil {
		var aead cipher.AEAD

		defer keyBuffer.Destroy()

		if aead, err = chacha20poly1305.NewX(keyBuffer.Bytes()); err == nil {
			var decrypted []byte

			if decrypted, err = aead.Open(nil, lh.nonce, encrypted, lh.makeAeadAdditional()); err == nil {
				return memguard.NewBufferFromBytes(decrypted), nil
			}
		}
	}

	return nil, err
}

func (lh lockerHeader) Encrypt(data Enclave, passphrase Enclave) ([]byte, error) {
	var keyBuffer *memguard.LockedBuffer
	var err error

	if keyBuffer, err = lh.makeArgon2idKey(passphrase); err == nil {
		var dataBuffer *memguard.LockedBuffer
		var aead cipher.AEAD

		defer keyBuffer.Destroy()

		if dataBuffer, err = data.Open(); err == nil {
			defer dataBuffer.Destroy()

			if aead, err = chacha20poly1305.NewX(keyBuffer.Bytes()); err == nil {
				return aead.Seal(nil, lh.nonce, dataBuffer.Bytes(), lh.makeAeadAdditional()), nil
			}
		}
	}

	return nil, err
}

func (lh lockerHeader) makeAeadAdditional() []byte {
	var b bytes.Buffer

	_ = binary.Write(&b, binary.BigEndian, lh.version)
	_ = binary.Write(&b, binary.BigEndian, lh.iterations)
	_ = binary.Write(&b, binary.BigEndian, lh.memoryKb)
	_ = binary.Write(&b, binary.BigEndian, lh.threads)
	_ = binary.Write(&b, binary.BigEndian, lh.keyLength)
	_ = binary.Write(&b, binary.BigEndian, uint16(len(lh.salt)))
	return b.Bytes()
}

func (lh lockerHeader) makeArgon2idKey(passphrase Enclave) (*memguard.LockedBuffer, error) {
	var lb *memguard.LockedBuffer
	var err error

	if lb, err = passphrase.Open(); err == nil {
		var result *memguard.LockedBuffer

		defer lb.Destroy()
		result = memguard.NewBufferFromBytes(
			argon2.IDKey(lb.Bytes(), lh.salt, lh.iterations, lh.memoryKb, lh.threads, lh.keyLength))
		return result, nil
	}

	return nil, err
}

type lockerLink struct {
	LockerHandle string `json:"lockerHandle"`
	LinkOem      string `json:"oem"`
	LinkHandle   string `json:"handle"`
	LinkVersion  string `json:"version"`
	Properties   string `json:"properties"`
}

func (ll lockerLink) GetPublishing() (string, string, string) {
	return ll.LinkOem, ll.LinkHandle, ll.LinkVersion
}

func (ll lockerLink) decodeProperties(passphrase Enclave) (map[string]string, error) {
	if ll.Properties != "" {
		var err error
		var b []byte

		if b, err = base64.StdEncoding.DecodeString(ll.Properties); err == nil {
			var header *lockerHeader
			var cipherBytes []byte

			if header, cipherBytes, err = readLockerData(bytes.NewReader(b)); err == nil {
				var propBytes *memguard.LockedBuffer

				if propBytes, err = header.Decrypt(cipherBytes, passphrase); err == nil {
					var properties map[string]string

					defer propBytes.Destroy()

					if err = json.Unmarshal(propBytes.Bytes(), &properties); err == nil {
						return properties, nil
					}
				}
			}
		}

		return nil, err
	}

	return map[string]string{}, nil
}

func (ll lockerLink) encodeProperties(properties map[string]string, passphrase Enclave) (string, error) {
	if len(properties) > 0 {
		var b []byte
		var err error

		if b, err = json.Marshal(properties); err == nil {
			var header = newLockerHeader()
			var enclaved = memguard.NewEnclave(b)

			if b, err = header.Encrypt(enclaved, passphrase); err == nil {
				return base64.StdEncoding.EncodeToString(b), nil
			}
		}

		return "", err
	}

	return "", nil
}

func (ll lockerLink) withProperty(key string, value Enclave, passphrase Enclave) (*lockerLink, error) {
	var decoded map[string]string
	var err error

	if decoded, err = ll.decodeProperties(passphrase); err == nil {
		var lb *memguard.LockedBuffer

		if lb, err = value.Open(); err == nil {
			var encoded string

			defer lb.Destroy()
			decoded[key] = lb.String()

			if encoded, err = ll.encodeProperties(decoded, passphrase); err == nil {
				return &lockerLink{
					LockerHandle: ll.LockerHandle,
					LinkOem:      ll.LinkOem,
					LinkHandle:   ll.LinkHandle,
					LinkVersion:  ll.LinkVersion,
					Properties:   encoded,
				}, nil
			}
		}
	}

	return nil, err
}

type SecuredLockerState struct {
	SecuredLockerTracking
	state *task.State
}

type SecuredLockerTracking struct {
	current     *memguard.LockedBuffer
	currentPath string
}

func (slt *SecuredLockerTracking) Close(path string) error {
	if path != slt.currentPath {
		return os.Rename(slt.currentPath, path)
	}

	return nil
}

func (slt *SecuredLockerTracking) Destroy() {
	if slt.current != nil {
		defer slt.current.Destroy()
	}
}

func (slt *SecuredLockerTracking) IsOpened() bool {
	return slt.current != nil
}

func (slt *SecuredLockerTracking) LookupSource(accountUrl, handle string) (RemoteLink, error) {
	var body *lockerBody
	var err error

	if body, err = slt.unfold(); err == nil {
		var account *lockerAccount

		if account, err = body.findAccount(accountUrl); err == nil {
			var link *lockerLink

			if link, err = account.findSource(handle); err == nil {
				return link, nil
			}
		}
	}

	return nil, fmt.Errorf("data source [%s] for account [%s] does not exist", handle, accountUrl)
}

func (slt *SecuredLockerTracking) Read(path string, passphrase Enclave) error {
	var fd *os.File
	var err error

	if fd, err = os.Open(path); err == nil {
		var header *lockerHeader
		var encrypted []byte

		defer filez.CloseSilently(fd)

		if header, encrypted, err = readLockerData(fd); err == nil {
			if slt.current, err = header.Decrypt(encrypted, passphrase); err == nil {
				return nil
			}
		}
	}

	return err
}

func (slt *SecuredLockerTracking) Write(path string, passphrase Enclave) error {
	var fd *os.File
	var err error

	if fd, err = os.OpenFile(path, os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0660); err == nil {
		var root *lockerBody
		var header *lockerHeader

		defer filez.CloseSilently(fd)
		header = newLockerHeader()

		if root, err = slt.unfold(); err == nil {
			var bodyByes []byte

			if bodyByes, err = json.Marshal(root); err == nil {
				var bodyEnclave = memguard.NewEnclave(bodyByes)
				var encrypted []byte

				if encrypted, err = header.Encrypt(bodyEnclave, passphrase); err == nil {
					if err = writeLockerData(header, fd, encrypted); err == nil {
						slt.currentPath = path
						return nil
					}
				}
			}
		}
	}

	return err
}

func (slt *SecuredLockerTracking) addSource(accountUrl string, link *lockerLink) error {
	var body *lockerBody
	var err error

	if body, err = slt.unfold(); err == nil {
		var account *lockerAccount

		if account, err = body.findAccount(accountUrl); err != nil {
			body.Accounts = append(body.Accounts, lockerAccount{
				AccountUrl: accountUrl,
			})
			account = &body.Accounts[len(body.Accounts)-1]
		}

		if i := slices.IndexFunc(account.DataSources, func(l lockerLink) bool {
			return link.LockerHandle == l.LockerHandle
		}); i >= 0 {
			return fmt.Errorf("data source [%s] for account [%s] is already defined", link.LockerHandle, accountUrl)
		}

		account.DataSources = append(account.DataSources, *link)
		slt.current = slt.fold(body)
		return nil
	}

	return err
}

func (slt *SecuredLockerTracking) fold(body *lockerBody) *memguard.LockedBuffer {
	var b, err = json.Marshal(body)

	// There is no way json.Marshal ever returns an error here, unless lockerBody was tempered with
	panicz.PanicIfError(err)
	return memguard.NewBufferFromBytes(b)
}

func (slt *SecuredLockerTracking) unfold() (*lockerBody, error) {
	var result lockerBody
	var err error

	if slt.current == nil {
		return &lockerBody{}, nil
	}

	if err = json.Unmarshal(slt.current.Bytes(), &result); err == nil {
		return &result, nil
	}

	return nil, err
}

func (slt *SecuredLockerTracking) updateSource(accountUrl, handle, key string, value, passphrase Enclave) error {
	var body *lockerBody
	var err error

	if body, err = slt.unfold(); err == nil {
		var account *lockerAccount

		if account, err = body.findAccount(accountUrl); err == nil {
			if i := slices.IndexFunc(account.DataSources, func(l lockerLink) bool {
				return handle == l.LockerHandle
			}); i >= 0 {
				var link *lockerLink

				if link, err = account.DataSources[i].withProperty(key, value, passphrase); err == nil {
					var updatedBody = body.withAccount(account.withSource(link))

					slt.current = slt.fold(updatedBody)
					return nil
				}

				return err
			}

			return fmt.Errorf("data source [%s] for account [%s] does not exist", handle, accountUrl)
		}
	}

	return err
}

func NewSecuredLockerState(state *task.State) *SecuredLockerState {
	var internal, ok = state.Internal.(SecuredLockerTracking)
	var current *memguard.LockedBuffer
	var path string

	if ok {
		current = internal.current
		path = internal.currentPath
	}

	return &SecuredLockerState{
		SecuredLockerTracking: SecuredLockerTracking{
			current:     current,
			currentPath: path,
		},
		state: state,
	}
}

func newEmptyEnclave() Enclave {
	return &emptyEnclave{}
}

// newLockerHeader initializes a lockerHeader with expected defaults, an initialization vector for chacha20poly1305 and a salt for the argon2id key
func newLockerHeader() *lockerHeader {
	var nonce = make([]byte, chacha20poly1305.NonceSizeX)
	var salt = make([]byte, 16)
	var err error

	// these will never return errors, except on legacy Linux systems
	_, err = rand.Read(salt)
	panicz.PanicIfError(err)
	_, err = rand.Read(nonce)
	panicz.PanicIfError(err)
	return &lockerHeader{
		version:    DefaultVersion,
		iterations: DefaultIterations,
		memoryKb:   DefaultMemoryKb,
		threads:    DefaultThreads,
		keyLength:  DefaultKeyLength,
		nonce:      nonce,
		salt:       salt,
	}
}

// readLockerData retrieves the lockerHeader of a locker io.Reader with the parameters to build an argon2id key and the initialization vector of the chacha20poly1305 encryption algorithm. It also returns encrypted bytes.
func readLockerData(in io.Reader) (*lockerHeader, []byte, error) {
	var header = &lockerHeader{}
	var saltLength uint16
	var encrypted bytes.Buffer
	var err error

	if err = binary.Read(in, binary.BigEndian, &header.version); err != nil {
		return nil, nil, err
	}

	if err = binary.Read(in, binary.BigEndian, &header.iterations); err != nil {
		return nil, nil, err
	}

	if err = binary.Read(in, binary.BigEndian, &header.memoryKb); err != nil {
		return nil, nil, err
	}

	if err = binary.Read(in, binary.BigEndian, &header.threads); err != nil {
		return nil, nil, err
	}

	if err = binary.Read(in, binary.BigEndian, &header.keyLength); err != nil {
		return nil, nil, err
	}

	if err = binary.Read(in, binary.BigEndian, &saltLength); err != nil {
		return nil, nil, err
	}

	header.salt = make([]byte, saltLength)

	if _, err = io.ReadFull(in, header.salt); err != nil {
		return nil, nil, err
	}

	header.nonce = make([]byte, chacha20poly1305.NonceSizeX)

	if _, err = io.ReadFull(in, header.nonce); err != nil {
		return nil, nil, err
	}

	// ignore copy errors, it will always succeed at this point
	if _, _ = io.Copy(&encrypted, in); encrypted.Len() > 0 {
		return header, encrypted.Bytes(), nil
	}

	return nil, nil, errorLockerContentEmpty
}

// writeLockerData will write locker data according to the format used under readLockerData treating encrypted as already secured
func writeLockerData(header *lockerHeader, out io.Writer, encrypted []byte) error {
	var err error

	if err = binary.Write(out, binary.BigEndian, header.version); err != nil {
		return err
	}

	if err = binary.Write(out, binary.BigEndian, header.iterations); err != nil {
		return err
	}

	if err = binary.Write(out, binary.BigEndian, header.memoryKb); err != nil {
		return err
	}

	if err = binary.Write(out, binary.BigEndian, header.threads); err != nil {
		return err
	}

	if err = binary.Write(out, binary.BigEndian, header.keyLength); err != nil {
		return err
	}

	if err = binary.Write(out, binary.BigEndian, uint16(len(header.salt))); err != nil {
		return err
	}

	if _, err = out.Write(header.salt); err != nil {
		return err
	}

	if _, err = out.Write(header.nonce); err != nil {
		return err
	}

	_, err = out.Write(encrypted)
	return err
}
