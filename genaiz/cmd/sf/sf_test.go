package sf

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func TestBaseExecutor_validate(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOption = cli.NewOptionBuilder().
		WithKeys(&schema.Keys{Doc: "key"}).
		BuildStringOption()
	var testExecutor = &BaseExecutor{
		Ledger: testLedger,
	}

	assert.ErrorIs(t, testExecutor.validateConnector(testOption), errInvalidConnectorType)
	testViper.Set(testOption.Key, shared.FunctionTypeTrigger)
	assert.ErrorIs(t, testExecutor.validateConnector(testOption), errInvalidConnectorType)
	testViper.Set(testOption.Key, shared.FunctionTypeFunction)
	assert.ErrorIs(t, testExecutor.validateConnector(testOption), errInvalidConnectorType)
	testViper.Set(testOption.Key, shared.FunctionTypeConnector)
	assert.NoError(t, testExecutor.validateConnector(testOption))
}

func TestCli_ContainerPrefix(t *testing.T) {
	var testSfCli = NewSfCli(nil, nil, nil)
	var testOption = &config.StringOption{
		Option: config.Option{Param: "ws"},
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var expectedDockerRepo = "namespace/repo"
	var expectedWorkspace = "workSpace"
	var actual string

	testLedger.InitWorkspace(testOption)
	testViper.Set(testOption.Param, expectedWorkspace)
	testViper.Set(testSfCli.optionDockerRepo.Key, expectedDockerRepo)
	actual = cast.ToString(testSfCli.ContainerPrefix(testLedger))
	assert.Equal(t, "workSpace-namespace-repo", actual)
}

func TestCli_DefaultRunImage(t *testing.T) {
	var testSfCli = NewSfCli(nil, nil, nil)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var expectedDockerRepo = "namespace/repo"
	var expectedDockerVersion = "version"
	var actual string

	testViper.Set(testSfCli.optionDockerRepo.Key, expectedDockerRepo)
	testViper.Set(testSfCli.optionDockerVersion.Key, expectedDockerVersion)
	actual = cast.ToString(testSfCli.DefaultRunImage(testLedger))
	assert.Equal(t, "namespace/repo:version", actual)
}

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
	assert.NotEmpty(t, testSfCli.optionDockerRepo)
	assert.NotEmpty(t, testSfCli.optionDockerVersion)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerContext)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerFile)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerRepo)
	assert.Contains(t, testCliDefiners, testSfCli.optionDockerVersion)
}

func TestEnvOptions_makeEnvMap(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")
	var testEnvOptions = &EnvOptions{
		optionEnvFile: cli.Options.Docker.EnvFile().
			WithKeys(&schema.Keys{Doc: "key"}).
			BuildStringOption(),
		optionEnvVars: cli.Options.Docker.EnvVar().
			WithKeys(&schema.Keys{Doc: "vars"}).
			BuildListOption(),
	}
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var expectedKey1 = "KEY_1"
	var expectedKey2 = "KEY_2"
	var expectedValue1 = "VALUE.1"
	var expectedValue2 = "VALUE.2"
	var expectedType = "type"

	testViper.Set(schema.Genaiz.Function.Publish.Type.Doc, expectedType)
	testViper.Set(testEnvOptions.optionEnvFile.Key, testFile)
	testViper.Set(testEnvOptions.optionEnvVars.Key, []string{
		expectedKey1 + "=" + expectedValue1,
		expectedKey2 + "=" + expectedValue2,
	})
	result, err := testEnvOptions.makeEnvMap(testLedger)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(result))
	assert.Equal(t, expectedValue1, result[expectedKey1])
	assert.Equal(t, expectedValue2, result[expectedKey2])
	assert.Equal(t, expectedType, result["SF_TYPE"])
}

func TestEnvOptions_makeEnvMap_PermissionDenied(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")

	var fd *os.File
	var err error

	if fd, err = os.OpenFile(testFile, os.O_CREATE|os.O_TRUNC, 0222); err == nil {
		defer filez.CloseSilently(fd)
		var testEnvOptions = &EnvOptions{
			optionEnvFile: cli.Options.Docker.EnvFile().
				WithKeys(&schema.Keys{Doc: "key"}).
				BuildStringOption(),
		}
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()

		testViper.Set(testEnvOptions.optionEnvFile.Key, testFile)
		_, err = testEnvOptions.makeEnvMap(testLedger)
		assert.Error(t, err)
		assert.False(t, errorz.IsPathError(err))
	}
}

func TestEnvOptions_parseEnvFile(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)
		var expectedKey = "expected_key"
		var expectedValue = "VALUE"

		if _, err = fd.Write([]byte(expectedKey + "=" + expectedValue)); err == nil {
			var testEnvOption = &EnvOptions{}
			var result, actual = testEnvOption.parseEnvFile(testFile)

			assert.NoError(t, actual)
			assert.Equal(t, 1, len(result))
			assert.Equal(t, expectedValue, result[expectedKey])
		}
	}

	assert.NoError(t, err)
}

func TestEnvOptions_parseEnvFile_ParseError(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte("not_a_valid_key_pair")); err == nil {
			var testEnvOption = &EnvOptions{}
			var result, actual = testEnvOption.parseEnvFile(testFile)

			assert.Empty(t, result)
			assert.Error(t, actual)
			assert.False(t, errorz.IsPathError(actual))
		}
	}

	assert.NoError(t, err)
}

func TestEnvOptions_parseEnvFile_PathError(t *testing.T) {
	var testDir = t.TempDir()
	var testEnvOption = &EnvOptions{}
	var result, actual = testEnvOption.parseEnvFile(filepath.Join(testDir, "not_exist"))

	assert.Empty(t, result)
	assert.Error(t, actual)
	assert.True(t, errorz.IsPathError(actual))
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

func TestNewSfCli(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), "oem", "handle")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var err error

	if err = os.MkdirAll(testDir, 0750); err == nil {
		testLedger.WorkDir = testDir
		assert.Equal(t, "oem/handle", testCli.optionDockerRepo.DefaultGetter(testLedger))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewSfCli_invalidOem(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), ".oem", "handle")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var err error

	if err = os.MkdirAll(testDir, 0750); err == nil {
		testLedger.WorkDir = testDir
		assert.Empty(t, testCli.optionDockerRepo.DefaultGetter(testLedger))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestNewSfCli_invalidHandle(t *testing.T) {
	var testDir = filepath.Join(t.TempDir(), "oem", ".handle")
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testCli = NewSfCli(nil, nil, nil)
	var err error

	if err = os.MkdirAll(testDir, 0750); err == nil {
		testLedger.WorkDir = testDir
		assert.Empty(t, testCli.optionDockerRepo.DefaultGetter(testLedger))
	} else {
		assert.Fail(t, err.Error())
	}
}
