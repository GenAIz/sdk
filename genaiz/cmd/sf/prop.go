package sf

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cli/options"
	"genaiz.com/genaiz/cmd/sf/prop"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
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

	if newProp, err = pse.Build(pse.Ledger, &broker.PropSpec{Key: key}); err == nil {
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
		if updated, err := pse.Build(pse.Ledger, &list[i]); err == nil {
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
	var plan = task.NewPlan("PropSpec", pse.Ledger.Logger)

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
	options.PropSpecOptions
}

func NewProp(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var addSpecOptions = NewAddSpecOptions()
	var editSpecOptions = NewEditSpecOptions()
	var addSpecCmd = prop.NewAddSpec(newPropAddExecutorFactory(ledger, sfCli, addSpecOptions), config.Validation.ValidateEnvKey)
	var rmPropCmd = prop.NewRemoveSpec(newPropRemoveExecutorFactory(ledger, sfCli))
	var editPropCmd = prop.NewEditSpec(newPropEditExecutorFactory(ledger, sfCli, editSpecOptions))
	var propCmd = &cobra.Command{
		Use:     "prop",
		Aliases: []string{"pr"},
		Short:   "Manages property specifications for Smart Functions",
	}

	propCmd.AddCommand(addSpecCmd)
	propCmd.AddCommand(editPropCmd)
	propCmd.AddCommand(rmPropCmd)
	propCmd.AddCommand(prop.NewEnv(ledger, &sfCli.BaseCli))
	ledger.Register(addSpecCmd, addSpecOptions.Definers()...)
	ledger.Register(editPropCmd, editSpecOptions.Definers()...)
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
	var propSpecOptions = options.NewPropSpecAddOptionsBuilder().
		WithOptionDefaultValue(cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.DefaultValue).
			BuildStringOption()).
		WithOptionDescription(cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Description).
			BuildStringOption()).
		WithOptionEnumValue(cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.EnumValue).
			BuildListOption()).
		WithOptionName(cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption()).
		WithOptionType(cli.Options.PropSpecs.Type().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecAdd.Type).
			WithDefaultValue("STRING").
			BuildStringOption())

	return &PropSpecOptions{
		PropSpecOptions: propSpecOptions.Build(),
	}
}

func NewEditSpecOptions() *PropSpecOptions {
	var propSpecOptions = options.NewPropSpecEditOptionsBuilder().
		WithOptionDefaultValue(cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.DefaultValue).
			// That is for allowing default-value="" for clearing it
			WithDefaultValue(cli.DefaultValueForNil).
			BuildStringOption()).
		WithOptionDescription(cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Description).
			WithDefaultValue(cli.DefaultValueForNil).
			BuildStringOption()).
		WithOptionEnumAdd(cli.Options.PropSpecs.EnumAddValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumAddValue).
			BuildListOption()).
		WithOptionEnumRemove(cli.Options.PropSpecs.EnumRemoveValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumRemoveValue).
			BuildListOption()).
		WithOptionEnumValue(cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.EnumValue).
			BuildListOption()).
		WithOptionName(cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.Function.Publish.PropSpecEdit.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption())

	return &PropSpecOptions{
		PropSpecOptions: propSpecOptions.Build(),
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
