package sf

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
)

func TestCli_allDefiners(t *testing.T) {
	var testSfCli = NewSfCli(nil, nil, nil)
	var testCliDefiners = testSfCli.allDefiners()

	assert.NotEmpty(t, testSfCli.optionDockerContext)
	assert.NotEmpty(t, testSfCli.optionDockerFile)
	assert.NotEmpty(t, testSfCli.optionDockerTag)
	assert.NotEmpty(t, testSfCli.optionDockerVersion)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerContext)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerFile)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerTag)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerVersion)
}

func TestNewSf(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testSf = NewSf(testLedger, nil, nil, nil)
	var testSubCommand = &cobra.Command{
		Run: func(cmd *cobra.Command, args []string) {

		},
	}
	var testDockerContextOption = newOptionDockerContext()
	var testDockerFileOption = newOptionDockerFile()
	var expectedFile = "dockerFile"

	assert.NotEmpty(t, testSf.Commands())
	assert.NotEmpty(t, testSf.PersistentFlags())
	assert.NoError(t, testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.Set(expectedFile))
	testSubCommand.Flags().AddFlagSet(testSf.PersistentFlags())
	testLedger.WorkDir = "/tmp"
	testSf.PersistentPreRun(testSubCommand, []string{})
	assert.EqualValues(t, "/tmp", testSf.PersistentFlags().Lookup(testDockerContextOption.Param).Value.String())
	assert.EqualValues(t, "/tmp/dockerFile", testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.String())
}

func Test_newOptionDockerTag_DefaultSetter(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = newOptionDockerTag()
	var expectedTag = "test"

	testLedger.WorkDir = filepath.Join("/tmp", "genaiz", expectedTag)
	assert.EqualValues(t, expectedTag, testOptions.DefaultSetter(testLedger))
}
