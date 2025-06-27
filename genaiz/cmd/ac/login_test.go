package ac

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-it/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestLoginExecutor_LoginExistingSession(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testRefreshOption = newOptionRefresh()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionRefresh: testRefreshOption,
		sessionTaskFactory: func() *task.Task[broker.Broker] {
			return &task.Task[broker.Broker]{
				Name: "test-session",
				OnPrepare: func(params *broker.Broker, state *task.State) error {
					return nil
				},
				OnComplete: func(params *broker.Broker, state *task.State) error {
					state.Output = expectedSession
					return nil
				},
			}
		},
	}

	defer patch.Unpatch()
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedSession)
}

func TestLoginExecutor_LoginExpiredSession(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPasswordOption = newOptionPassword()
	var testRefreshOption = newOptionRefresh()
	var testUsernameOption = newOptionUsername("test")
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionPassword: testPasswordOption,
		optionRefresh:  testRefreshOption,
		optionUsername: testUsernameOption,

		loginTaskFactory: func() *task.Task[broker.LoginParams] {
			return &task.Task[broker.LoginParams]{
				Name: "test-login",
				OnPrepare: func(params *broker.LoginParams, state *task.State) error {
					return nil
				},
				OnComplete: func(params *broker.LoginParams, state *task.State) error {
					state.Output = expectedSession
					return nil
				},
			}
		},
		sessionTaskFactory: func() *task.Task[broker.Broker] {
			return &task.Task[broker.Broker]{
				Name: "test-session",
				OnPrepare: func(params *broker.Broker, state *task.State) error {
					return errors.New("expired")
				},
			}
		},
	}

	defer patch.Unpatch()
	testViper.Set(testPasswordOption.Key, "password")
	testViper.Set(testUsernameOption.Key, "username")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedSession)
}

func TestLoginExecutor_LoginRefresh(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPasswordOption = newOptionPassword()
	var testRefreshOption = newOptionRefresh()
	var testUsernameOption = newOptionUsername("test")
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionPassword: testPasswordOption,
		optionRefresh:  testRefreshOption,
		optionUsername: testUsernameOption,

		loginTaskFactory: func() *task.Task[broker.LoginParams] {
			return &task.Task[broker.LoginParams]{
				Name: "test-login",
				OnPrepare: func(params *broker.LoginParams, state *task.State) error {
					return nil
				},
				OnComplete: func(params *broker.LoginParams, state *task.State) error {
					state.Output = expectedSession
					return nil
				},
			}
		},
	}

	defer patch.Unpatch()
	testViper.Set(testRefreshOption.Key, true)
	testViper.Set(testPasswordOption.Key, "password")
	testViper.Set(testUsernameOption.Key, "username")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedSession)
}

func TestLoginExecutor_queryPassword(t *testing.T) {
	var buff bytes.Buffer
	var testPasswordOption = newOptionPassword()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithInput(os.Stdin).
		WithOutput(io.Writer(&buff)).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionPassword: testPasswordOption,
	}

	assert.Empty(t, testExecutor.queryPassword())
}

func TestLoginExecutor_queryUsername(t *testing.T) {
	var buff bytes.Buffer
	var expectedUsername = "username"
	var testUsernameOption = newOptionUsername("test")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithInput(strings.NewReader(expectedUsername)).
		WithOutput(io.Writer(&buff)).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionUsername: testUsernameOption,
	}

	assert.EqualValues(t, expectedUsername, testExecutor.queryUsername())
}

func TestNewLogin_InvalidArgs(t *testing.T) {
	var loginCompleted = false
	var testLedger = config.NewBuilder().Build()
	var testLogin = NewLogin(testLedger)

	testLedger.Logger = &logrus.Logger{}
	testLogin.PostRun = func(cmd *cobra.Command, args []string) {
		loginCompleted = true
	}
	assert.Error(t, testLogin.Execute())
	assert.False(t, loginCompleted)
}

func TestNewLogin_InvalidBrokerAddr(t *testing.T) {
	var loginCompleted = false
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLogin = NewLogin(testLedger)
	var testPasswordOption = newOptionPassword()
	var testRefreshOption = newOptionRefresh()
	var testUsernameOption = newOptionUsername("Login")

	defer patch.Unpatch()
	testViper.Set(testRefreshOption.Key, true)
	testViper.Set(testPasswordOption.Key, "password")
	testViper.Set(testUsernameOption.Key, "username")
	testLedger.Logger = &logrus.Logger{}
	testLogin.SetArgs([]string{"localhost:1"})
	testLogin.PostRun = func(cmd *cobra.Command, args []string) {
		loginCompleted = true
	}
	assert.NoError(t, testLogin.Execute())
	assert.True(t, loginCompleted)
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}
