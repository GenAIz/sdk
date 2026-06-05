package cli

import (
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
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

func Test_OptionsAccountsHost(t *testing.T) {
	var testOption = Options.Accounts.Host().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.Empty(t, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
}

func Test_OptionsNoBrowser(t *testing.T) {
	var testOption = Options.Accounts.NoBrowser().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.Empty(t, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsAccountsPassword(t *testing.T) {
	var testOption = Options.Accounts.Password().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Account.Login.Password.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Account.Login.Password.Env, testOption.Env)
}

func Test_OptionsAccountsRefresh(t *testing.T) {
	var testOption = Options.Accounts.Refresh().BuildBoolOption()

	assert.Equal(t, schema.Genaiz.Account.Login.Refresh.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Account.Login.Refresh.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsAccountsUsername(t *testing.T) {
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

func Test_OptionsDataLinksDescription(t *testing.T) {
	var testOption = Options.DataLinks.Description().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("A valid description"))
}

func Test_OptionsDataLinksHandle(t *testing.T) {
	var testOption = Options.DataLinks.Handle().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("valid-handle"))
	assert.False(t, testOption.Validator("__notValid"))
}

func Test_OptionsDataLinksName(t *testing.T) {
	var testOption = Options.DataLinks.Name().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("valid name"))
	assert.False(t, testOption.Validator("this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for this name is too long for "))
}

func Test_OptionsDataLinksNoValidation(t *testing.T) {
	var testOption = Options.DataLinks.NoValidation().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDataLinksOem(t *testing.T) {
	var testOption = Options.DataLinks.Oem().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("valid.oem"))
	assert.False(t, testOption.Validator("invalid.oem."))
}

func Test_OptionsDataLinksPublishedVersion(t *testing.T) {
	var testOption = Options.DataLinks.PublishedVersion().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("1.0.0"))
	assert.False(t, testOption.Validator("1"))
}

func Test_OptionsDataLinksSequence(t *testing.T) {
	var testOption = Options.DataLinks.Sequence().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("1"))
	assert.False(t, testOption.Validator("01"))
	assert.False(t, testOption.Validator("A"))
}

func Test_OptionsDataLinksUserDefined(t *testing.T) {
	var testOption = Options.DataLinks.UserDefined().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsDataLinksVersion(t *testing.T) {
	var testOption = Options.DataLinks.Version().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("1.0.0"))
	assert.False(t, testOption.Validator("1"))
}

func Test_OptionsDataPortsDesc(t *testing.T) {
	var testOption = Options.DataPorts.Desc().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a description"))
}

func Test_OptionsDataPortsName(t *testing.T) {
	var testOption = Options.DataPorts.Name().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a description"))
}

func Test_OptionsDockerContainerName(t *testing.T) {
	var testOption = Options.Docker.ContainerName().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsDockerContainerPrefix(t *testing.T) {
	var testOption = Options.Docker.ContainerPrefix().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
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

func Test_OptionsDockerEnvFile(t *testing.T) {
	var expectedDir = t.TempDir()
	var testOption = Options.Docker.EnvFile().BuildStringOption()
	var testLedger = config.NewBuilder().WithViper(viper.New()).Build()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	testLedger.WorkDir = expectedDir
	assert.Equal(t, filepath.Join(expectedDir, ".env"), testOption.DefaultGetter(testLedger))
}

func Test_OptionsDockerEnvVar(t *testing.T) {
	var testOption = Options.Docker.EnvVar().BuildListOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, testOption.Validator("invalid"))
	assert.True(t, testOption.Validator("CLASSIC_KEY="))
	assert.True(t, testOption.Validator("CLASSIC_KEY=$CLASSIC_VALUE"))
	assert.False(t, testOption.Validator(".hidden_key=..."))
	assert.False(t, testOption.Validator("kebab-key=..."))
	assert.True(t, testOption.Validator("_escaped_value_=value\\=escaped"))
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

func Test_OptionsDockerImage(t *testing.T) {
	var testOption = Options.Docker.Image().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsDockerLabel(t *testing.T) {
	var testOption = Options.Docker.Label().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Label.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Label.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerLegacy(t *testing.T) {
	var testOption = Options.Docker.Legacy().BuildBoolOption()

	assert.Equal(t, schema.Genaiz.Function.Build.LegacyBuilder.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.LegacyBuilder.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerNoCache(t *testing.T) {
	var testOption = Options.Docker.NoCache().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerPlatform(t *testing.T) {
	var testOption = Options.Docker.Platform().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
}

func Test_OptionsDockerPreserve(t *testing.T) {
	var testOption = Options.Docker.Preserve().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerPrune(t *testing.T) {
	var testOption = Options.Docker.Prune().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Prune.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Prune.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerReplace(t *testing.T) {
	var testOption = Options.Docker.Replace().BuildBoolOption()

	assert.NotEmpty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsDockerTag(t *testing.T) {
	var testOption = Options.Docker.Repository().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Repository.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Repository.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsDockerVersion(t *testing.T) {
	var testOption = Options.Docker.Version().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Function.Build.Version.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Function.Build.Version.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
}

func Test_OptionsFunctionsArches(t *testing.T) {
	var testOption = Options.Functions.Arches().BuildListOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(shared.ArchTypeArm))
	assert.True(t, testOption.Validator(shared.ArchTypeArm64))
	assert.True(t, testOption.Validator(shared.ArchTypeX86))
	assert.True(t, testOption.Validator(shared.ArchTypeX86_64))
	assert.False(t, testOption.Validator("invalid"))
}

func Test_OptionsFunctionsHandle(t *testing.T) {
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

func Test_OptionsFunctionsMountLog(t *testing.T) {
	var testOption = Options.Functions.MountLog().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.True(t, testOption.Validator(filepath.Join(testDir, "created")))
}

func Test_OptionsFunctionsMountInput(t *testing.T) {
	var testOption = Options.Functions.MountInput().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.False(t, testOption.Validator(filepath.Join(testDir, "_not_exist")))
}

func Test_OptionsFunctionsMountOutput(t *testing.T) {
	var testOption = Options.Functions.MountOutput().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.True(t, testOption.Validator(filepath.Join(testDir, "created")))
}

func Test_OptionsFunctionsMountVar(t *testing.T) {
	var testOption = Options.Functions.MountVar().BuildStringOption()
	var testDir = t.TempDir()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(""))
	assert.True(t, testOption.Validator(testDir))
	assert.True(t, testOption.Validator(filepath.Join(testDir, "created")))
}

func Test_OptionsFunctionsName(t *testing.T) {
	var testOption = Options.Functions.Name().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a name"))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionsFunctionsNoPropSync(t *testing.T) {
	var testOption = Options.Functions.NoPropSync().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsFunctionsOem(t *testing.T) {
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

func Test_OptionsFunctionsRebuild(t *testing.T) {
	var testOption = Options.Functions.Rebuild().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsFunctionsRecipe(t *testing.T) {
	var testOption = Options.Functions.Recipe().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsFunctionsType(t *testing.T) {
	var testOption = Options.Functions.Type().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.Equal(t, shared.FunctionTypeFunction, testOption.DefaultValue)
	assert.True(t, testOption.Validator(shared.FunctionTypeConnector))
	assert.True(t, testOption.Validator(shared.FunctionTypeFunction))
	assert.True(t, testOption.Validator(shared.FunctionTypeTrigger))
}

func Test_OptionsFunctionsVersion(t *testing.T) {
	var testOption = Options.Functions.Version().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("0.0.0"))
	assert.False(t, testOption.Validator("01.0"))
	assert.False(t, testOption.Validator("1.00"))
	assert.False(t, testOption.Validator("1.1"))
}

func Test_OptionsModesInteractive(t *testing.T) {
	var testOption = Options.Modes.Interactive().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsPrinterJsonPrinter(t *testing.T) {
	var testOption = Options.Printer.JsonPrinter().BuildBoolOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsPropSpecsDefaultValue(t *testing.T) {
	var testOption = Options.PropSpecs.DefaultValue().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsPropSpecsDescription(t *testing.T) {
	var testOption = Options.PropSpecs.Description().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a simple description"))
}

func Test_OptionsPropSpecsEnumAddValue(t *testing.T) {
	var testOption = Options.PropSpecs.EnumAddValue().BuildListOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(" a value"))
}

func Test_OptionsPropSpecsEnumRemoveValue(t *testing.T) {
	var testOption = Options.PropSpecs.EnumRemoveValue().BuildListOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsPropSpecsEnumValue(t *testing.T) {
	var testOption = Options.PropSpecs.EnumValue().BuildListOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(" a value"))
}

func Test_OptionsPropSpecsName(t *testing.T) {
	var testOption = Options.PropSpecs.Name().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsPropSpecsSecret(t *testing.T) {
	var testOption = Options.PropSpecs.Secret().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsPropSpecsType(t *testing.T) {
	var testOption = Options.PropSpecs.Type().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, testOption.Validator("invalid"))
	assert.True(t, testOption.Validator(broker.PropSpecTypeBoolean))
	assert.True(t, testOption.Validator(broker.PropSpecTypeDouble))
	assert.True(t, testOption.Validator(broker.PropSpecTypeEnum))
	assert.True(t, testOption.Validator(broker.PropSpecTypeInt))
	assert.True(t, testOption.Validator(broker.PropSpecTypeString))
}

func Test_OptionsProxyInactive(t *testing.T) {
	var testOption = Options.Proxies.Inactive().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsProxyTcp(t *testing.T) {
	var testOption = Options.Proxies.Tcp().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsProxyUdp(t *testing.T) {
	var testOption = Options.Proxies.Udp().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsSolutionsBroker(t *testing.T) {
	var testOption = Options.Solutions.Broker().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsSolutionsDescription(t *testing.T) {
	var testOption = Options.Solutions.Description().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.Validator)
}

func Test_OptionsSolutionsFunctionArches(t *testing.T) {
	var testOption = Options.Solutions.FunctionArches().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.True(t, testOption.Validator(shared.ArchTypeArm))
	assert.True(t, testOption.Validator(shared.ArchTypeArm64))
	assert.True(t, testOption.Validator(shared.ArchTypeX86))
	assert.True(t, testOption.Validator(shared.ArchTypeX86_64))
	assert.False(t, testOption.Validator("invalid"))
}

func Test_OptionsSolutionsFunctionDesc(t *testing.T) {
	var testOption = Options.Solutions.FunctionDesc().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Validator)
}

func Test_OptionsSolutionsFunctionHandle(t *testing.T) {
	var testOption = Options.Solutions.FunctionHandle().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.True(t, testOption.Validator("test-handle_test.2"))
	assert.False(t, testOption.Validator("-test-handle"))
	assert.False(t, testOption.Validator("test-handle-"))
	assert.False(t, testOption.Validator("test-handle."))
	assert.False(t, testOption.Validator(".test-handle"))
	assert.False(t, testOption.Validator("_test-handle"))
	assert.False(t, testOption.Validator("test-handle_"))
	assert.False(t, testOption.Validator(""))
}

func Test_OptionsSolutionsFunctionName(t *testing.T) {
	var testOption = Options.Solutions.FunctionName().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.True(t, testOption.Validator("a name"))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionsSolutionsFunctionOem(t *testing.T) {
	var testOption = Options.Solutions.FunctionOem().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
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

func Test_OptionsSolutionsFunctionType(t *testing.T) {
	var testOption = Options.Solutions.FunctionType().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.Equal(t, shared.FunctionTypeFunction, testOption.DefaultValue)
	assert.True(t, testOption.Validator(shared.FunctionTypeConnector))
	assert.True(t, testOption.Validator(shared.FunctionTypeFunction))
	assert.True(t, testOption.Validator(shared.FunctionTypeTrigger))
}

func Test_OptionsSolutionsFunctionVersion(t *testing.T) {
	var testOption = Options.Solutions.FunctionVersion().BuildStringOption()

	assert.NotEmpty(t, testOption.Key)
	// Needs to be empty, there can be several functions, and they would collide on the cmd line
	assert.Empty(t, testOption.Param)
	assert.True(t, testOption.Validator("0.0.0"))
	assert.False(t, testOption.Validator("01.0"))
	assert.False(t, testOption.Validator("1.00"))
	assert.False(t, testOption.Validator("1.1"))
}

func Test_OptionsSolutionsHandle(t *testing.T) {
	var testOption = Options.Solutions.Handle().BuildStringOption()

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

func Test_OptionsSolutionsLogFormat(t *testing.T) {
	var testOption = Options.Solutions.LogFormat().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Solution.Log.Format.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Solution.Log.Format.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, cast.ToString(testOption.DefaultValue))
}

func Test_OptionsSolutionsLogLevel(t *testing.T) {
	var testOption = Options.Solutions.LogLevel().BuildStringOption()

	assert.Equal(t, schema.Genaiz.Solution.Log.Level.Doc, testOption.Key)
	assert.Equal(t, schema.Genaiz.Solution.Log.Level.Env, testOption.Env)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, cast.ToString(testOption.DefaultValue))
}

func Test_OptionsSolutionsName(t *testing.T) {
	var testOption = Options.Solutions.Name().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Short)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a name"))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionsSolutionsOem(t *testing.T) {
	var testOption = Options.Solutions.Oem().BuildStringOption()

	assert.NotEmpty(t, testOption.Param)
	assert.Empty(t, testOption.Validator)
}

func Test_OptionsSolutionsVersion(t *testing.T) {
	var testOption = Options.Solutions.Version().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
}

func Test_OptionsSolutionsWorkflowDesc(t *testing.T) {
	var testOption = Options.Solutions.WorkflowDesc().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
	assert.NotEmpty(t, testOption.Validator)
}

func Test_OptionsSolutionsWorkflowHandle(t *testing.T) {
	var testOption = Options.Solutions.WorkflowHandle().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
	assert.True(t, testOption.Validator(testOption.DefaultValue))
	assert.False(t, testOption.Validator("-test-handle"))
	assert.False(t, testOption.Validator("test-handle-"))
	assert.False(t, testOption.Validator("test-handle."))
	assert.False(t, testOption.Validator(".test-handle"))
	assert.False(t, testOption.Validator("_test-handle"))
	assert.False(t, testOption.Validator("test-handle_"))
	assert.False(t, testOption.Validator(""))
}

func Test_OptionsSolutionsWorkflowName(t *testing.T) {
	var testOption = Options.Solutions.WorkflowName().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.DefaultValue)
	assert.True(t, testOption.Validator(testOption.DefaultValue))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionsWorkflowDescription(t *testing.T) {
	var testOption = Options.Workflows.Description().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a valid description"))
}

func Test_OptionsWorkflowHandle(t *testing.T) {
	var testOption = Options.Workflows.Handle().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, testOption.Validator("--not-valid--"))
	assert.True(t, testOption.Validator("function-37"))
}

func Test_OptionsWorkflowName(t *testing.T) {
	var testOption = Options.Workflows.Name().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("a valid name"))
	assert.False(t, testOption.Validator("a name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too longa name too long"))
}

func Test_OptionsWorkflowNoLinkValidation(t *testing.T) {
	var testOption = Options.Workflows.NoLinkValidation().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkflowNoPropSync(t *testing.T) {
	var testOption = Options.Workflows.NoPropSync().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkflowNoPropValidation(t *testing.T) {
	var testOption = Options.Workflows.NoPropValidation().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkflowSFHandle(t *testing.T) {
	var testOption = Options.Workflows.SfHandle().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, testOption.Validator("--not-valid--"))
	assert.True(t, testOption.Validator("function-37"))
}

func Test_OptionsWorkflowSfOem(t *testing.T) {
	var testOption = Options.Workflows.SfOem().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("com.genaiz"))
	assert.False(t, testOption.Validator("com..genaiz"))
}

func Test_OptionsWorkflowSfSequence(t *testing.T) {
	var testOption = Options.Workflows.SfSequence().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("37"))
	assert.False(t, testOption.Validator("notASequence"))
}

func Test_OptionsWorkflowSfSerialized(t *testing.T) {
	var testOption = Options.Workflows.SfSerialized().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsWorkflowSfVersion(t *testing.T) {
	var testOption = Options.Workflows.SfVersion().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator("1.1.37"))
	assert.False(t, testOption.Validator("1.2"))
}

func Test_OptionsWorkspacesAccount(t *testing.T) {
	var testOption = Options.Workspaces.Account().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
}

func Test_OptionsWorkspacesDateMonthly(t *testing.T) {
	var testOption = Options.Workspaces.DateMonthly().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkspacesDateToday(t *testing.T) {
	var testOption = Options.Workspaces.DateToday().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkspacesDateWeekly(t *testing.T) {
	var testOption = Options.Workspaces.DateWeekly().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkspacesDescription(t *testing.T) {
	var testOption = Options.Workspaces.Description().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.NotEmpty(t, testOption.Validator)
}

func Test_OptionsWorkspacesOwnerOnly(t *testing.T) {
	var testOption = Options.Workspaces.OwnerOnly().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.False(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkspacesRcEnabled(t *testing.T) {
	var testOption = Options.Workspaces.RcEnabled().BuildBoolOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, cast.ToBool(testOption.DefaultValue))
}

func Test_OptionsWorkspacesVisibility(t *testing.T) {
	var testOption = Options.Workspaces.Visibility().BuildStringOption()

	assert.Empty(t, testOption.Key)
	assert.NotEmpty(t, testOption.Param)
	assert.NotEmpty(t, testOption.Usage)
	assert.True(t, testOption.Validator(broker.WorkspaceVisibilityOrg))
	assert.True(t, testOption.Validator(broker.WorkspaceVisibilityPrivate))
	assert.False(t, testOption.Validator("InvalidVisibility"))
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
