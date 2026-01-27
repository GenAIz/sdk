package dk

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cli/options"
	"genaiz.com/genaiz/cmd/dk/prop"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type EditLinkTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type PropSpecExecutor struct {
	BaseExecutor
	*PropSpecOptions

	addedPropSpec   *broker.PropSpec
	editedPropSpec  *broker.PropSpec
	removedPropSpec *broker.PropSpec
	editedLink      *broker.DataLink

	dataLinksWriterFactory DataLinksWriterFactory
	editLinkTaskFactory    EditLinkTaskFactory
}

func (pse *PropSpecExecutor) Add(handleArg, key string) error {
	var writer *DataLinksWriter
	var err error

	pse.setOptions(pse.Ledger, handleArg)

	if writer, err = pse.loadDataLinkWriter(); err == nil {
		var oem = pse.Ledger.GetString(pse.optionOem)
		var handle = pse.Ledger.GetString(pse.optionHandle)
		var version = pse.Ledger.GetString(pse.optionVersion)
		var allSpecs []broker.PropSpec
		var addedSpec *broker.PropSpec

		if pse.editedLink = writer.GetDataLink(oem, handle, version); pse.editedLink != nil {
			allSpecs = append(allSpecs, pse.editedLink.PropSpecs...)
			allSpecs = append(allSpecs, pse.editedLink.SecretSpecs...)

			if slices.ContainsFunc(allSpecs, func(specs broker.PropSpec) bool {
				return strings.EqualFold(key, specs.Key)
			}) {
				return fmt.Errorf("the key [%s] already exists", key)
			}

			if addedSpec, err = pse.Build(pse.Ledger, &broker.PropSpec{Key: key}); err == nil {
				if pse.IsSecret(pse.Ledger) {
					pse.editedLink.SecretSpecs = append(pse.editedLink.SecretSpecs, *addedSpec)
				} else {
					pse.editedLink.PropSpecs = append(pse.editedLink.PropSpecs, *addedSpec)
				}

				pse.addedPropSpec = addedSpec
				pse.Cli.Exec(pse.Ledger, pse)
				return nil
			}

			return err
		}

		return fmt.Errorf("data link [%s/%s:%s] not found", oem, handle, version)
	}

	return err
}

func (pse *PropSpecExecutor) Display() {
	var propSpec *broker.PropSpec

	if pse.addedPropSpec != nil {
		_, _ = fmt.Println("Adding the following property specification:")
		propSpec = pse.addedPropSpec
	} else if pse.editedPropSpec != nil {
		_, _ = fmt.Println("Editing the following property specification:")
		propSpec = pse.editedPropSpec
	} else if pse.removedPropSpec != nil {
		_, _ = fmt.Println("Removing the following property specification:")
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
	} else {
		_, _ = fmt.Printf("No action would be taken, no property specification could be found\n")
	}
}

func (pse *PropSpecExecutor) Edit(handleArg, key string) error {
	var writer *DataLinksWriter
	var err error

	pse.setOptions(pse.Ledger, handleArg)

	if writer, err = pse.loadDataLinkWriter(); err == nil {
		var oem = pse.Ledger.GetString(pse.optionOem)
		var handle = pse.Ledger.GetString(pse.optionHandle)
		var version = pse.Ledger.GetString(pse.optionVersion)

		if pse.editedLink = writer.GetDataLink(oem, handle, version); pse.editedLink != nil {
			var editClosure func(edited *broker.PropSpec) error
			var editedSpec *broker.PropSpec

			if editedSpec = pse.editedLink.FindPropSpec(key); editedSpec != nil {
				editClosure = func(edited *broker.PropSpec) error {
					pse.editedLink.ReplacePropSpec(edited)
					return nil
				}
			} else if editedSpec = pse.editedLink.FindSecretSpec(key); editedSpec != nil {
				editClosure = func(edited *broker.PropSpec) error {
					if edited.Value != "" {
						return fmt.Errorf("secret specs cannot have default values")
					}

					pse.editedLink.ReplaceSecretSpec(edited)
					return nil
				}
			} else {
				return fmt.Errorf("the key [%s] could not be found", key)
			}

			if editedSpec, err = pse.Build(pse.Ledger, editedSpec); err == nil {
				if err = editClosure(editedSpec); err == nil {
					pse.editedPropSpec = editedSpec
					pse.Cli.Exec(pse.Ledger, pse)
					return nil
				}
			}

			return err
		}

		return fmt.Errorf("data link [%s/%s:%s] not found", oem, handle, version)
	}

	return err
}

