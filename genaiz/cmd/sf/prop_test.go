package sf

import (
	"bytes"
	"io"
	"os"
	"path/filepath"
	"regexp"
	"testing"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cli/options"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
)

func TestPropSpecExecutor_Add(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedTypeOption = cli.Options.PropSpecs.Type().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
		BuildStringOption()
	var expectedNameOption = cli.Options.PropSpecs.Name().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Name).
		BuildStringOption()
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.DefaultValue).
		BuildStringOption()
	var testExecutor = newPropAddExecutorFactory(testLedger, testCli, NewAddSpecOptions())(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "37"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testViper.Set(expectedTypeOption.Key, broker.PropSpecTypeInt)
	testViper.Set(expectedNameOption.Key, expectedName)
	testViper.Set(expectedDescOption.Key, expectedDescription)
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	assert.NoError(t, testExecutor.Add(expectedKey))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(expectedTypeOption.Param+`:[\s\t]*`+broker.PropSpecTypeInt), actual)
	assert.Regexp(t, regexp.MustCompile(expectedNameOption.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(expectedDescOption.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(expectedDefaultOption.Param+`:[\s\t]*`+expectedDefaultValue), actual)
}

func TestPropSpecExecutor_Add_IllegalDefaultValue(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedTypeOption = cli.Options.PropSpecs.Type().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
		BuildStringOption()
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.DefaultValue).
		BuildStringOption()
	var testExecutor = newPropAddExecutorFactory(testLedger, testCli, NewAddSpecOptions())(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "notBoolean"
	var expectedDescription = "expectedDescription"

	testViper.Set(expectedTypeOption.Key, broker.PropSpecTypeBoolean)
	testViper.Set(expectedDescOption.Key, expectedDescription)
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	assert.Error(t, testExecutor.Add(expectedKey))
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Add_KeyExistsError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedTypeOption = cli.Options.PropSpecs.Type().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
		BuildStringOption()
	var expectedNameOption = cli.Options.PropSpecs.Name().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Name).
		BuildStringOption()
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.DefaultValue).
		BuildStringOption()
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, NewAddSpecOptions())
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testViper.Set(expectedTypeOption.Key, broker.PropSpecTypeInt)
	testViper.Set(expectedNameOption.Key, expectedName)
	testViper.Set(expectedDescOption.Key, expectedDescription)
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key": expectedKey,
		},
	})
	assert.Error(t, testExecutor.Add(expectedKey))
}

func TestPropSpecExecutor_Display_Nothing(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewAddSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, testOptions)

	testExecutor.Display()
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Edit(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedNameOption = cli.Options.PropSpecs.Name().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Name).
		BuildStringOption()
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
		BuildStringOption()
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, NewEditSpecOptions())
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testViper.Set(expectedNameOption.Key, expectedName)
	testViper.Set(expectedDescOption.Key, expectedDescription)
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": broker.PropSpecTypeString,
		},
	})
	assert.NoError(t, testExecutor.Edit(expectedKey))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(expectedNameOption.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(expectedDescOption.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(expectedDefaultOption.Param+`:[\s\t]*`+expectedDefaultValue), actual)
}

