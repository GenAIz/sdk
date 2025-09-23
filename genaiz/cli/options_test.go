package cli

import (
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

func Test_OptionsConfigsNoUpdate(t *testing.T) {
	var testOption = Options.Configs.NoUpdate().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsConfigsSolutionPath(t *testing.T) {
	var testOption = Options.Configs.SolutionPath().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsConfigsType(t *testing.T) {
	var testOption = Options.Configs.Type().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(shared.ConfigTypeJson))
	assert.True(t, testOption.Validator(shared.ConfigTypeNone))
	assert.True(t, testOption.Validator(shared.ConfigTypeToml))
	assert.True(t, testOption.Validator(shared.ConfigTypeYaml))
	assert.False(t, testOption.Validator("invalid"))
}

func Test_OptionFunctionsArches(t *testing.T) {
	var testOption = Options.Functions.Arches().BuildListOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(layout.ArchTypeArm))
	assert.True(t, testOption.Validator(layout.ArchTypeArm64))
	assert.True(t, testOption.Validator(layout.ArchTypeX86))
	assert.True(t, testOption.Validator(layout.ArchTypeX86_64))
	assert.False(t, testOption.Validator("invalid"))
}

func Test_OptionFunctionsHandle(t *testing.T) {
	var testOption = Options.Functions.Handle().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("test-handle_test.2"))
	assert.False(t, testOption.Validator("-test-handle"))
	assert.False(t, testOption.Validator("test-handle-"))
	assert.False(t, testOption.Validator("test-handle."))
	assert.False(t, testOption.Validator(".test-handle"))
	assert.False(t, testOption.Validator("_test-handle"))
	assert.False(t, testOption.Validator("test-handle_"))
	assert.False(t, testOption.Validator(""))
}

func Test_OptionFunctionsMountInput(t *testing.T) {
	var testOption = Options.Functions.MountInput().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.False(t, testOption.Validator(filepath.Join(testDir, "_not_exist")))
}

func Test_OptionFunctionsMountOutput(t *testing.T) {
	var testOption = Options.Functions.MountOutput().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.False(t, testOption.Validator(filepath.Join(testDir, "_not_exist")))
}

func Test_OptionFunctionsName(t *testing.T) {
	var testOption = Options.Functions.Name().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a name"))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionFunctionsOem(t *testing.T) {
	var testOption = Options.Functions.Oem().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("com.oem.test"))
	assert.False(t, testOption.Validator("-test-handle"))
	assert.False(t, testOption.Validator("test-handle-"))
	assert.False(t, testOption.Validator("test-handle."))
	assert.False(t, testOption.Validator(".test-handle"))
	assert.False(t, testOption.Validator("_test-handle"))
	assert.False(t, testOption.Validator("test-handle_"))
	assert.False(t, testOption.Validator("test-handle_"))
	assert.False(t, testOption.Validator(""))
}

func Test_OptionFunctionsRebuild(t *testing.T) {
	var testOption = Options.Functions.Rebuild().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionFunctionsRecipe(t *testing.T) {
	var testOption = Options.Functions.Recipe().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionFunctionsType(t *testing.T) {
	var testOption = Options.Functions.Type().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.Equal(t, layout.FunctionTypeFunction, testOption.DefaultValue)
	assert.True(t, testOption.Validator(layout.FunctionTypeConnector))
	assert.True(t, testOption.Validator(layout.FunctionTypeFunction))
	assert.True(t, testOption.Validator(layout.FunctionTypeTrigger))
}

func Test_OptionFunctionsVersion(t *testing.T) {
	var testOption = Options.Functions.Version().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("0.0.0"))
	assert.False(t, testOption.Validator("01.0"))
	assert.False(t, testOption.Validator("1.00"))
	assert.False(t, testOption.Validator("1.1"))
}

func Test_OptionModesInteractive(t *testing.T) {
	var testOption = Options.Modes.Interactive().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func TestOptionBuilder_Validated(t *testing.T) {
	var called bool
	var expectedValue = "value"
	var testKey = "key"
	var testOption = NewOptionBuilder().
		WithKeys(&schema.Keys{Doc: testKey}).
		WithValidator(func(value any) bool {
			called = true
			return false
		}).
		Validated(false).
		BuildStringOption()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	testViper.Set(testKey, expectedValue)
	assert.Equal(t, expectedValue, testLedger.GetString(testOption))
	assert.False(t, called)
}

func TestOptionBuilder_WithDefaultGetter(t *testing.T) {
	var expectedValue = "value"
	var testKey = "key"
	var testOption = NewOptionBuilder().
		WithKeys(&schema.Keys{Doc: testKey}).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return expectedValue
		}).BuildStringOption()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	assert.Equal(t, expectedValue, testLedger.GetString(testOption))
}

func TestOptionBuilder_WithDefaultSetter(t *testing.T) {
	var expectedValue = "value"
	var testKey = "key"
	var testOption = NewOptionBuilder().
		WithKeys(&schema.Keys{Doc: testKey}).
		WithDefaultSetter(func(ledger *config.Ledger) any {
			return expectedValue
		}).BuildStringOption()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()

	testLedger.Register(&cobra.Command{}, testOption)
	testLedger.InitDefaults()
	assert.Equal(t, expectedValue, testLedger.GetString(testOption))
}

func TestOptionBuilder_WithKeys(t *testing.T) {
	var testKeys = &schema.Keys{Doc: "expectedDocKey", Env: "expectedEnvKey"}
	var testOption = NewOptionBuilder().
		WithKeys(testKeys).
		BuildStringOption()

	assert.Equal(t, testKeys.Doc, testOption.Key)
	assert.Equal(t, testKeys.Env, testOption.Env)
}
