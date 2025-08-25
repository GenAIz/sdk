package config

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/task/shared"
)

type configStruct struct {
	Key   string
	Value string
}

func TestBuilder_WithTemplates(t *testing.T) {
	var expectedPath = "/tmp"
	var testLedger = NewBuilder().WithTemplates(expectedPath).Build()

	assert.Contains(t, testLedger.TemplatePaths, expectedPath)
}

func TestLedger_backupConfigsInvalidUserPath(t *testing.T) {
	var _, testLedger = newTestConfigs()

	testLedger.UserPath = "/notValid"
	assert.ErrorIs(t, testLedger.backupConfigs(), os.ErrNotExist)
}

func TestLedger_backupConfigs(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testLedger.UserPath = "/tmp"
	_ = testLedger.makeConfigs(&testStruct)
	assert.NoError(t, testLedger.backupConfigs())
	assert.NoError(t, testLedger.backupConfigs())
	assert.NoError(t, os.Remove("/tmp/"+defaultConfigName+".yaml.back"))
}

func TestLedger_makeConfigsNoOverwriting(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testLedger.UserPath = "/tmp"
	_ = testLedger.makeConfigs(&testStruct)
	assert.Error(t, testLedger.makeConfigs(&testStruct))
	assert.NoError(t, os.Remove("/tmp/"+defaultConfigName+".yaml"))
}

func TestLedger_rollbackConfigsInvalidUserPath(t *testing.T) {
	var _, testLedger = newTestConfigs()

	testLedger.UserPath = "/notValid"
	assert.ErrorIs(t, testLedger.rollbackConfigs(), os.ErrNotExist)
}

func TestLedger_rollbackConfigs(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testLedger.UserPath = "/tmp"
	_ = testLedger.makeConfigs(&testStruct)
	assert.NoError(t, testLedger.backupConfigs())
	assert.NoError(t, testLedger.rollbackConfigs())
	assert.NoError(t, os.Remove("/tmp/"+defaultConfigName+".yaml"))
}

func TestLedger_AddConfigOption(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testPath = "/tmp"
	var expectedFile = filepath.Join(testPath, testLedger.ConfigName+".yaml")
	var _, err = os.Create(expectedFile)
	var testOption = &StringOption{
		Option{
			Key: "key",
			DefaultGetter: func(ledger *Ledger) any {
				return testPath
			},
		},
	}

	assert.NoError(t, err)
	testLedger.Init()
	testLedger.AddConfigOption(testOption)
	testLedger.InitDefaults()
	assert.EqualValues(t, expectedFile, testLedger.viper.ConfigFileUsed())
	assert.NoError(t, os.Remove(expectedFile))
}

func TestLedger_AddConfigOptionInvalidFile(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testPath = "/tmp"
	var testOption = &StringOption{
		Option{
			Key: "key",
			DefaultGetter: func(ledger *Ledger) any {
				return testPath
			},
		},
	}

	testLedger.Init()
	testLedger.AddConfigOption(testOption)
	testLedger.InitDefaults()
	assert.Empty(t, testLedger.viper.ConfigFileUsed())
}

func TestLedger_ChangeWorkDirEmptyDir(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var expectedWorkDir = testLedger.WorkDir
	var testOption = &StringOption{}

	testLedger.ChangeWorkDir(testOption)
	assert.EqualValues(t, expectedWorkDir, testLedger.WorkDir)
}

func TestLedger_ChangeWorkDirOptionNil(t *testing.T) {
	var _, testLedger = newTestConfigs()

	assert.Panics(t, func() {
		testLedger.ChangeWorkDir(nil)
	})
}

func TestLedger_ChangeWorkDir(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var currentWorkDir = testLedger.WorkDir
	var expectedWorkDir = "/tmp"
	var testOption = &StringOption{
		Option: Option{
			Key:          "TMP",
			DefaultValue: expectedWorkDir,
		},
	}

	testLedger.Register(&cobra.Command{}, testOption)
	testLedger.InitDefaults()
	testLedger.ChangeWorkDir(testOption)
	assert.EqualValues(t, expectedWorkDir, testLedger.WorkDir)

	// reset the work dir
	panicz.PanicIfError(os.Chdir(currentWorkDir))
}

