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

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
)

func TestPublishExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewPublishOptions()
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var expectedBroker = "broker"
	var expectedHandle = "handle"
	var expectedFqdn = "fqdn.genaiz.com"
	var expectedName = "name-publish"
	var expectedOem = "oem"
	var expectedType = layout.FunctionTypeFunction
	var expectedVersion = "version"
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: testOptions,

		brokerAddr: expectedBroker,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionFqdn.Key, expectedFqdn)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionType.Key, expectedType)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionArches.Param+`:[\s\t]*\[`+strings.Join(expectedArches, " ")+`\]`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionFqdn.Param+`:[\s\t]*`+expectedFqdn), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionRebuild.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionNoUpdate.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+expectedType), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`broker:[\s\t]*`+expectedBroker), actual)
}

func TestPublishExecutor_PretendNoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_PretendNoRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testExecutor.Pretend()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_PretendRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_PretendRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testExecutor.Pretend()
		assert.True(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_ProceedNoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory:      newInitTaskCompleteStub(&calledInit),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testLedger.Logger = &logrus.Logger{}
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_ProceedNoRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory:      newInitTaskCompleteStub(&calledInit),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testLedger.Logger = &logrus.Logger{}
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_ProceedRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory:      newInitTaskCompleteStub(&calledInit),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testLedger.Logger = &logrus.Logger{}
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_ProceedRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    NewSfCli(nil, nil, nil),
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory:      newInitTaskCompleteStub(&calledInit),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testProvisionParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := filez.CreateRecursiveTemp("/tmp/.genaiz", "GDockerfile"); err == nil {
		var fileName = tmpFile.Name()

		defer filez.RemoveSilently("/tmp/.genaiz")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionFqdn.Key, "test.genaiz.com")
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testLedger.Logger = &logrus.Logger{}
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
	} else {
		assert.NoError(t, err)
	}
}

