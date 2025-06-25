package sf

import (
	"bytes"
	"io"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/layout"
)

func TestInitWriter_BuildArches(t *testing.T) {
	var expectedArches = []string{"test"}
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionArches: newOptionArches("_test"),
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithArches(expectedArches).BuildArches()

	assert.EqualValues(t, testWriter.optionArches.Key, actualKey)
	assert.ElementsMatch(t, expectedArches, actualValue)

	_, actualValue = testWriter.WithArches(nil).BuildArches()

	assert.ElementsMatch(t, expectedArches, actualValue)
}

func TestInitWriter_BuildFqdn(t *testing.T) {
	var expectedFqdn = "test.genaiz.com"
	var expectedTag = expectedFqdn[strings.Index(expectedFqdn, ".")+1:]
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionFqdn:   newOptionFqdn("_test"),
			optionHandle: newOptionHandle("_test"),
		},
		vp:      testViper,
		baseTag: newOptionDockerTag(),
	}
	var actualKey, actualValue = testWriter.WithFqdn(expectedFqdn).BuildFqdn()

	assert.EqualValues(t, testWriter.optionFqdn.Key, actualKey)
	assert.EqualValues(t, expectedFqdn, actualValue)

	_, actualValue = testWriter.WithFqdn("").BuildFqdn()

	assert.EqualValues(t, expectedFqdn, actualValue)
	assert.EqualValues(t, expectedTag, testViper.GetString(testWriter.baseTag.Key))
}

func TestInitWriter_BuildHandle(t *testing.T) {
	var expectedHandle = "test-handle"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionFqdn:   newOptionFqdn("_test"),
			optionHandle: newOptionHandle("_test"),
		},
		vp:      testViper,
		baseTag: newOptionDockerTag(),
	}
	var actualKey, actualValue = testWriter.WithHandle(expectedHandle).BuildHandle()

	assert.EqualValues(t, testWriter.optionHandle.Key, actualKey)
	assert.EqualValues(t, expectedHandle, actualValue)

	_, actualValue = testWriter.WithHandle("").BuildHandle()

	assert.EqualValues(t, expectedHandle, actualValue)
	assert.EqualValues(t, expectedHandle, testViper.GetString(testWriter.baseTag.Key))
}

func TestInitWriter_BuildInput(t *testing.T) {
	var expectedInput = "input"
	var testWriter = &InitWriter{
		RunOptions: &RunOptions{
			optionMountInput: newOptionMountInput("_test", false),
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithInput(expectedInput).BuildInput()

	assert.EqualValues(t, testWriter.optionMountInput.Key, actualKey)
	assert.EqualValues(t, expectedInput, actualValue)

	_, actualValue = testWriter.WithInput("").BuildInput()

	assert.EqualValues(t, expectedInput, actualValue)
}

func TestInitWriter_BuildName(t *testing.T) {
	var expectedName = "name"
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionName: newOptionName(newOptionHandle("_test"), "_test"),
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
			optionOem: newOptionOem("_test"),
		},
		vp: viper.New(),
	}
	var actualKey, actualValue = testWriter.WithOem(expectedOem).BuildOem()

	assert.EqualValues(t, testWriter.optionOem.Key, actualKey)
	assert.EqualValues(t, expectedOem, actualValue)

	_, actualValue = testWriter.WithOem("").BuildOem()

	assert.EqualValues(t, expectedOem, actualValue)
}

func TestInitWriter_BuildOutput(t *testing.T) {
	var expectedOutput = "output"
	var outputOption = newOptionMountOutput("_test", false)
	var testWriter = &InitWriter{
		RunOptions: &RunOptions{
			optionMountOutput: outputOption,
			optionMountVar:    newOptionMountVar("_test", outputOption),
			optionMountLog:    newOptionMountLog("_test", outputOption),
		},
		vp: viper.New(),
	}
	var actualValues = testWriter.WithOutput(expectedOutput).BuildOutput()

	assert.EqualValues(t, actualValues[testWriter.optionMountOutput.Key], expectedOutput)
	assert.EqualValues(t, actualValues[testWriter.optionMountLog.Key], filepath.Join(expectedOutput, "log"))
	assert.EqualValues(t, actualValues[testWriter.optionMountVar.Key], filepath.Join(expectedOutput, "var"))

	actualValues = testWriter.WithOutput("").BuildOutput()

	assert.EqualValues(t, actualValues[testWriter.optionMountOutput.Key], expectedOutput)
	assert.EqualValues(t, actualValues[testWriter.optionMountLog.Key], filepath.Join(expectedOutput, "log"))
	assert.EqualValues(t, actualValues[testWriter.optionMountVar.Key], filepath.Join(expectedOutput, "var"))
}

func TestInitWriter_BuildType(t *testing.T) {
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionType: newOptionType("_test"),
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
	var expectedVersion = "version"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionVersion: newOptionVersion("_test"),
		},
		vp:          testViper,
		baseVersion: newOptionDockerVersion(),
	}
	var actualKey, actualValue = testWriter.WithVersion(expectedVersion).BuildVersion()

	assert.EqualValues(t, testWriter.optionVersion.Key, actualKey)
	assert.EqualValues(t, expectedVersion, actualValue)

	_, actualValue = testWriter.WithVersion("").BuildVersion()

	assert.EqualValues(t, expectedVersion, actualValue)
	assert.EqualValues(t, "latest", testViper.GetString(testWriter.baseVersion.Key))
}

