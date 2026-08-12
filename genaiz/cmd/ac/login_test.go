package ac

import (
	"bytes"
	"errors"
	"io"
	"os"
	"strings"
	"testing"

	"github.com/awnumar/memguard"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestLoginExecutor_Login_ExistingSession(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionRefresh: cli.Options.Accounts.Refresh().BuildBoolOption(),
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
	assert.Contains(t, cast.ToStringSlice(patch.CalledWith)[0], "logged in")
}

func TestLoginExecutor_Login_ExpiredSession(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPasswordOption = cli.Options.Accounts.Password().BuildStringOption()
	var testUsernameOption = cli.Options.Accounts.Username().
		WithKeys(&schema.Genaiz.Account.Login.Username).
		BuildStringOption()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionPassword:  testPasswordOption,
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),
		optionUsername:  testUsernameOption,

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
		oidcTaskFactory: newOidcNotSupportedTaskFactory(),
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
	assert.Contains(t, strings.ToLower(cast.ToStringSlice(patch.CalledWith)[0]), "logged in")
}

func TestLoginExecutor_Login_Refresh(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testPasswordOption = cli.Options.Accounts.Password().BuildStringOption()
	var testRefreshOption = cli.Options.Accounts.Refresh().BuildBoolOption()
	var testUsernameOption = cli.Options.Accounts.Username().
		WithKeys(&schema.Genaiz.Account.Login.Username).
		BuildStringOption()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionPassword:  testPasswordOption,
		optionRefresh:   testRefreshOption,
		optionUsername:  testUsernameOption,

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
		oidcTaskFactory: newOidcNotSupportedTaskFactory(),
	}

	defer patch.Unpatch()
	testViper.Set(testRefreshOption.Key, true)
	testViper.Set(testPasswordOption.Key, "password")
	testViper.Set(testUsernameOption.Key, "username")
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, strings.ToLower(cast.ToStringSlice(patch.CalledWith)[0]), "logged in")
}

func TestLoginExecutor_Login_Oidc(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionPassword:  cli.Options.Accounts.Password().BuildStringOption(),
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Login.Username).
			BuildStringOption(),

		oidcTaskFactory: newOidcCompletedTaskFactory(),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionRefresh.Key, true)
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, strings.ToLower(cast.ToStringSlice(patch.CalledWith)[0]), "logged in")
}

func TestLoginExecutor_Login_OidcError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHost = "expectedAddr"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionPassword:  cli.Options.Accounts.Password().BuildStringOption(),
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),
		optionUsername: cli.Options.Accounts.Username().
			WithKeys(&schema.Genaiz.Account.Login.Username).
			BuildStringOption(),

		oidcTaskFactory: func() *task.Task[broker.OidcParams] {
			return &task.Task[broker.OidcParams]{
				Name: "oidc-error",
				OnPrepare: func(params *broker.OidcParams, state *task.State) error {
					return errors.New("expected")
				},
			}
		},
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionRefresh.Key, true)
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLoginExecutor_queryPassword(t *testing.T) {
	var buff bytes.Buffer
	var testPasswordOption = cli.Options.Accounts.Password().BuildStringOption()
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

func TestLoginExecutor_queryPassword_fromEnv(t *testing.T) {
	var buff bytes.Buffer
	var testPasswordOption = cli.Options.Accounts.Password().BuildStringOption()
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
	var expectedPassword = "testPassword"
	var err error

	if err = os.Setenv(testPasswordOption.Env, expectedPassword); err == nil {
		var locked *memguard.LockedBuffer

		defer func() { _ = os.Unsetenv(testPasswordOption.Env) }()

		if locked, err = testExecutor.queryPassword().Open(); err == nil {
			defer locked.Destroy()
			assert.Equal(t, expectedPassword, locked.String())
			return
		}
	}

	assert.Fail(t, err.Error())
}

func TestLoginExecutor_queryUsername(t *testing.T) {
	var buff bytes.Buffer
	var expectedUsername = "username"
	var testUsernameOption = cli.Options.Accounts.Username().BuildStringOption()
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
	var testPasswordOption = cli.Options.Accounts.Password().BuildStringOption()
	var testRefreshOption = cli.Options.Accounts.Refresh().BuildBoolOption()
	var testUsernameOption = cli.Options.Accounts.Username().
		WithKeys(&schema.Genaiz.Account.Login.Username).
		BuildStringOption()

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

func newOidcCompletedTaskFactory() func() *task.Task[broker.OidcParams] {
	return func() *task.Task[broker.OidcParams] {
		return &task.Task[broker.OidcParams]{
			Name: "test-oidc",
			OnPrepare: func(params *broker.OidcParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.OidcParams, state *task.State) error {
				return nil
			},
		}
	}
}

func newOidcNotSupportedTaskFactory() func() *task.Task[broker.OidcParams] {
	return func() *task.Task[broker.OidcParams] {
		return &task.Task[broker.OidcParams]{
			Name: "test-oidc",
			OnPrepare: func(params *broker.OidcParams, state *task.State) error {
				return broker.ErrorOidcNotSupported
			},
		}
	}
}
