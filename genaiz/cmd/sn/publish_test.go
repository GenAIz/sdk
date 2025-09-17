package sn

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/docker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

var (
	testSolution = &testSnDoc{
		Solution: &broker.Solution{
			Handle: "handle",
			Workflows: []broker.Workflow{
				{
					Name:   "Test Workflow",
					Handle: "testWf",
					Nodes: []broker.WorkflowNode{
						{
							Handle: "emptyNode",
						},
						{
							Handle: "testNode",
							Sf: &broker.WorkflowNodeFunction{
								Oem:     "testOem",
								Handle:  "testHandle",
								Version: "0.1.0",
							},
						},
					},
				},
			},
		},
	}

	testFunction = &testSfDoc{
		Sf: &testSf{
			Publish: &broker.Function{
				Oem:     "testOem",
				Handle:  "testHandle",
				Type:    layout.FunctionTypeFunction,
				Version: "0.1.0",
			},
		},
	}
)

type testSf struct {
	Publish *broker.Function
}

type testSfDoc struct {
	Sf *testSf
}

type testSnDoc struct {
	Solution *broker.Solution
}

func TestPublishExecutor_Display(t *testing.T) {
	var testDir = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader: config.NewSolutionReader(testLedger),
	}
	var err error
	var fd *os.File

	testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
	testLedger.Logger = logrus.New()

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		var expectedVersion = "v1"
		var expectedSolution = &testSnDoc{Solution: &broker.Solution{Version: expectedVersion}}
		var data []byte
		var actual string

		if data, err = yaml.Marshal(expectedSolution); err == nil {
			var expectedHandle = "handle"
			var td = &testSfDoc{Sf: &testSf{Publish: &broker.Function{Handle: expectedHandle}}}
			var fnd1 *os.File
			var fnd2 *os.File

			defer filez.CloseSilently(fd)
			_, err = fd.Write(data)
			panicz.PanicIfError(err)

			panicz.PanicIfError(os.MkdirAll(filepath.Join(testDir, "function1"), 0750))
			fnd1, err = os.Create(filepath.Join(testDir, "function1", testLedger.ConfigName+"."+shared.ConfigTypeYaml))
			panicz.PanicIfError(err)
			defer filez.CloseSilently(fnd1)

			panicz.PanicIfError(os.MkdirAll(filepath.Join(testDir, "function2"), 0750))
			fnd2, err = os.Create(filepath.Join(testDir, "function2", testLedger.ConfigName+"."+shared.ConfigTypeYaml))
			panicz.PanicIfError(err)
			defer filez.CloseSilently(fnd2)
			data, err = yaml.Marshal(td)
			_, err = fnd2.Write(data)
			panicz.PanicIfError(err)

			testExecutor.Display()
			actual = testOutput.String()
			assert.NotEmpty(t, actual)
			assert.Contains(t, actual, fd.Name())
			assert.Contains(t, actual, expectedVersion)
			assert.NotContains(t, actual, fnd1.Name())
			assert.Contains(t, actual, fnd2.Name())
			return
		}
	}

	assert.NoError(t, err)
}

