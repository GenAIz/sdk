package sc

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
)

func TestPrintExecutor_PrintSchema(t *testing.T) {
	var testExecutor = &PrintExecutor{}
	var stdoutRestore = os.Stdout

	r, w, _ := os.Pipe()
	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, testExecutor.PrintSchema())

	_ = w.Close()
	out, _ := io.ReadAll(r)
	assert.NotEmpty(t, out)
}

func TestPrintExecutor_PrintSchema_File(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "testOut")
	var testExecutor = &PrintExecutor{
		outputFile: testFile,
	}

	assert.NoError(t, testExecutor.PrintSchema())

	if b, err := os.ReadFile(testFile); err == nil {
		assert.NotEmpty(t, b)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewPrint_BadPath(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "badPath", "testOut")
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testCmd = NewPrint()

	defer patch.Unpatch()
	testCmd.SetArgs([]string{testFile})
	assert.NoError(t, testCmd.Execute())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}
