package layout

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type stubWriter struct {
	arches               []string
	configFile           string
	dest                 string
	handle               string
	input                string
	inputPorts           []broker.DataPort
	inputPortRemoved     *broker.DataPort
	name                 string
	oem                  string
	outboundProxies      []broker.Proxy
	outboundProxyRemoved *broker.Proxy
	output               map[string]string
	outputPorts          []broker.DataPort
	outputPortRemoved    *broker.DataPort
	propSpecs            []broker.PropSpec
	rmPropSpec           *broker.PropSpec
	sfType               string
	sources              []string
	stores               []string
	version              string
	writeErr             error
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

func (s *stubWriter) BuildInputPorts() (string, []broker.DataPort) {
	return "", s.inputPorts
}

func (s *stubWriter) BuildInputPortRemoved() (string, *broker.DataPort) {
	return "", s.inputPortRemoved
}

func (s *stubWriter) BuildName() (string, string) {
	return "", s.name
}

func (s *stubWriter) BuildOem() (string, string) {
	return "", s.oem
}

func (s *stubWriter) BuildOutboundProxies() (string, []broker.Proxy) {
	return "", s.outboundProxies
}

func (s *stubWriter) BuildOutboundProxyRemoved() (string, *broker.Proxy) {
	return "", s.outboundProxyRemoved
}

func (s *stubWriter) BuildOutput() map[string]string {
	return s.output
}

func (s *stubWriter) BuildOutputPorts() (string, []broker.DataPort) {
	return "", s.outputPorts
}

func (s *stubWriter) BuildOutputPortRemoved() (string, *broker.DataPort) {
	return "", s.outputPortRemoved
}

func (s *stubWriter) BuildPropSpecs() (string, []broker.PropSpec) {
	return "", s.propSpecs
}

func (s *stubWriter) BuildPropSpecRemoved() (string, *broker.PropSpec) {
	return "", s.rmPropSpec
}

func (s *stubWriter) BuildSources() (string, []string) {
	return "", s.sources
}

func (s *stubWriter) BuildStores() (string, []string) {
	return "", s.stores
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

func (s *stubWriter) WithInputPorts(inputPorts []broker.DataPort) ConfigWriter {
	s.inputPorts = inputPorts
	return s
}

func (s *stubWriter) WithInputPortRemoved(inputPort *broker.DataPort) ConfigWriter {
	s.inputPortRemoved = inputPort
	return s
}

func (s *stubWriter) WithLog(logDir string) ConfigWriter {
	if s.output == nil {
		s.output = make(map[string]string)
	}

	s.output["log"] = logDir
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

func (s *stubWriter) WithOutboundProxies(proxies []broker.Proxy) ConfigWriter {
	s.outboundProxies = proxies
	return s
}

func (s *stubWriter) WithOutboundProxyRemoved(proxy *broker.Proxy) ConfigWriter {
	s.outboundProxyRemoved = proxy
	return s
}

func (s *stubWriter) WithOutput(output string) ConfigWriter {
	if s.output == nil {
		s.output = make(map[string]string)
	}

	s.output["output"] = output
	return s
}

func (s *stubWriter) WithOutputPorts(outputPorts []broker.DataPort) ConfigWriter {
	s.outputPorts = outputPorts
	return s
}

func (s *stubWriter) WithOutputPortRemoved(outputPort *broker.DataPort) ConfigWriter {
	s.outputPortRemoved = outputPort
	return s
}

func (s *stubWriter) WithPropSpecs(specs []broker.PropSpec) ConfigWriter {
	s.propSpecs = specs
	return s
}

func (s *stubWriter) WithPropSpecRemoved(spec *broker.PropSpec) ConfigWriter {
	s.rmPropSpec = spec
	return s
}

func (s *stubWriter) WithSources(sources []string) ConfigWriter {
	s.sources = sources
	return s
}

func (s *stubWriter) WithStores(stores []string) ConfigWriter {
	s.stores = stores
	return s
}

func (s *stubWriter) WithType(sfType string) ConfigWriter {
	s.sfType = sfType
	return s
}

func (s *stubWriter) WithVar(varDir string) ConfigWriter {
	if s.output == nil {
		s.output = make(map[string]string)
	}

	s.output["var"] = varDir
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

func Test_handleLayoutInitContext(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		CreateParams: CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigFolder: t.TempDir(),
				ConfigName:   "test",
				ConfigType:   lang.Ref(shared.ConfigTypeYaml),
			},
		},
	}

	assert.NoError(t, handleLayoutInitContext(testParams, testState))
	assert.Equal(t, testParams.GetConfigPath(), testState.Output)
}

func Test_handleLayoutInitContext_dirError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		CreateParams: CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigFolder: t.TempDir(),
				ConfigName:   "test",
				ConfigType:   lang.Ref(shared.ConfigTypeYaml),
			},
		},
	}
	var expectedPath = testParams.GetConfigPath()
	var err error

	if err = os.MkdirAll(expectedPath, 0750); err == nil {
		assert.ErrorIs(t, handleLayoutInitContext(testParams, testState), shared.ErrorConfigFileInvalid)
		assert.Empty(t, testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutInitContext_existsError(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &InitParams{
		CreateParams: CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigFolder: t.TempDir(),
				ConfigName:   "test",
				ConfigType:   lang.Ref(shared.ConfigTypeYaml),
			},
		},
	}
	var expectedPath = testParams.GetConfigPath()
	var err error

	if _, err = os.Create(expectedPath); err == nil {
		assert.ErrorIs(t, handleLayoutInitContext(testParams, testState), shared.ErrorConfigFileExists)
		assert.Equal(t, expectedPath, testState.Output)
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_handleLayoutInitContext_knownOutput(t *testing.T) {
	assert.NoError(t, handleLayoutInitContext(&InitParams{}, &task.State{Output: "output"}))
}

func Test_handleLayoutInitCreate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
	}
	var testParams = &InitParams{
		Arches:      []shared.ArchType{shared.ArchTypeX86},
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

func Test_handleLayoutInitCreate_noConfig(t *testing.T) {
	assert.ErrorIs(t, errorNoConfigFile, handleLayoutInitCreate(nil, &InitParams{}, &task.State{}))
}

func Test_handleLayoutInitPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{
		Arches:      []shared.ArchType{shared.ArchTypeX86},
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

func Test_handleLayoutInitPretend_dataSources(t *testing.T) {
	var expectedDataSources = []string{"expectedDataSource", "expectedDataSource2"}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{DataSources: expectedDataSources}
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

	assert.Contains(t, output, expectedDataSources[0])
	assert.Contains(t, output, expectedDataSources[1])
}

func Test_handleLayoutInitPretend_dataStores(t *testing.T) {
	var expectedDataStores = []string{"expectedDataStore", "expectedDataStore2"}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{DataStores: expectedDataStores}
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

	assert.Contains(t, output, expectedDataStores[0])
	assert.Contains(t, output, expectedDataStores[1])
}

func Test_handleLayoutInitPretend_inputPorts(t *testing.T) {
	var expectedInputPort = broker.DataPort{
		Handle:      "expectedHandle",
		Description: "expectedDescription",
		Name:        "expectedName",
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{InputPorts: []broker.DataPort{expectedInputPort}}
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

	assert.Contains(t, output, expectedInputPort.Handle)
	assert.Contains(t, output, expectedInputPort.Description)
	assert.Contains(t, output, expectedInputPort.Name)
}

func Test_handleLayoutInitPretend_inputPortsRemoval(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var expectedRemovePort = &broker.DataPort{Handle: "expectedRemoveHandle"}
	var testParams = &InitParams{}
	var testWriter = &stubWriter{inputPortRemoved: expectedRemovePort}
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

	assert.Contains(t, output, expectedRemovePort.Handle)
}

func Test_handleLayoutInitPretend_noConfig(t *testing.T) {
	assert.ErrorIs(t, errorNoConfigFile, handleLayoutInitPretend(nil, &InitParams{}, &task.State{}))
}

func Test_handleLayoutInitPretend_outboundProxies(t *testing.T) {
	var expectedProxy = broker.Proxy{
		Host: "expectedHost",
		Port: 37,
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{OutboundProxies: []broker.Proxy{expectedProxy}}
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

	assert.Contains(t, output, expectedProxy.Host)
	assert.Contains(t, output, cast.ToString(expectedProxy.Port))
}

func Test_handleLayoutInitPretend_outboundProxiesRemoval(t *testing.T) {
	var expectedOutboundProxy = &broker.Proxy{Host: "expectedHost", Port: 37}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testWriter = &stubWriter{outboundProxyRemoved: expectedOutboundProxy}
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	assert.NoError(t, handleLayoutInitPretend(testWriter, &InitParams{}, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, expectedOutboundProxy.Host)
	assert.Contains(t, output, cast.ToString(expectedOutboundProxy.Port))
}

func Test_handleLayoutInitPretend_outputPorts(t *testing.T) {
	var expectedOutputPort = broker.DataPort{
		Handle:      "expectedHandle",
		Description: "expectedDescription",
		Name:        "expectedName",
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{OutputPorts: []broker.DataPort{expectedOutputPort}}
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

	assert.Contains(t, output, expectedOutputPort.Handle)
	assert.Contains(t, output, expectedOutputPort.Description)
	assert.Contains(t, output, expectedOutputPort.Name)
}

func Test_handleLayoutInitPretend_outputPortRemoval(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var expectedRemovePort = &broker.DataPort{Handle: "expectedRemoveHandle"}
	var testParams = &InitParams{}
	var testWriter = &stubWriter{outputPortRemoved: expectedRemovePort}
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

	assert.Contains(t, output, expectedRemovePort.Handle)
}

func Test_handleLayoutInitPretend_propSpecs(t *testing.T) {
	var expectedPropSpec = broker.PropSpec{
		Key:         "expectedKey",
		Description: "expectedDescription",
		Name:        "expectedName",
		Type:        broker.PropSpecTypeString,
		Value:       "expectedDefaultValue",
	}
	var expectedEnumPropSpec = broker.PropSpec{
		Key:    "expectedEnumKey",
		Name:   "expectedEnumName",
		Type:   broker.PropSpecTypeEnum,
		Values: []string{"value1"},
	}
	var expectedNoValueSpec = broker.PropSpec{
		Key:  "expectedNoValueKey",
		Name: "expectedNoValueName",
		Type: broker.PropSpecTypeBoolean,
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var testParams = &InitParams{
		PropSpecs: []broker.PropSpec{
			expectedPropSpec, expectedEnumPropSpec, expectedNoValueSpec,
		},
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

	assert.Contains(t, output, expectedPropSpec.Key)
	assert.Contains(t, output, expectedPropSpec.Name)
	assert.Contains(t, output, expectedPropSpec.Type)
	assert.Contains(t, output, expectedPropSpec.Value)
	assert.Contains(t, output, expectedEnumPropSpec.Key)
	assert.Contains(t, output, expectedEnumPropSpec.Name)
	assert.Contains(t, output, expectedEnumPropSpec.Type)
	assert.Contains(t, output, expectedEnumPropSpec.Values[0])
	assert.Contains(t, output, expectedNoValueSpec.Key)
	assert.Contains(t, output, expectedNoValueSpec.Name)
	assert.Contains(t, output, expectedNoValueSpec.Type)
}

func Test_handleLayoutInitPretend_propSpecsRemoval(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.yaml",
	}
	var expectedRemoveSpec = &broker.PropSpec{Key: "expectedRemoveKey"}
	var testParams = &InitParams{}
	var testWriter = &stubWriter{rmPropSpec: expectedRemoveSpec}
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

	assert.Contains(t, output, expectedRemoveSpec.Key)
}

func Test_handleLayoutInitUpdate(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
		Error:  shared.ErrorConfigFileExists,
	}
	var testParams = &InitParams{
		Arches:      []shared.ArchType{shared.ArchTypeX86},
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

func Test_handleLayoutInitUpdate_createError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "test.json",
		Error:  shared.ErrorConfigFileExists,
	}
	var testWriter = &stubWriter{writeErr: expectedError}

	assert.ErrorIs(t, handleLayoutInitUpdate(testWriter, &InitParams{}, testState), expectedError)
}

func Test_handleLayoutInitUpdate_noConfig(t *testing.T) {
	assert.NoError(t, handleLayoutInitUpdate(nil, &InitParams{}, &task.State{}))
}