func (pse *PropSpecExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pse.makeConfigParams(pse.optionConfigType, pse.optionUserDefined); err == nil {
		var params = pse.makeDataLinkParams(*configParams)
		var writer = pse.dataLinksWriterFactory(pse.Ledger, configParams.GetConfigPath())

		pse.Ledger.DisplayChangeDir()
		pse.editLinkTaskFactory(writer).Pretend(params, pse.Ledger.Logger)
		return
	}

	lang.HandleExit(err)
}

func (pse *PropSpecExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pse.makeConfigParams(pse.optionConfigType, pse.optionUserDefined); err == nil {
		var params = pse.makeDataLinkParams(*configParams)
		var writer = pse.dataLinksWriterFactory(pse.Ledger, configParams.GetConfigPath())
		var plan = task.NewPlan("DataLink", pse.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, pse.editLinkTaskFactory(writer))
		return
	}

	lang.HandleExit(err)
}

func (pse *PropSpecExecutor) Remove(handleArg, key string) error {
	var writer *DataLinksWriter
	var err error

	pse.setOptions(pse.Ledger, handleArg)

	if writer, err = pse.loadDataLinkWriter(); err == nil {
		var oem = pse.Ledger.GetString(pse.optionOem)
		var handle = pse.Ledger.GetString(pse.optionHandle)
		var version = pse.Ledger.GetString(pse.optionVersion)

		if pse.editedLink = writer.GetDataLink(oem, handle, version); pse.editedLink != nil {
			if pse.removedPropSpec = pse.editedLink.RemovePropSpec(key); pse.removedPropSpec == nil {
				pse.removedPropSpec = pse.editedLink.RemoveSecretSpec(key)
			}

			if pse.removedPropSpec != nil {
				pse.Cli.Exec(pse.Ledger, pse)
			}

			return nil
		}

		return fmt.Errorf("data link [%s/%s:%s] not found", oem, handle, version)
	}

	return err
}

func (pse *PropSpecExecutor) loadDataLinkWriter() (*DataLinksWriter, error) {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = pse.makeConfigParams(pse.optionConfigType, pse.optionUserDefined); err != nil {
		return nil, err
	}

	return pse.dataLinksWriterFactory(pse.Ledger, configParams.GetConfigPath()), nil
}

func (pse *PropSpecExecutor) makeDataLinkParams(configParams shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: pse.Ledger.AuthFile,
		},
		ConfigParams: configParams,
		DataLink:     pse.editedLink,
	}
}

type PropSpecOptions struct {
	options.PropSpecOptions
	BaseOptions
}

func (pso PropSpecOptions) Definers() []config.Definer {
	var result []config.Definer

	if pso.PropSpecOptions != nil {
		result = pso.PropSpecOptions.Definers()
	}

	result = append(result, pso.BaseOptions.allDefiners()...)
	return result
}

func NewProp(ledger *config.Ledger, dkCLi *Cli) *cobra.Command {
	var addSpecOptions = NewAddSpecOptions(ledger)
	var editSpecOptions = NewEditSpecOptions()
	var rmSpecOptions = NewRemoveSpecOptions()
	var addSpecCmd = prop.NewAddSpec(newPropAddExecutorFactory(ledger, dkCLi, addSpecOptions), config.Validation.ValidateEnvKey)
	var editPropCmd = prop.NewEditSpec(newPropEditExecutorFactory(ledger, dkCLi, editSpecOptions))
	var rmPropCmd = prop.NewRemoveSpec(newPropRemoveExecutorFactory(ledger, dkCLi, rmSpecOptions))
	var propCmd = &cobra.Command{
		Use:     "prop",
		Aliases: []string{"pr"},
		Short:   "Manages property specifications for Data Links",
	}

	propCmd.AddCommand(addSpecCmd)
	propCmd.AddCommand(editPropCmd)
	propCmd.AddCommand(rmPropCmd)
	ledger.Register(addSpecCmd, addSpecOptions.Definers()...)
	ledger.Register(editPropCmd, editSpecOptions.Definers()...)
	ledger.Register(rmPropCmd, rmSpecOptions.Definers()...)
	return propCmd
}

func NewPropSpecExecutor(ctx context.Context, ledger *config.Ledger, dkCli *Cli, options *PropSpecOptions) *PropSpecExecutor {
	return &PropSpecExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     dkCli,
			Context: ctx,
			Ledger:  ledger,
		},
		PropSpecOptions: options,

		dataLinksWriterFactory: NewDataLinksWriter,
		editLinkTaskFactory:    broker.NewDataLinkEditTask,
	}
}

