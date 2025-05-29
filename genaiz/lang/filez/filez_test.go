package filez

import (
	"errors"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/lang/panicz"
)

func TestDoIfPathExist(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expectedError = errors.New("expected")
	var testCall = func() error {
		return expectedError
	}

	assert.EqualValues(t, expectedError, DoIfPathExist(cwd, testCall))
}

func TestDoIfPathExist_NotExist(t *testing.T) {
	var testCall = func() error {
		return errors.New("expected")
	}

	assert.NoError(t, DoIfPathExist("/_not_exist", testCall))
}

func TestFirstNamedFile(t *testing.T) {
	var testName = "name"
	var expectedContent = "valid"
	var back, _ = os.Getwd()
	var dir, _ = os.MkdirTemp("/tmp", "genait")
	var testPath string
	var testBytes []byte
	var err error

	assert.NoError(t, os.Chdir(dir))
	assert.NoError(t, os.Mkdir("testDir", 0750))
	assert.NoError(t, os.WriteFile("testFile", []byte("invalid"), 0640))
	assert.NoError(t, os.WriteFile(testName+".txt", []byte(expectedContent), 0640))

	testPath, err = FirstNamedFile(testName)
	assert.NoError(t, err)

	if testBytes, err = os.ReadFile(testPath); err == nil {
		assert.EqualValues(t, expectedContent, string(testBytes))
	} else {
		assert.Fail(t, "could not read test file")
	}

	panicz.PanicIfError(os.RemoveAll(dir))
	panicz.PanicIfError(os.Chdir(back))
}

func TestFirstNamedFileUnderInvalidDir(t *testing.T) {
	var _, err = FirstNamedFileUnder("/invalid", "name")

	assert.Error(t, err)
}

func TestFromWorkDir(t *testing.T) {
	var testPath, _ = filepath.Abs("test")

	assert.EqualValues(t, "./test", FromWorkDir(testPath))
}

func TestFromWorkDir_NotParent(t *testing.T) {
	assert.EqualValues(t, "/tmp", FromWorkDir("/tmp"))
}
