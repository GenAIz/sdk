package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
)

var (
	testInitKeys          = &schema.Keys{Doc: "Init"}
	testInitOptionArches  = cli.Options.Functions.Arches().WithKeys(testInitKeys).BuildListOption()
	testInitOptionHandle  = cli.Options.Functions.Handle().WithKeys(testInitKeys).BuildStringOption()
	testInitOptionName    = cli.Options.Functions.Name().WithKeys(testInitKeys).WithDefaultGetter(func(ledger *config.Ledger) any { return ledger.GetString(testInitOptionHandle) }).BuildStringOption()
	testInitOptionOem     = cli.Options.Functions.Oem().WithKeys(testInitKeys).BuildStringOption()
	testInitOptionType    = cli.Options.Functions.Type().WithKeys(testInitKeys).BuildStringOption()
	testInitOptionVersion = cli.Options.Functions.Type().WithKeys(testInitKeys).BuildStringOption()
)

func TestInitWriter_BuildArches(t *testing.T) {
	var expectedArches = []string{"test"}
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionArches: testInitOptionArches,
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithArches(expectedArches).BuildArches()

	assert.EqualValues(t, testWriter.optionArches.Key, actualKey)
	assert.ElementsMatch(t, expectedArches, actualValue)

	_, actualValue = testWriter.WithArches(nil).BuildArches()

	assert.ElementsMatch(t, expectedArches, actualValue)
}

func TestInitWriter_BuildHandle(t *testing.T) {
	var expectedHandle = "test-handle"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionOem:    testInitOptionOem,
			optionHandle: testInitOptionHandle,
		},
		vp:           testViper,
		buildTagKeys: &schema.Genaiz.Function.Build.Tag,
	}
	var actualKey, actualValue = testWriter.WithHandle(expectedHandle).BuildHandle()

	assert.EqualValues(t, testWriter.optionHandle.Key, actualKey)
	assert.EqualValues(t, expectedHandle, actualValue)

	_, actualValue = testWriter.WithHandle("").BuildHandle()

	assert.EqualValues(t, expectedHandle, actualValue)
	assert.EqualValues(t, expectedHandle, testViper.GetString(testWriter.buildTagKeys.Doc))
}

func TestInitWriter_BuildInput(t *testing.T) {
	var expectedInput = "input"
	var testWriter = &InitWriter{
		RunOptions: &RunOptions{
			optionMountInput: cli.Options.Functions.MountInput().
				WithKeys(&schema.Genaiz.Function.Test.MountInput).
				BuildStringOption(),
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithInput(expectedInput).BuildInput()

	assert.EqualValues(t, testWriter.optionMountInput.Key, actualKey)
	assert.EqualValues(t, expectedInput, actualValue)

	_, actualValue = testWriter.WithInput("").BuildInput()

	assert.EqualValues(t, expectedInput, actualValue)
}

func TestInitWriter_BuildInputPorts(t *testing.T) {
	var expectedDataPort = broker.DataPort{Handle: "portHandle"}
	var testPorts = []broker.DataPort{expectedDataPort}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishInputPortsKeys: &schema.Genaiz.Function.Publish.InputPorts,
		vp:                    testViper,
	}
	var actualPorts []broker.DataPort
	var actualKey string

	actualKey, actualPorts = testWriter.WithInputPorts(nil).BuildInputPorts()
	assert.Equal(t, testWriter.publishInputPortsKeys.Doc, actualKey)
	assert.Empty(t, actualPorts)

	actualKey, actualPorts = testWriter.WithInputPorts(testPorts).BuildInputPorts()
	assert.Equal(t, testWriter.publishInputPortsKeys.Doc, actualKey)
	assert.Equal(t, testPorts, actualPorts)
}

func TestInitWriter_BuildInputPortRemoved(t *testing.T) {
	var expectedDataPort = &broker.DataPort{Handle: "portHandle"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishInputPortsKeys: &schema.Genaiz.Function.Publish.OutputPorts,
		vp:                    testViper,
	}
	var actualPort *broker.DataPort
	var actualKey string

	actualKey, actualPort = testWriter.BuildInputPortRemoved()
	assert.Equal(t, testWriter.publishInputPortsKeys.Doc, actualKey)
	assert.Empty(t, actualPort)

	actualKey, actualPort = testWriter.WithInputPortRemoved(expectedDataPort).BuildInputPortRemoved()
	assert.Equal(t, testWriter.publishInputPortsKeys.Doc, actualKey)
	assert.Equal(t, expectedDataPort, actualPort)
}

func TestInitWriter_BuildName(t *testing.T) {
	var expectedName = "name"
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionName: testInitOptionName,
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithName(expectedName).BuildName()

	assert.EqualValues(t, testWriter.optionName.Key, actualKey)
	assert.EqualValues(t, expectedName, actualValue)

	_, actualValue = testWriter.WithName("").BuildName()

	assert.EqualValues(t, expectedName, actualValue)
}

func TestInitWriter_BuildOem(t *testing.T) {
	var expectedOem = "oem"
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionHandle: testInitOptionHandle,
			optionOem:    testInitOptionOem,
		},
		vp:           viper.New(),
		buildTagKeys: &schema.Genaiz.Function.Build.Tag,
	}
	var actualKey, actualValue = testWriter.WithOem(expectedOem).BuildOem()

	assert.EqualValues(t, testWriter.optionOem.Key, actualKey)
	assert.EqualValues(t, expectedOem, actualValue)

	_, actualValue = testWriter.WithOem("").BuildOem()

	assert.EqualValues(t, expectedOem, actualValue)
}

