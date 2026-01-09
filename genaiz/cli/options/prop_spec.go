package options

import (
	"errors"
	"slices"
	"strings"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/broker"
)

var (
	ErrorPropSpecTypeConflict     = errors.New("properties spec type is invalid")
	ErrorPropSpecTypeEnumConflict = errors.New("the property spec type does not allow enum values")
	ErrorPropSpecTypeEnumInvalid  = errors.New("the property spec type enum should have enum values")
)

type PropSpecOptions interface {
	Build(*config.Ledger, *broker.PropSpec) (*broker.PropSpec, error)

	Definers() []config.Definer

	IsSecret(*config.Ledger) bool
}

type propSpecOptions struct {
	optionDefaultValue *config.StringOption
	optionDescription  *config.StringOption
	optionEnumAdd      *config.ListOption
	optionEnumRemove   *config.ListOption
	optionEnumValue    *config.ListOption
	optionName         *config.StringOption
	optionSecret       *config.BoolOption
	optionType         *config.StringOption
}

type PropSpecOptionsBuilder interface {
	Build() PropSpecOptions

	WithOptionDefaultValue(*config.StringOption) PropSpecOptionsBuilder

	WithOptionDescription(*config.StringOption) PropSpecOptionsBuilder

	WithOptionEnumAdd(option *config.ListOption) PropSpecOptionsBuilder

	WithOptionEnumRemove(option *config.ListOption) PropSpecOptionsBuilder

	WithOptionEnumValue(*config.ListOption) PropSpecOptionsBuilder

	WithOptionName(*config.StringOption) PropSpecOptionsBuilder

	WithOptionSecret(*config.BoolOption) PropSpecOptionsBuilder

	WithOptionType(*config.StringOption) PropSpecOptionsBuilder
}

type PropSpecAddOptions struct {
	propSpecOptions
}

func (pso *PropSpecAddOptions) Build(ledger *config.Ledger, previous *broker.PropSpec) (*broker.PropSpec, error) {
	var defaultValue = ledger.GetString(pso.optionDefaultValue)
	var enumValues = ledger.GetList(pso.optionEnumValue)
	var name = ledger.GetString(pso.optionName)
	var specType = ledger.GetString(pso.optionType)
	var result *broker.PropSpec

	if strings.EqualFold(specType, broker.PropSpecTypeEnum) &&
		len(enumValues) == 0 {
		return nil, ErrorPropSpecTypeEnumInvalid
	}

	if !strings.EqualFold(specType, broker.PropSpecTypeEnum) &&
		len(enumValues) > 0 {
		return nil, ErrorPropSpecTypeEnumConflict
	}

	if name == "" {
		name = previous.Key
	}

	result = &broker.PropSpec{
		Key:         previous.Key,
		Description: ledger.GetString(pso.optionDescription),
		Name:        name,
		Type:        specType,
		Values:      enumValues,
	}

	if !pso.IsSecret(ledger) && defaultValue != "" {
		if err := result.Validate(defaultValue); err != nil {
			return nil, err
		}

		result.Value = defaultValue
	}

	return result, nil
}

func (pso *PropSpecAddOptions) Definers() []config.Definer {
	var result = []config.Definer{
		pso.optionDefaultValue,
		pso.optionDescription,
		pso.optionEnumValue,
		pso.optionName,
		pso.optionType,
	}

	if pso.optionSecret != nil {
		result = append(result, pso.optionSecret)
	}

	return result
}

func (pso *PropSpecAddOptions) IsSecret(ledger *config.Ledger) bool {
	if pso.optionSecret != nil {
		return ledger.GetBool(pso.optionSecret)
	}

	return false
}

type PropSpecEditOptions struct {
	propSpecOptions
}

