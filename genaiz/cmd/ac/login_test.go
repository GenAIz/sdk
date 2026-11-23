package ac

import (
	"bytes"
	"errors"
	"io"
	"strings"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestLoginExecutor_Login_LoggedIn(t *testing.T) {
	var expectedHost = "expectedAddr"
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithOutput(io.Writer(testOutput)).
		WithLogger(getTestLogger()).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger:       testLedger,
		LoginOptions: NewLoginOptions(),

		sessionTaskFactory: newSessionCompletedTaskFactory(expectedHost),
	}

	testLedger.InitLogging()
	testExecutor.Login(expectedHost)
	actual := testOutput.String()
	assert.Contains(t, actual, "Already logged in to")
	assert.Contains(t, actual, expectedHost)
}

func TestLoginExecutor_Login_Oidc(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithLogger(getTestLogger()).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger:       testLedger,
		LoginOptions: NewLoginOptions(),

		oidcTaskFactory: newOidcCompletedTaskFactory(),
	}

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionRefresh.Key, true)
	testLedger.InitLogging()
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, strings.ToLower(cast.ToStringSlice(patch.CalledWith)[0]), "logged in")
}

func TestLoginExecutor_Login_OidcError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var expectedHost = "expectedAddr"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithLogger(getTestLogger()).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger:       testLedger,
		LoginOptions: NewLoginOptions(),

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
	testLedger.InitLogging()
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, patch.CalledWith)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestLoginExecutor_Login_Version(t *testing.T) {
	var capturedParams broker.OidcParams
	var expectedHost = "expectedAddr"
	var expectedPrefix = "prefix"
	var expectedVersion = "2"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithLogger(getTestLogger()).
		WithViper(testViper).
		Build()
	var testExecutor = &LoginExecutor{
		Ledger:       testLedger,
		LoginOptions: NewLoginOptions(),

		oidcTaskFactory: newOidcCapturedTaskFactory(&capturedParams),
	}

	testViper.Set(testExecutor.optionRefresh.Key, true)
	testViper.Set(testExecutor.optionVersion.Key, expectedVersion)
	testViper.Set(testExecutor.optionPrefix.Key, expectedPrefix)
	testLedger.InitLogging()
	testExecutor.Login(expectedHost)
	assert.NotEmpty(t, capturedParams)
	assert.Equal(t, expectedPrefix, capturedParams.ApiPrefix)
	assert.True(t, strings.HasSuffix(capturedParams.ApiVersion, expectedVersion))
}

func TestNewLogin_Unsupported(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithLogger(getTestLogger()).
		WithViper(testViper).
		Build()
	var testLogin = NewLogin(testLedger)

	defer patch.Unpatch()
	testLedger.InitLogging()
	testLogin.Run(testLogin, []string{"testHost"})
	assert.True(t, patch.Called)
	assert.Equal(t, patch.CalledWith, 1)
}

func newOidcCapturedTaskFactory(captured *broker.OidcParams) func() *task.Task[broker.OidcParams] {
	return func() *task.Task[broker.OidcParams] {
		return &task.Task[broker.OidcParams]{
			Name: "test-oidc",
			OnPrepare: func(params *broker.OidcParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.OidcParams, state *task.State) error {
				*captured = *params
				return nil
			},
		}
	}
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

func newSessionCompletedTaskFactory(expected string) func() *task.Task[broker.Broker] {
	return func() *task.Task[broker.Broker] {
		return &task.Task[broker.Broker]{
			Name: "test-session",
			OnPrepare: func(params *broker.Broker, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.Broker, state *task.State) error {
				state.Output = expected
				return nil
			},
		}
	}
}