func TestInitWriter_BuildOutboundProxies(t *testing.T) {
	var expectedProxy = broker.Proxy{Host: "expectedHost"}
	var testProxies = []broker.Proxy{expectedProxy}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishOutboundProxiesKeys: &schema.Genaiz.Function.Publish.OutboundProxies,
		vp:                         testViper,
	}
	var actualProxies []broker.Proxy
	var actualKey string

	actualKey, actualProxies = testWriter.WithOutboundProxies(nil).BuildOutboundProxies()
	assert.Equal(t, testWriter.publishOutboundProxiesKeys.Doc, actualKey)
	assert.Empty(t, actualProxies)

	actualKey, actualProxies = testWriter.WithOutboundProxies(testProxies).BuildOutboundProxies()
	assert.Equal(t, testWriter.publishOutboundProxiesKeys.Doc, actualKey)
	assert.Equal(t, testProxies, actualProxies)
}

func TestInitWriter_BuildOutboundProxyRemoved(t *testing.T) {
	var expectedProxy = &broker.Proxy{Host: "expectedHost"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishOutboundProxiesKeys: &schema.Genaiz.Function.Publish.OutboundProxies,
		vp:                         testViper,
	}
	var actualProxy *broker.Proxy
	var actualKey string

	actualKey, actualProxy = testWriter.BuildOutboundProxyRemoved()
	assert.Equal(t, testWriter.publishOutboundProxiesKeys.Doc, actualKey)
	assert.Empty(t, actualProxy)

	actualKey, actualProxy = testWriter.WithOutboundProxyRemoved(expectedProxy).BuildOutboundProxyRemoved()
	assert.Equal(t, testWriter.publishOutboundProxiesKeys.Doc, actualKey)
	assert.Equal(t, expectedProxy, actualProxy)
}

func TestInitWriter_BuildOutput(t *testing.T) {
	var expectedOutput = "output"
	var expectedLog = "log"
	var expectedVar = "var"
	var testWriter = &InitWriter{
		RunOptions: &RunOptions{
			optionMountOutput: cli.Options.Functions.MountOutput().
				WithKeys(&schema.Genaiz.Function.Start.MountOutput).
				BuildStringOption(),
			optionMountVar: cli.Options.Functions.MountVar().
				WithKeys(&schema.Genaiz.Function.Start.MountVar).
				BuildStringOption(),
			optionMountLog: cli.Options.Functions.MountLog().
				WithKeys(&schema.Genaiz.Function.Start.MountLog).
				BuildStringOption(),
		},
		vp: viper.New(),
	}
	var actualValues = testWriter.WithOutput(expectedOutput).WithLog(expectedLog).WithVar(expectedVar).BuildOutput()

	assert.Equal(t, actualValues[testWriter.optionMountOutput.Key], expectedOutput)
	assert.Equal(t, actualValues[testWriter.optionMountLog.Key], expectedLog)
	assert.Equal(t, actualValues[testWriter.optionMountVar.Key], expectedVar)

	actualValues = testWriter.WithOutput("").WithLog("").WithVar("").BuildOutput()

	assert.Equal(t, actualValues[testWriter.optionMountOutput.Key], expectedOutput)
	assert.Equal(t, actualValues[testWriter.optionMountLog.Key], expectedLog)
	assert.Equal(t, actualValues[testWriter.optionMountVar.Key], expectedVar)
}

