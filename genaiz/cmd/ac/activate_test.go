package ac

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestActivateExecutor_Activate(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedToken = "token"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &ActivateExecutor{
		ActivateOptions:     NewActivateOptions(),
		hostAddr:            "expectedHost",
		ledger:              testLedger,
		activateTaskFactory: newActivateCompletedTaskFactory(expectedToken),
	}

	defer patch.Unpatch()
	testLedger.InitLogging()
	testExecutor.Activate()
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, testExecutor.hostAddr)
}

func TestActivateExecutor_Activate_HostAndUser(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedToken = "token"
	var expectedUser = "user"
	var expectedHost = "host"
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testExecutor = &ActivateExecutor{
		ActivateOptions:     NewActivateOptions(),
		hostAddr:            fmt.Sprintf("%s@%s", expectedUser, expectedHost),
		ledger:              testLedger,
		activateTaskFactory: newActivateCompletedTaskFactory(expectedToken),
	}

	defer patch.Unpatch()
	testLedger.InitLogging()
	testExecutor.Activate()
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedHost)
	assert.Contains(t, patch.CalledWith, expectedUser)
}

func TestActivateExecutor_Activate_Username(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedToken = "token"
	var expectedUser = "user"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testExecutor = &ActivateExecutor{
		ActivateOptions:     NewActivateOptions(),
		hostAddr:            "expectedHost",
		ledger:              testLedger,
		activateTaskFactory: newActivateCompletedTaskFactory(expectedToken),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionUsername.Key, expectedUser)
	testLedger.InitLogging()
	testExecutor.Activate()
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedUser)
	assert.Contains(t, patch.CalledWith, testExecutor.hostAddr)
}

func TestNewActivate(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testLedger = config.NewBuilder().
		WithUserPath(t.TempDir()).
		WithViper(viper.New()).
		Build()
	var testCmd = NewActivate(testLedger)

	defer patch.Unpatch()
	testLedger.InitLogging()
	testCmd.Run(testCmd, []string{"hostString"})
	assert.Equal(t, 1, patch.CalledWith)
}

func newActivateCompletedTaskFactory(expectedToken string) func() *task.Task[broker.Broker] {
	return func() *task.Task[broker.Broker] {
		return &task.Task[broker.Broker]{
			Name: "test-activate",
			OnPrepare: func(params *broker.Broker, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.Broker, state *task.State) error {
				state.Output = expectedToken
				return nil
			},
		}
	}
}
