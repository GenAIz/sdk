package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"testing"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

func Test_OptionsAccountsHost(t *testing.T) {
	var testOption = Options.Accounts.Host().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.Empty(t, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
}

func Test_OptionAccountsPassword(t *testing.T) {
	var testOption = Options.Accounts.Password().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Account.Login.Password.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Account.Login.Password.Env, testOption.Env)
}

func Test_OptionAccountsRefresh(t *testing.T) {
	var testOption = Options.Accounts.Refresh().BuildBoolOption()

	assert.Equal(t, schema.Genaiz.Account.Login.Refresh.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Account.Login.Refresh.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionAccountsUsername(t *testing.T) {
	var testOption = Options.Accounts.Username().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.Empty(t, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
}

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

func Test_OptionsDockerContextPath(t *testing.T) {
	var expectedDir = t.TempDir()
	var testOption = Options.Docker.ContextPath().BuildStringOption()
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()

	assert.Equal(t, schema.Genaiz.Function.Build.Context.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Context.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	testLedger.WorkDir = expectedDir
	assert.Equal(t, expectedDir, testOption.DefaultGetter(testLedger))
	assert.True(t, testOption.Validator(expectedDir))
	assert.False(t, testOption.Validator(filepath.Join(expectedDir, "_not_exist")))
}

func Test_OptionsDockerFilePath(t *testing.T) {
	var testDir = t.TempDir()
	var expectedFile = filepath.Join(testDir, "Dockerfile")
	var testOption = Options.Docker.FilePath().BuildStringOption()
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()

	if fd, err := os.Create(expectedFile); err == nil {
		defer filez.CloseSilently(fd)
		assert.Equal(t, schema.Genaiz.Function.Build.File.Doc, testOption.Key)
		assert.Equal(t, schema.Genaiz.Function.Build.File.Env, testOption.Env)
		assert.NotEmpty(t, testOption.Param)
		assert.NotEmpty(t, testOption.Short)
		assert.NotEmpty(t, testOption.Usage)
		testLedger.WorkDir = testDir
		assert.Equal(t, expectedFile, testOption.DefaultGetter(testLedger))
		assert.True(t, testOption.Validator(expectedFile))
		assert.False(t, testOption.Validator(filepath.Join(testDir, "_not_exist")))
	} else {
		assert.Fail(t, err.Error())
	}
}

func Test_OptionDockerLabel(t *testing.T) {
	var testOption = Options.Docker.Label().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Label.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Label.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionDockerLegacy(t *testing.T) {
	var testOption = Options.Docker.Legacy().BuildBoolOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Legacy.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Legacy.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionDockerNoCache(t *testing.T) {
	var testOption = Options.Docker.NoCache().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionDockerPrune(t *testing.T) {
	var testOption = Options.Docker.Prune().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Prune.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Prune.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionDockerTag(t *testing.T) {
	var testDir = t.TempDir()
	var testOption = Options.Docker.Tag().BuildStringOption()
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()
	var expectedTag = fmt.Sprintf("%s/%s", filepath.Base(filepath.Dir(testDir)), filepath.Base(testDir))

	assert.Equal(t, schema.Genaiz.Function.Build.Tag.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Tag.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	testLedger.WorkDir = testDir
	assert.Equal(t, expectedTag, testOption.DefaultSetter(testLedger))
}

func Test_OptionDockerVersion(t *testing.T) {
	var testOption = Options.Docker.Version().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Version.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Version.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
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

func Test_OptionSolutionsBroker(t *testing.T) {
	var testOption = Options.Solutions.Broker().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionSolutionsLogFormat(t *testing.T) {
	var testOption = Options.Solutions.LogFormat().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Solution.Log.Format.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Solution.Log.Format.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, cast.ToString(testOption.DefaultValue))
}

func Test_OptionSolutionsLogLevel(t *testing.T) {
	var testOption = Options.Solutions.LogLevel().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Solution.Log.Level.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Solution.Log.Level.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, cast.ToString(testOption.DefaultValue))
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
