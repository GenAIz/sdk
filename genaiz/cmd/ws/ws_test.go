package ws

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
)

type stubPrinter struct {
	err        error
	printError interface{}
	printOut   interface{}
}

func (s *stubPrinter) Error(i interface{}) error {
	s.printError = i
	return s.err
}

func (s *stubPrinter) Print(i interface{}) error {
	s.printOut = i
	return s.err
}

type stubPrinterParametric struct {
	defaultPrinter bool
	printer        cli.Printer
}

func (s stubPrinterParametric) IsDefault() bool {
	return s.defaultPrinter
}

func (s stubPrinterParametric) Printer() cli.Printer {
	return s.printer
}

func TestNewWs(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewWs(testLedger, nil, nil, nil)

	assert.Equal(t, 3, len(testCmd.Commands()))
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
