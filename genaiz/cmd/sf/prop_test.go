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
	var testOptions = NewAddSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newPropAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "37"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
	assert.NoError(t, testExecutor.Add(expectedKey))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionType.Param+`:[\s\t]*`+broker.PropSpecTypeInt), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDefaultValue.Param+`:[\s\t]*`+expectedDefaultValue), actual)
}

func TestPropSpecExecutor_Add_IllegalDefaultValue(t *testing.T) {
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
	var testExecutor = newPropAddExecutorFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "notBoolean"
	var expectedDescription = "expectedDescription"

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeBoolean)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
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
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testLedger.Register(testCmd, testOptions.addDefiners()...)
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
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

	testLedger.Register(testCmd, testOptions.addDefiners()...)
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
	var testOptions = NewEditSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, testOptions)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testLedger.Register(testCmd, testOptions.editDefiners()...)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
	testViper.Set(testExecutor.innerPropSpecs.Key, []interface{}{
		map[string]interface{}{
			"key":  expectedKey,
			"type": broker.PropSpecTypeString,
		},
	})
	assert.NoError(t, testExecutor.Edit(expectedKey))
	actual := testOutput.String()
	assert.Regexp(t, regexp.MustCompile(testOptions.optionName.Param+`:[\s\t]*`+expectedName), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDescription.Param+`:[\s\t]*`+expectedDescription), actual)
	assert.Regexp(t, regexp.MustCompile(testOptions.optionDefaultValue.Param+`:[\s\t]*`+expectedDefaultValue), actual)
}

func TestPropSpecExecutor_Edit_IllegalDefaultValue(t *testing.T) {
	var testOutput = new(bytes.Buffer)
	var testViper = viper.New()
	var testLedger = config.NewBuilder().
		WithViper(testViper).
		WithOutput(io.Writer(testOutput)).
		Build()
	var testOptions = NewEditSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, testOptions)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "value3"

	testLedger.Register(testCmd, testOptions.editDefiners()...)
	testViper.Set(testOptions.optionEnumValue.Key, []string{"value1", "value2"})
	testViper.Set(testOptions.optionEnumAdd.Key, expectedDefaultValue)
	testViper.Set(testOptions.optionEnumRemove.Key, expectedDefaultValue)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
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
	var testOptions = NewEditSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, testOptions)
	var expectedKey = "expectedKey"

	testLedger.Register(testCmd, testOptions.editDefiners()...)
	testViper.Set(testOptions.optionEnumValue.Key, []string{"value1"})
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
	var testOptions = NewEditSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = newPropEditExecutorFactory(testLedger, testCli, testOptions)(testCmd)
	var expectedKey = "expectedKey"
	var expectedDefaultValue = "expectedDefaultValue"
	var expectedName = "expectedName"
	var expectedDescription = "expectedDescription"

	testLedger.Register(testCmd, testOptions.editDefiners()...)
	testViper.Set(testOptions.optionName.Key, expectedName)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
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
	var testOptions = NewEditSpecOptions()
	var testCli = &Cli{
		BaseCli: cli.BaseCli{
			Dry: func(ledger *config.Ledger) bool {
				return true
			},
		},
	}
	var testCmd = &cobra.Command{}
	var testExecutor = NewPropSpecExecutor(testCmd.Context(), testLedger, testCli, testOptions)
	var expectedKey = "expectedKey"

	testLedger.Register(testCmd, testOptions.editDefiners()...)
	testViper.Set(testOptions.optionDefaultValue.Key, cli.DefaultValueForNil)
	testViper.Set(testOptions.optionDescription.Key, cli.DefaultValueForNil)
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
	var testOptions = NewAddSpecOptions()
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
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
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
	var testOptions = NewAddSpecOptions()
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
		filez.CloseSilently(fd)
		testLedger.WorkDir = testDir
		testLedger.InitLogging()
		testLedger.Register(&cobra.Command{}, testOptions.addDefiners()...)
		testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
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