func TestLedger_DisplayChangeDir(t *testing.T) {
	var buff, _, testLedger = newTestConfigsWithBuffer()
	var expectedWorkDir = "/tmp"
	var testOption = &StringOption{
		Option{
			Key:          "TMP",
			DefaultValue: expectedWorkDir,
		},
	}

	testLedger.Register(&cobra.Command{}, testOption)
	testLedger.InitDefaults()
	testLedger.ChangeWorkDir(testOption)
	testLedger.DisplayChangeDir()
	assert.Contains(t, buff.String(), expectedWorkDir)
}

func TestLedger_DisplayOptions(t *testing.T) {
	var buff, _, testLedger = newTestConfigsWithBuffer()
	var expectedOne = "one"
	var expectedTwo = "two"
	var testOption1 = &StringOption{
		Option{
			Param:        "paramA",
			DefaultValue: expectedOne,
		},
	}
	var testOption2 = &StringOption{
		Option{
			Param:        "paramB",
			DefaultValue: expectedTwo,
		},
	}

	testLedger.Register(&cobra.Command{}, testOption1, testOption2)
	testLedger.InitDefaults()
	testLedger.DisplayOptions(&testOption2.Option, &testOption1.Option)

	if s := buff.String(); s != "" {
		assert.Contains(t, s, expectedOne)
		assert.Contains(t, s, expectedTwo)
		// tests the sorting
		assert.True(t, strings.Index(s, testOption1.Param) < strings.Index(s, testOption2.Param))
	} else {
		assert.Fail(t, "could not read output")
	}
}

func TestLedger_DisplayOptionsWithMap(t *testing.T) {
	var buff, _, testLedger = newTestConfigsWithBuffer()
	var expectedKey = "key"
	var expectedValue = "value"
	var testMap = map[string]string{expectedKey: expectedValue}

	testLedger.DisplayOptionsWithMap(&testMap)

	if s := buff.String(); s != "" {
		assert.Contains(t, s, expectedKey)
		assert.Contains(t, s, expectedValue)
		// tests the sorting
		assert.True(t, strings.Index(s, expectedKey) < strings.Index(s, expectedValue))
	} else {
		assert.Fail(t, "could not read output")
	}
}

func TestLedger_FromWorkDirAbs(t *testing.T) {
	var expectedParam = "param"
	var expectedValue = "/path"
	var _, testLedger = newTestConfigs()
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, expectedValue, "")
	testLedger.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestLedger_FromWorkDirFailLookup(t *testing.T) {
	var expectedParam = "param"
	var _, testLedger = newTestConfigs()
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}

	testLedger.FromWorkDir(testOption, testFlags)
	assert.False(t, testFlags.HasFlags())
}

func TestLedger_FromWorkDirLocalValue(t *testing.T) {
	var expectedParam = "param"
	var _, testLedger = newTestConfigs()
	var expectedValue = testLedger.WorkDir + "/path"
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, "path", "")
	testLedger.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestLedger_FromWorkDirRelativeValue(t *testing.T) {
	var expectedParam = "param"
	var _, testLedger = newTestConfigs()
	var expectedValue, _ = filepath.Abs(testLedger.WorkDir + "/../path")
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, "../path", "")
	testLedger.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestLedger_GetKey(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &Option{Key: "key"}

	testViper.SetDefault(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.Get(testOption))
}

func TestLedger_GetParam(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &Option{Param: "param"}

	testViper.SetDefault(testOption.Param, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.Get(testOption))
}

func TestLedger_GetDefaultValue(t *testing.T) {
	var expectedValue = "expected"
	var testViper, testLedger = newTestConfigs()
	var testOption = &Option{
		Param:        "param",
		DefaultValue: "value",
		DefaultGetter: func(ledger *Ledger) any {
			return expectedValue
		},
	}

	testViper.SetDefault(testOption.Param, "value")
	assert.EqualValues(t, expectedValue, testLedger.Get(testOption))
}

