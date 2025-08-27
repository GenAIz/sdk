package layout

import (
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
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
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigName:   "notValid/name",
			ConfigFolder: "/tmp/.create_layout_test",
		},
	}
	var testState = &task.State{Logger: logrus.New()}
	var cwd, err = os.Getwd()

	panicz.PanicIfError(err)
	assert.Error(t, handleLayoutCreate(testParams, testState))
	defer func() { _ = os.Chdir(cwd) }()
	defer filez.RemoveSilently(testParams.ConfigFolder)
}

func Test_handleLayoutCreate_ConfigTypeNone(t *testing.T) {
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigFolder: "/tmp/.create_layout_test",
		},
	}
	var testState = &task.State{Logger: logrus.New()}
	var back, err = os.Getwd()
	var expected string

	panicz.PanicIfError(err)
	assert.NoError(t, handleLayoutCreate(testParams, testState))
	expected, err = os.Getwd()
	assert.NoError(t, err)
	assert.Equal(t, expected, testParams.ConfigFolder)
	defer func() { _ = os.Chdir(back) }()
	defer filez.RemoveSilently(testParams.ConfigFolder)
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
	var expectedOutput = "/tmp/.create_layout_test/test.json"
	var testParams = &CreateParams{}
	var testState = &task.State{Logger: logrus.New(), Output: expectedOutput}
	var cwd, err = os.Getwd()

	panicz.PanicIfError(err)
	assert.NoError(t, handleLayoutCreate(testParams, testState))
	defer func() { _ = os.Chdir(cwd) }()
	defer filez.RemoveSilently(filepath.Dir(expectedOutput))
}

func Test_handleLayoutCreateContext(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   "name",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: "/tmp",
		},
	}

	assert.NoError(t, handleLayoutCreateContext(testParams, testState))
	assert.Equal(t, "/tmp/name.json", testState.Output)
}

func Test_handleLayoutCreateContext_ConfigTypeNone(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{}

	assert.NoError(t, handleLayoutCreateContext(testParams, testState))
}

func Test_handleLayoutCreateContext_ContextAlreadyExists(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   "name",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: "/tmp/.layout_create_test",
		},
	}

	if _, err := filez.CreateRecursive(testParams.ConfigFolder, testParams.ConfigName+"."+shared.ConfigTypeJson); err == nil {
		defer filez.RemoveSilently(testParams.ConfigFolder)

		assert.Error(t, handleLayoutCreateContext(testParams, testState))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutCreateContext_ContextNotWriteable(t *testing.T) {
	var testState = &task.State{Logger: logrus.New()}
	var testParams = &CreateParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   "name",
			ConfigType:   lang.Ref(shared.ConfigTypeJson),
			ConfigFolder: "/_not_writeable_",
		},
	}

	assert.Error(t, handleLayoutCreateContext(testParams, testState))
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
