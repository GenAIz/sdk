package dirz

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"

	"github.com/stretchr/testify/assert"
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

func TestAnchorWorkingFile_Local(t *testing.T) {
	var expected = "testFile"

	assert.Equal(t, expected, AnchorWorkingFile(expected))
}

func TestAnchorWorkingFile_Relative(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expected = "testFile"
	var testPath = "/tmp/_genaiz_anchoring"

	assert.NoError(t, os.MkdirAll(testPath, 0750))
	assert.NoError(t, os.Chdir(testPath))

	defer func() {
		panicz.PanicIfError(os.Chdir(cwd))
		filez.RemoveSilently(testPath)
	}()

	assert.Equal(t, expected, AnchorWorkingFile("../"+expected))
	assert.Equal(t, expected, AnchorWorkingFile("./"+expected))
}

func TestAnchorWorkingFile_Suffix(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expected = "testFile"
	var testPath = "/tmp/_genaiz_anchoring"

	assert.NoError(t, os.MkdirAll(testPath, 0750))
	assert.NoError(t, os.Chdir(testPath))

	defer func() {
		panicz.PanicIfError(os.Chdir(cwd))
		filez.RemoveSilently(testPath)
	}()

	assert.Equal(t, expected, AnchorWorkingFile(filepath.Join(testPath, expected)))
}

func TestChangeWorkingDir_CurrentDir(t *testing.T) {
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	reset, err = ChangeWorkingDir()
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
	reset, err = ChangeWorkingDir(".")
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestChangeWorkingDir_NotReadable(t *testing.T) {
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	reset, err = ChangeWorkingDir("/opt/_genaiz_not_readable")
	assert.Error(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestChangeWorkingDir(t *testing.T) {
	var expected = "/tmp/_genaiz_readable/test"
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	assert.NoError(t, os.MkdirAll(expected, 0750))
	defer filez.RemoveSilently("/tmp/_genaiz_readable")
	reset, err = ChangeWorkingDir(expected)
	assert.NoError(t, err)
	actual, _ = os.Getwd()
	assert.Equal(t, expected, actual)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestCreateWorkingDir_CurrentDir(t *testing.T) {
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	reset, err = CreateWorkingDir()
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
	reset, err = CreateWorkingDir(".")
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestCreateWorkingDir_NotWritable(t *testing.T) {
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	reset, err = CreateWorkingDir("/opt/_genaiz_not_writable")
	assert.Error(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestCreateWorkingDir(t *testing.T) {
	var expected = "/tmp/_genaiz_writeable/test"
	var cwd, _ = os.Getwd()
	var actual string
	var reset func()
	var err error

	reset, err = CreateWorkingDir(expected)
	defer filez.RemoveSilently("/tmp/_genaiz_writeable")
	assert.NoError(t, err)
	actual, _ = os.Getwd()
	assert.Equal(t, expected, actual)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, cwd, actual)
}

func TestOptionalWorkingDir_NoArgs(t *testing.T) {
	var cwd string
	var err error

	if cwd, err = os.Getwd(); err == nil {
		var fn = OptionalWorkingDir()
		var actual string

		if actual, err = fn(); err == nil {
			assert.Equal(t, cwd, actual)
		}
	}

	assert.NoError(t, err)
}

func TestOptionalWorkingDir(t *testing.T) {
	var fn = OptionalWorkingDir("path", "test")
	var actual string
	var err error

	if actual, err = fn(); err == nil {
		assert.True(t, strings.HasSuffix(actual, "path/test"))
	}

	assert.NoError(t, err)
}
