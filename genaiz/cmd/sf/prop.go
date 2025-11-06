package sf

import (
	"context"
	"errors"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/sf/prop"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

const (
	DefaultValueForNil = "__internal_nil"
)

var (
	ErrorTypeConflict     = errors.New("properties spec type is invalid")
	ErrorTypeEnumConflict = errors.New("the property spec type does not allow enum values")
)

type PropSpecExecutor struct {
	BaseExecutor
	*PropSpecOptions

	innerPropSpecs   *config.Option
	addedPropSpec    *broker.PropSpec
	editedPropSpec   *broker.PropSpec
	removedPropSpec  *broker.PropSpec
	updatedPropSpecs []broker.PropSpec

	initTaskFactory InitTaskFactory
}

func (pse *PropSpecExecutor) Add(key string) error {
	var propSpecs = pse.Ledger.Get(pse.innerPropSpecs)
	var list = broker.ListPropSpecs(propSpecs)
	var newProp *broker.PropSpec
	var err error

	if slices.ContainsFunc(list, func(specs broker.PropSpec) bool {
		return strings.EqualFold(key, specs.Key)
	}) {
		return fmt.Errorf("the key [%s] already exists", key)
	}

	if newProp, err = pse.makeAddPropSpec(key); err == nil {
		list = append(list, *newProp)
		pse.addedPropSpec = newProp
		pse.updatedPropSpecs = list
		pse.Cli.Exec(pse.Ledger, pse)
		return nil
	}

	return err
}

func (pse *PropSpecExecutor) Display() {
	var propSpec *broker.PropSpec

	if pse.addedPropSpec != nil {
		_, _ = fmt.Printf("Adding the following property specification:\n")
		propSpec = pse.addedPropSpec
	} else if pse.editedPropSpec != nil {
		_, _ = fmt.Printf("Editing the following property specification:\n")
		propSpec = pse.editedPropSpec
	} else if pse.removedPropSpec != nil {
		_, _ = fmt.Printf("Removing the following property specification:\n")
		propSpec = pse.removedPropSpec
	}

	if propSpec != nil {
		var detailsMap = map[string]string{}

		detailsMap["key"] = propSpec.Key
		detailsMap["description"] = propSpec.Description
		detailsMap["name"] = propSpec.Name
		detailsMap["default-value"] = propSpec.Value
		detailsMap["enum-values"] = strings.Join(propSpec.Values, ", ")
		detailsMap["type"] = propSpec.Type
		pse.Ledger.DisplayOptionsWithMap(&detailsMap)
	}
}

func (pse *PropSpecExecutor) Edit(key string) error {
	var propSpecs = pse.Ledger.Get(pse.innerPropSpecs)
	var list = broker.ListPropSpecs(propSpecs)

	if i := slices.IndexFunc(list, func(spec broker.PropSpec) bool {
		return strings.EqualFold(key, spec.Key)
	}); i >= 0 {
		if updated, err := pse.makeEditPropSpec(&list[i]); err == nil {
			pse.editedPropSpec = updated
			pse.updatedPropSpecs = slices.Replace(list, i, i+1, *updated)
			pse.Cli.Exec(pse.Ledger, pse)
			return nil
		} else {
			return err
		}
	}

	return fmt.Errorf("the key [%s] could not be found", key)
}

func (pse *PropSpecExecutor) Pretend() {
	var params = pse.makeInitParams()
	var builder = makeInitBuilder(pse.Ledger, pse.Cli)

	pse.Ledger.DisplayChangeDir()
	builder.WithPropSpecRemoved(pse.removedPropSpec)
	pse.initTaskFactory(builder).Pretend(params, pse.Ledger.Logger)
}

func (pse *PropSpecExecutor) Proceed() {
	var builder = makeInitBuilder(pse.Ledger, pse.Cli)
	var params = pse.makeInitParams()
	var plan = task.NewPlan("Init", pse.Ledger.Logger)

	plan.PrintReportsOnly = true
	builder.WithPropSpecRemoved(pse.removedPropSpec)
	task.Single(plan, params, pse.initTaskFactory(builder))
}

func (pse *PropSpecExecutor) Remove(key string) error {
	var propSpecs = pse.Ledger.Get(pse.innerPropSpecs)
	var list = broker.ListPropSpecs(propSpecs)
	var updated []broker.PropSpec

	if i := slices.IndexFunc(list, func(spec broker.PropSpec) bool {
		return strings.EqualFold(key, spec.Key)
	}); i >= 0 {
		pse.removedPropSpec = &list[i]

		for j, spec := range list {
			if j != i {
				updated = append(updated, spec)
			}
		}

		pse.updatedPropSpecs = updated
		pse.Cli.Exec(pse.Ledger, pse)
		return nil
	}

	return fmt.Errorf("property key [%s] does not exist in propSpecs", key)
}

func (pse *PropSpecExecutor) makeAddPropSpec(key string) (*broker.PropSpec, error) {
	var defaultValue = pse.Ledger.GetString(pse.optionDefaultValue)
	var enumValues = pse.Ledger.GetList(pse.optionEnumValue)
	var name = pse.Ledger.GetString(pse.optionName)
	var result *broker.PropSpec

	if name == "" {
		name = key
	}

	result = &broker.PropSpec{
		Key:         key,
		Description: pse.Ledger.GetString(pse.optionDescription),
		Name:        name,
		Type:        pse.Ledger.GetString(pse.optionType),
		Value:       defaultValue,
		Values:      enumValues,
	}

	if result.Value != "" {
		if err := result.Validate(result.Value); err != nil {
			return nil, err
		}
	}

	return result, nil
}

func (pse *PropSpecExecutor) makeEditPropSpec(propSpec *broker.PropSpec) (*broker.PropSpec, error) {
	var addEnumValue = pse.Ledger.GetList(pse.optionEnumAdd)
	var defaultValue = pse.Ledger.GetString(pse.optionDefaultValue)
	var description = pse.Ledger.GetString(pse.optionDescription)
	var enumValues = pse.Ledger.GetList(pse.optionEnumValue)
	var name = pse.Ledger.GetString(pse.optionName)
	var rmEnumValue = pse.Ledger.GetList(pse.optionEnumRemove)
	var result = *propSpec

	// Edge case where the type was edited manually to something invalid in the config
	if !config.AnyOfEnumerated(broker.PropSpecTypes)(result.Type) {
		return nil, ErrorTypeConflict
	}

	if defaultValue != DefaultValueForNil {
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

	if len(result.Values) > 0 && result.Type != broker.PropSpecTypeEnum {
		return nil, ErrorTypeEnumConflict
	}

	if name != "" {
		result.Name = name
	}

	if description != DefaultValueForNil {
		result.Description = description
	}

	return &result, nil
}

func (pse *PropSpecExecutor) makeInitParams() *layout.InitParams {
	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: pse.Ledger.ConfigName,
			},
		},
		PropSpecs: pse.updatedPropSpecs,
	}
}

