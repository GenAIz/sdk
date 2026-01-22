package dk

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
)

func TestBaseExecutor_getConfigPath(t *testing.T) {
	var testUserPath = t.TempDir()
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testOption = &config.BoolOption{Option: config.Option{Key: "key"}}
	var testExecutor *BaseExecutor

	t.Chdir(testWorkDir)
	testExecutor = &BaseExecutor{
		Ledger: config.NewBuilder().
			WithViper(testViper).
			WithUserPath(testUserPath).
			Build(),
	}

	assert.Equal(t, testWorkDir, testExecutor.getConfigPath(testOption))
	testViper.Set(testOption.Key, "True")
	assert.Equal(t, testUserPath+"/.config/genaiz", testExecutor.getConfigPath(testOption))
}

func TestBaseExecutor_makeConfigParams(t *testing.T) {
	var testWorkDir = t.TempDir()
	var testViper = viper.New()
	var testLedgerBuilder = config.NewBuilder().WithViper(testViper)
	var testTypeOption = &config.StringOption{Option: config.Option{Key: "type"}}
	var testUserOption = &config.BoolOption{Option: config.Option{Key: "user"}}
	var testExecutor *BaseExecutor

	t.Chdir(testWorkDir)
	testExecutor = &BaseExecutor{
		Ledger: testLedgerBuilder.Build(),
	}
	actual, err := testExecutor.makeConfigParams(testTypeOption, testUserOption)
	assert.NotNil(t, actual)
	assert.NoError(t, err)
	assert.Equal(t, actual.ConfigFolder, testWorkDir)
}

func TestBaseExecutor_makeConfigParams_InvalidConfigType(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testTypeOption = &config.StringOption{Option: config.Option{Key: "type"}}
	var testUserOption = &config.BoolOption{Option: config.Option{Key: "user"}}
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}

	testViper.Set(testTypeOption.Key, "notValid")
	actual, err := testExecutor.makeConfigParams(testTypeOption, testUserOption)
	assert.Nil(t, actual)
	assert.Error(t, err)
}

func TestBaseExecutor_parseDataLinkArgument(t *testing.T) {
	var expectedOem = "oem"
	var expectedHandle = "handle"
	var expectedVersion = "version"
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}

	actualOem, actualHandle, actualVersion := testExecutor.parseDataLinkArgument(
		fmt.Sprintf("%s/%s:%s", expectedOem, expectedHandle, expectedVersion))
	assert.Equal(t, expectedOem, actualOem)
	assert.Equal(t, expectedHandle, actualHandle)
	assert.Equal(t, expectedVersion, actualVersion)
	actualOem, actualHandle, actualVersion = testExecutor.parseDataLinkArgument(
		fmt.Sprintf("%s/%s", expectedOem, expectedHandle))
	assert.Equal(t, expectedOem, actualOem)
	assert.Equal(t, expectedHandle, actualHandle)
	assert.Empty(t, actualVersion)
	actualOem, actualHandle, actualVersion = testExecutor.parseDataLinkArgument(
		fmt.Sprintf("%s:%s", expectedHandle, expectedVersion))
	assert.Empty(t, actualOem)
	assert.Equal(t, expectedHandle, actualHandle)
	assert.Equal(t, expectedVersion, actualVersion)
	actualOem, actualHandle, actualVersion = testExecutor.parseDataLinkArgument(expectedHandle)
	assert.Empty(t, actualOem)
	assert.Equal(t, expectedHandle, actualHandle)
	assert.Empty(t, actualVersion)
}

func TestNewDk(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCmd = NewDk(testLedger, nil, nil, nil)

	assert.Equal(t, 3, len(testCmd.Commands()))
}

func Test_newDataLinksWriter(t *testing.T) {
	var testOutput = filepath.Join(t.TempDir(), "Genaiz.yaml")
	var testLinks = []broker.DataLink{
		{
			Handle:      "handle",
			Oem:         "oem",
			Version:     "version",
			PropSpecs:   []broker.PropSpec{},
			SecretSpecs: []broker.PropSpec{},
		},
	}
	var fd *os.File
	var err error

	if fd, err = os.Create(testOutput); err == nil {
		var dataLinks = map[string][]broker.DataLink{"dataLinks": testLinks}
		var lb []byte

		defer filez.CloseSilently(fd)

		if lb, err = yaml.Marshal(dataLinks); err == nil {
			if _, err = fd.Write(lb); err == nil {
				var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
				var testWriter *dataLinksWriter

				filez.CloseSilently(fd)
				testLedger.InitLogging()
				testWriter = newDataLinksWriter(testLedger, testOutput)
				assert.Equal(t, &testLinks[0], testWriter.GetDataLink(testLinks[0].Oem, testLinks[0].Handle, testLinks[0].Version))
				return
			}
		}
	}

	assert.NoError(t, err)
}