func TestPropSpecExecutor_Edit_IllegalDefaultValue(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedEnumValueAddOption = cli.Options.PropSpecs.EnumAddValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumAddValue).
		BuildStringOption()
	var expectedEnumValueRmOption = cli.Options.PropSpecs.EnumRemoveValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumRemoveValue).
		BuildStringOption()
	var expectedEnumValueOption = cli.Options.PropSpecs.EnumValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumValue).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
		BuildStringOption()
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, NewEditSpecOptions())
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "value3"

	testViper.Set(expectedEnumValueAddOption.Key, expectedDefaultValue)
	testViper.Set(expectedEnumValueRmOption.Key, expectedDefaultValue)
	testViper.Set(expectedEnumValueOption.Key, []string{"value1", "value2"})
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": broker.PropSpecTypeEnum,
		},
	})
	assert.ErrorIs(t, testExecutor.Edit(expectedKey), broker.ErrorPropIllegalEnum)
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Edit_IllegalEnumValue(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedEnumValueOption = cli.Options.PropSpecs.EnumValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumValue).
		BuildStringOption()
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, NewEditSpecOptions())
	var expectedKey = "expectedKey"

	testViper.Set(expectedEnumValueOption.Key, []string{"value1"})
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": broker.PropSpecTypeBoolean,
		},
	})
	assert.ErrorIs(t, testExecutor.Edit(expectedKey), options.ErrorPropSpecTypeEnumConflict)
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Edit_KeyNotFoundError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedNameOption = cli.Options.PropSpecs.Name().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Name).
		BuildStringOption()
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
		BuildStringOption()
	var testExecutor = newPropEditExecutorFactory(testLedger, testCli, NewEditSpecOptions())(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testViper.Set(expectedNameOption.Key, expectedName)
	testViper.Set(expectedDescOption.Key, expectedDescription)
	testViper.Set(expectedDefaultOption.Key, expectedDefaultValue)
	assert.Error(t, testExecutor.Edit(expectedKey))
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Edit_TypeConflictError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var expectedDescOption = cli.Options.PropSpecs.Description().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Description).
		BuildStringOption()
	var expectedDefaultOption = cli.Options.PropSpecs.DefaultValue().
		WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
		BuildStringOption()
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, NewEditSpecOptions())
	var expectedKey = "expectedKey"

	testViper.Set(expectedDescOption.Key, cli.DefaultValueForNil)
	testViper.Set(expectedDefaultOption.Key, cli.DefaultValueForNil)
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": "_invalid",
		},
	})
	assert.ErrorIs(t, testExecutor.Edit(expectedKey), options.ErrorPropSpecTypeConflict)
	actual := testOutput.String()
	assert.Empty(t, actual)
}

func TestPropSpecExecutor_Pretend(t *testing.T) {
	var calledInit bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    NewSfCli(nil, nil, nil),
			Ledger: testLedger,
		},

		updatedPropSpecs: []broker.PropSpec{
			{
				Key:   "expectedKey",
				Value: "expectedValue",
			},
		},

		initTaskFactory: newInitTaskPretendStub(&calledInit),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		var expectedTypeOption = cli.Options.PropSpecs.Type().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
			BuildStringOption()

		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testViper.Set(expectedTypeOption.Key, broker.PropSpecTypeInt)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Pretend()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestPropSpecExecutor_Proceed(t *testing.T) {
	var calledInit bool
	var testDir = t.TempDir()
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var expectedSpecs = []broker.PropSpec{
		{
			Key:   "expectedKey",
			Value: "expectedValue",
		},
	}
	var testExecutor = &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Cli:    NewSfCli(nil, nil, nil),
			Ledger: testLedger,
		},

		updatedPropSpecs: expectedSpecs,

		initTaskFactory: newInitTaskCompleteStub(func(actual *layout.InitParams) {
			calledInit = true
			assert.Equal(t, expectedSpecs, actual.PropSpecs)
		}),
	}

	if fd, err := os.Create(filepath.Join(testDir, "Dockerfile")); err == nil {
		var expectedTypeOption = cli.Options.PropSpecs.Type().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
			BuildStringOption()

		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.InitLogging()
		testViper.Set(expectedTypeOption.Key, broker.PropSpecTypeInt)
		testViper.Set(testExecutor.Cli.optionDockerTag.Key, "tag/tag")
		testExecutor.Proceed()
		assert.True(t, calledInit)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestPropSpecExecutor_Remove(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, nil)
	var expectedKey = "expectedKey"

	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": broker.PropSpecTypeString,
		},
		map[string]interface{}{
			"key":  "anotherOne",
			"type": broker.PropSpecTypeBoolean,
		},
	})
	assert.NoError(t, testExecutor.Remove(expectedKey))
	actual := testOutput.String()
	assert.Contains(t, actual, expectedKey)
}

func TestPropSpecExecutor_Remove_KeyNotFoundError(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newPropRemoveExecutorFactory(testLedger, testCli)(testCmd)
	var expectedKey = "expectedKey"

	assert.Error(t, testExecutor.Remove(expectedKey))
	actual := testOutput.String()
	assert.Empty(t, actual)
}
