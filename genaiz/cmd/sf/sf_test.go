package sf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestCli_ParentConfigType(t *testing.T) {
	var testDir = t.TempDir()
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{Option: config.Option{Key: "Test"}}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFunc = testSfCli.ParentConfigType(testOption)
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeJson)); err == nil {
		defer filez.CloseSilently(fd)

		_, err = fd.Write([]byte("{}"))
		panicz.PanicIfError(err)
		testViper.Set(testOption.Key, testDir)
		testFunc(testLedger)
		assert.Equal(t, shared.ConfigTypeJson, *testSfCli.parentConfigType)
		testFunc(testLedger)
		assert.Equal(t, shared.ConfigTypeJson, *testSfCli.parentConfigType)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestCli_ParentConfigType_MissingParent(t *testing.T) {
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{Option: config.Option{Key: "Test"}}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFunc = testSfCli.ParentConfigType(testOption)

	testFunc(testLedger)
	assert.Equal(t, shared.ConfigTypeYaml, *testSfCli.parentConfigType)
}

func TestCli_ParentConfigType_ParsingError(t *testing.T) {
	var testDir = t.TempDir()
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{Option: config.Option{Key: "Test"}}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFunc = testSfCli.ParentConfigType(testOption)
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeJson)); err == nil {
		var patch = mock.Patches{T: t}.OsExit(func(int) {})

		defer patch.Unpatch()
		defer filez.CloseSilently(fd)

		testViper.Set(testOption.Key, testDir)
		testFunc(testLedger)
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestCli_ParentOem(t *testing.T) {
	var testDir = t.TempDir()
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{Option: config.Option{Key: "Test"}}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFunc = testSfCli.ParentOem(testOption)
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		var expectedOem = "expectOem"
		var testSolution = &struct{ Solution *broker.Solution }{Solution: &broker.Solution{Oem: expectedOem}}
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(testSolution); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(testOption.Key, testDir)
				testFunc(testLedger)
				assert.Equal(t, expectedOem, testSfCli.parentSolution.Oem)
				testFunc(testLedger)
				assert.Equal(t, expectedOem, testSfCli.parentSolution.Oem)
			}
		}
	}

	assert.NoError(t, err)
}

func TestCli_ParentVersion(t *testing.T) {
	var testDir = t.TempDir()
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{Option: config.Option{Key: "Test"}}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testFunc = testSfCli.ParentVersion(testOption)
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, testLedger.ConfigName+"."+shared.ConfigTypeYaml)); err == nil {
		var expectedVersion = "expectVersion"
		var testSolution = &struct{ Solution *broker.Solution }{Solution: &broker.Solution{Version: expectedVersion}}
		var testBytes []byte

		defer filez.CloseSilently(fd)

		if testBytes, err = yaml.Marshal(testSolution); err == nil {
			if _, err = fd.Write(testBytes); err == nil {
				testViper.Set(testOption.Key, testDir)
				testFunc(testLedger)
				assert.Equal(t, expectedVersion, testSfCli.parentSolution.Version)
				testFunc(testLedger)
				assert.Equal(t, expectedVersion, testSfCli.parentSolution.Version)
			}
		}
	}

	assert.NoError(t, err)
}

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
	var testDockerContextOption = cli.Options.Docker.ContextPath().BuildStringOption()
	var testDockerFileOption = cli.Options.Docker.FilePath().BuildStringOption()
	var expectedFile = "dockerFile"

	assert.NotEmpty(t, testSf.Commands())
	assert.NotEmpty(t, testSf.PersistentFlags())
	assert.NoError(t, testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.Set(expectedFile))
	testSubCommand.Flags().AddFlagSet(testSf.PersistentFlags())
	testLedger.WorkDir = os.TempDir()
	testSf.PersistentPreRun(testSubCommand, []string{})
	assert.EqualValues(t, testLedger.WorkDir, testSf.PersistentFlags().Lookup(testDockerContextOption.Param).Value.String())
	assert.EqualValues(t, filepath.Join(testLedger.WorkDir, "dockerFile"),
		testSf.PersistentFlags().Lookup(testDockerFileOption.Param).Value.String())
}
