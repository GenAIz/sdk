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
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	easy "github.com/t-tomalak/logrus-easy-formatter"

	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/lang/panicz"
)

type stringValue string

func newStringValue(val string, p *string) *stringValue {
	*p = val
	return (*stringValue)(p)
}

func (s *stringValue) Set(val string) error {
	*s = stringValue(val)
	return nil
}

func (s *stringValue) Type() string {
	return "string"
}

func (s *stringValue) String() string {
	return string(*s)
}

func TestOption_bindDefValueNilFlag(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.bindDefValue(nil)
	})
}

func TestOption_bindDefValueNilValue(t *testing.T) {
	var expectedValue = "value"
	var testFlag = &pflag.Flag{DefValue: expectedValue}
	var testOption = &Option{}

	testOption.bindDefValue(testFlag)
	assert.EqualValues(t, testFlag.DefValue, expectedValue)
}

func TestOption_bindDefValueInteger(t *testing.T) {
	var testFlag = &pflag.Flag{DefValue: "value"}
	var testOption = &Option{DefaultValue: 1}

	testOption.bindDefValue(testFlag)
	assert.EqualValues(t, testFlag.DefValue, cast.ToString(testOption.DefaultValue))
}

func TestOption_bindKeyFlagNilFlag(t *testing.T) {
	var testOption = &Option{Key: "key"}
	var testViper = viper.New()

	assert.Panics(t, func() {
		testOption.bindKeyFlag(testViper, nil)
	})
}

func TestOption_bindKeyFlagViperKey(t *testing.T) {
	var testViper = viper.New()
	var expectedValue = "keyValue"
	var flagValue string
	var testOption = &Option{Key: "key", Param: "param"}
	var testFlag = &pflag.Flag{
		Name:  testOption.Key,
		Value: newStringValue(expectedValue, &flagValue),
	}

	testOption.bindKeyFlag(testViper, testFlag)
	assert.EqualValues(t, expectedValue, testViper.GetString(testOption.Key))
}

func TestOption_bindKeyFlagViperParam(t *testing.T) {
	var testViper = viper.New()
	var expectedValue = "paramValue"
	var flagValue string
	var testOption = &Option{Param: "param"}
	var testFlag = &pflag.Flag{
		Name:  testOption.Param,
		Value: newStringValue(expectedValue, &flagValue),
	}

	testOption.bindKeyFlag(testViper, testFlag)
	assert.EqualValues(t, expectedValue, testViper.GetString(testOption.Param))
}

func TestOption_bindKeyEnvViperNil(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.bindKeyEnv(nil)
	})
}

func TestOption_bindKeyEnvNoKey(t *testing.T) {
	var testOption = &Option{}
	var testViper = viper.New()

	testOption.bindKeyEnv(testViper)
	assert.Empty(t, testViper.AllKeys())
}

func TestOption_bindKeyEnv(t *testing.T) {
	var testOption = &Option{Key: "key", Env: "env"}
	var testViper = viper.New()
	var expectedValue = "expected"

	testOption.bindKeyEnv(testViper)

	if err := os.Setenv("env", expectedValue); err != nil {
		t.Errorf("could not set environment variable [%s]", testOption.Env)
	}

	assert.EqualValues(t, expectedValue, testViper.GetString(testOption.Key))
}

func TestOption_GetEnvKeyWithEnv(t *testing.T) {
	var testOption = &Option{Env: "env"}

	assert.EqualValues(t, testOption.Env, testOption.GetEnvKey())
}

func TestOption_GetEnvKeyWithKey(t *testing.T) {
	var testOption = &Option{Key: "test.key"}
	var expected = "TEST_KEY"

	assert.EqualValues(t, expected, testOption.GetEnvKey())
}

func TestOption_GetEnvKeyWithNothing(t *testing.T) {
	var testOption = &Option{}

	assert.Empty(t, testOption.GetEnvKey())
}

func TestOption_DefaultRepoNil(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.Default(nil)
	})
}

func TestOption_DefaultDefaultParamValue(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var expectedDefault = "defaultValue"
	var testOption = &StringOption{
		Option: Option{
			Param:        "param",
			DefaultValue: "defaultValue",
		},
	}

	testOption.Default(repo)
	assert.EqualValues(t, expectedDefault, repo.GetString(testOption))
}

func TestOption_DefaultDefaultSetterKeyValue(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var expectedDefault = "defaultValue"
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			DefaultSetter: func(repo *Repo) any {
				return expectedDefault
			},
		},
	}

	testOption.Default(repo)
	assert.EqualValues(t, expectedDefault, repo.GetString(testOption))
}

func TestOption_DefinedRepoNil(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.Defined(nil, nil)
	})
}

func TestOption_DefinedNoFlag(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var flagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedValue = "envValue"
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Env: "env",
		},
	}

	testOption.Defined(repo, flagSet)

	if err := os.Setenv(testOption.Env, expectedValue); err != nil {
		t.Errorf("could not set environment variable %s", testOption.Env)
	}

	assert.EqualValues(t, expectedValue, repo.GetString(testOption))
}

func TestOption_DefinedWithFlag(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var flagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedValue = "flagValue"
	var testOption = &StringOption{
		Option: Option{
			Key:   "key",
			Param: "param",
		},
	}

	testOption.Defined(repo, flagSet)

	if err := flagSet.Lookup(testOption.Param).Value.Set(expectedValue); err != nil {
		t.Errorf("could not set value [%s]", err)
	}

	assert.EqualValues(t, expectedValue, repo.GetString(testOption))
}

