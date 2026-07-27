package ac

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
)

type stubCliPrinterParametric struct {
	stubPrinter cli.Printer
	stubDefault *bool
}

func (s stubCliPrinterParametric) Printer() cli.Printer {
	if s.stubPrinter == nil {
		return &stubCliPrinter{}

	}

	return s.stubPrinter
}

func (s stubCliPrinterParametric) IsDefault() bool {
	if s.stubDefault == nil {
		return true
	}

	return *s.stubDefault
}

type stubCliPrinter struct {
	errorError   error
	errorPayload interface{}
	printError   error
	printPayload interface{}
}

func (s *stubCliPrinter) Error(i interface{}) error {
	s.errorPayload = i
	return s.errorError
}

func (s *stubCliPrinter) Print(i interface{}) error {
	s.printPayload = i
	return s.printError
}

func TestNewAc(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testAc = NewAc(testLedger)

	assert.Equal(t, 5, len(testAc.Commands()))
	assert.NoError(t, testAc.Execute())
}
