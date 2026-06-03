package config

import (
	"fmt"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
)

func TestAccountParams_BrokerParams(t *testing.T) {
	var testViper = viper.New()
	var testLedger = NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOption = &StringOption{Option: Option{Key: "key"}}
	var testParametric = NewAccountParams(testLedger, testOption)
	var expectedAccount = "account"

	testViper.Set(testOption.Key, expectedAccount)
	actual := testParametric.BrokerParams()
	assert.Equal(t, testLedger.AuthFile, actual.AuthFile)
	assert.Equal(t, expectedAccount, actual.HostAddr)
}

func TestAccountParams_BrokerParams_NoAccount(t *testing.T) {
	var testViper = viper.New()
	var testLedger = NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOption = &StringOption{Option: Option{Key: "key"}}
	var testParametric = NewAccountParams(testLedger, testOption)
	var actual = testParametric.BrokerParams()

	assert.Equal(t, testLedger.AuthFile, actual.AuthFile)
}

func TestAccountParams_BrokerParams_UserValue(t *testing.T) {
	var testViper = viper.New()
	var testLedger = NewBuilder().
		WithViper(testViper).
		WithUserPath(t.TempDir()).
		Build()
	var testOption = &StringOption{Option: Option{Key: "key"}}
	var testParametric = NewAccountParams(testLedger, testOption)
	var expectedAccount = "account"
	var expectedUsername = "username"

	testViper.Set(testOption.Key, fmt.Sprintf("%s@%s", expectedUsername, expectedAccount))
	actual := testParametric.BrokerParams()
	assert.Equal(t, testLedger.AuthFile, actual.AuthFile)
	assert.Equal(t, expectedAccount, actual.HostAddr)
	assert.Equal(t, expectedUsername, actual.Username)
}
