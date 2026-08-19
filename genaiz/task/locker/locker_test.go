package locker

import (
	"bytes"
	"crypto/rand"
	"encoding/binary"
	"errors"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/stretchr/testify/assert"
	"golang.org/x/crypto/chacha20poly1305"

	"genaiz.com/genaiz/task"
)

type stubEnclave struct {
	openBuffer *memguard.LockedBuffer
	openError  error
	size       int
}

func (se stubEnclave) Open() (*memguard.LockedBuffer, error) {
	return se.openBuffer, se.openError
}

func (se stubEnclave) Size() int {
	return se.size
}

type stubWriter struct {
	MatchError []byte
	WriteError error
	WriteSize  int
}

func (s *stubWriter) Write(writeBytes []byte) (n int, err error) {
	if bytes.Equal(writeBytes, s.MatchError) {
		return 0, s.WriteError
	}

	return len(writeBytes), nil
}

func TestLockerHeader_Decrypt_OpenFail(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = lockerHeader{}
	var testEnclave = &stubEnclave{
		openError: expectedError,
	}

	actual, err := testHeader.Decrypt([]byte{}, testEnclave)
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, actual)
}

func TestLockerHeader_Encrypt_OpenFail(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = lockerHeader{}
	var testEnclave = &stubEnclave{
		openError: expectedError,
	}

	actual, err := testHeader.Encrypt(memguard.NewEnclave([]byte("hello")), testEnclave)
	assert.ErrorIs(t, err, expectedError)
	assert.Nil(t, actual)
}

func TestNewSecuredLockerState(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("expected"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.True(t, actual.IsOpened())
	assert.Equal(t, actual.current, expectedBuffer)
	assert.Equal(t, actual.currentPath, expectedPath)
}

func TestNewSecuredLockerState_Close(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.NoError(t, actual.Close(expectedPath))
}

func TestNewSecuredLockerState_Destroy(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var actual = NewSecuredLockerState(testState)

	assert.NotEqual(t, 0, actual.current.Size())
	actual.Destroy()
	assert.Equal(t, 0, actual.current.Size())
}

func TestNewSecuredLockerState_isOpen(t *testing.T) {
	var actual = NewSecuredLockerState(&task.State{})

	assert.False(t, actual.IsOpened())
}

func TestNewSecuredLockerState_unfold(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("{}"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var testLocker = NewSecuredLockerState(testState)

	actual, err := testLocker.unfold()
	assert.NoError(t, err)
	assert.Empty(t, actual.Accounts)
}

func TestNewSecuredLockerState_unfold_JsonError(t *testing.T) {
	var expectedBuffer = memguard.NewBufferFromBytes([]byte("expected"))
	var expectedPath = t.TempDir()
	var testState = &task.State{
		Internal: SecuredLockerTracking{
			current:     expectedBuffer,
			currentPath: expectedPath,
		},
	}
	var testLocker = NewSecuredLockerState(testState)

	actual, err := testLocker.unfold()
	assert.Empty(t, actual)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorCopyEmpty(t *testing.T) {
	var testNonce = make([]byte, chacha20poly1305.NonceSizeX)
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	if _, err = rand.Read(testNonce); err == nil {
		testBytes = append(testBytes, []byte{37}...)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
		testBytes = append(testBytes, []byte{3}...)
		testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
		testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
		testBytes = append(testBytes, testSalt...)
		testBytes = append(testBytes, testNonce...)
		testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
		assert.Nil(t, testHeader)
		assert.Empty(t, actualBytes)
		assert.ErrorIs(t, err, errorLockerContentEmpty)
		return
	}

	assert.Fail(t, err.Error())
}

func Test_readLockerData_ErrorIterations(t *testing.T) {
	var testReader = strings.NewReader(string([]byte{27}))
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testHeader, actualBytes, err = readLockerData(testReader)
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorKeyLength(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorMemoryKb(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorNonce(t *testing.T) {
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
	testBytes = append(testBytes, testSalt...)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorSalt(t *testing.T) {
	var testSalt = []byte("salt is good")
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testBytes = binary.BigEndian.AppendUint16(testBytes, uint16(len(testSalt)))
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorSaltLength(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testBytes = append(testBytes, []byte{3}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2048)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorThreads(t *testing.T) {
	var testBytes []byte
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testBytes = append(testBytes, []byte{37}...)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 2)
	testBytes = binary.BigEndian.AppendUint32(testBytes, 1024)
	testHeader, actualBytes, err = readLockerData(strings.NewReader(string(testBytes)))
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_readLockerData_ErrorVersion(t *testing.T) {
	var testReader = strings.NewReader(string([]byte{}))
	var testHeader *lockerHeader
	var actualBytes []byte
	var err error

	testHeader, actualBytes, err = readLockerData(testReader)
	assert.Nil(t, testHeader)
	assert.Empty(t, actualBytes)
	assert.Error(t, err)
}

func Test_writeLockerData_ErrorIterations(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		iterations: 2,
	}
	var testWriter = &stubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.iterations)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorKeyLength(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		keyLength: 37,
	}
	var testWriter = &stubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.keyLength)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorMemoryKb(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		memoryKb: 1024,
	}
	var testWriter = &stubWriter{
		MatchError: make([]byte, 4),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint32(testWriter.MatchError, testHeader.memoryKb)
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorNonce(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		nonce: []byte("N once init"),
	}
	var testWriter = &stubWriter{
		MatchError: testHeader.nonce,
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorSaltLength(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		salt: []byte("salt is good"),
	}
	var testWriter = &stubWriter{
		MatchError: make([]byte, 2),
		WriteError: expectedError,
	}

	binary.BigEndian.PutUint16(testWriter.MatchError, uint16(len(testHeader.salt)))
	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorSalt(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		salt: []byte("salt is good"),
	}
	var testWriter = &stubWriter{
		MatchError: testHeader.salt,
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorThreads(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		threads: 37,
	}
	var testWriter = &stubWriter{
		MatchError: []byte{testHeader.threads},
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}

func Test_writeLockerData_ErrorVersion(t *testing.T) {
	var expectedError = errors.New("expected")
	var testHeader = &lockerHeader{
		version: 37,
	}
	var testWriter = &stubWriter{
		MatchError: []byte{testHeader.version},
		WriteError: expectedError,
	}

	assert.ErrorIs(t, writeLockerData(testHeader, testWriter, []byte("not encrypted obviously")), expectedError)
}
