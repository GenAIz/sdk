package ac

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestLogoutExecutor_Logout(t *testing.T) {
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var expectedSession = "expectedSession"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &LogoutExecutor{
		Ledger: testLedger,

		optionHost:     newOptionHost(),
		optionUsername: newOptionUsername("Test"),

		logoutTaskFactory: func() *task.Task[broker.LoginParams] {
			return &task.Task[broker.LoginParams]{
				Name: "test-logout",
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
	testLedger.Logger = &logrus.Logger{}
	testExecutor.Logout()
	assert.NotEmpty(t, patch.CalledWith)
	assert.Contains(t, patch.CalledWith, expectedSession)
}

func TestNewLogout(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testLogout = NewLogout(testLedger)

	if tmpFile, err := os.CreateTemp("", "genaiz.auth"); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(i int) {})

		defer filez.RemoveSilently(tmpFile.Name())
		defer patch.Unpatch()
		testLedger.Logger = &logrus.Logger{}
		testLedger.AuthFile = tmpFile.Name()
		assert.NoError(t, testLogout.Execute())
		assert.NotEmpty(t, patch.CalledWith)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}