func TestNewPublish(t *testing.T) {
	var publishCompleted = false
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
	var testPublish = NewPublish(testLedger, testCli)
	var expectedVersion = "version"
	var expectedHandle = "handle"
	var expectedHost = "host"

	testViper.Set(newOptionHandle("Publish").Key, expectedHandle)
	testViper.Set(newOptionVersion("Publish").Key, expectedVersion)
	testPublish.SetArgs([]string{expectedHost})
	testPublish.PostRun = func(cmd *cobra.Command, args []string) {
		publishCompleted = true
	}

	assert.NoError(t, testPublish.Execute())
	assert.True(t, publishCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedHandle)
		assert.Contains(t, actual, expectedHost)
		assert.Contains(t, actual, expectedVersion)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewPublishOptions(t *testing.T) {
	var expectedCmd = "Publish"
	var expectedArches = newOptionArches(expectedCmd)
	var expectedFqdn = newOptionFqdn(expectedCmd)
	var expectedHandle = newOptionHandle(expectedCmd)
	var expectedName = newOptionName(expectedHandle, expectedCmd)
	var expectedOem = newOptionOem(expectedCmd)
	var expectedType = newOptionType(expectedCmd)
	var expectedVersion = newOptionVersion(expectedCmd)
	var testOptions = NewPublishOptions()

	assert.True(t, expectedArches.Equals(&testOptions.optionArches.Option))
	assert.True(t, expectedFqdn.Equals(&testOptions.optionFqdn.Option))
	assert.True(t, expectedHandle.Equals(&testOptions.optionHandle.Option))
	assert.True(t, expectedName.Equals(&testOptions.optionName.Option))
	assert.True(t, expectedOem.Equals(&testOptions.optionOem.Option))
	assert.True(t, expectedType.Equals(&testOptions.optionType.Option))
	assert.True(t, expectedVersion.Equals(&testOptions.optionVersion.Option))
}

func TestNewPublishOptions_GetDefaultName(t *testing.T) {
	var cwd, _ = os.Getwd()
	var expectedName = filepath.Base(cwd)
	var expectedCmd = "_test"
	var testHandleOption = newOptionHandle(expectedCmd)
	var testNameOption = newOptionName(testHandleOption, expectedCmd)
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()

	testLedger.Register(&cobra.Command{}, testHandleOption)
	testLedger.InitDefaults()
	assert.EqualValues(t, expectedName, testNameOption.DefaultGetter(testLedger))
}

func TestNewPublishOptions_ValidateArches(t *testing.T) {
	var testOption = newOptionArches("_test")

	assert.False(t, testOption.Validator("not a valid arch"))

	for _, arch := range layout.ArchTypes.Values {
		assert.True(t, testOption.Validator(arch))
	}

	assert.True(t, testOption.Validator(layout.ArchTypes.Values))
}

func TestNewPublishOptions_ValidateFqdn(t *testing.T) {
	var testOption = newOptionFqdn("_test")

	assert.False(t, testOption.Validator("not a valid domain"))
	assert.False(t, testOption.Validator(".com"))
	assert.False(t, testOption.Validator("a.a"))
	assert.False(t, testOption.Validator("a%.acorn"))
	assert.False(t, testOption.Validator("abcdefghijklmnopqrstuvwxyzAbcdefghijklmnopqrstuvwxyzABcdefghijklmnopqrstuvwxyz.info"))
	assert.True(t, testOption.Validator("dev.genaiz.com"))
	assert.True(t, testOption.Validator("dev.genaiz.com"))
	assert.True(t, testOption.Validator("genaiz.com"))
}

func TestNewPublishOptions_ValidateHandle(t *testing.T) {
	var testOption = newOptionHandle("_test")

	assert.False(t, testOption.Validator("not a valid handle"))

	for _, c := range "`~!@#$%^&*()+=\\][}{'\";:>,<?|" {
		assert.False(t, testOption.Validator(string(c)))
	}

	assert.True(t, testOption.Validator("abcdefghijklmnopqrstuvwxyxABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789-._"))
}

func TestNewPublishOptions_ValidateType(t *testing.T) {
	var testOption = newOptionType("_test")

	assert.False(t, testOption.Validator("not valid ever"))

	for _, ct := range layout.FunctionTypes.Values {
		assert.True(t, testOption.Validator(ct))
	}

	// mutually exclusive, does not accept multiple types
	assert.False(t, testOption.Validator(layout.FunctionTypes))
}

func Test_makePublishInitParams_WithArches(t *testing.T) {
	var testOptions = newPublishOptions("Publish")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var actualParams *layout.InitParams

	testViper.Set(testOptions.optionFqdn.Key, "dev.genaiz.com")
	testViper.Set(testOptions.optionHandle.Key, "handle")
	testViper.Set(testOptions.optionType.Key, layout.FunctionTypeTrigger)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	actualParams = makePublishInitParams(testLedger, testOptions)
	assert.EqualValues(t, expectedArches, actualParams.Arches)
}

func Test_newPublishOptions_Rebuild(t *testing.T) {
	var expectedRebuildOption = config.BoolOption{}
	var testOptions = newPublishOptions("Publish", &expectedRebuildOption)

	assert.Same(t, &expectedRebuildOption, testOptions.optionRebuild)
}

func Test_newPublishOptions_RebuildAndNoUpdate(t *testing.T) {
	var expectedNoUpdateOption = config.BoolOption{}
	var expectedRebuildOption = config.BoolOption{}
	var testOptions = newPublishOptions("Publish", &expectedRebuildOption, &expectedNoUpdateOption)

	assert.Same(t, &expectedNoUpdateOption, testOptions.optionNoUpdate)
	assert.Same(t, &expectedRebuildOption, testOptions.optionRebuild)
}

func newTaskPretendStub[T any](flag *bool, paramType *T) func() *task.Task[T] {
	return func() *task.Task[T] {
		return &task.Task[T]{
			OnPrepare: func(params *T, state *task.State) error {
				return nil
			},
			OnPretend: func(params *T, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}

func newTaskProceedStub[T any](flag *bool, paramType *T) func() *task.Task[T] {
	return func() *task.Task[T] {
		return &task.Task[T]{
			OnPrepare: func(params *T, state *task.State) error {
				return nil
			},
			OnComplete: func(params *T, state *task.State) error {
				*flag = true
				return nil
			},
		}
	}
}
