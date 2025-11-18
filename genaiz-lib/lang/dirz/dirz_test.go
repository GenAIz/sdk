package dirz

import (
	"errors"
	"os"
	"path/filepath"
	"strings"
	"testing"

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
	var testDir = t.TempDir()
	var testCall = func() error {
		return errors.New("expected")
	}

	assert.NoError(t, DoIfPathExist(filepath.Join(testDir, "_not_exist"), testCall))
}

func TestAnchorWorkingFile_Local(t *testing.T) {
	var expected = "testFile"

	assert.Equal(t, expected, AnchorWorkingFile(expected))
}

func TestAnchorWorkingFile_Relative(t *testing.T) {
	var testDir = t.TempDir()
	var expected = "testFile"

	t.Chdir(testDir)
	assert.Equal(t, expected, AnchorWorkingFile("../"+expected))
	assert.Equal(t, expected, AnchorWorkingFile("./"+expected))
}

func TestAnchorWorkingFile_Suffix(t *testing.T) {
	var testDir = t.TempDir()
	var expected = "testFile"

	t.Chdir(testDir)
	assert.Equal(t, expected, AnchorWorkingFile(filepath.Join(testDir, expected)))
}

func TestChangeWorkingDir_CurrentDir(t *testing.T) {
	var testDir = t.TempDir()
	var actual string
	var reset func()
	var err error

	t.Chdir(testDir)
	reset, err = ChangeWorkingDir()
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, testDir, actual)
	reset, err = ChangeWorkingDir(".")
	assert.NoError(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, testDir, actual)
}

func TestChangeWorkingDir_NotReadable(t *testing.T) {
	var testDir = t.TempDir()
	var actual string
	var reset func()
	var err error

	t.Chdir(testDir)
	reset, err = ChangeWorkingDir(filepath.Join(testDir, "_not_readable"))
	assert.Error(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, testDir, actual)
}

func TestChangeWorkingDir(t *testing.T) {
	if expected, err := filepath.EvalSymlinks(t.TempDir()); err == nil {
		var cwd, _ = os.Getwd()
		var actual string
		var reset func()

		reset, err = ChangeWorkingDir(expected)
		assert.NoError(t, err)
		actual, _ = os.Getwd()
		assert.Contains(t, actual, expected)
		reset()
		actual, _ = os.Getwd()
		assert.Equal(t, cwd, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestCreateWorkingDir_CurrentDir(t *testing.T) {
	if testDir, err := filepath.EvalSymlinks(t.TempDir()); err == nil {
		var actual string
		var reset func()

		t.Chdir(testDir)
		reset, err = CreateWorkingDir()
		assert.NoError(t, err)
		reset()
		actual, _ = os.Getwd()
		assert.Equal(t, testDir, actual)
		reset, err = CreateWorkingDir(".")
		assert.NoError(t, err)
		reset()
		actual, _ = os.Getwd()
		assert.Equal(t, testDir, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestCreateWorkingDir_NotWritable(t *testing.T) {
	var testDir = t.TempDir()
	var actual string
	var reset func()
	var err error

	t.Chdir(testDir)
	reset, err = CreateWorkingDir("/opt/_genaiz_not_writable")
	assert.Error(t, err)
	reset()
	actual, _ = os.Getwd()
	assert.Equal(t, testDir, actual)
}

func TestFirstParentName(t *testing.T) {
	var expectedParent = "parent"

	assert.Equal(t, expectedParent, FirstParentName("parent/child/friends"))
	assert.Equal(t, expectedParent, FirstParentName("/parent/friends"))
}

func TestCreateWorkingDir(t *testing.T) {
	if expected, err := filepath.EvalSymlinks(t.TempDir()); err == nil {
		var cwd, _ = os.Getwd()
		var actual string
		var reset func()

		reset, err = CreateWorkingDir(expected)
		assert.NoError(t, err)
		actual, _ = os.Getwd()
		assert.Equal(t, expected, actual)
		reset()
		actual, _ = os.Getwd()
		assert.Equal(t, cwd, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestOptionalWorkingDir_NoArgs(t *testing.T) {
	var testDir = t.TempDir()
	var cwd string
	var err error

	t.Chdir(testDir)

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

func TestWorkingDirBase(t *testing.T) {
	var testDir = t.TempDir()

	t.Chdir(testDir)
	assert.True(t, strings.HasSuffix(testDir, WorkingDirBase()))
}

func TestWorkingDirParent(t *testing.T) {
	var testDir = t.TempDir()

	t.Chdir(testDir)
	assert.Equal(t, filepath.Dir(testDir), WorkingDirParent())
}
