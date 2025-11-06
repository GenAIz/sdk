package prop

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/mock"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

func TestEnvExecutor_Display(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(testOutput).
		Build()
	var testExecutor = &EnvExecutor{
		EnvOptions: NewEnvOptions(),
		Ledger:     testLedger,

		key:   "expectedKey",
		value: "expectedValue",
	}
	var expectedContext = "expectedContext"
	var expectedEnvFile = "expectedFile"

	testViper.Set(testExecutor.optionContext.Key, expectedContext)
	testViper.Set(testExecutor.optionEnvFile.Key, expectedEnvFile)
	testExecutor.Display()
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionContext.Param+`:[\s\t]*`+expectedContext), actual)
	assert.Regexp(t, regexp.MustCompile(testExecutor.optionEnvFile.Param+`:[\s\t]*`+expectedEnvFile), actual)
	assert.Regexp(t, regexp.MustCompile(`key:[\s\t]*`+testExecutor.key), actual)
	assert.Regexp(t, regexp.MustCompile(`value:[\s\t]*`+testExecutor.value), actual)
}

func TestEnvExecutor_Pretend(t *testing.T) {
	var testDir = t.TempDir()
	var fd *os.File
	var err error

	if fd, err = os.Create(filepath.Join(testDir, ".env")); err == nil {
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

		if _, err = fd.Write([]byte(fmt.Sprintf("%s=\"%s\"", testExecutor.key, "valueToReplace"))); err == nil {
			var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})

			defer patch.Unpatch()
			testLedger.WorkDir = testDir
			testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
				map[string]interface{}{
					"key":  testExecutor.key,
					"type": broker.PropSpecTypeString,
				},
			})
			testExecutor.Pretend()
			assert.NotEmpty(t, patch.CalledWith)
			params := cast.ToStringSlice(patch.CalledWith)
			assert.Equal(t, testExecutor.key, params[1])
			assert.Equal(t, testExecutor.value, params[2])
			assert.Equal(t, testOptions.optionEnvFile.DefaultGetter(testLedger), params[3])
		}
	}

	assert.NoError(t, err)
}

func TestEnvExecutor_Pretend_PathError(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.FmtPrintf(func(format string, a ...any) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testLedger.WorkDir = testDir
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  testExecutor.key,
			"type": broker.PropSpecTypeString,
		},
	})
	testExecutor.Pretend()
	assert.NotEmpty(t, patch.CalledWith)
	params := cast.ToStringSlice(patch.CalledWith)
	assert.Equal(t, testExecutor.key, params[1])
	assert.Equal(t, testExecutor.value, params[2])
	assert.Equal(t, testOptions.optionEnvFile.DefaultGetter(testLedger), params[3])
}

func TestEnvExecutor_Pretend_PermissionError(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), "not_exist")
	var fd *os.File
	var err error

	if fd, err = os.OpenFile(testFile, os.O_CREATE|os.O_TRUNC, 0222); err == nil {
		defer filez.CloseSilently(fd)
		var patch = mock.Patches{T: t}.OsExit(func(int) {})
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

		defer patch.Unpatch()
		testViper.Set(testOptions.optionEnvFile.Key, testFile)
		testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
			map[string]interface{}{
				"key":  testExecutor.key,
				"type": broker.PropSpecTypeString,
			},
		})
		testExecutor.Pretend()
		assert.True(t, patch.Called)
		assert.EqualValues(t, 1, patch.CalledWith)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestEnvExecutor_Pretend_ValidationError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testExecutor.Pretend()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed(t *testing.T) {
	var testFile = filepath.Join(t.TempDir(), ".env")
	var fd *os.File
	var err error

	if fd, err = os.Create(testFile); err == nil {
		var testViper = viper.New()
		var testLedger = config.NewBuilder().
			WithViper(testViper).
			Build()
		var testOptions = NewEnvOptions()
		var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")
		var testKeyPair = fmt.Sprintf("%s=\"%s\"\n", testExecutor.key, "notRightValue")
		var expectedOther = fmt.Sprintf("%s=\"%s\"\n", "otherKey", "otherValue")

		defer filez.CloseSilently(fd)

		if _, err = fd.Write([]byte(testKeyPair + expectedOther)); err == nil {
			var content []byte

			testViper.Set(testExecutor.optionEnvFile.Key, testFile)
			testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
				map[string]interface{}{
					"key":  testExecutor.key,
					"type": broker.PropSpecTypeString,
				},
			})
			testExecutor.Proceed()

			if content, err = os.ReadFile(testFile); err == nil {
				var expectedKeyPair = fmt.Sprintf("%s=\"%s\"\n", testExecutor.key, testExecutor.value)

				actual := string(content)
				assert.Contains(t, actual, expectedOther)
				assert.Contains(t, actual, expectedKeyPair)
			}
		}
	}

	assert.NoError(t, err)
}

func TestEnvExecutor_Proceed_CreateFileError(t *testing.T) {
	var testDir = t.TempDir()
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testViper.Set(testExecutor.optionEnvFile.Key, filepath.Join(testDir, "not_exist", ".env"))
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  testExecutor.key,
			"type": broker.PropSpecTypeString,
		},
	})
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestEnvExecutor_Proceed_ValidationError(t *testing.T) {
	var patch = mock.Patches{T: t}.OsExit(func(int) {})
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testOptions = NewEnvOptions()
	var testExecutor = NewEnvExecutor(testLedger, testOptions, "expectedKey", "expectedValue")

	defer patch.Unpatch()
	testExecutor.Proceed()
	assert.True(t, patch.Called)
	assert.EqualValues(t, 1, patch.CalledWith)
}

func TestNewEnv(t *testing.T) {
	var testDir = t.TempDir()
	var testFile = filepath.Join(testDir, ".env")
	var envCompleted = false
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(testOutput).
		Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
		return true
	}})
	var expectedKey = "EXPECTED_KEY"
	var expectedValue = "expectedValue"

	testEnv.PostRun = func(cmd *cobra.Command, args []string) {
		envCompleted = true
	}
	testViper.Set(schema.Genaiz.Function.Env.File.Doc, testFile)
	testViper.Set(schema.Genaiz.Function.Env.Context.Doc, testDir)
	testEnv.SetArgs([]string{expectedKey, expectedValue})
	// Execute will change the work dir, so we need to reset it, by changing it with testing first
	t.Chdir(testDir)
	assert.NoError(t, testEnv.Execute())
	assert.True(t, envCompleted)

	if actual := testOutput.String(); actual != "" {
		assert.Contains(t, actual, testFile)
		assert.Contains(t, actual, testDir)
		assert.Contains(t, actual, expectedKey)
		assert.Contains(t, actual, expectedValue)
	} else {
		assert.Fail(t, "no --dry content")
	}
}

func TestNewEnv_Validation(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		Build()
	var testEnv = NewEnv(testLedger, &cli.BaseCli{Dry: func(ledger *config.Ledger) bool {
		return true
	}})

	assert.Error(t, testEnv.ValidateArgs([]string{"..NotValid..", "value"}))
	assert.NoError(t, testEnv.ValidateArgs([]string{"MY_KEY", "value"}))
}