func TestInitWriter_BuildOutputPorts(t *testing.T) {
	var expectedDataPort = broker.DataPort{Handle: "portHandle"}
	var testPorts = []broker.DataPort{expectedDataPort}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishOutputPortsKeys: &schema.Genaiz.Function.Publish.OutputPorts,
		vp:                     testViper,
	}
	var actualPorts []broker.DataPort
	var actualKey string

	actualKey, actualPorts = testWriter.WithOutputPorts(nil).BuildOutputPorts()
	assert.Equal(t, testWriter.publishOutputPortsKeys.Doc, actualKey)
	assert.Empty(t, actualPorts)

	actualKey, actualPorts = testWriter.WithOutputPorts(testPorts).BuildOutputPorts()
	assert.Equal(t, testWriter.publishOutputPortsKeys.Doc, actualKey)
	assert.Equal(t, testPorts, actualPorts)
}

func TestInitWriter_BuildOutputPortRemoved(t *testing.T) {
	var expectedDataPort = &broker.DataPort{Handle: "portHandle"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishOutputPortsKeys: &schema.Genaiz.Function.Publish.OutputPorts,
		vp:                     testViper,
	}
	var actualPort *broker.DataPort
	var actualKey string

	actualKey, actualPort = testWriter.BuildOutputPortRemoved()
	assert.Equal(t, testWriter.publishOutputPortsKeys.Doc, actualKey)
	assert.Empty(t, actualPort)

	actualKey, actualPort = testWriter.WithOutputPortRemoved(expectedDataPort).BuildOutputPortRemoved()
	assert.Equal(t, testWriter.publishOutputPortsKeys.Doc, actualKey)
	assert.Equal(t, expectedDataPort, actualPort)
}

func TestInitWriter_BuildPropSpecs(t *testing.T) {
	var expectedPropSpec = broker.PropSpec{Key: "expectedPropKey"}
	var testSpecs = []broker.PropSpec{expectedPropSpec}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishPropSpecsKeys: &schema.Genaiz.Function.Publish.PropSpecs,
		vp:                   testViper,
	}
	var actualSpecs []broker.PropSpec
	var actualKey string

	actualKey, actualSpecs = testWriter.WithPropSpecs(nil).BuildPropSpecs()
	assert.Equal(t, testWriter.publishPropSpecsKeys.Doc, actualKey)
	assert.Empty(t, actualSpecs)

	actualKey, actualSpecs = testWriter.WithPropSpecs(testSpecs).BuildPropSpecs()
	assert.Equal(t, testWriter.publishPropSpecsKeys.Doc, actualKey)
	assert.Equal(t, testSpecs, actualSpecs)
}

func TestInitWriter_BuildRemovedPropSpec(t *testing.T) {
	var expectedPropSpec = broker.PropSpec{Key: "expectedPropKey"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishPropSpecsKeys: &schema.Genaiz.Function.Publish.PropSpecs,
		vp:                   testViper,
	}
	var actualSpecs *broker.PropSpec
	var actualKey string

	actualKey, actualSpecs = testWriter.BuildPropSpecRemoved()
	assert.Equal(t, testWriter.publishPropSpecsKeys.Doc, actualKey)
	assert.Empty(t, actualSpecs)

	actualKey, actualSpecs = testWriter.WithPropSpecRemoved(&expectedPropSpec).BuildPropSpecRemoved()
	assert.Equal(t, testWriter.publishPropSpecsKeys.Doc, actualKey)
	assert.Equal(t, expectedPropSpec, *actualSpecs)
}

func TestInitWriter_BuildSources(t *testing.T) {
	var expectedSources = []string{"expectedLink"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishSourcesKeys: &schema.Genaiz.Function.Publish.DataSources,
		vp:                 testViper,
	}
	var actualSources []string
	var actualKey string

	actualKey, actualSources = testWriter.WithSources(nil).BuildSources()
	assert.Equal(t, testWriter.publishSourcesKeys.Doc, actualKey)
	assert.Empty(t, actualSources)

	actualKey, actualSources = testWriter.WithSources(expectedSources).BuildSources()
	assert.Equal(t, testWriter.publishSourcesKeys.Doc, actualKey)
	assert.Equal(t, expectedSources, actualSources)

	testWriter.vp.Set(testWriter.publishSourcesKeys.Doc, "invalid slice")
	actualKey, actualSources = testWriter.BuildSources()
	assert.Empty(t, actualSources)
}

