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

func TestNodesExecutor_Add(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Add(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
}

func TestNodesExecutor_AddOnlyVersion(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfVersion.Key, "0.0.1")
	testExecutor.Add(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
	assert.NotContains(t, actual, "version")
}

func TestNodesExecutor_AddOnlyVersionOem(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfOem.Key, "genaiz.com")
	testViper.Set(testOptions.optionSfVersion.Key, "0.0.1")
	testExecutor.Add(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
	assert.NotContains(t, actual, "version")
	assert.NotContains(t, actual, "oem")
}

func TestNodesExecutor_AddWithSf(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var expectedSfHandle = "handle"
	var expectedSfOem = "genaiz.com"
	var expectedSfVersion = "0.0.1"
	var expectedSfRc = "4"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testViper.Set(testOptions.optionSfSerialized.Key, fmt.Sprintf("%s/%s:%s-rc%s", expectedSfOem, expectedSfHandle, expectedSfVersion, expectedSfRc))
	testExecutor.Add(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.handle:[\s\t]*`+expectedSfHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.oem:[\s\t]*`+expectedSfOem), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.version:[\s\t]*`+expectedSfVersion), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.seq:[\s\t]*`+expectedSfRc), actual)
}

func TestNodesExecutor_Display(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var expectedSfHandle = "_sf_handle"
	var expectedSfOem = "_sf_oem"
	var expectedSfVersion = "_sf_version"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddNodesOptions()
	var testCmd = &cobra.Command{}
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions: testOptions,
		addNode: &broker.WorkflowNode{
			Handle: expectedNode,
			Sf: &broker.WorkflowNodeFunction{
				Handle:  expectedSfHandle,
				Oem:     expectedSfOem,
				Version: expectedSfVersion,
				Seq:     0,
			},
		},
		workflowArg: expectedWorkflow,
	}

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.description:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.handle:[\s\t]*`+expectedNode), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.name:[\s\t]*`), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.handle:[\s\t]*`+expectedSfHandle), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.oem:[\s\t]*`+expectedSfOem), actual)
	assert.Regexp(t, regexp.MustCompile(`node.add.sf.version:[\s\t]*`+expectedSfVersion), actual)
}

func TestNodesExecutor_Pretend(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions: testOptions,
		workflowArg:  "workflow",
		addNode: &broker.WorkflowNode{
			Handle: "handle",
		},
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.True(t, calledWorkflow)
}

func TestNodesExecutor_PretendInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskPretendStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_PretendInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewAddNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskPretendStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Pretend()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_Proceed(t *testing.T) {
	var calledWorkflow bool
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowArg:           "workflow",
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	testLedger.InitLogging()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.True(t, calledWorkflow)
}

func TestNodesExecutor_ProceedInvalidConfigType(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:        testOptions,
		workflowTaskFactory: newWorkflowTaskCompleteStub(&calledWorkflow),
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_ProceedInvalidParams(t *testing.T) {
	var calledWorkflow bool
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = NewRemoveNodesOptions()
	var testExecutor = &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Ledger: testLedger,
		},
		NodesOptions:          testOptions,
		workflowTaskFactory:   newWorkflowTaskCompleteStub(&calledWorkflow),
		workflowWriterFactory: newWorkflowWriterStub,
	}

	defer patch.Unpatch()
	testLedger.Register(&cobra.Command{}, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Proceed()
	assert.False(t, calledWorkflow)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNodesExecutor_Remove(t *testing.T) {
	var expectedWorkflow = "workflow"
	var expectedNode = "node"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewRemoveNodesOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newNodesRemoveExecutorFactory(testLedger, testCli, testOptions)(testCmd)

	testLedger.Register(testCmd, testOptions.removeDefiners()...)
	testViper.Set(testOptions.optionConfigType.Key, shared.ConfigTypeJson)
	testExecutor.Remove(expectedWorkflow, expectedNode)
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionConfigType.Param+`:[\s\t]*`+shared.ConfigTypeJson), actual)
	assert.Regexp(t, regexp.MustCompile(`workflow:[\s\t]*`+expectedWorkflow), actual)
	assert.Regexp(t, regexp.MustCompile(`node.remove\[0\].handle:[\s\t]*`+expectedNode), actual)
}

func TestNewNodes(t *testing.T) {
	var testLedger = config.NewLedger()
	var testCmd = NewNodes(testLedger, &Cli{})

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func Test_newOptionSfDeserialized_dashSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOption = newOptionSfSerialized("test")
	var testOption = newOptionSfDeserialized(serializedOption, "test")
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOption.Key, fmt.Sprintf("%s/%s:%s-rc-%d", expectedOem, expectedHandle, expectedVersion, expectedSeq))
	actual = testOption.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func Test_newOptionSfDeserialized_dotSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOption = newOptionSfSerialized("test")
	var testOption = newOptionSfDeserialized(serializedOption, "test")
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOption.Key, fmt.Sprintf("%s/%s:%s-rc.%d", expectedOem, expectedHandle, expectedVersion, expectedSeq))
	actual = testOption.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func Test_newOptionSfDeserialized_noOem(t *testing.T) {
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var expectedSeq = 37
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOption = newOptionSfSerialized("test")
	var testOption = newOptionSfDeserialized(serializedOption, "test")
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOption.Key, fmt.Sprintf("%s:%s-rc%d", expectedHandle, expectedVersion, expectedSeq))
	actual = testOption.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
	assert.Equal(t, expectedSeq, actual.Seq)
}

func Test_newOptionSfDeserialized_noSeq(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var expectedVersion = "0.0.1"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOption = newOptionSfSerialized("test")
	var testOption = newOptionSfDeserialized(serializedOption, "test")
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOption.Key, fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion))
	actual = testOption.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
	assert.Equal(t, expectedVersion, actual.Version)
}

func Test_newOptionSfDeserialized_noVersion(t *testing.T) {
	var expectedOem = "genaiz.com"
	var expectedHandle = "handle_test"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var serializedOption = newOptionSfSerialized("test")
	var testOption = newOptionSfDeserialized(serializedOption, "test")
	var actual *broker.WorkflowNodeFunction

	testViper.Set(serializedOption.Key, fmt.Sprintf("%s/%s", expectedOem, expectedHandle))
	actual = testOption.DefaultGetter(testLedger).(*broker.WorkflowNodeFunction)
	assert.Equal(t, expectedOem, actual.Oem)
	assert.Equal(t, expectedHandle, actual.Handle)
}

func Test_validateArgNodes(t *testing.T) {
	assert.NoError(t, validateArgNodes("handle", "more-handle"))
	assert.Error(t, validateArgNodes("valid.handle", "_invalid_handle"))
}
