package lk

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestNewLk(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewLk(testLedger, nil, nil, nil)

	assert.Equal(t, 1, len(testCmd.Commands()))
}
