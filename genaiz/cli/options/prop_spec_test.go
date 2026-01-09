package options

import (
	"testing"

	"github.com/spf13/viper"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
)

func TestPropSpecAddOptions_Build(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
			optionSecret:       &config.BoolOption{Option: config.Option{Key: "secret"}},
			optionType:         &config.StringOption{Option: config.Option{Key: "type"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey"}
	var expectedDefaultValue = "37"
	var expectedDescription = "expectedDescription"

	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionSecret.Key, "False")
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, expectedDefaultValue, actual.Value)
	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, testPrevious.Key, actual.Name)
	assert.Empty(t, actual.Values)
	assert.Equal(t, broker.PropSpecTypeInt, actual.Type)
	assert.Equal(t, testPrevious.Key, actual.Key)
}

func TestPropSpecAddOptions_Build_ErrorEnumNoValues(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
			optionSecret:       &config.BoolOption{Option: config.Option{Key: "secret"}},
			optionType:         &config.StringOption{Option: config.Option{Key: "type"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey"}

	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeEnum)
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, ErrorPropSpecTypeEnumInvalid)
	assert.Nil(t, actual)
}

func TestPropSpecAddOptions_Build_ErrorInvalidDefaultValue(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
			optionSecret:       &config.BoolOption{Option: config.Option{Key: "secret"}},
			optionType:         &config.StringOption{Option: config.Option{Key: "type"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey"}

	testViper.Set(testOptions.optionDefaultValue.Key, "notAnInt")
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeInt)
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, broker.ErrorPropIllegalInt)
	assert.Nil(t, actual)
}

func TestPropSpecAddOptions_Build_ErrorValuesNoEnum(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
			optionSecret:       &config.BoolOption{Option: config.Option{Key: "secret"}},
			optionType:         &config.StringOption{Option: config.Option{Key: "type"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey"}

	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeString)
	testViper.Set(testOptions.optionEnumValue.Key, []string{"invalid"})
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, ErrorPropSpecTypeEnumConflict)
	assert.Nil(t, actual)
}

func TestPropSpecAddOptions_Build_SecretNoName(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
			optionSecret:       &config.BoolOption{Option: config.Option{Key: "secret"}},
			optionType:         &config.StringOption{Option: config.Option{Key: "type"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey"}
	var expectedDefaultValue = "expectedDefault"
	var expectedDescription = "expectedDescription"
	var expectedEnumValue = []string{"expectedValue"}

	testViper.Set(testOptions.optionDefaultValue.Key, expectedDefaultValue)
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionEnumValue.Key, expectedEnumValue)
	testViper.Set(testOptions.optionSecret.Key, "True")
	testViper.Set(testOptions.optionType.Key, broker.PropSpecTypeEnum)
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Empty(t, actual.Value)
	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, testPrevious.Key, actual.Name)
	assert.Equal(t, expectedEnumValue, actual.Values)
	assert.Equal(t, broker.PropSpecTypeEnum, actual.Type)
	assert.Equal(t, testPrevious.Key, actual.Key)
}

func TestPropSpecAddOptions_Definers(t *testing.T) {
	var testOptions = &PropSpecAddOptions{}

	assert.Equal(t, 5, len(testOptions.Definers()))
	testOptions.optionSecret = &config.BoolOption{Option: config.Option{Key: "key"}}
	assert.Equal(t, 6, len(testOptions.Definers()))
}

func TestPropSpecAddOptions_IsSecret(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testEditOptions = &PropSpecAddOptions{
		propSpecOptions{
			optionSecret: &config.BoolOption{Option: config.Option{Key: "key"}},
		},
	}

	assert.False(t, testEditOptions.IsSecret(testLedger))
	testViper.Set(testEditOptions.optionSecret.Key, "True")
	assert.True(t, testEditOptions.IsSecret(testLedger))
}

func TestPropSpecAddOption_IsSecret_NoOption(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testEditOptions = &PropSpecAddOptions{}

	// should always be false
	assert.False(t, testEditOptions.IsSecret(testLedger))
}

