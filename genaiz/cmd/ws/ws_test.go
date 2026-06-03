package ws

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewWs(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWs(testLedger, nil, nil, nil)

	assert.Equal(t, 1, len(testCmd.Commands()))
}

func TestValidation_ArgsWorkspaceName(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWs(testLedger, nil, nil, nil)
	var testValidation = NewWsValidation()

	assert.NoError(t, testValidation.ArgsWorkspaceName(0)(testCmd, []string{"testName"}))
}

func TestValidation_ArgsWorkspaceName_NameError(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWs(testLedger, nil, nil, nil)
	var testValidation = NewWsValidation()
	var testString = "ABCDEFGHIJKLMNOPQRSTUVWXYZ"

	for i := 0; i < 10; i++ {
		testString += testString
	}

	assert.Error(t, testValidation.ArgsWorkspaceName(0)(testCmd, []string{testString}))
}

func TestValidation_ArgsWorkspaceName_OutOfRange(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWs(testLedger, nil, nil, nil)
	var testValidation = NewWsValidation()

	assert.Panics(t, func() { _ = testValidation.ArgsWorkspaceName(1)(testCmd, []string{"testString"}) })
}
