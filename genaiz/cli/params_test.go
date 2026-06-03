package cli

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
)

func TestPrinterParametric_DefaultPrinter(t *testing.T) {
	var testKey = "testKey"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = NewOptionBuilder().WithKeys(&schema.Keys{Doc: testKey}).BuildBoolOption()
	var testParametric PrinterParametric

	testViper.Set(testKey, "false")
	testParametric = NewPrinterParam(testLedger, testOption)
	assert.True(t, testParametric.IsDefault())
	assert.NotNil(t, testParametric.Printer())
}

func TestPrinterParametric_JsonPrinter(t *testing.T) {
	var testKey = "testKey"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = NewOptionBuilder().WithKeys(&schema.Keys{Doc: testKey}).BuildBoolOption()
	var testParametric PrinterParametric

	testViper.Set(testKey, "true")
	testParametric = NewPrinterParam(testLedger, testOption)
	assert.False(t, testParametric.IsDefault())
	assert.NotNil(t, testParametric.Printer())
}