func TestPropSpecEditOptions_Build(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var expectedDescription = "expectedDescription"
	var expectedEnumValues = []string{"expected"}
	var expectedEnumAdds = []string{"add1", "rm1"}
	var expectedName = "expectedName"
	var testPrevious = &broker.PropSpec{
		Key:  "previousKey",
		Type: broker.PropSpecTypeEnum,
	}

	testViper.Set(testOptions.optionDefaultValue.Key, expectedEnumValues[0])
	testViper.Set(testOptions.optionDescription.Key, expectedDescription)
	testViper.Set(testOptions.optionEnumAdd.Key, expectedEnumAdds)
	testViper.Set(testOptions.optionEnumRemove.Key, expectedEnumAdds[1])
	testViper.Set(testOptions.optionEnumValue.Key, expectedEnumValues)
	testViper.Set(testOptions.optionName.Key, expectedName)
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, testPrevious.Key, actual.Key)
	assert.Equal(t, testPrevious.Type, actual.Type)
	assert.Equal(t, []string{expectedEnumValues[0], expectedEnumAdds[0]}, actual.Values)
	assert.Equal(t, expectedName, actual.Name)
	assert.Equal(t, expectedDescription, actual.Description)
	assert.Equal(t, expectedEnumValues[0], actual.Value)
}

func TestPropSpecEditOptions_Build_ErrorEnumNoValues(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey", Type: broker.PropSpecTypeEnum}

	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, ErrorPropSpecTypeEnumInvalid)
	assert.Nil(t, actual)
}

func TestPropSpecEditOptions_Build_ErrorInvalidDefaultValue(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey", Type: broker.PropSpecTypeInt}

	testViper.Set(testOptions.optionDefaultValue.Key, "notAnInt")
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, broker.ErrorPropIllegalInt)
	assert.Nil(t, actual)
}

func TestPropSpecEditOptions_Build_ErrorInvalidType(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey", Type: "NotValid"}

	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, ErrorPropSpecTypeConflict)
	assert.Nil(t, actual)
}

func TestPropSpecEditOptions_Build_ErrorValuesNoEnum(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var testPrevious = &broker.PropSpec{Key: "previousKey", Type: broker.PropSpecTypeString}

	testViper.Set(testOptions.optionEnumValue.Key, []string{"invalid"})
	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.ErrorIs(t, err, ErrorPropSpecTypeEnumConflict)
	assert.Nil(t, actual)
}

func TestPropSpecEditOptions_Build_NoChanges(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionDefaultValue: &config.StringOption{Option: config.Option{Key: "defaultValue"}},
			optionDescription:  &config.StringOption{Option: config.Option{Key: "description"}},
			optionEnumAdd:      &config.ListOption{Option: config.Option{Key: "enumAdd"}},
			optionEnumRemove:   &config.ListOption{Option: config.Option{Key: "enumRemove"}},
			optionEnumValue:    &config.ListOption{Option: config.Option{Key: "enumValue"}},
			optionName:         &config.StringOption{Option: config.Option{Key: "name"}},
		},
	}
	var expectedEnumValues = []string{"expected"}
	var testPrevious = &broker.PropSpec{
		Key:    "previousKey",
		Type:   broker.PropSpecTypeEnum,
		Values: expectedEnumValues,
	}

	actual, err := testOptions.Build(testLedger, testPrevious)
	assert.NoError(t, err)
	assert.NotNil(t, actual)
	assert.Equal(t, testPrevious.Key, actual.Key)
	assert.Equal(t, testPrevious.Type, actual.Type)
	assert.Equal(t, testPrevious.Values, actual.Values)
	assert.Empty(t, actual.Name)
	assert.Empty(t, actual.Description)
	assert.Empty(t, actual.Value)
}

func TestPropSpecEditOptions_Definers(t *testing.T) {
	var testOptions = &PropSpecEditOptions{}

	assert.Equal(t, 6, len(testOptions.Definers()))
}

func TestPropSpecEditOptions_IsSecret(t *testing.T) {
	var testViper = viper.New()
	var testLedger = config.NewBuilder().WithViper(testViper).Build()
	var testEditOptions = &PropSpecEditOptions{
		propSpecOptions{
			optionSecret: &config.BoolOption{Option: config.Option{Key: "key"}},
		},
	}

	assert.False(t, testEditOptions.IsSecret(testLedger))
	testViper.Set(testEditOptions.optionSecret.Key, "True")
	// Should always be false for edit
	assert.False(t, testEditOptions.IsSecret(testLedger))
}

