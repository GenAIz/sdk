package layout

import (
	"errors"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type stubWriter struct {
	arches     []string
	configFile string
	dest       string
	handle     string
	input      string
	name       string
	oem        string
	output     map[string]string
	sfType     string
	version    string
	writeErr   error
}

func (s *stubWriter) Write(dest string) error {
	s.dest = dest
	return s.writeErr
}

func (s *stubWriter) BuildArches() (string, []string) {
	return "", s.arches
}

func (s *stubWriter) BuildHandle() (string, string) {
	return "", s.handle
}

func (s *stubWriter) BuildInput() (string, string) {
	return "", s.input
}

func (s *stubWriter) BuildName() (string, string) {
	return "", s.name
}

func (s *stubWriter) BuildOem() (string, string) {
	return "", s.oem
}

func (s *stubWriter) BuildOutput() map[string]string {
	return s.output
}

func (s *stubWriter) BuildType() (string, string) {
	return "", s.sfType
}

func (s *stubWriter) BuildVersion() (string, string) {
	return "", s.version
}

func (s *stubWriter) WithConfigFile(configFile string) ConfigWriter {
	s.configFile = configFile
	return s
}

func (s *stubWriter) WithArches(arches []string) ConfigWriter {
	s.arches = arches
	return s
}

func (s *stubWriter) WithHandle(handle string) ConfigWriter {
	s.handle = handle
	return s
}

func (s *stubWriter) WithInput(input string) ConfigWriter {
	s.input = input
	return s
}

func (s *stubWriter) WithName(name string) ConfigWriter {
	s.name = name
	return s
}

func (s *stubWriter) WithOem(oem string) ConfigWriter {
	s.oem = oem
	return s
}

func (s *stubWriter) WithOutput(output string) ConfigWriter {
	if s.output == nil {
		s.output = make(map[string]string)
	}

	s.output[output] = output
	return s
}

func (s *stubWriter) WithType(sfType string) ConfigWriter {
	s.sfType = sfType
	return s
}

func (s *stubWriter) WithVersion(version string) ConfigWriter {
	s.version = version
	return s
}

func TestNewInitTask(t *testing.T) {
	var testTask = NewInitTask(nil)

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnIncomplete)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleLayoutInitContext_initialized(t *testing.T) {
	assert.NoError(t, handleLayoutInitContext(&InitParams{}, &task.State{Output: "test.yaml"}))
}

func Test_handleLayoutInitContext_needsConfig(t *testing.T) {
	var actual = &task.State{Logger: logrus.New()}
	var testParams = &InitParams{
		CreateParams: CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigType: lang.Ref(shared.ConfigTypeJson),
				ConfigName: "test",
			},
		},
	}

	assert.NoError(t, handleLayoutInitContext(testParams, actual))
	assert.Equal(t, "test.json", actual.Output)
}

func Test_handleLayoutInitContext_noConfig(t *testing.T) {
	var actual = &task.State{Logger: logrus.New()}
	var testParams = &InitParams{
		CreateParams: CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: "test",
			},
		},
	}

	assert.ErrorIs(t, errorNeedsConfigFile, handleLayoutInitContext(testParams, actual))
	assert.Equal(t, "", actual.Output)
}

func Test_handleLayoutInitContext(t *testing.T) {
	var expectedDir = "/tmp/genaiz.init"
	var expectedFile = "test.json"
	var reset func()
	var err error

	if reset, err = dirz.CreateWorkingDir(expectedDir); err == nil {
		defer reset()
		defer filez.RemoveSilently(expectedDir)
		var actual = &task.State{Logger: logrus.New()}
		var testParams = &InitParams{
			CreateParams: CreateParams{
				ConfigParams: shared.ConfigParams{
					ConfigName: "test",
				},
			},
		}

		if _, err = os.Create(filepath.Join(expectedDir, expectedFile)); err == nil {
			assert.ErrorIs(t, errorNeedsConfigFile, handleLayoutInitContext(testParams, actual))
			assert.Equal(t, expectedFile, actual.Output)
		}
	}

	assert.NoError(t, err)
}