func TestInitWriter_BuildStores(t *testing.T) {
	var expectedStores = []string{"expectedLink"}
	var testViper = viper.New()
	var testWriter = &InitWriter{
		publishStoresKeys: &schema.Genaiz.Function.Publish.DataStores,
		vp:                testViper,
	}
	var actualStores []string
	var actualKey string

	actualKey, actualStores = testWriter.WithStores(nil).BuildStores()
	assert.Equal(t, testWriter.publishStoresKeys.Doc, actualKey)
	assert.Empty(t, actualStores)

	actualKey, actualStores = testWriter.WithStores(expectedStores).BuildStores()
	assert.Equal(t, testWriter.publishStoresKeys.Doc, actualKey)
	assert.Equal(t, expectedStores, actualStores)

	testWriter.vp.Set(testWriter.publishStoresKeys.Doc, "invalid slice")
	actualKey, actualStores = testWriter.BuildStores()
	assert.Empty(t, actualStores)
}

func TestInitWriter_BuildType(t *testing.T) {
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionType: testInitOptionType,
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithType(layout.FunctionTypeConnector).BuildType()

	assert.EqualValues(t, testWriter.optionType.Key, actualKey)
	assert.EqualValues(t, layout.FunctionTypeConnector, actualValue)

	_, actualValue = testWriter.WithType("").BuildType()

	assert.EqualValues(t, layout.FunctionTypeConnector, actualValue)
}

func TestInitWriter_BuildVersion(t *testing.T) {
	var expectedVersion = "0.0.0"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionVersion: testInitOptionVersion,
		},
		vp:               testViper,
		buildVersionKeys: &schema.Genaiz.Function.Build.Version,
	}
	var actualKey, actualValue = testWriter.WithVersion(expectedVersion).BuildVersion()

	assert.EqualValues(t, testWriter.optionVersion.Key, actualKey)
	assert.EqualValues(t, expectedVersion, actualValue)

	_, actualValue = testWriter.WithVersion("").BuildVersion()

	assert.EqualValues(t, expectedVersion, actualValue)
	assert.EqualValues(t, "latest", testViper.GetString(testWriter.buildVersionKeys.Doc))
}

func TestInitWriter_Write(t *testing.T) {
	var expectedFile = filepath.Join(t.TempDir(), ".init_test", "init_write_test.yaml")
	var expectedOem = "genaiz.com"
	var expectedHandle = "test-handle"
	var expectedVersion = "0.0.0"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionOem:     testInitOptionOem,
			optionHandle:  testInitOptionHandle,
			optionVersion: testInitOptionVersion,
		},
		vp:               testViper,
		buildTagKeys:     &schema.Genaiz.Function.Build.Tag,
		buildVersionKeys: &schema.Genaiz.Function.Build.Version,
	}
	var testFolder = filepath.Dir(expectedFile)

	if fd, err := filez.CreateRecursive(testFolder, filepath.Base(expectedFile)); err == nil {
		defer filez.CloseSilently(fd)

		assert.NoError(t, testWriter.
			WithOem(expectedOem).
			WithHandle(expectedHandle).
			WithVersion(expectedVersion).
			Write(expectedFile))

		assert.NotPanics(t, func() { testWriter.WithConfigFile(expectedFile) })

		assert.EqualValues(t, "latest", testViper.GetString(testWriter.buildVersionKeys.Doc))
		assert.EqualValues(t, expectedOem+"/"+expectedHandle, testViper.GetString(testWriter.buildTagKeys.Doc))
	} else {
		assert.NoError(t, err)
	}

	defer filez.RemoveSilently(testFolder)
}

func TestInitWriter_WriteInvalidFile(t *testing.T) {
	var invalidFile = filepath.Join(t.TempDir(), ".init_test", "init_write_invalid.yaml")
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionHandle:  testInitOptionHandle,
			optionVersion: testInitOptionVersion,
		},
		vp:               viper.New(),
		buildTagKeys:     &schema.Genaiz.Function.Build.Tag,
		buildVersionKeys: &schema.Genaiz.Function.Build.Version,
	}

	assert.Panics(t, func() { testWriter.WithConfigFile(invalidFile) })
}

func TestInitExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewInitOptions(NewSfCli(nil, nil, nil))
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var expectedHandle = "handle"
	var expectedName = "name-init"
	var expectedOem = "oem"
	var expectedVersion = "0.0.0"
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionArches.Param+`:[\s\t]*\[`+strings.Join(expectedArches, " ")+`\]`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+layout.FunctionTypeFunction), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
}

