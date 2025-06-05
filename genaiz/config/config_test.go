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

	"genaiz.com/genaiz/lang/panicz"
)

type configStruct struct {
	Key   string
	Value string
}

func TestBuilder_WithTemplates(t *testing.T) {
	var expectedPath = "/tmp"
	var testRepo = NewBuilder().WithTemplates(expectedPath).Build()

	assert.Contains(t, testRepo.TemplatePaths, expectedPath)
}

func TestRepo_backupConfigsInvalidUserPath(t *testing.T) {
	var _, testRepo = newTestConfigs()

	testRepo.UserPath = "/notValid"
	assert.ErrorIs(t, testRepo.backupConfigs(), os.ErrNotExist)
}

func TestRepo_backupConfigs(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testRepo.UserPath = "/tmp"
	_ = testRepo.makeConfigs(&testStruct)
	assert.NoError(t, testRepo.backupConfigs())
	assert.NoError(t, testRepo.backupConfigs())
	assert.NoError(t, os.Remove("/tmp/genaiz.yaml.back"))
}

func TestRepo_makeConfigsNoOverwriting(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testRepo.UserPath = "/tmp"
	_ = testRepo.makeConfigs(&testStruct)
	assert.Error(t, testRepo.makeConfigs(&testStruct))
	assert.NoError(t, os.Remove("/tmp/genaiz.yaml"))
}

func TestRepo_rollbackConfigsInvalidUserPath(t *testing.T) {
	var _, testRepo = newTestConfigs()

	testRepo.UserPath = "/notValid"
	assert.ErrorIs(t, testRepo.rollbackConfigs(), os.ErrNotExist)
}

func TestRepo_rollbackConfigs(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testRepo.UserPath = "/tmp"
	_ = testRepo.makeConfigs(&testStruct)
	assert.NoError(t, testRepo.backupConfigs())
	assert.NoError(t, testRepo.rollbackConfigs())
	assert.NoError(t, os.Remove("/tmp/genaiz.yaml"))
}

func TestRepo_AddConfigOption(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testPath = "/tmp"
	var expectedFile = filepath.Join(testPath, testRepo.ConfigName+".yaml")
	var _, err = os.Create(expectedFile)
	var testOption = &StringOption{
		Option{
			Key: "key",
			DefaultGetter: func(repo *Repo) any {
				return testPath
			},
		},
	}

	assert.NoError(t, err)
	testRepo.Init()
	testRepo.AddConfigOption(testOption)
	testRepo.InitDefaults()
	assert.NoError(t, testRepo.viper.ReadInConfig())
	assert.EqualValues(t, expectedFile, testRepo.viper.ConfigFileUsed())
	assert.NoError(t, os.Remove(expectedFile))
}

func TestRepo_ChangeWorkDirEmptyDir(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var expectedWorkDir = testRepo.WorkDir
	var testOption = &StringOption{}

	testRepo.ChangeWorkDir(testOption)
	assert.EqualValues(t, expectedWorkDir, testRepo.WorkDir)
}

func TestRepo_ChangeWorkDirOptionNil(t *testing.T) {
	var _, testRepo = newTestConfigs()

	assert.Panics(t, func() {
		testRepo.ChangeWorkDir(nil)
	})
}

func TestRepo_ChangeWorkDir(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var currentWorkDir = testRepo.WorkDir
	var expectedWorkDir = "/tmp"
	var testOption = &StringOption{
		Option: Option{
			Key:          "TMP",
			DefaultValue: expectedWorkDir,
		},
	}

	testRepo.Register(&cobra.Command{}, testOption)
	testRepo.InitDefaults()
	testRepo.ChangeWorkDir(testOption)
	assert.EqualValues(t, expectedWorkDir, testRepo.WorkDir)

	// reset the work dir
	panicz.PanicIfError(os.Chdir(currentWorkDir))
}

