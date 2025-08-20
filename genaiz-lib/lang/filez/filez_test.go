package filez

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCloseSilently(t *testing.T) {
	if fd, err := os.CreateTemp("/tmp", "genaiz-filez"); err == nil {
		assert.NotPanics(t, func() { CloseSilently(fd) })
		assert.NoError(t, os.Remove(fd.Name()))
	} else {
		assert.Fail(t, "failed with error", err)
	}

	assert.NotPanics(t, func() { CloseSilently(nil) })
}

func TestCreateRecursive_Err(t *testing.T) {
	var expectedDir = "///.genaiz"

	_, err := CreateRecursive(expectedDir, "FilezTest.txt")
	assert.Error(t, err)
}

func TestCreateRecursive(t *testing.T) {
	var expectedDir = "/tmp/.genaiz"
	var dirHandle, err = CreateRecursive(expectedDir, "FilezTest.txt")

	assert.NoError(t, err)
	assert.NotEmpty(t, dirHandle.Name())
	RemoveSilently(expectedDir)
}

func TestCreateRecursiveTemp(t *testing.T) {
	var expectedDir = "/tmp/.genaiz"
	var dirHandle, err = CreateRecursiveTemp(expectedDir, "FilezTest")

	assert.NoError(t, err)
	assert.NotEmpty(t, dirHandle.Name())
	RemoveSilently(expectedDir)
}

func TestFirstNamedFile(t *testing.T) {
	var testName = "name"
	var expectedContent = "valid"
	var back, _ = os.Getwd()
	var dir, _ = os.MkdirTemp("/tmp", "genait")
	var testPath string
	var testBytes []byte
	var err error

	defer RemoveSilently(dir)
	defer func() { _ = os.Chdir(back) }()

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
}

func TestFirstNamedFileUnder_InvalidDir(t *testing.T) {
	var _, err = FirstNamedFileUnder("/invalid", "name")

	assert.Error(t, err)
}

func TestFirstNamedFileUnder_NotFound(t *testing.T) {
	var testName = "name"
	var back, _ = os.Getwd()
	var dir, _ = os.MkdirTemp("/tmp", "genait")
	var err error

	defer RemoveSilently(dir)
	defer func() { _ = os.Chdir(back) }()

	assert.NoError(t, os.Chdir(dir))
	assert.NoError(t, os.Mkdir("testDir", 0750))

	_, err = FirstNamedFile(testName)
	assert.Error(t, err)
	assert.Equal(t, "not found", err.Error())
}

func TestFromWorkDir(t *testing.T) {
	var testPath, _ = filepath.Abs("test")

	assert.EqualValues(t, "./test", FromWorkDir(testPath))
}

func TestFromWorkDir_NotParent(t *testing.T) {
	assert.EqualValues(t, "/tmp", FromWorkDir("/tmp"))
}

func TestGetFileType(t *testing.T) {
	var expectedType = "txt"

	assert.Empty(t, GetFileType("without_any_extension"))
	assert.EqualValues(t, expectedType, GetFileType("test.txt"))
	assert.EqualValues(t, expectedType, GetFileType("folder/.hiddenFolder/test.txt"))
}

func TestIsReadable(t *testing.T) {
	var dir string
	var fd *os.File
	var err error

	dir, err = os.MkdirTemp("/tmp", "genait")
	assert.NoError(t, err)
	fd, err = os.CreateTemp(dir, "testIsReadable")
	assert.NoError(t, err)
	defer RemoveSilently(dir)

	assert.NoError(t, IsReadable(fd.Name()))
	assert.Error(t, IsReadable(fd.Name()+".notexist"))
}