func Test_handleLayoutInitCreate_noConfig(t *testing.T) {
	assert.ErrorIs(t, errorNoConfigFile, handleLayoutInitCreate(nil, &InitParams{}, &task.State{}))
}

func Test_handleLayoutInitCreate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
	}
	var testParams = &InitParams{
		Arches:      []ArchType{ArchTypeX86},
		Handle:      "handle",
		Name:        "name",
		Type:        "type",
		MountInput:  "input",
		MountOutput: "output",
		OEM:         "oem",
		Version:     "version",
	}
	var testWriter = &stubWriter{}

	assert.NoError(t, handleLayoutInitCreate(testWriter, testParams, testState))
	assert.Equal(t, testParams.Arches, testWriter.arches)
	assert.Equal(t, testParams.Handle, testWriter.handle)
	assert.Equal(t, testParams.Name, testWriter.name)
	assert.Equal(t, testParams.Type, testWriter.sfType)
	assert.Equal(t, testParams.MountInput, testWriter.input)
	assert.Equal(t, testParams.MountOutput, testWriter.output[testParams.MountOutput])
	assert.Equal(t, testParams.OEM, testWriter.oem)
	assert.Equal(t, testParams.Version, testWriter.version)
	assert.Equal(t, testState.Output, testWriter.dest)
}

func Test_handleLayoutInitPretend_noConfig(t *testing.T) {
	assert.ErrorIs(t, errorNoConfigFile, handleLayoutInitPretend(nil, &InitParams{}, &task.State{}))
}

func Test_handleLayoutInitPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{
		Arches:      []ArchType{ArchTypeX86},
		Handle:      "handle",
		Name:        "name",
		Type:        "type",
		MountInput:  "input",
		MountOutput: "output",
		OEM:         "oem",
		Version:     "version",
	}
	var testWriter = &stubWriter{}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleLayoutInitPretend(testWriter, testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.Arches[0])
	assert.Contains(t, output, testParams.Handle)
	assert.Contains(t, output, testParams.Name)
	assert.Contains(t, output, testParams.Type)
	assert.Contains(t, output, testParams.MountInput)
	assert.Contains(t, output, testParams.MountOutput)
	assert.Contains(t, output, testParams.OEM)
	assert.Contains(t, output, testParams.Version)
}

func Test_handleLayoutInitUpdate_createError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
	}
	var testWriter = &stubWriter{writeErr: expectedError}

	assert.ErrorIs(t, handleLayoutInitUpdate(testWriter, &InitParams{}, testState), expectedError)
}

func Test_handleLayoutInitUpdate_noConfig(t *testing.T) {
	assert.NoError(t, handleLayoutInitUpdate(nil, &InitParams{}, &task.State{}))
}

func Test_handleLayoutInitUpdate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
	}
	var testParams = &InitParams{
		Arches:      []ArchType{ArchTypeX86},
		Handle:      "handle",
		Name:        "name",
		Type:        "type",
		MountInput:  "input",
		MountOutput: "output",
		OEM:         "oem",
		Version:     "version",
	}
	var testWriter = &stubWriter{}

	assert.NoError(t, handleLayoutInitUpdate(testWriter, testParams, testState))
	assert.Equal(t, testParams.Arches, testWriter.arches)
	assert.Equal(t, testParams.Handle, testWriter.handle)
	assert.Equal(t, testParams.Name, testWriter.name)
	assert.Equal(t, testParams.Type, testWriter.sfType)
	assert.Equal(t, testParams.MountInput, testWriter.input)
	assert.Equal(t, testParams.MountOutput, testWriter.output[testParams.MountOutput])
	assert.Equal(t, testParams.OEM, testWriter.oem)
	assert.Equal(t, testParams.Version, testWriter.version)
	assert.Empty(t, testState.Output)
}