func TestRepo_DisplayChangeDir(t *testing.T) {
	var buff, _, testRepo = newTestConfigsWithBuffer()
	var expectedWorkDir = "/tmp"
	var testOption = &StringOption{
		Option{
			Key:          "TMP",
			DefaultValue: expectedWorkDir,
		},
	}

	testRepo.Register(&cobra.Command{}, testOption)
	testRepo.InitDefaults()
	testRepo.ChangeWorkDir(testOption)
	testRepo.DisplayChangeDir()
	assert.Contains(t, buff.String(), expectedWorkDir)
}

func TestRepo_DisplayOptions(t *testing.T) {
	var buff, _, testRepo = newTestConfigsWithBuffer()
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

	testRepo.Register(&cobra.Command{}, testOption1, testOption2)
	testRepo.InitDefaults()
	testRepo.DisplayOptions(&testOption2.Option, &testOption1.Option)

	if s := buff.String(); s != "" {
		assert.Contains(t, s, expectedOne)
		assert.Contains(t, s, expectedTwo)
		// tests the sorting
		assert.True(t, strings.Index(s, testOption1.Param) < strings.Index(s, testOption2.Param))
	} else {
		assert.Fail(t, "could not read output")
	}
}

func TestRepo_DisplayOptionsWithMap(t *testing.T) {
	var buff, _, testRepo = newTestConfigsWithBuffer()
	var expectedKey = "key"
	var expectedValue = "value"
	var testMap = map[string]string{expectedKey: expectedValue}

	testRepo.DisplayOptionsWithMap(&testMap)

	if s := buff.String(); s != "" {
		assert.Contains(t, s, expectedKey)
		assert.Contains(t, s, expectedValue)
		// tests the sorting
		assert.True(t, strings.Index(s, expectedKey) < strings.Index(s, expectedValue))
	} else {
		assert.Fail(t, "could not read output")
	}
}

func TestRepo_FromWorkDirAbs(t *testing.T) {
	var expectedParam = "param"
	var expectedValue = "/path"
	var _, testRepo = newTestConfigs()
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, expectedValue, "")
	testRepo.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestRepo_FromWorkDirFailLookup(t *testing.T) {
	var expectedParam = "param"
	var _, testRepo = newTestConfigs()
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}

	testRepo.FromWorkDir(testOption, testFlags)
	assert.False(t, testFlags.HasFlags())
}

func TestRepo_FromWorkDirLocalValue(t *testing.T) {
	var expectedParam = "param"
	var _, testRepo = newTestConfigs()
	var expectedValue = testRepo.WorkDir + "/path"
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, "path", "")
	testRepo.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestRepo_FromWorkDirRelativeValue(t *testing.T) {
	var expectedParam = "param"
	var _, testRepo = newTestConfigs()
	var expectedValue, _ = filepath.Abs(testRepo.WorkDir + "/../path")
	var testFlags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var testOption = &StringOption{
		Option{
			Param: expectedParam,
		},
	}
	var testValue string

	testFlags.StringVar(&testValue, expectedParam, "../path", "")
	testRepo.FromWorkDir(testOption, testFlags)
	assert.EqualValues(t, expectedValue, testValue)
}

