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

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
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
	var testOptions = NewPublishOptions(NewSfCli(nil, nil, nil))
	var expectedArches = []string{layout.ArchTypeArm64, layout.ArchTypeX86}
	var expectedBroker = "broker"
	var expectedHandle = "handle"
	var expectedName = "name-publish"
	var expectedOem = "oem"
	var expectedType = layout.FunctionTypeFunction
	var expectedVersion = "0.0.0"
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionArches.Key, expectedArches)
	testViper.Set(testOptions.optionBroker.Key, expectedBroker)
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
	assert.Regexp(t, regexp.MustCompile(`broker:[\s\t]*`+expectedBroker), actual)
}

func TestPublishExecutor_PretendNoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
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

func TestPublishExecutor_PretendRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
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

func TestPublishExecutor_PretendRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory:     newBuildTaskPretendStub(&calledBuild),
		initTaskFactory:      newInitTaskPretendStub(&calledInit),
		inspectTaskFactory:   newTaskPretendStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskPretendStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskPretendStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskPretendStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
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

func TestPublishExecutor_ProceedNoRebuildNoUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var expectedVersion = "0.0.0"
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedVersion, params.Version)
		}),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerVersion.Key, expectedVersion)
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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
	var expectedArches = []string{layout.ArchTypeX86, layout.ArchTypeArm64}
	var expectedHandle = "handleNoRebuildUpdate"
	var testDir = filepath.Join(t.TempDir(), expectedHandle)
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedArches, params.Arches)
			assert.Equal(t, expectedHandle, params.Handle)
		}),
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

			t.Chdir(testDir)
			testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
			testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
			testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
			testViper.Set(testOptions.optionArches.Key, expectedArches)
			testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
			testViper.Set(testExecutor.optionOem.Key, "oem")
			testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
			testLedger.Logger = &logrus.Logger{}
			testExecutor.Proceed()
			assert.False(t, calledBuild)
			assert.True(t, calledInspect)
			assert.True(t, calledProvision)
			assert.True(t, calledPush)
			assert.True(t, calledPublish)
			assert.True(t, calledInit)
		}
	}

	assert.NoError(t, err)
}

func TestPublishExecutor_ProceedRebuildUpdate(t *testing.T) {
	var calledBuild, calledInspect, calledProvision, calledPublish, calledInit, calledPush bool
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
		}),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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
	var expectedArches = []string{layout.ArchTypeX86, layout.ArchTypeArm64}
	var testDir = t.TempDir()
	var testBuildParams = &docker.BuildParams{}
	var testProvisionParams = &broker.ProvisionParams{}
	var testPublishParams = &broker.PublishParams{}
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewPublishOptions(testCli)
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
			Cli:    testCli,
		},
		PublishOptions: testOptions,

		buildTaskFactory: newBuildTaskCompleteStub(&calledBuild),
		initTaskFactory: newInitTaskCompleteStub(func(params *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedArches, params.Arches)
		}),
		inspectTaskFactory:   newTaskProceedStub(&calledInspect, testBuildParams),
		provisionTaskFactory: newTaskProceedStub(&calledProvision, testProvisionParams),
		publishTaskFactory:   newTaskProceedStub(&calledPublish, testPublishParams),
		pushTaskFactory:      newTaskProceedStub(&calledPush, &docker.PushParams{}),
	}

	if tmpFile, err := os.Create(filepath.Join(testDir, "GDockerfile")); err == nil {
		var fileName = tmpFile.Name()

		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testViper.Set(testExecutor.Cli.optionDockerFile.Key, fileName)
		testViper.Set(testExecutor.Cli.optionDockerContext.Key, filepath.Dir(fileName))
		testViper.Set(testOptions.optionArches.Key, expectedArches)
		testViper.Set(testOptions.optionRebuild.Key, true)
		testViper.Set(testOptions.optionNoUpdate.Key, true)
		testViper.Set(testOptions.optionType.Key, layout.FunctionTypeFunction)
		testViper.Set(testOptions.optionHandle.Key, "test-genaiz")
		testViper.Set(testExecutor.optionOem.Key, "oem")
		testViper.Set(testExecutor.optionVersion.Key, "0.0.0")
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
	var testPublish = NewPublish(testLedger, testCli)
	var expectedVersion = "0.0.0"
	var expectedHandle = "handle"
	var expectedHost = "host"

	testViper.Set(schema.Genaiz.Solution.Publish.Broker.Doc, expectedHost)
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