func Test_NewPropSpecAddOptionsBuilder(t *testing.T) {
	var expectedDefaultOption = &config.StringOption{Option: config.Option{Key: "defaultOption"}}
	var expectedDescriptionOption = &config.StringOption{Option: config.Option{Key: "descriptionOption"}}
	var expectedEnumAddOption = &config.ListOption{Option: config.Option{Key: "enumAddOption"}}
	var expectedEnumRmOption = &config.ListOption{Option: config.Option{Key: "enumRmOption"}}
	var expectedEnumValueOption = &config.ListOption{Option: config.Option{Key: "enumValueOption"}}
	var expectedNameOption = &config.StringOption{Option: config.Option{Key: "nameOption"}}
	var expectedSecretOption = &config.BoolOption{Option: config.Option{Key: "secretOption"}}
	var expectedTypeOption = &config.StringOption{Option: config.Option{Key: "typeOption"}}
	var testOptions = NewPropSpecAddOptionsBuilder().
		WithOptionDefaultValue(expectedDefaultOption).
		WithOptionDescription(expectedDescriptionOption).
		WithOptionEnumAdd(expectedEnumAddOption).
		WithOptionEnumRemove(expectedEnumRmOption).
		WithOptionEnumValue(expectedEnumValueOption).
		WithOptionName(expectedNameOption).
		WithOptionSecret(expectedSecretOption).
		WithOptionType(expectedTypeOption).
		Build()
	var testDefiners = testOptions.Definers()

	assert.Contains(t, testDefiners, expectedDefaultOption)
	assert.Contains(t, testDefiners, expectedDescriptionOption)
	assert.NotContains(t, testDefiners, expectedEnumAddOption)
	assert.NotContains(t, testDefiners, expectedEnumRmOption)
	assert.Contains(t, testDefiners, expectedEnumValueOption)
	assert.Contains(t, testDefiners, expectedNameOption)
	assert.Contains(t, testDefiners, expectedSecretOption)
	assert.Contains(t, testDefiners, expectedTypeOption)
}

func Test_NewPropSpecEditOptionsBuilder(t *testing.T) {
	var expectedDefaultOption = &config.StringOption{Option: config.Option{Key: "defaultOption"}}
	var expectedDescriptionOption = &config.StringOption{Option: config.Option{Key: "descriptionOption"}}
	var expectedEnumAddOption = &config.ListOption{Option: config.Option{Key: "enumAddOption"}}
	var expectedEnumRmOption = &config.ListOption{Option: config.Option{Key: "enumRmOption"}}
	var expectedEnumValueOption = &config.ListOption{Option: config.Option{Key: "enumValueOption"}}
	var expectedNameOption = &config.StringOption{Option: config.Option{Key: "nameOption"}}
	var expectedSecretOption = &config.BoolOption{Option: config.Option{Key: "secretOption"}}
	var expectedTypeOption = &config.StringOption{Option: config.Option{Key: "typeOption"}}
	var testOptions = NewPropSpecEditOptionsBuilder().
		WithOptionDefaultValue(expectedDefaultOption).
		WithOptionDescription(expectedDescriptionOption).
		WithOptionEnumAdd(expectedEnumAddOption).
		WithOptionEnumRemove(expectedEnumRmOption).
		WithOptionEnumValue(expectedEnumValueOption).
		WithOptionName(expectedNameOption).
		WithOptionSecret(expectedSecretOption).
		WithOptionType(expectedTypeOption).
		Build()
	var testDefiners = testOptions.Definers()

	assert.Contains(t, testDefiners, expectedDefaultOption)
	assert.Contains(t, testDefiners, expectedDescriptionOption)
	assert.Contains(t, testDefiners, expectedEnumAddOption)
	assert.Contains(t, testDefiners, expectedEnumRmOption)
	assert.Contains(t, testDefiners, expectedEnumValueOption)
	assert.Contains(t, testDefiners, expectedNameOption)
	assert.NotContains(t, testDefiners, expectedSecretOption)
	assert.NotContains(t, testDefiners, expectedTypeOption)
}