func TestLedger_GetEnvPlaceholder(t *testing.T) {
	var expectedValue = "expected"
	var testViper, testLedger = newTestConfigs()
	var testOption = &Option{
		Param:        "param",
		DefaultValue: "$value",
	}

	_ = os.Setenv("value", expectedValue)
	testViper.SetDefault(testOption.Param, "$value")
	assert.EqualValues(t, expectedValue, testLedger.Get(testOption))
}

func TestLedger_GetBoolNoResult(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &BoolOption{}

	assert.False(t, testLedger.GetBool(testOption))
}

func TestLedger_GetBool(t *testing.T) {
	var testViper, testLedger = newTestConfigs()
	var testOption = &BoolOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, "true")
	assert.True(t, testLedger.GetBool(testOption))
}

func TestLedger_GetConfigType(t *testing.T) {
	var expectedKey = "key"
	var testViper, testLedger = newTestConfigs()
	var testOption = StringOption{
		Option: Option{
			Key: expectedKey,
		},
	}

	testViper.Set(expectedKey, shared.ConfigTypeToml)
	actual, err := testLedger.GetConfigType(&testOption)
	assert.NoError(t, err)
	assert.EqualValues(t, shared.ConfigTypeToml, *actual)
}

func TestLedger_GetConfigType_Invalid(t *testing.T) {
	var expectedKey = "key"
	var testViper, testLedger = newTestConfigs()
	var testOption = StringOption{
		Option: Option{
			Key: expectedKey,
		},
	}

	testViper.Set(expectedKey, "")
	actual, err := testLedger.GetConfigType(&testOption)
	assert.Error(t, err)
	assert.Empty(t, actual)
}

func TestLedger_GetList(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, []string{expectedValue, "two"})
	assert.Contains(t, testLedger.GetList(testOption), expectedValue)
}

func TestLedger_GetListAsString(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue+" two")
	assert.Contains(t, testLedger.GetList(testOption), expectedValue)
}

func TestLedger_GetListEmptyString(t *testing.T) {
	var testViper, testLedger = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, "")
	assert.Empty(t, testLedger.GetList(testOption))
}

func TestLedger_GetListNil(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	assert.Empty(t, testLedger.GetList(testOption))
}

func TestLedger_GetListSingle(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue)
	assert.Contains(t, testLedger.GetList(testOption), expectedValue)
}

func TestLedger_GetString(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestLedger_GetStringInvalid(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Validator: func(value any) bool {
				return false
			},
		},
	}

	testLedger.validationHandler = func(e interface{}) {
		assert.NotEmpty(t, e)
	}
	testLedger.GetString(testOption)
}

func TestLedger_GetStringValid(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Validator: func(value any) bool {
				return true
			},
		},
	}

	testLedger.validationHandler = func(e interface{}) {
		assert.Fail(t, "not expecting error")
	}
	testViper.Set(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestLedger_GetValue(t *testing.T) {
	var testKey = "key"
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()

	testViper.Set(testKey, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.GetValue(testKey))
}

func TestLedger_GetWorkspaceEmpty(t *testing.T) {
	var _, testLedger = newTestConfigs()

	assert.Empty(t, testLedger.GetWorkspace())
}

func TestLedger_GetWorkspace(t *testing.T) {
	var expectedValue = "value"
	var testViper, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Param: "param",
		},
	}

	testViper.Set(testOption.Param, expectedValue)
	testLedger.InitWorkspace(testOption)
	assert.EqualValues(t, expectedValue, testLedger.GetWorkspace())
}

func TestLedger_InitNoConfig(t *testing.T) {
	var buff bytes.Buffer
	var _, testLedger = newTestConfigs()

	testLedger.UserPath = "/tmp"
	testLedger.Init()
	testLedger.LoggerFactory = func(ledger *Ledger) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testLedger.InitLogging()
	assert.Contains(t, buff.String(), "Could not")
}

