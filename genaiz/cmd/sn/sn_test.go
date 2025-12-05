package sn

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
)

func TestBaseExecutor_makeConfigParams_InvalidConfigType(t *testing.T) {
	var testOption = cli.NewOptionBuilder().
		WithKeys(&schema.Keys{Doc: "testKey"}).
		BuildStringOption()
	var testViper = viper.New()
	var testExecutor = &BaseExecutor{
		Ledger: config.NewBuilder().
			WithViper(testViper).
			Build(),
	}

	testViper.Set(testOption.Key, "invalid")
	actual, err := testExecutor.makeConfigParams(testOption)
	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestNewSn(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewSn(testLedger, nil, nil, nil)

	assert.Equal(t, 2, len(testCmd.Commands()))
}

func TestNewSnCli(t *testing.T) {
	var expectedConfirm = "confirm"
	var expectedDry = "dry"
	var expectedPretend = "pretend"
	var actual string
	var testCli = NewSnCli(func(*config.Ledger, ...func()) bool {
		actual = expectedConfirm
		return true
	}, func(ledger *config.Ledger) bool {
		actual = expectedDry
		return true
	}, func(ledger *config.Ledger) bool {
		actual = expectedPretend
		return true
	})

	assert.True(t, testCli.Confirm(nil))
	assert.Equal(t, expectedConfirm, actual)
	assert.True(t, testCli.Dry(nil))
	assert.Equal(t, expectedDry, actual)
	assert.True(t, testCli.Pretend(nil))
	assert.Equal(t, expectedPretend, actual)
}