type PropSpecOptions struct {
	optionDefaultValue *config.StringOption
	optionDescription  *config.StringOption
	optionEnumAdd      *config.ListOption
	optionEnumRemove   *config.ListOption
	optionEnumValue    *config.ListOption
	optionName         *config.StringOption
	optionType         *config.StringOption
}

func (pso PropSpecOptions) addDefiners() []config.Definer {
	return []config.Definer{
		pso.optionDefaultValue,
		pso.optionDescription,
		pso.optionEnumValue,
		pso.optionName,
		pso.optionType,
	}
}

func (pso PropSpecOptions) editDefiners() []config.Definer {
	return []config.Definer{
		pso.optionDefaultValue,
		pso.optionDescription,
		pso.optionEnumAdd,
		pso.optionEnumRemove,
		pso.optionEnumValue,
		pso.optionName,
	}
}

func NewProp(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var addSpecOptions = NewAddSpecOptions()
	var editSpecOptions = NewEditSpecOptions()
	var addSpecCmd = prop.NewAddSpec(newPropAddExecutorFactory(ledger, sfCli, addSpecOptions), validateSpecKey)
	var rmPropCmd = prop.NewRemoveSpec(newPropRemoveExecutorFactory(ledger, sfCli))
	var editPropCmd = prop.NewEditSpec(newPropEditExecutorFactory(ledger, sfCli, editSpecOptions))
	var propCmd = &cobra.Command{
		Use:     "prop",
		Aliases: []string{"pr"},
		Short:   "Manages property specifications for Smart Function",
	}

	propCmd.AddCommand(addSpecCmd)
	propCmd.AddCommand(editPropCmd)
	propCmd.AddCommand(rmPropCmd)
	propCmd.AddCommand(prop.NewEnv(ledger, &sfCli.BaseCli))
	ledger.Register(addSpecCmd, addSpecOptions.addDefiners()...)
	ledger.Register(editPropCmd, editSpecOptions.editDefiners()...)
	return propCmd
}

func NewPropSpecExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, options *PropSpecOptions) *PropSpecExecutor {
	return &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     sfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		PropSpecOptions: options,

		innerPropSpecs: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecs).
			BuildOption(),
		initTaskFactory: layout.NewInitTask,
	}
}

func NewAddSpecOptions() *PropSpecOptions {
	return &PropSpecOptions{
		optionDefaultValue: cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.DefaultValue).
			BuildStringOption(),
		optionDescription: cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Description).
			BuildStringOption(),
		optionEnumValue: cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.EnumValue).
			BuildListOption(),
		optionName: cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption(),
		optionType: cli.Options.PropSpecs.Type().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
			WithDefaultValue("STRING").
			BuildStringOption(),
	}
}

func NewEditSpecOptions() *PropSpecOptions {
	return &PropSpecOptions{
		optionDefaultValue: cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
			// That is for allowing default-value="" for clearing it
			WithDefaultValue(DefaultValueForNil).
			BuildStringOption(),
		optionDescription: cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Description).
			WithDefaultValue(DefaultValueForNil).
			BuildStringOption(),
		optionEnumAdd: cli.Options.PropSpecs.EnumAddValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumAddValue).
			BuildListOption(),
		optionEnumRemove: cli.Options.PropSpecs.EnumRemoveValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumRemoveValue).
			BuildListOption(),
		optionEnumValue: cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumValue).
			BuildListOption(),
		optionName: cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption(),
	}
}

func newPropAddExecutorFactory(ledger *config.Ledger, sfCli *Cli, options *PropSpecOptions) prop.AddExecutorFactory {
	return func(cmd *cobra.Command) prop.AddExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, sfCli, options)
	}
}

func newPropEditExecutorFactory(ledger *config.Ledger, sfCli *Cli, options *PropSpecOptions) prop.EditExecutorFactory {
	return func(cmd *cobra.Command) prop.EditExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, sfCli, options)
	}
}

func newPropRemoveExecutorFactory(ledger *config.Ledger, sfCli *Cli) prop.RemoveExecutorFactory {
	return func(cmd *cobra.Command) prop.RemoveExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, sfCli, nil)
	}
}

func validateSpecKey(key string) error {
	if !config.Validation.EnvKey(key) {
		return fmt.Errorf("[%s] is not a valid environment key", key)
	}

	return nil
}
