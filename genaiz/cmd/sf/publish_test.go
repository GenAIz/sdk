package sf

import (
	"bytes"
	"errors"
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
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type stubCliPrinterParametric struct {
	cliPrinter cli.Printer
}

func (s stubCliPrinterParametric) Printer() cli.Printer {
	return s.cliPrinter
}

func (s stubCliPrinterParametric) IsDefault() bool {
	return true
}

type stubCliPrinter struct {
	errorError   error
	errorPayload interface{}
	printError   error
	printPayload interface{}
}

func (s *stubCliPrinter) Error(i interface{}) error {
	s.errorPayload = i
	return s.errorError
}

func (s *stubCliPrinter) Print(i interface{}) error {
	s.printPayload = i
	return s.printError
}

func TestPublishExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewPublishOptions(NewSfCli(nil, nil, nil))
	var expectedArches = []string{shared.ArchTypeArm64, shared.ArchTypeX86}
	var expectedBroker = "broker"
	var expectedHandle = "handle"
	var expectedName = "name-publish"
	var expectedOem = "oem"
	var expectedType = shared.FunctionTypeFunction
	var expectedVersion = "0.0.0"
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionAccount.Key, expectedBroker)
	testViper.Set(testOptions.optionHandle.Key, expectedHandle)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionOem.Key, expectedOem)
	testViper.Set(testOptions.optionType.Key, expectedType)
	testViper.Set(testOptions.optionVersion.Key, expectedVersion)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionArches.Param+`:[\s\t]*\[`+strings.Join(expectedArches, " ")+`\]`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionHandle.Param+`:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionRebuild.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionNoUpdate.Param+`:[\s\t]*false`), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionOem.Param+`:[\s\t]*`+expectedOem), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+expectedType), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionVersion.Param+`:[\s\t]*`+expectedVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`account:[\s\t]*`+expectedBroker), actual)
}

func TestPublishExecutor_Pretend_NoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()
		var expectedExtraKey = "extra"
		var expectedExtraValue = 37

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testViper.Set(testExecutor.innerExtras.Key, map[string]any{expectedExtraKey: expectedExtraValue})
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

func TestPublishExecutor_Pretend_NoRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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

func TestPublishExecutor_Pretend_RebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionRebuild.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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

func TestPublishExecutor_Pretend_RebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionRebuild.Key, true)
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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

func TestPublishExecutor_Proceed_JsonPrinter(t *testing.T) {
	var calledBuild, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedVersion = "0.0.0"
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCliPrinter = &stubCliPrinter{}
	var testFunction = &broker.Function{Id: 37}
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedVersion, params.Version)
		}),
		getTaskFactory:       newFunctionGetCompleteStub(testFunction, nil),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerVersion.Key, expectedVersion)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionJsonPrinter.Key, true)
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
		assert.True(t, calledId)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_Proceed_JsonPrinterError(t *testing.T) {
	var calledBuild, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedVersion = "0.0.0"
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCliPrinter = &stubCliPrinter{}
	var testFunction = &broker.Function{Id: 37}
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		printerParams: &stubCliPrinterParametric{
			cliPrinter: testCliPrinter,
		},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedVersion, params.Version)
		}),
		getTaskFactory:       newFunctionGetCompleteStub(testFunction, errors.New("test")),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerVersion.Key, expectedVersion)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionJsonPrinter.Key, true)
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
		assert.True(t, calledId)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_Proceed_NoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledGet, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedVersion = "0.0.0"
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedVersion, params.Version)
		}),
		getTaskFactory:       newTaskProceedStub(&calledGet, testGetParams),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerVersion.Key, expectedVersion)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testLedger.InitLogging()
		testExecutor.Proceed()
		assert.False(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.False(t, calledInit)
		assert.True(t, calledId)
		assert.True(t, calledGet)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_Proceed_NoRebuildUpdate(t *testing.T) {
	var calledBuild, calledGet, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedArches = []string{shared.ArchTypeX86, shared.ArchTypeArm64}
	var expectedHandle = "handleNoRebuildUpdate"
	var testDir = filepath.Join(t.TempDir(), expectedHandle)
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedArches, params.Arches)
			assert.Equal(t, expectedHandle, params.Handle)
		}),
		getTaskFactory:       newTaskProceedStub(&calledGet, testGetParams),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}
	var err error

	if err = os.MkdirAll(testDir, 0750); err == nil {
		var tmpFile *os.File

		if tmpFile, err = os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
			var fileName = tmpFile.Name()

			defer filez.CloseSilently(tmpFile)
			t.Chdir(testDir)
			testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
			testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
			testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
			testViper.Set(testExecutor.optionArches.Key, expectedArches)
			testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
			testViper.Set(testExecutor.optionOem.Key, "oem")
			testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
			testLedger.InitLogging()
			testExecutor.Proceed()
			assert.False(t, calledBuild)
			assert.True(t, calledInspect)
			assert.True(t, calledProvision)
			assert.True(t, calledPush)
			assert.True(t, calledPublish)
			assert.True(t, calledInit)
			assert.True(t, calledId)
			assert.True(t, calledGet)
		}
	}

	assert.NoError(t, err)
}

func TestPublishExecutor_Proceed_RebuildUpdate(t *testing.T) {
	var calledBuild, calledGet, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
		}),
		getTaskFactory:       newTaskProceedStub(&calledGet, testGetParams),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionRebuild.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testLedger.InitLogging()
		testLedger.Logger.Level = logrus.DebugLevel
		testExecutor.Proceed()
		assert.True(t, calledBuild)
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledInit)
		assert.True(t, calledId)
		assert.True(t, calledGet)
	} else {
		assert.NoError(t, err)
	}
}

func TestPublishExecutor_Proceed_RebuildNoUpdate(t *testing.T) {
	var calledBuild, calledGet, calledId, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedArches = []string{shared.ArchTypeX86, shared.ArchTypeArm64}
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testGetParams = &broker.GetParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerDataSources:     &config.ListOption{Option: config.Option{Key: "innerDataSources"}},
		innerDataStores:      &config.ListOption{Option: config.Option{Key: "innerDataStores"}},
		innerExtras:          &config.Option{Key: "innerExtras"},
		innerInputPorts:      &config.Option{Key: "innerInputPorts"},
		innerOutputPorts:     &config.Option{Key: "innerOutputPorts"},
		innerOutboundProxies: &config.Option{Key: "innerOutboundProxies"},
		innerPropSpecs:       &config.Option{Key: "innerSpecs"},
		innerResultValues:    &config.ListOption{Option: config.Option{Key: "innerResultValues"}},

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedArches, params.Arches)
		}),
		getTaskFactory:       newTaskProceedStub(&calledGet, testGetParams),
		idTaskFactory:        newTaskProceedStub(&calledId, testGetParams),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		defer filez.CloseSilently(tmpFile)
		testViper.Set(testExecutor.Cli.optionDockerRepo.Key, "namespace/repo")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testExecutor.optionArches.Key, expectedArches)
		testViper.Set(testExecutor.optionRebuild.Key, true)
		testViper.Set(testExecutor.optionNoUpdate.Key, true)
		testViper.Set(testExecutor.optionType.Key, shared.FunctionTypeFunction)
		testViper.Set(testExecutor.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
		testLedger.InitLogging()
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

func TestPublishExecutor_makeProvisionExtras(t *testing.T) {
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: NewPublishOptions(testCli),

		innerExtras: &config.Option{Key: "innerExtras"},
	}
	var expectedExtraKey = "extra"
	var expectedExtraValue = 37
	var actual map[string]any

	testViper.Set(testExecutor.innerExtras.Key, map[string]any{expectedExtraKey: expectedExtraValue})
	actual = testExecutor.makeProvisionExtras()
	assert.Equal(t, expectedExtraValue, actual[expectedExtraKey])
}

func TestNewPublish(t *testing.T) {
	var publishCompleted = false
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
		optionDockerRepo:    cli.Options.Docker.Repository().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
	var testPublish = NewPublish(testLedger, testCli)
	var expectedVersion = "0.0.0"
	var expectedHandle = "handle"
	var expectedHost = "host"

	testViper.Set(schema.Genaiz.Function.Publish.Account.Doc, expectedHost)
	testViper.Set(schema.Genaiz.Function.Publish.Handle.Doc, expectedHandle)
	testViper.Set(schema.Genaiz.Function.Publish.Version.Doc, expectedVersion)
	testPublish.SetArgs([]string{})
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

func newTaskPretendStub[T any](flag *bool, paramType *T) func() *task.Task[T] {
	_ = paramType // suppress warning on inference type

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

func newFunctionGetCompleteStub(expectedFunction *broker.Function, expectedError error) func() *task.Task[broker.GetParams] {
	return func() *task.Task[broker.GetParams] {
		return &task.Task[broker.GetParams]{
			OnPrepare: func(params *broker.GetParams, state *task.State) error { return nil },
			OnComplete: func(params *broker.GetParams, state *task.State) error {
				state.Internal = expectedFunction
				return expectedError
			},
		}
	}
}

func newTaskProceedStub[T any](flag *bool, paramType *T) func() *task.Task[T] {
	_ = paramType // suppress warning on inference type

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
