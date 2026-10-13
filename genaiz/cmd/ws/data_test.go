package ws

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewData(t *testing.T) {
	var testLedger = config.NewBuilder().
		WithViper(viper.New()).
		Build()
	var testCli = &Cli{}
	var testData = NewData(testLedger, testCli)

	assert.Equal(t, 2, len(testData.Commands()))
}