func TestInitWriter_Write(t *testing.T) {
	var expectedFile = "/tmp/.genaiz/init_write_test.yaml"
	var expectedFqdn = "genaiz.com"
	var expectedHandle = "test-handle"
	var expectedVersion = "version"
	var testViper = viper.New()
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionFqdn:    newOptionVersion("_test"),
			optionHandle:  newOptionHandle("_test"),
			optionVersion: newOptionVersion("_test"),
		},
		vp:          testViper,
		baseTag:     newOptionDockerTag(),
		baseVersion: newOptionDockerVersion(),
	}
	var testFolder = filepath.Dir(expectedFile)

	if _, err := filez.CreateRecursive(testFolder, filepath.Base(expectedFile)); err == nil {
		assert.NoError(t, testWriter.
			WithFqdn(expectedFqdn).
			WithHandle(expectedHandle).
			WithVersion(expectedVersion).
			Write(expectedFile))

		assert.NotPanics(t, func() { testWriter.WithConfigFile(expectedFile) })

		assert.EqualValues(t, "latest", testViper.GetString(testWriter.baseVersion.Key))
		assert.EqualValues(t, expectedFqdn+"/"+expectedHandle, testViper.GetString(testWriter.baseTag.Key))
	} else {
		assert.NoError(t, err)
	}

	defer filez.RemoveSilently(testFolder)
}

func TestInitWriter_WriteInvalidFile(t *testing.T) {
	var invalidFile = "/tmp/.genaiz/init_write_invalid.yaml"
	var testWriter = &InitWriter{
		PublishOptions: &PublishOptions{
			optionFqdn:    newOptionVersion("_test"),
			optionHandle:  newOptionHandle("_test"),
			optionVersion: newOptionVersion("_test"),
		},
		vp:          viper.New(),
		baseTag:     newOptionDockerTag(),
		baseVersion: newOptionDockerVersion(),
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
	var testOptions = NewInitOptions()
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var expectedHandle = "handle"
	var expectedFqdn = "fqdn.genaiz.com"
	var expectedName = "name-init"
	var expectedOem = "oem"
	var expectedVersion = "version"
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		InitOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionFqdn.Key, expectedFqdn)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionArches.Param+`:[\s\t]*\[`+strings.Join(expectedArches, " ")+`\]`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionFqdn.Param+`:[\s\t]*`+expectedFqdn), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+layout.FunctionTypeFunction), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
}

func TestInitExecutor_Pretend(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledInit = false
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},

		InitOptions:     NewInitOptions(),
		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionFqdn.Key, "test.genaiz.com")
	testViper.Set(testExecutor.optionHandle.Key, "init-pretend")
	testViper.Set(newOptionArches("Init").Key, layout.ArchTypeArm64)
	testExecutor.Pretend()
	assert.True(t, calledInit)
}

func TestInitExecutor_Proceed(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var calledInit = false
	var testExecutor = &InitExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		InitOptions: NewInitOptions(),

		initTaskFactory: newInitTaskCompleteStub(&calledInit),
	}

	testLedger.Logger = logrus.New()
	testViper.Set(testExecutor.optionType.Key, layout.FunctionTypeFunction)
	testViper.Set(testExecutor.optionFqdn.Key, "test.genaiz.com")
	testViper.Set(testExecutor.optionHandle.Key, "init-pretend")
	testExecutor.Proceed()
	assert.True(t, calledInit)
}

func TestNewInit(t *testing.T) {
	var buildCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithOutput(io.Writer(testOutput)).WithViper(testViper).Build()
	var testCli = &Cli{
		Dry: func(ledger *config.Ledger) bool {
			return true
		},
		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
	var testInit = NewInit(testLedger, testCli)
	var expectedFqdn = "init.genaiz.com"
	var expectedHandle = "init-handle"

	testViper.Set(newOptionFqdn("Init").Key, expectedFqdn)
	testViper.Set(newOptionHandle("Init").Key, expectedHandle)
	testViper.Set(newOptionType("Init").Key, layout.FunctionTypeFunction)
	testInit.PostRun = func(cmd *cobra.Command, args []string) {
		buildCompleted = true
	}

	assert.NoError(t, testInit.Execute())
	assert.True(t, buildCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedFqdn)
		assert.Contains(t, actual, expectedHandle)
		assert.Contains(t, actual, layout.FunctionTypeFunction)
	} else {
		assert.Fail(t, "no --dry content")
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

func newInitTaskCompleteStub(flag *bool) InitTaskFactory {
	return func(builder layout.ConfigWriter) *task.Task[layout.InitParams] {
		return &task.Task[layout.InitParams]{
			Name: "init_test",
			OnPrepare: func(params *layout.InitParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *layout.InitParams, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
