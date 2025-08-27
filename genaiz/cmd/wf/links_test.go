package wf

import (
	"bytes"
	"fmt"
	"io"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestLinksExecutor_Add(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedLink = "link"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewLinksOptions("test")
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newLinksAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Add(expectedWorkflow, []string{expectedLink})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`add-0:[\s\t]*`+expectedLink), actual)
}

func TestLinksExecutor_Display(t *testing.T) {
	var expectedWorkflow = "workflow"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		workflowArg:  expectedWorkflow,
		LinksOptions: testOptions,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
}

func TestLinksExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestLinksExecutor_PretendInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_PretendInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.InitLogging()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestLinksExecutor_ProceedInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_ProceedInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewLinksOptions("test")
	var testExecutor = &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		LinksOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLinksExecutor_Remove(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedLink = "link"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewLinksOptions("test")
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newLinksRemoveExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(&cobra.Command{}, testOptions.allDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Remove(expectedWorkflow, []string{expectedLink})
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`rm-0:[\s\t]*`+expectedLink), actual)
}

func TestNewLinks(t *testing.T) {
	var testLedger = config.NewLedger()
	var testCmd = NewLinks(testLedger, &Cli{})

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func Test_parseArgsLinks(t *testing.T) {
	var expectedInvalid = "invalid"
	var expectedLeft = "left"
	var expectedPort = "port"
	var expectedRight = "right"
	var testLinks = []string{
		expectedInvalid + "[nil]",
		fmt.Sprintf("%s[%s]:%s", expectedLeft, expectedPort, expectedRight),
	}
	var actualLinks = parseArgsLinks(testLinks...)

	assert.Equal(t, 1, len(actualLinks))
	assert.Equal(t, expectedLeft, actualLinks[0].LhsNode)
	assert.Equal(t, expectedPort, actualLinks[0].LhsNodePort)
	assert.Equal(t, expectedRight, actualLinks[0].RhsNode)
}

func Test_validateArgsLinks(t *testing.T) {
	assert.Error(t, validateArgLinks([]string{"valid:valid", "notValid[]"}))
	assert.NoError(t, validateArgLinks([]string{"valid1:valid2[port]"}))
}

func newWorkflowWriterStub(*config.Ledger, string) *workflowWriter {
	var stub = &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter(),
	}

	stub.WithCurrent(&broker.Solution{
		Workflows: []broker.Workflow{
			{
				Handle: "workflow",
			},
		},
	})
	return stub
}