func TestLedger_Init(t *testing.T) {
	var buff bytes.Buffer
	var _, testLedger = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testLedger.UserPath = "/tmp"
	assert.NoError(t, testLedger.makeConfigs(&testStruct))

	testLedger.Init()
	testLedger.LoggerFactory = func(ledger *Ledger) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testLedger.InitLogging()
	assert.Contains(t, buff.String(), "Using")
	assert.NoError(t, os.Remove("/tmp/"+testLedger.ConfigName+".yaml"))
}

func TestLedger_InitLogging(t *testing.T) {
	var testLedger = NewLedger()

	testLedger.LogDebug("TestLedger_InitLogging")
	testLedger.InitLogging()
	assert.Empty(t, testLedger.loggers)
}

func TestLedger_InitValue(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var _, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Key: expectedKey,
		},
	}

	testLedger.InitValue(testOption, expectedValue)
	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestLedger_InitValueAlreadySet(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "other"
	var testViper, testLedger = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Key: expectedKey,
		},
	}

	testViper.Set(expectedKey, expectedValue)
	testLedger.InitValue(testOption, "value")
	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestLedger_LogDebug(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LoggerFactory = func(ledger *Ledger) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testLedger.InitLogging()
	testLedger.LogDebug("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_LogDebugNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LogDebug("%s", expectedString)
	assert.NotEmpty(t, testLedger.loggers)
	testLedger.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_LogInfo(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LoggerFactory = func(ledger *Ledger) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.InfoLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testLedger.InitLogging()
	testLedger.LogInfo("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_LogInfoNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LogInfo("%s", expectedString)
	assert.NotEmpty(t, testLedger.loggers)
	testLedger.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_LogError(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LoggerFactory = func(ledger *Ledger) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.ErrorLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testLedger.InitLogging()
	testLedger.LogError("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_LogErrorNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testLedger = newTestConfigs()

	testLedger.LogError("%s", expectedString)
	assert.NotEmpty(t, testLedger.loggers)
	testLedger.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.ErrorLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestLedger_QueryMandatory(t *testing.T) {
	var buff, _, testLedger = newTestConfigsWithInput(strings.NewReader("input"))
	var expectedInput = "input"
	var expectedOutput = "test"

	assert.EqualValues(t, expectedInput, testLedger.QueryMandatory("test"))
	assert.EqualValues(t, expectedOutput, buff.String())
}

func TestLedger_QuerySecret(t *testing.T) {
	var buff, _, testLedger = newTestConfigsWithInput(os.Stdin)

	assert.Empty(t, testLedger.QuerySecret("secret"))
	assert.EqualValues(t, "secret\n", buff.String())
}

func TestLedger_ToWorkDir(t *testing.T) {
	var expectedFlag = "flag"
	var expectedPath = "path"
	var testOption = &StringOption{Option{Param: expectedFlag}}
	var testFlagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var _, testLedger = newTestConfigs()
	var testValue = testFlagSet.String(expectedFlag, expectedPath, "usage")

	testLedger.ToWorkDir(testOption, testFlagSet)
	assert.EqualValues(t, testLedger.WorkDir, *testValue)
}

func TestLedger_ValidateValidatorNil(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &Option{Key: "key"}

	assert.True(t, testLedger.Validate(testOption))
}

func TestLedger_Validate(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &Option{
		Key: "key",
		Validator: func(value any) bool {
			return false
		},
	}

	assert.False(t, testLedger.Validate(testOption))
}

func newTestConfigs() (*viper.Viper, *Ledger) {
	var v = viper.New()

	return v, NewBuilder().
		WithViper(v).
		Build()
}

func newTestConfigsWithBuffer() (*bytes.Buffer, *viper.Viper, *Ledger) {
	var v = viper.New()
	var buff bytes.Buffer

	return &buff, v, NewBuilder().
		WithInput(io.Reader(&buff)).
		WithOutput(io.Writer(&buff)).
		WithViper(v).
		Build()
}

func newTestConfigsWithInput(input io.Reader) (*bytes.Buffer, *viper.Viper, *Ledger) {
	var v = viper.New()
	var buff bytes.Buffer

	return &buff, v, NewBuilder().
		WithInput(input).
		WithOutput(io.Writer(&buff)).
		WithViper(v).
		Build()
}