func TestPublishExecutor_Display_EmptySolution(t *testing.T) {
	var testDir = t.TempDir()
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader: config.NewSolutionReader(testLedger),
	}

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		var expectedVersion = "version"
		var actual string

		testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
		testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
		testExecutor.Display()
		actual = testOutput.String()
		assert.NotEmpty(t, actual)
		assert.Contains(t, actual, fd.Name())
		assert.Contains(t, actual, expectedVersion)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestPublishExecutor_Display_invalidConfigType(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
		},
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "invalid")
	testExecutor.Display()
	assert.Empty(t, testOutput.String())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPublishExecutor_Display_invalidFile(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
		},
		solutionReader: config.NewSolutionReader(testLedger),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
	testExecutor.Display()
	assert.Empty(t, testOutput.String())
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPublishExecutor_Pretend(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledInspect, calledProvision, calledPublish, calledPush, calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testHandleOption = newOptionHandle("test")
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType:  newOptionConfigType("test"),
			optionDescription: newOptionDescription(testHandleOption, "test"),
			optionHandle:      testHandleOption,
			optionName:        newOptionName(testHandleOption, "test"),
			optionOem:         newOptionOem("test"),
			optionVersion:     newOptionVersion("test"),
		},
		cmd:                        &cobra.Command{},
		solutionReader:             config.NewSolutionReader(testLedger),
		inspectTaskFactory:         newTaskPretendStub(&calledInspect, &docker.BuildParams{}),
		provisionTaskFactory:       newTaskPretendStub(&calledProvision, &broker.ProvisionParams{}),
		publishTaskFactory:         newTaskPretendStub(&calledPublish, &broker.PublishParams{}),
		pushTaskFactory:            newTaskPretendStub(&calledPush, &docker.PushParams{}),
		solutionPublishTaskFactory: newTaskPretendStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		var data []byte
		var fnd1 *os.File

		data, err = yaml.Marshal(testSolution)
		panicz.PanicIfError(err)
		_, err = fd.Write(data)
		panicz.PanicIfError(err)

		panicz.PanicIfError(os.MkdirAll(filepath.Join(testDir, "function1"), 0750))
		fnd1, err = os.Create(filepath.Join(testDir, "function1", testLedger.ConfigName+"."+shared.ConfigTypeYaml))
		panicz.PanicIfError(err)
		defer filez.CloseSilently(fnd1)
		data, err = yaml.Marshal(testFunction)
		panicz.PanicIfError(err)
		_, err = fnd1.Write(data)
		panicz.PanicIfError(err)

		testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
		testExecutor.Pretend()
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledSolutionPublish)
		assert.False(t, patch.Called)
	}
}

func TestPublishExecutor_Pretend_FileNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader:             config.NewSolutionReader(testLedger),
		solutionPublishTaskFactory: newTaskPretendStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
	testExecutor.Pretend()
	assert.False(t, calledSolutionPublish)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPublishExecutor_Pretend_SolutionNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader:             config.NewSolutionReader(testLedger),
		solutionPublishTaskFactory: newTaskPretendStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
		testExecutor.Pretend()
		assert.False(t, calledSolutionPublish)
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	}
}

func TestPublishExecutor_Proceed(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledInspect, calledProvision, calledPublish, calledPush, calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testHandleOption = newOptionHandle("test")
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType:  newOptionConfigType("test"),
			optionDescription: newOptionDescription(testHandleOption, "test"),
			optionHandle:      testHandleOption,
			optionName:        newOptionName(testHandleOption, "test"),
			optionOem:         newOptionOem("test"),
			optionVersion:     newOptionVersion("test"),
		},
		cmd:                        &cobra.Command{},
		solutionReader:             config.NewSolutionReader(testLedger),
		inspectTaskFactory:         newTaskProceedStub(&calledInspect, &docker.BuildParams{}),
		provisionTaskFactory:       newTaskProceedStub(&calledProvision, &broker.ProvisionParams{}),
		publishTaskFactory:         newTaskProceedStub(&calledPublish, &broker.PublishParams{}),
		pushTaskFactory:            newTaskProceedStub(&calledPush, &docker.PushParams{}),
		solutionPublishTaskFactory: newTaskProceedStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		var data []byte
		var fnd1 *os.File

		data, err = yaml.Marshal(testSolution)
		panicz.PanicIfError(err)
		_, err = fd.Write(data)
		panicz.PanicIfError(err)

		panicz.PanicIfError(os.MkdirAll(filepath.Join(testDir, "function1"), 0750))
		fnd1, err = os.Create(filepath.Join(testDir, "function1", testLedger.ConfigName+"."+shared.ConfigTypeYaml))
		panicz.PanicIfError(err)
		defer filez.CloseSilently(fnd1)
		data, err = yaml.Marshal(testFunction)
		panicz.PanicIfError(err)
		_, err = fnd1.Write(data)
		panicz.PanicIfError(err)

		testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
		testLedger.Logger = logrus.New()
		testExecutor.Proceed()
		assert.True(t, calledInspect)
		assert.True(t, calledProvision)
		assert.True(t, calledPush)
		assert.True(t, calledPublish)
		assert.True(t, calledSolutionPublish)
		assert.False(t, patch.Called)
	}
}

