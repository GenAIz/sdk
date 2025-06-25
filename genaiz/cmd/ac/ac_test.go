package ac

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewAc(t *testing.T) {
	var testViper = viper.New()
	var testRepo = config.NewBuilder().WithViper(testViper).Build()
	var testAc = NewAc(testRepo)

	assert.NoError(t, testAc.Execute())
}