func TestRepo_GetKey(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &Option{Key: "key"}

	testViper.SetDefault(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetParam(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &Option{Param: "param"}

	testViper.SetDefault(testOption.Param, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetDefaultValue(t *testing.T) {
	var expectedValue = "expected"
	var testViper, testRepo = newTestConfigs()
	var testOption = &Option{
		Param:        "param",
		DefaultValue: "value",
		DefaultGetter: func(repo *Repo) any {
			return expectedValue
		},
	}

	testViper.SetDefault(testOption.Param, "value")
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetEnvPlaceholder(t *testing.T) {
	var expectedValue = "expected"
	var testViper, testRepo = newTestConfigs()
	var testOption = &Option{
		Param:        "param",
		DefaultValue: "$value",
	}

	_ = os.Setenv("value", expectedValue)
	testViper.SetDefault(testOption.Param, "$value")
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetBoolNoResult(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testOption = &BoolOption{}

	assert.False(t, testRepo.GetBool(testOption))
}

func TestRepo_GetBool(t *testing.T) {
	var testViper, testRepo = newTestConfigs()
	var testOption = &BoolOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, "true")
	assert.True(t, testRepo.GetBool(testOption))
}

func TestRepo_GetList(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, []string{expectedValue, "two"})
	assert.Contains(t, testRepo.GetList(testOption), expectedValue)
}

func TestRepo_GetListAsString(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue+" two")
	assert.Contains(t, testRepo.GetList(testOption), expectedValue)
}

func TestRepo_GetListEmptyString(t *testing.T) {
	var testViper, testRepo = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, "")
	assert.Empty(t, testRepo.GetList(testOption))
}

func TestRepo_GetListNil(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	assert.Empty(t, testRepo.GetList(testOption))
}

func TestRepo_GetListSingle(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &ListOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue)
	assert.Contains(t, testRepo.GetList(testOption), expectedValue)
}

func TestRepo_GetString(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetString(testOption))
}

func TestRepo_GetStringInvalid(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Validator: func(value any) bool {
				return false
			},
		},
	}

	testRepo.validationHandler = func(e interface{}) {
		assert.NotEmpty(t, e)
	}
	testRepo.GetString(testOption)
}

func TestRepo_GetStringValid(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Validator: func(value any) bool {
				return true
			},
		},
	}

	testRepo.validationHandler = func(e interface{}) {
		assert.Fail(t, "not expecting error")
	}
	testViper.Set(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetString(testOption))
}

func TestRepo_GetValue(t *testing.T) {
	var testKey = "key"
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()

	testViper.Set(testKey, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetValue(testKey))
}

func TestRepo_GetWorkspaceEmpty(t *testing.T) {
	var _, testRepo = newTestConfigs()

	assert.Empty(t, testRepo.GetWorkspace())
}

func TestRepo_GetWorkspace(t *testing.T) {
	var expectedValue = "value"
	var testViper, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Param: "param",
		},
	}

	testViper.Set(testOption.Param, expectedValue)
	testRepo.InitWorkspace(testOption)
	assert.EqualValues(t, expectedValue, testRepo.GetWorkspace())
}

func TestRepo_InitNoConfig(t *testing.T) {
	var buff bytes.Buffer
	var _, testRepo = newTestConfigs()

	testRepo.UserPath = "/tmp"
	testRepo.Init()
	testRepo.LoggerFactory = func(repo *Repo) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testRepo.InitLogging()
	assert.Contains(t, buff.String(), "Could not")
}

func TestRepo_Init(t *testing.T) {
	var buff bytes.Buffer
	var _, testRepo = newTestConfigs()
	var testStruct = configStruct{Key: "key", Value: "value"}

	testRepo.UserPath = "/tmp"
	assert.NoError(t, testRepo.makeConfigs(&testStruct))

	testRepo.Init()
	testRepo.LoggerFactory = func(repo *Repo) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testRepo.InitLogging()
	assert.Contains(t, buff.String(), "Using")
}

func TestRepo_InitLogging(t *testing.T) {
	var testRepo = NewRepo()

	testRepo.LogDebug("TestRepo_InitLogging")
	testRepo.InitLogging()
	assert.Empty(t, testRepo.loggers)
}

func TestRepo_InitValue(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "value"
	var _, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Key: expectedKey,
		},
	}

	testRepo.InitValue(testOption, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetString(testOption))
}

func TestRepo_InitValueAlreadySet(t *testing.T) {
	var expectedKey = "key"
	var expectedValue = "other"
	var testViper, testRepo = newTestConfigs()
	var testOption = &StringOption{
		Option{
			Key: expectedKey,
		},
	}

	testViper.Set(expectedKey, expectedValue)
	testRepo.InitValue(testOption, "value")
	assert.EqualValues(t, expectedValue, testRepo.GetString(testOption))
}