func TestBoolOption_DefinedFlagsNil(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var testOption = &BoolOption{}

	assert.Panics(t, func() {
		testOption.Defined(repo, nil)
	})
}

func TestBoolOption_DefinedRepoNil(t *testing.T) {
	var testOption = &BoolOption{}

	assert.Panics(t, func() {
		testOption.Defined(nil, nil)
	})
}

func TestBoolOption_Defined(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())
	var flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedParam = "param"
	var expectedShort = "p"
	var expectedUsage = "usage"
	var testOption = &BoolOption{
		Option: Option{
			Param: expectedParam,
			Short: expectedShort,
			Usage: expectedUsage,
		},
	}

	testOption.Defined(repo, flags)

	if flag := flags.Lookup(expectedParam); flag == nil {
		t.Errorf("could not find defined flag [%s]", expectedParam)
	} else {
		assert.EqualValues(t, expectedShort, flag.Shorthand)
		assert.EqualValues(t, expectedUsage, expectedUsage)
		assert.False(t, cast.ToBool(flag.Value.String()))
	}
}

func TestRepo_ChangeWorkDirEmptyDir(t *testing.T) {
	var testRepo = NewRepoWithViper(viper.New())
	var expectedWorkDir = testRepo.WorkDir
	var testOption = &StringOption{}

	testRepo.ChangeWorkDir(testOption)
	assert.EqualValues(t, expectedWorkDir, testRepo.WorkDir)
}

func TestRepo_ChangeWorkDirOptionNil(t *testing.T) {
	var repo = NewRepoWithViper(viper.New())

	assert.Panics(t, func() {
		repo.ChangeWorkDir(nil)
	})
}

func TestRepo_ChangeWorkDir(t *testing.T) {
	var testRepo = NewRepoWithViper(viper.New())
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
	var buff bytes.Buffer
	var testRepo = NewRepoWith(viper.New(), io.Reader(&buff), io.Writer(&buff))
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
	var buff bytes.Buffer
	var expectedOne = "one"
	var expectedTwo = "two"
	var testRepo = NewRepoWith(viper.New(), io.Reader(&buff), io.Writer(&buff))
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

func TestRepo_DisplayOptionsNoOptions(t *testing.T) {
	var buff bytes.Buffer
	var testRepo = NewRepoWith(viper.New(), io.Reader(&buff), io.Writer(&buff))

	testRepo.DisplayOptions()
	assert.Empty(t, buff)
}

func TestRepo_LogDebug(t *testing.T) {
	var buff bytes.Buffer
	var expectedString = "arg"
	var testRepo = NewRepoWithViper(viper.New())

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
	var testRepo = NewRepoWithViper(viper.New())

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
	var testRepo = NewRepoWithViper(viper.New())

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
	var testRepo = NewRepoWithViper(viper.New())

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
	var testRepo = NewRepoWithViper(viper.New())

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
	var testRepo = NewRepoWithViper(viper.New())

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

func TestRepo_FromWorkDirAbs(t *testing.T) {
	var expectedParam = "param"
	var expectedValue = "/path"
	var testRepo = NewRepoWithViper(viper.New())
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
	var testRepo = NewRepoWithViper(viper.New())
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
	var testRepo = NewRepoWithViper(viper.New())
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
	var testRepo = NewRepoWithViper(viper.New())
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
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
	var testOption = &Option{Key: "key"}

	testViper.SetDefault(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetParam(t *testing.T) {
	var expectedValue = "value"
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
	var testOption = &Option{Param: "param"}

	testViper.SetDefault(testOption.Param, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetDefaultValue(t *testing.T) {
	var expectedValue = "expected"
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
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
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
	var testOption = &Option{
		Param:        "param",
		DefaultValue: "$value",
	}

	_ = os.Setenv("value", expectedValue)
	testViper.SetDefault(testOption.Param, "$value")
	assert.EqualValues(t, expectedValue, testRepo.Get(testOption))
}

func TestRepo_GetBoolNoResult(t *testing.T) {
	var testRepo = NewRepoWithViper(viper.New())
	var testOption = &BoolOption{}

	assert.False(t, testRepo.GetBool(testOption))
}

func TestRepo_GetBool(t *testing.T) {
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
	var testOption = &BoolOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, "true")
	assert.True(t, testRepo.GetBool(testOption))
}

func TestRepo_GetString(t *testing.T) {
	var expectedValue = "value"
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
		},
	}

	testViper.Set(testOption.Key, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetString(testOption))
}

func TestRepo_GetStringInvalid(t *testing.T) {
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
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
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
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
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)

	testViper.Set(testKey, expectedValue)
	assert.EqualValues(t, expectedValue, testRepo.GetValue(testKey))
}

func TestRepo_GetWorkspaceEmpty(t *testing.T) {
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)

	assert.Empty(t, testRepo.GetWorkspace())
}

func TestRepo_GetWorkspace(t *testing.T) {
	var expectedValue = "value"
	var testViper = viper.New()
	var testRepo = NewRepoWithViper(testViper)
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
	var testRepo = NewRepoWithViper(viper.New())

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
	assert.Contains(t, buff.String(), testRepo.UserPath)
}
