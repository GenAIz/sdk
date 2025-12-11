package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
)

func TestDataExecutor_Add(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var expectedHandle = "portHandle"
	var expectedName = "portName"
	var expectedDesc = "portDesc"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testInputOptions = NewDataOptionsInput()
	var testOutputOptions = NewDataOptionsOutput()
	var testExecutor = newDataExecutorFactory(testLedger, testCli, testInputOptions, testOutputOptions)(&cobra.Command{})

	testViper.Set(testOutputOptions.optionDesc.Key, expectedDesc)
	testViper.Set(testOutputOptions.optionName.Key, expectedName)
	assert.NoError(t, testExecutor.Add(outputPortType, expectedHandle))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testInputOptions.optionDesc.Param+`:[\s\t]*`+expectedDesc), actual)
	assert.Regexp(t, regexp.MustCompile(testInputOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
}

func TestDataExecutor_Add_InvalidHandle(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var expectedHandle = "portHandle/notValid"
	var expectedDesc = "portDesc"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		inputOptions: NewDataOptionsInput(),

		updatedPorts: map[string][]broker.DataPort{},
	}

	testViper.Set(testExecutor.inputOptions.optionDesc.Key, expectedDesc)
	assert.Error(t, testExecutor.Add(inputPortType, expectedHandle))
	assert.Empty(t, testExecutor.addedPort)
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestDataExecutor_Add_NoName(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var expectedHandle = "portHandle"
	var expectedDesc = "portDesc"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},
		inputOptions: NewDataOptionsInput(),

		updatedPorts: map[string][]broker.DataPort{},
	}

	testViper.Set(testExecutor.inputOptions.optionDesc.Key, expectedDesc)
	assert.NoError(t, testExecutor.Add(inputPortType, expectedHandle))
	assert.Equal(t, expectedHandle, testExecutor.addedPort.Handle)
	assert.Equal(t, expectedDesc, testExecutor.addedPort.Description)
	assert.Equal(t, 1, len(testExecutor.updatedPorts))
	assert.Equal(t, expectedHandle, testExecutor.updatedPorts[inputPortType][0].Handle)
	assert.Equal(t, expectedDesc, testExecutor.updatedPorts[inputPortType][0].Description)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+expectedHandle), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.inputOptions.optionDesc.Param+`:[\s\t]*`+expectedDesc), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.inputOptions.optionName.Param+`:[\s\t]*`+expectedHandle), actual)
}

func TestDataExecutor_Add_PanicOnDataType(t *testing.T) {
	var testExecutor = &DataExecutor{}

	assert.Panics(t, func() {
		_ = testExecutor.Add("invalid", "handle")
	})
}

func TestDataExecutor_Add_PortNotFound(t *testing.T) {
	var expectedDataPort = broker.DataPort{Handle: "portHandle"}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}

	testViper.Set(inputPortsWriter.portOption.Key, []interface{}{
		map[string]interface{}{
			"handle": expectedDataPort.Handle,
		},
	})
	assert.Error(t, testExecutor.Add(inputPortType, expectedDataPort.Handle))
}

func TestDataExecutor_Display_Nothing(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testInputOptions = NewDataOptionsInput()
	var testOutputOptions = NewDataOptionsOutput()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewDataExecutor(testCmd.Context(), testLedger, testCli, testInputOptions, testOutputOptions)

	testExecutor.Display()
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestDataExecutor_Init(t *testing.T) {
	var expectedHandle = "portHandle"
	var testDir = t.TempDir()
	var testPath = filepath.Join(testDir, "input", expectedHandle)
	var expectedRunInput = filepath.Join(testDir, "input")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}
	var actual string
	var err error

	if err = os.MkdirAll(filepath.Join(expectedRunInput, expectedHandle), 0750); err == nil {
		testViper.Set(inputPortsWriter.runOption.Key, expectedRunInput)
		actual, err = testExecutor.Init(inputPortType, testPath)
		assert.NoError(t, err)
		assert.Equal(t, expectedHandle, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestDataExecutor_Init_PanicOnDataType(t *testing.T) {
	var testExecutor = &DataExecutor{}

	assert.Panics(t, func() {
		_, _ = testExecutor.Init("invalid", "handle")
	})
}

func TestDataExecutor_Init_PathIsFile(t *testing.T) {
	var testDir = t.TempDir()
	var expectedRunInput = filepath.Join(testDir, "input")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}
	var actual string
	var err error

	if err = os.MkdirAll(expectedRunInput, 0750); err == nil {
		var testPath = filepath.Join(testDir, "input", "portHandle")
		var fd *os.File

		if fd, err = os.Create(testPath); err == nil {
			defer filez.CloseSilently(fd)
			testViper.Set(inputPortsWriter.runOption.Key, expectedRunInput)
			actual, err = testExecutor.Init(inputPortType, testPath)
			assert.Error(t, err)
			assert.Empty(t, actual)
			return
		}
	}

	assert.NoError(t, err)
}

func TestDataExecutor_Init_PathNotExist(t *testing.T) {
	var testDir = t.TempDir()
	var expectedRunInput = filepath.Join(testDir, "input")
	var expectedHandle = "portHandle"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}

	testViper.Set(inputPortsWriter.runOption.Key, expectedRunInput)
	actual, err := testExecutor.Init(inputPortType, filepath.Join(testDir, "input", expectedHandle))
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual)
}

