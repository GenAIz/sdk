package wf

import (
	"bytes"
	"io"
	"regexp"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/shared"
)

func TestDeleteExecutor_Display(t *testing.T) {
	var expected = "workflow"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewDeleteOptions(NewWfCli(nil, nil, nil))
	var testExecutor = &DeleteExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},

		workflowArg:   expected,
		DeleteOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expected), actual)
}

func TestDeleteExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DeleteExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DeleteOptions: NewDeleteOptions(NewWfCli(nil, nil, nil)),

		workflowArg: "test",

		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testLedger.InitLogging()
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestDeleteExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DeleteExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DeleteOptions: NewDeleteOptions(NewWfCli(nil, nil, nil)),

		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	testViper.Set(testExecutor.optionConfigType.Key, "yaml")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestDeleteExecutor_Proceed_InvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &DeleteExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		DeleteOptions: NewDeleteOptions(NewWfCli(nil, nil, nil)),

		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionConfigType.Key, "invalid")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewDelete(t *testing.T) {
	var deleteCompleted = false
	var testOutput = new(bytes.Buffer)
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
	var testDelete = NewDelete(testLedger, testCli)
	var expectedFolder = "test-folder"

	testDelete.PostRun = func(cmd *cobra.Command, args []string) {
		deleteCompleted = true
	}
	testDelete.SetArgs([]string{expectedFolder})
	assert.NoError(t, testDelete.Execute())
	assert.True(t, deleteCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, expectedFolder)
	} else {
		assert.Fail(t, "no --dry content")
	}
}