func TestRepo_LogDebug(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LoggerFactory = func(repo *Repo) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.DebugLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testRepo.InitLogging()
	testRepo.LogDebug("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_LogDebugNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LogDebug("%s", expectedString)
	assert.NotEmpty(t, testRepo.loggers)
	testRepo.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.DebugLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_LogInfo(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LoggerFactory = func(repo *Repo) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.InfoLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testRepo.InitLogging()
	testRepo.LogInfo("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_LogInfoNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LogInfo("%s", expectedString)
	assert.NotEmpty(t, testRepo.loggers)
	testRepo.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.InfoLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_LogError(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LoggerFactory = func(repo *Repo) *logrus.Logger {
		return &logrus.Logger{
			Out:   io.Writer(&buff),
			Level: logrus.ErrorLevel,
			Formatter: &easy.Formatter{
				TimestampFormat: time.DateTime,
				LogFormat:       "%msg%",
			},
		}
	}
	testRepo.InitLogging()
	testRepo.LogError("%s", expectedString)
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_LogErrorNoLogger(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var _, testRepo = newTestConfigs()

	testRepo.LogError("%s", expectedString)
	assert.NotEmpty(t, testRepo.loggers)
	testRepo.loggers[0](&logrus.Logger{
		Out:   io.Writer(&buff),
		Level: logrus.ErrorLevel,
		Formatter: &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "%msg%",
		},
	})
	assert.True(t, strings.HasSuffix(buff.String(), expectedString))
}

func TestRepo_QueryMandatory(t *testing.T) {
	var buff, _, testRepo = newTestConfigsWithInput(strings.NewReader("input"))
	var expectedInput = "input"
	var expectedOutput = "test"

	assert.EqualValues(t, expectedInput, testRepo.QueryMandatory("test"))
	assert.EqualValues(t, expectedOutput, buff.String())
}

func TestRepo_QuerySecret(t *testing.T) {
	var buff, _, testRepo = newTestConfigsWithInput(os.Stdin)

	assert.Empty(t, testRepo.QuerySecret("secret"))
	assert.EqualValues(t, "secret\n", buff.String())
}

func TestRepo_ToWorkDir(t *testing.T) {
	var expectedFlag = "flag"
	var expectedPath = "path"
	var testOption = &StringOption{Option{Param: expectedFlag}}
	var testFlagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var _, testRepo = newTestConfigs()
	var testValue = testFlagSet.String(expectedFlag, expectedPath, "usage")

	testRepo.ToWorkDir(testOption, testFlagSet)
	assert.EqualValues(t, testRepo.WorkDir, *testValue)
}

func TestRepo_ValidateValidatorNil(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testOption = &Option{Key: "key"}

	assert.True(t, testRepo.Validate(testOption))
}

func TestRepo_Validate(t *testing.T) {
	var _, testRepo = newTestConfigs()
	var testOption = &Option{
		Key: "key",
		Validator: func(value any) bool {
			return false
		},
	}

	assert.False(t, testRepo.Validate(testOption))
}

func newTestConfigs() (*viper.Viper, *Repo) {
	var v = viper.New()

	return v, NewBuilder().
		WithViper(v).
		Build()
}

func newTestConfigsWithBuffer() (*bytes.Buffer, *viper.Viper, *Repo) {
	var v = viper.New()
	var buff bytes.Buffer

	return &buff, v, NewBuilder().
		WithInput(io.Reader(&buff)).
		WithOutput(io.Writer(&buff)).
		WithViper(v).
		Build()
}

func newTestConfigsWithInput(input io.Reader) (*bytes.Buffer, *viper.Viper, *Repo) {
	var v = viper.New()
	var buff bytes.Buffer

	return &buff, v, NewBuilder().
		WithInput(input).
		WithOutput(io.Writer(&buff)).
		WithViper(v).
		Build()
}