func TestDataExecutor_Init_PathNotWriteable(t *testing.T) {
	var testDir = t.TempDir()
	var expectedRunInput = filepath.Join(testDir, "input")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}
	var actual string
	var err error

	if err = os.MkdirAll(expectedRunInput, 0222); err == nil {
		testViper.Set(inputPortsWriter.runOption.Key, expectedRunInput)
		actual, err = testExecutor.Init(inputPortType, filepath.Join(testDir, "input", "portHandle"))
		assert.Error(t, err)
		assert.Empty(t, actual)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestDataExecutor_Init_RunFolderDiverge(t *testing.T) {
	var expectedHandle = "portHandle"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}

	testViper.Set(inputPortsWriter.runOption.Key, "noWhere/common")
	actual, err := testExecutor.Init(inputPortType, expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual)
}

func TestDataExecutor_Init_RunFolderNotParent(t *testing.T) {
	var testDir = t.TempDir()
	var expectedRunInput = filepath.Join(testDir, "input")
	var expectedHandle = filepath.Join(testDir, "portHandle")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}

	testViper.Set(inputPortsWriter.runOption.Key, expectedRunInput)
	actual, err := testExecutor.Init(inputPortType, expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual)
}

func TestDataExecutor_Init_RunFolderNotSet(t *testing.T) {
	var expectedHandle = "portHandle"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
	}

	actual, err := testExecutor.Init(inputPortType, expectedHandle)
	assert.NoError(t, err)
	assert.Equal(t, expectedHandle, actual)
}

func TestDataExecutor_Pretend_MultiPortTypes(t *testing.T) {
	var calledInit bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    NewSfCli(nil, nil, nil),
			Ledger: testLedger,
		},

		inputOptions:  NewDataOptionsInput(),
		outputOptions: NewDataOptionsOutput(),

		updatedPorts: map[string][]broker.DataPort{
			inputPortType:  {{Handle: "inputHandle"}},
			outputPortType: {{Handle: "outputHandle"}},
		},

		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		defer filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Pretend()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestDataExecutor_Proceed_MultiPortTypes(t *testing.T) {
	var calledInit bool
	var expectedInputPort = broker.DataPort{Handle: "inputHandle"}
	var expectedOutputPort = broker.DataPort{Handle: "outputHandle"}
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    NewSfCli(nil, nil, nil),
			Ledger: testLedger,
		},

		inputOptions:  NewDataOptionsInput(),
		outputOptions: NewDataOptionsOutput(),

		updatedPorts: map[string][]broker.DataPort{
			inputPortType:  {expectedInputPort},
			outputPortType: {expectedOutputPort},
		},

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {
			calledInit = true
			assert.Equal(t, []broker.DataPort{expectedInputPort}, actual.InputPorts)
			assert.Equal(t, []broker.DataPort{expectedOutputPort}, actual.OutputPorts)
		}),
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		defer filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.InitLogging()
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Proceed()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestDataExecutor_Remove(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var expectedHandle = "portHandle"
	var testHandle = "toRemove"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testExecutor = &DataExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    testCli,
			Ledger: testLedger,
		},

		updatedPorts: make(map[string][]broker.DataPort),
	}

	testViper.Set(inputPortsWriter.portOption.Key, []interface{}{
		map[string]interface{}{"handle": expectedHandle},
		map[string]interface{}{"handle": testHandle},
	})
	err := testExecutor.Remove(inputPortType, testHandle)
	assert.NoError(t, err)
	assert.True(t, strings.EqualFold(testHandle, testExecutor.removedPort.Handle))
	assert.Equal(t, 1, len(testExecutor.updatedPorts))
	assert.True(t, strings.EqualFold(expectedHandle, testExecutor.updatedPorts[inputPortType][0].Handle))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(`handle:[\s\t]*`+strings.ToLower(testHandle)), actual)
}

func TestDataExecutor_Remove_PanicOnDataType(t *testing.T) {
	var testExecutor = &DataExecutor{}

	assert.Panics(t, func() {
		_ = testExecutor.Remove("invalid", "handle")
	})
}

func TestNewData(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = NewData(testLedger, testCli)

	assert.Empty(t, testCmd.Run)
	assert.Equal(t, 4, len(testCmd.Commands()))
}