func TestPublishExecutor_Proceed_FileNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader:             config.NewSolutionReader(testLedger),
		solutionPublishTaskFactory: newTaskPretendStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
	testExecutor.Proceed()
	assert.False(t, calledSolutionPublish)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestPublishExecutor_Proceed_SolutionNotFound(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var calledSolutionPublish bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PublishExecutor{
		BaseExecutor: BaseExecutor{
			Ledger:     testLedger,
			folderPath: testDir,
		},
		PublishOptions: &PublishOptions{
			optionConfigType: newOptionConfigType("test"),
			optionVersion:    newOptionVersion("test"),
		},
		solutionReader:             config.NewSolutionReader(testLedger),
		solutionPublishTaskFactory: newTaskPretendStub(&calledSolutionPublish, &broker.SolutionPublishParams{}),
	}

	defer patch.Unpatch()

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)
		testViper.Set(testExecutor.optionConfigType.Key, shared.ConfigTypeYaml)
		testExecutor.Proceed()
		assert.False(t, calledSolutionPublish)
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	}
}

func TestNewFunctionOptions(t *testing.T) {
	var testOptions = NewFunctionOptions(&broker.Solution{})

	assert.NotEmpty(t, testOptions.optionArches)
	assert.NotEmpty(t, testOptions.optionDescription)
	assert.NotEmpty(t, testOptions.optionHandle)
	assert.NotEmpty(t, testOptions.optionName)
	assert.NotEmpty(t, testOptions.optionOem)
	assert.NotEmpty(t, testOptions.optionType)
	assert.NotEmpty(t, testOptions.optionVersion)
}

func TestNewPublish(t *testing.T) {
	var createCompleted = false
	var testOutput = new(bytes.Buffer)
	var testDir = t.TempDir()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(testOutput).
		WithViper(testViper).
		Build()
	var testCmd = NewPublish(testLedger, testCli)

	if fd, err := os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		defer filez.CloseSilently(fd)

		testViper.Set(newOptionConfigType("publish").Key, shared.ConfigTypeYaml)
		testCmd.PostRun = func(cmd *cobra.Command, args []string) {
			createCompleted = true
		}
		testCmd.SetArgs([]string{"host"})
		t.Chdir(testDir)
		assert.NoError(t, testCmd.Execute())
		assert.True(t, createCompleted)
	}
}

func TestNewPublish_DisappearingWorkdir(t *testing.T) {
	var testDir = t.TempDir()
	var testCli = &Cli{}
	var testLedger = config.NewLedger()
	var testCmd = NewPublish(testLedger, testCli)
	var patch = mock.Patches{T: t}.OsExit(func(int) {})

	defer patch.Unpatch()
	t.Chdir(testDir)
	panicz.PanicIfError(os.Remove(testDir))
	testCmd.Run(testCmd, []string{})
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewPublishOptions(t *testing.T) {
	var testOptions = NewPublishOptions()

	assert.NotEmpty(t, testOptions.optionConfigType)
	assert.NotEmpty(t, testOptions.optionDescription)
	assert.NotEmpty(t, testOptions.optionHandle)
	assert.NotEmpty(t, testOptions.optionName)
	assert.NotEmpty(t, testOptions.optionOem)
	assert.NotEmpty(t, testOptions.optionVersion)
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