func TestInitExecutor_Pretend(t *testing.T) {
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledInit = false
	var testCli = NewSfCli(nil, nil, nil)
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},

		InitOptions:     NewInitOptions(testCli),
		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.optionConfigType.Key, "yaml")
		testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "init-pretend")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testViper.Set(testInitOptionArches.Key, layout.ArchTypeArm64)
		testExecutor.Pretend()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestInitExecutor_Pretend_InvalidConfigType(t *testing.T) {
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledInit = false
	var testCli = NewSfCli(nil, nil, nil)
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},

		InitOptions:     NewInitOptions(testCli),
		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(int) {})

		defer patch.Unpatch()
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(testExecutor.optionConfigType.Key, "invalidType")
		testExecutor.Pretend()
		assert.False(t, calledInit)
		assert.NotEmpty(t, patch.CalledWith)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestInitExecutor_Proceed(t *testing.T) {
	var calledInit bool
	var expectedHandle = "handleProceed"
	var testDir = filepath.Join(t.TempDir(), expectedHandle)
	var expectedArches = []string{layout.ArchTypeX86, layout.ArchTypeArm}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		InitOptions: NewInitOptions(testCli),

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedArches, actual.Arches)
		}),
	}
	var err error

	if err = os.MkdirAll(testDir, 0750); err == nil {
		var fd *os.File

		if fd, err = os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
			filez.CloseSilently(fd)
			testLedger.Logger = logrus.New()
			testLedger.WorkDir = testDir
			testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
			testViper.Set(testExecutor.optionArches.Key, expectedArches)
			testViper.Set(testExecutor.optionConfigType.Key, "yaml")
			testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
			testViper.Set(testExecutor.optionOem.Key, "oem")
			testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
			testExecutor.Proceed()
			assert.True(t, calledInit)
		}
	}

	assert.NoError(t, err)
}

func TestInitExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledInit = false
	var testCli = NewSfCli(nil, nil, nil)
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},

		InitOptions:     NewInitOptions(testCli),
		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(int) {})

		defer patch.Unpatch()
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(testExecutor.optionConfigType.Key, "invalidType")
		testExecutor.Proceed()
		assert.False(t, calledInit)
		assert.NotEmpty(t, patch.CalledWith)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestInitExecutor_Proceed_InvalidOem(t *testing.T) {
	var expectedHandle = "handleProceed"
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testDir = filepath.Join(t.TempDir(), expectedHandle)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		InitOptions: NewInitOptions(testCli),

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {}),
	}
	var err error

	defer patch.Unpatch()

	if err = os.MkdirAll(testDir, 0750); err == nil {
		t.Chdir(testDir)
		testLedger.Logger = logrus.New()
		testViper.Set(testExecutor.optionConfigType.Key, "yaml")
		testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testExecutor.Proceed()
	}

	assert.NoError(t, err)
	assert.True(t, patch.Called)
	assert.Equal(t, 1, patch.CalledWith)
}

func TestNewInit(t *testing.T) {
	var initCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testInit = NewInit(testLedger, testCli)
	var expectedHandle = "init-handle"
	var expectedOem = "init-oem"

	testViper.Set(schema.Genaiz.Function.Init.Handle.Doc, expectedHandle)
	testViper.Set(schema.Genaiz.Function.Init.Oem.Doc, expectedOem)
	testViper.Set(schema.Genaiz.Function.Init.Type.Doc, layout.FunctionTypeFunction)
	testInit.PostRun = func(cmd *cobra.Command, args []string) {
		initCompleted = true
	}

	assert.NoError(t, testInit.Execute())
	assert.True(t, initCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedHandle)
		assert.Contains(t, actual, expectedOem)
		assert.Contains(t, actual, layout.FunctionTypeFunction)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func newInitTaskCompleteStub(checks func(params *layout.InitParams)) InitTaskFactory {
	return func(builder layout.ConfigWriter) *task.Task[layout.InitParams] {
		return &task.Task[layout.InitParams]{
			Name: "init_test",
			OnPrepare: func(params *layout.InitParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *layout.InitParams, state *task.State) error {
				checks(params)
				return nil
			},
		}
	}
}

func newInitTaskPretendStub(flag *bool) InitTaskFactory {
	return func(layout.ConfigWriter) *task.Task[layout.InitParams] {
		return &task.Task[layout.InitParams]{
			OnPrepare: func(params *layout.InitParams, state *task.State) error {
				return nil
			},
			OnPretend: func(params *layout.InitParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
