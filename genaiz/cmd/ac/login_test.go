package ac

import (
	"errors"
	"strings"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestLoginExecutor_Login_Oidc(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedHost = "expectedAddr"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &LoginExecutor{
		Ledger: testLedger,

		optionNoBrowser: cli.Options.Accounts.NoBrowser().BuildBoolOption(),
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),

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
		optionRefresh:   cli.Options.Accounts.Refresh().BuildBoolOption(),

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