func NewAddSpecOptions(ledger *config.Ledger) *PropSpecOptions {
	var secretOption = cli.Options.PropSpecs.Secret().
		WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Secret).
		BuildBoolOption()
	var propSpecOptions = options.NewPropSpecAddOptionsBuilder().
		WithOptionDefaultValue(cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.DefaultValue).
			WithValidator(func(value any) bool {
				return !ledger.GetBool(secretOption) || cast.ToString(value) == ""
			}).
			BuildStringOption()).
		WithOptionDescription(cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Description).
			BuildStringOption()).
		WithOptionEnumValue(cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.EnumValue).
			BuildListOption()).
		WithOptionName(cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption()).
		WithOptionSecret(secretOption).
		WithOptionType(cli.Options.PropSpecs.Type().
			WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Type).
			WithDefaultValue("STRING").
			BuildStringOption()).Build()

	return &PropSpecOptions{
		PropSpecOptions: propSpecOptions,
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.PropSpecAdd.Version).
				WithValidator(config.Validation.Version).
				Optional(true).
				BuildStringOption(),
		},
	}
}

func NewEditSpecOptions() *PropSpecOptions {
	var propSpecOptions = options.NewPropSpecEditOptionsBuilder().
		WithOptionDefaultValue(cli.Options.PropSpecs.DefaultValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.DefaultValue).
			// That is for allowing default-value="" for clearing it
			WithDefaultValue(cli.DefaultValueForNil).
			BuildStringOption()).
		WithOptionDescription(cli.Options.PropSpecs.Description().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.Description).
			WithDefaultValue(cli.DefaultValueForNil).
			BuildStringOption()).
		WithOptionEnumAdd(cli.Options.PropSpecs.EnumAddValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.EnumAddValue).
			BuildListOption()).
		WithOptionEnumRemove(cli.Options.PropSpecs.EnumRemoveValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.EnumRemoveValue).
			BuildListOption()).
		WithOptionEnumValue(cli.Options.PropSpecs.EnumValue().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.EnumValue).
			BuildListOption()).
		WithOptionName(cli.Options.PropSpecs.Name().
			WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.Name).
			WithValidator(config.Validation.Name).
			BuildStringOption()).Build()

	return &PropSpecOptions{
		PropSpecOptions: propSpecOptions,
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.PropSpecEdit.Version).
				WithValidator(config.Validation.Version).
				Optional(true).
				BuildStringOption(),
		},
	}
}

func NewRemoveSpecOptions() *PropSpecOptions {
	return &PropSpecOptions{
		BaseOptions: BaseOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.DataLink.PropSpecRemove.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionHandle: cli.Options.DataLinks.Handle().
				WithKeys(&schema.Genaiz.DataLink.PropSpecRemove.Handle).
				WithValidator(config.Validation.Handle).
				BuildStringOption(),
			optionOem: cli.Options.DataLinks.Oem().
				WithKeys(&schema.Genaiz.DataLink.PropSpecRemove.Oem).
				WithValidator(config.Validation.Oem).
				BuildStringOption(),
			optionUserDefined: cli.Options.DataLinks.UserDefined().
				WithKeys(&schema.Genaiz.DataLink.PropSpecRemove.UserDefined).
				WithDefaultValue("True").
				BuildBoolOption(),
			optionVersion: cli.Options.DataLinks.Version().
				WithKeys(&schema.Genaiz.DataLink.PropSpecRemove.Version).
				WithValidator(config.Validation.Version).
				Optional(true).
				BuildStringOption(),
		},
	}
}

func newPropAddExecutorFactory(ledger *config.Ledger, dkCli *Cli, options *PropSpecOptions) prop.AddExecutorFactory {
	return func(cmd *cobra.Command) prop.AddExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, dkCli, options)
	}
}

func newPropEditExecutorFactory(ledger *config.Ledger, dkCli *Cli, options *PropSpecOptions) prop.EditExecutorFactory {
	return func(cmd *cobra.Command) prop.EditExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, dkCli, options)
	}
}

func newPropRemoveExecutorFactory(ledger *config.Ledger, dkCli *Cli, options *PropSpecOptions) prop.RemoveExecutorFactory {
	return func(cmd *cobra.Command) prop.RemoveExecutor {
		return NewPropSpecExecutor(cmd.Context(), ledger, dkCli, options)
	}
}
