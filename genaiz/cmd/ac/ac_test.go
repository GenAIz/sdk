package ac

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewAc(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testAc = NewAc(testLedger)

	assert.Equal(t, 4, len(testAc.Commands()))
	assert.NoError(t, testAc.Execute())
}
