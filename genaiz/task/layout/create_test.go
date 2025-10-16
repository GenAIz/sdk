package layout

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

func TestNewCreateTask(t *testing.T) {
	var testTask = NewCreateTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleLayoutCreate(t *testing.T) {
	var testDir = t.TempDir()
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigName:   "notValid/name",
			ConfigFolder: filepath.Join(testDir, ".create_layout_test"),
		},
	}
	var testState = &task.State{Logger: logrus.New()}

	t.Chdir(testDir)
	assert.Error(t, handleLayoutCreate(testParams, testState))
}

func Test_handleLayoutCreate_ConfigTypeNone(t *testing.T) {
	if testDir, err := filepath.EvalSymlinks(t.TempDir()); err == nil {
		var testParams = &CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigFolder: filepath.Join(testDir, ".create_layout_test"),
			},
		}
		var testState = &task.State{Logger: logrus.New()}
		var expected string

		t.Chdir(testDir)
		assert.NoError(t, handleLayoutCreate(testParams, testState))
		expected, err = os.Getwd()
		assert.NoError(t, err)
		assert.Equal(t, expected, testParams.ConfigFolder)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutCreate_InvalidPath(t *testing.T) {
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: "/_not_allowed",
		},
	}
	var testState = &task.State{Logger: logrus.New()}

	assert.Error(t, handleLayoutCreate(testParams, testState))
}

func Test_handleLayoutCreate_StateOutput(t *testing.T) {
	var testDir = t.TempDir()
	var expectedOutput = filepath.Join(testDir, ".create_layout_test", "test.json")
	var testParams = &CreateParams{}
	var testState = &task.State{Logger: logrus.New(), Output: expectedOutput}

	t.Chdir(testDir)
	assert.NoError(t, handleLayoutCreate(testParams, testState))
}

func Test_handleLayoutCreateContext(t *testing.T) {
	var testDir = t.TempDir()
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   "name",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: testDir,
		},
	}

	assert.NoError(t, handleLayoutCreateContext(testParams, testState))
	assert.Equal(t, filepath.Join(testDir, "name.json"), testState.Output)
}

func Test_handleLayoutCreateContext_ConfigTypeNone(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{}

	assert.NoError(t, handleLayoutCreateContext(testParams, testState))
}

func Test_handleLayoutCreateContext_ContextAlreadyExists(t *testing.T) {
	var testDir = t.TempDir()
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   "name",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: filepath.Join(testDir, ".layout_create_test"),
		},
	}

	if fd, err := filez.CreateRecursive(testParams.ConfigFolder, testParams.ConfigName+"."+shared.ConfigTypeJson); err == nil {
		defer filez.CloseSilently(fd)

		assert.Error(t, handleLayoutCreateContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutCreateFile_ConfigTypeNone(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{}
	var fd, err = handleLayoutCreateFile("/_not_existing", testParams, testState)

	assert.Error(t, err)
	assert.Empty(t, fd)
}

func Test_handleLayoutCreatePretend(t *testing.T) {
	var expectedPath = "folderPath"
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: expectedPath,
		},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	testState.Output = "outputFile"
	assert.NoError(t, handleLayoutCreatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedPath)
	assert.Contains(t, output, testState.Output)
}

func Test_handleLayoutCreatePretend_NoStateOutput(t *testing.T) {
	var expectedPath = "folderPath"
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: expectedPath,
		},
	}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleLayoutCreatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedPath)
	assert.NotContains(t, output, "touch")
}