func (pso *PropSpecEditOptions) Build(ledger *config.Ledger, previous *broker.PropSpec) (*broker.PropSpec, error) {
	var addEnumValue = ledger.GetList(pso.optionEnumAdd)
	var defaultValue = ledger.GetString(pso.optionDefaultValue)
	var description = ledger.GetString(pso.optionDescription)
	var enumValues = ledger.GetList(pso.optionEnumValue)
	var name = ledger.GetString(pso.optionName)
	var rmEnumValue = ledger.GetList(pso.optionEnumRemove)
	var result = *previous

	// Edge case where the type was edited manually to something invalid in the config
	if !config.AnyOfEnumerated(broker.PropSpecTypes)(result.Type) {
		return nil, ErrorPropSpecTypeConflict
	}

	if defaultValue != cli.DefaultValueForNil {
		result.Value = defaultValue
	}

	if len(enumValues) > 0 {
		result.Values = enumValues
	}

	result.Values = append(result.Values, addEnumValue...)
	result.Values = slices.DeleteFunc(result.Values, func(s string) bool {
		return slices.Contains(rmEnumValue, s)
	})

	if result.Value != "" {
		if err := result.Validate(result.Value); err != nil {
			return nil, err
		}
	}

	if strings.EqualFold(result.Type, broker.PropSpecTypeEnum) &&
		len(result.Values) == 0 {
		return nil, ErrorPropSpecTypeEnumInvalid
	}

	if !strings.EqualFold(result.Type, broker.PropSpecTypeEnum) &&
		len(result.Values) > 0 {
		return nil, ErrorPropSpecTypeEnumConflict
	}

	if name != "" {
		result.Name = name
	}

	if description != cli.DefaultValueForNil {
		result.Description = description
	}

	return &result, nil
}

func (pso *PropSpecEditOptions) Definers() []config.Definer {
	return []config.Definer{
		pso.optionDefaultValue,
		pso.optionDescription,
		pso.optionEnumAdd,
		pso.optionEnumRemove,
		pso.optionEnumValue,
		pso.optionName,
	}
}

func (pso *PropSpecEditOptions) IsSecret(ledger *config.Ledger) bool {
	_ = ledger
	return false
}

type propSpecOptionsBuilder struct {
	optionDefaultValue *config.StringOption
	optionDescription  *config.StringOption
	optionEnumAdd      *config.ListOption
	optionEnumRemove   *config.ListOption
	optionEnumValue    *config.ListOption
	optionName         *config.StringOption
	optionSecret       *config.BoolOption
	optionType         *config.StringOption

	buildFunc func(builder *propSpecOptionsBuilder) PropSpecOptions
}

func (psb *propSpecOptionsBuilder) Build() PropSpecOptions {
	return psb.buildFunc(psb)
}

func (psb *propSpecOptionsBuilder) WithOptionDefaultValue(option *config.StringOption) PropSpecOptionsBuilder {
	psb.optionDefaultValue = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionDescription(option *config.StringOption) PropSpecOptionsBuilder {
	psb.optionDescription = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionEnumAdd(option *config.ListOption) PropSpecOptionsBuilder {
	psb.optionEnumAdd = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionEnumRemove(option *config.ListOption) PropSpecOptionsBuilder {
	psb.optionEnumRemove = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionEnumValue(option *config.ListOption) PropSpecOptionsBuilder {
	psb.optionEnumValue = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionName(option *config.StringOption) PropSpecOptionsBuilder {
	psb.optionName = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionSecret(option *config.BoolOption) PropSpecOptionsBuilder {
	psb.optionSecret = option
	return psb
}

func (psb *propSpecOptionsBuilder) WithOptionType(option *config.StringOption) PropSpecOptionsBuilder {
	psb.optionType = option
	return psb
}

func NewPropSpecAddOptionsBuilder() PropSpecOptionsBuilder {
	return &propSpecOptionsBuilder{
		buildFunc: func(builder *propSpecOptionsBuilder) PropSpecOptions {
			return &PropSpecAddOptions{
				propSpecOptions: propSpecOptions{
					optionDefaultValue: builder.optionDefaultValue,
					optionDescription:  builder.optionDescription,
					optionEnumValue:    builder.optionEnumValue,
					optionName:         builder.optionName,
					optionSecret:       builder.optionSecret,
					optionType:         builder.optionType,
				},
			}
		},
	}
}

func NewPropSpecEditOptionsBuilder() PropSpecOptionsBuilder {
	return &propSpecOptionsBuilder{
		buildFunc: func(builder *propSpecOptionsBuilder) PropSpecOptions {
			return &PropSpecEditOptions{
				propSpecOptions: propSpecOptions{
					optionDefaultValue: builder.optionDefaultValue,
					optionDescription:  builder.optionDescription,
					optionEnumAdd:      builder.optionEnumAdd,
					optionEnumRemove:   builder.optionEnumRemove,
					optionEnumValue:    builder.optionEnumValue,
					optionName:         builder.optionName,
				},
			}
		},
	}
}
