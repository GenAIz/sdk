package sf

import (
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/config"
)

func TestNewList(t *testing.T) {
	var testDir = t.TempDir()
	var testCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testList = NewList(testLedger, testCli)

	if fd, err := os.CreateTemp(testDir, "genaizDockerfile"); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(int) {})

		defer filez.RemoveSilently(fd.Name())
		defer patch.Unpatch()
		testViper.Set(testCli.optionDockerFile.Key, fd.Name())
		testLedger.Logger = logrus.New()
		assert.NoError(t, testList.Execute())
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.NoError(t, err)
	}
}
