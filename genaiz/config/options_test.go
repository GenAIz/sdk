package config

import (
	"os"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"
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

func TestOption_equals(t *testing.T) {
	var testOption = &Option{}
	var expectedOption = &Option{
		Key:          "key",
		Param:        "param",
		Short:        "shot",
		Usage:        "usage",
		Env:          "env",
		Pseudonyms:   []string{"previousKey"},
		DefaultValue: "default",
	}

	assert.False(t, expectedOption.Equals(testOption))
	testOption.Key = expectedOption.Key
	assert.False(t, expectedOption.Equals(testOption))
	testOption.Param = expectedOption.Param
	assert.False(t, expectedOption.Equals(testOption))
	testOption.Short = expectedOption.Short
	assert.False(t, expectedOption.Equals(testOption))
	testOption.Usage = expectedOption.Usage
	assert.False(t, expectedOption.Equals(testOption))
	testOption.Env = expectedOption.Env
	assert.False(t, expectedOption.Equals(testOption))
	testOption.Pseudonyms = expectedOption.Pseudonyms
	assert.False(t, expectedOption.Equals(testOption))
	testOption.DefaultValue = expectedOption.DefaultValue
	assert.True(t, expectedOption.Equals(testOption))
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

func TestOption_DefaultLedgerNil(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.Default(nil)
	})
}

func TestOption_DefaultDefaultParamValue(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var expectedDefault = "defaultValue"
	var testOption = &StringOption{
		Option: Option{
			Param:        "param",
			DefaultValue: "defaultValue",
		},
	}

	testOption.Default(testLedger)
	assert.EqualValues(t, expectedDefault, testLedger.GetString(testOption))
}

func TestOption_DefaultDefaultSetterKeyValue(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var expectedDefault = "defaultValue"
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			DefaultSetter: func(ledger *Ledger) any {
				return expectedDefault
			},
		},
	}

	testOption.Default(testLedger)
	assert.EqualValues(t, expectedDefault, testLedger.GetString(testOption))
}

func TestOption_DefinedLedgerNil(t *testing.T) {
	var testOption = &Option{}

	assert.Panics(t, func() {
		testOption.Defined(nil, nil)
	})
}

func TestOption_DefinedNoFlag(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var flagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedValue = "envValue"
	var testOption = &StringOption{
		Option: Option{
			Key: "key",
			Env: "env",
		},
	}

	testOption.Defined(testLedger, flagSet)

	if err := os.Setenv(testOption.Env, expectedValue); err != nil {
		t.Errorf("could not set environment variable %s", testOption.Env)
	}

	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestOption_DefinedWithFlag(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var flagSet = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedValue = "flagValue"
	var testOption = &StringOption{
		Option: Option{
			Key:   "key",
			Param: "param",
		},
	}

	testOption.Defined(testLedger, flagSet)

	if err := flagSet.Lookup(testOption.Param).Value.Set(expectedValue); err != nil {
		t.Errorf("could not set value [%s]", err)
	}

	assert.EqualValues(t, expectedValue, testLedger.GetString(testOption))
}

func TestBoolOption_DefinedFlagsNil(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &BoolOption{}

	assert.Panics(t, func() {
		testOption.Defined(testLedger, nil)
	})
}

func TestBoolOption_DefinedLedgerNil(t *testing.T) {
	var testOption = &BoolOption{}

	assert.Panics(t, func() {
		testOption.Defined(nil, nil)
	})
}

func TestBoolOption_Defined(t *testing.T) {
	var _, testLedger = newTestConfigs()
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

	testOption.Defined(testLedger, flags)

	if flag := flags.Lookup(expectedParam); flag == nil {
		t.Errorf("could not find defined flag [%s]", expectedParam)
	} else {
		assert.EqualValues(t, expectedShort, flag.Shorthand)
		assert.EqualValues(t, expectedUsage, expectedUsage)
		assert.False(t, cast.ToBool(flag.Value.String()))
	}
}

func TestListOption_DefinedFlagsNil(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var testOption = &ListOption{}

	assert.Panics(t, func() {
		testOption.Defined(testLedger, nil)
	})
}

func TestListOption_DefinedLedgerNil(t *testing.T) {
	var testOption = &ListOption{}

	assert.Panics(t, func() {
		testOption.Defined(nil, nil)
	})
}

func TestListOption_Defined(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var flags = pflag.NewFlagSet("test", pflag.ContinueOnError)
	var expectedParam = "param"
	var expectedShort = "p"
	var expectedUsage = "usage"
	var testOption = &ListOption{
		Option: Option{
			Param: expectedParam,
			Short: expectedShort,
			Usage: expectedUsage,
		},
	}

	testOption.Defined(testLedger, flags)

	if flag := flags.Lookup(expectedParam); flag == nil {
		t.Errorf("could not find defined flag [%s]", expectedParam)
	} else {
		var expectedOne = "one"
		var expectedTwo = "two"

		assert.NoError(t, flag.Value.Set(expectedOne))
		assert.NoError(t, flag.Value.Set(expectedTwo))
		assert.EqualValues(t, expectedShort, flag.Shorthand)
		assert.EqualValues(t, expectedUsage, expectedUsage)
		assert.Contains(t, flag.Value.String(), expectedOne)
		assert.Contains(t, flag.Value.String(), expectedTwo)
	}
}

func TestListValue_GetSlice(t *testing.T) {
	var testValue = newListValue()

	assert.Empty(t, testValue.GetSlice())
}

func TestListValue_Set(t *testing.T) {
	var expectedValue = "value"
	var testValue = newListValue()

	assert.NoError(t, testValue.Set(expectedValue))
	assert.Contains(t, testValue.GetSlice(), expectedValue)
}

func TestListValue_String(t *testing.T) {
	var expectedValue = "value"
	var testValue = newListValue()

	assert.NoError(t, testValue.Set("one"))
	assert.NoError(t, testValue.Set("two"))
	assert.NoError(t, testValue.Set(expectedValue))
	assert.Contains(t, testValue.GetSlice(), expectedValue)
}

func TestListValue_Type(t *testing.T) {
	assert.EqualValues(t, "stringSlice", newListValue().Type())
}

func TestMapOptionsByParam(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var expectedKey = "param"
	var expectedValue = "value"
	var testNoParamOption = &Option{Key: "key"}
	var testNilValueOption = &Option{Param: "nilParam"}
	var testValueOption = &Option{
		Param:        expectedKey,
		DefaultValue: expectedValue,
	}

	testLedger.Register(&cobra.Command{}, testValueOption)
	testLedger.InitDefaults()
	testOptions := MapOptionsByParam(testLedger, testNoParamOption, testNilValueOption, testValueOption)
	assert.NotEmpty(t, testOptions)
	assert.EqualValues(t, expectedValue, testOptions[expectedKey])
}

func TestMapOptionsByEnvKey(t *testing.T) {
	var _, testLedger = newTestConfigs()
	var expectedKey = "KEY"
	var expectedValue = "value"
	var testNoKeyOption = &Option{Param: "param"}
	var testNilValueOption = &Option{Key: "nilParam"}
	var testValueOption = &Option{
		Key:          expectedKey,
		DefaultValue: expectedValue,
	}

	testLedger.Register(&cobra.Command{}, testValueOption)
	testLedger.InitDefaults()
	testOptions := MapOptionsByEnvKey(testLedger, testNoKeyOption, testNilValueOption, testValueOption)
	assert.NotEmpty(t, testOptions)
	assert.EqualValues(t, expectedValue, testOptions[expectedKey])
}
