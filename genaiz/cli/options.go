package cli

import (
	"strings"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

var (
	Options = &cliOptions{
		Configs: configOptions{
			NoUpdate: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-update").
					WithDefaultValue("false").
					WithUsage("do not update local configuration values after successful completion of the command")
			},
			SolutionPath: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("solution-path").
					WithUsage("path of a folder containing a solution configuration file")
			},
			Type: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("config-type").
					WithUsage("sets the format of the configuration file created. Supported values are \"yaml\", \"toml\", \"json\" and \"none\"").
					WithValidator(config.AnyOfEnumerated(shared.ConfigTypes))
			},
		},
		Functions: functionOptions{
			Arches: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("arch").
					WithUsage("a list of architectures supported by the function. Supported: x86, x86_64, arm and arm64").
					WithValidator(config.AllFromEnumerated(layout.ArchTypes))
			},
			Handle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("handle").
					WithUsage("uniquely identifies the function for its publishing oem").
					WithValidator(config.Validation.Handle)
			},
			MountInput: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("mount-in").
					WithUsage("path of the input files folder").
					WithValidator(config.Validation.DirExists).
					Optional(true)
			},
			MountOutput: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("mount-out").
					WithUsage("path of the output files folder").
					WithValidator(config.Validation.DirExists).
					Optional(true)
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithShort("n").
					WithUsage("display name of the function to create").
					WithValidator(config.Validation.Name)
			},
			Oem: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("oem").
					WithUsage("uniquely identifies the publisher of the function").
					WithValidator(config.Validation.Oem)
			},
			Rebuild: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("rebuild").
					WithDefaultValue("false").
					WithUsage("forces a rebuild of the smart function before completing the command")
			},
			Recipe: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("recipe").
					WithShort("r").
					WithUsage("name of a recipe to use as base for the function")
			},
			Type: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("type").
					WithShort("t").
					WithDefaultValue("function").
					WithUsage("type of smart function to create, only \"connector\", \"function\" or \"trigger\" are supported").
					WithValidator(config.AnyOfEnumerated(layout.FunctionTypes))
			},
			Version: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("version").
					WithShort("v").
					WithUsage("initial version to use for the smart function").
					WithValidator(config.Validation.Version)
			},
		},
		Modes: modeOptions{
			Interactive: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("it").
					WithUsage("specify smart function value through input questions").
					WithDefaultValue("false")
			},
		},
	}
)

type cliOptions struct {
	Configs   configOptions
	Functions functionOptions
	Modes     modeOptions
}

type configOptions struct {
	NoUpdate     func() OptionBuilder
	SolutionPath func() OptionBuilder
	Type         func() OptionBuilder
}

type functionOptions struct {
	Arches      func() OptionBuilder
	Handle      func() OptionBuilder
	MountInput  func() OptionBuilder
	MountOutput func() OptionBuilder
	Name        func() OptionBuilder
	Oem         func() OptionBuilder
	Rebuild     func() OptionBuilder
	Recipe      func() OptionBuilder
	Type        func() OptionBuilder
	Version     func() OptionBuilder
}

type modeOptions struct {
	Interactive func() OptionBuilder
}

type OptionBuilder interface {
	BuildBoolOption() *config.BoolOption

	BuildListOption() *config.ListOption

	BuildStringOption() *config.StringOption

	Optional(optional bool) OptionBuilder

	Validated(validated bool) OptionBuilder

	WithDefaultGetter(func(*config.Ledger) any) OptionBuilder

	WithDefaultSetter(func(*config.Ledger) any) OptionBuilder

	WithDefaultValue(string) OptionBuilder

	WithKeys(*schema.Keys) OptionBuilder

	WithParam(string) OptionBuilder

	WithShort(string) OptionBuilder

	WithUsage(string) OptionBuilder

	WithValidator(config.Validates) OptionBuilder
}

type optionBuilder struct {
	defaultGetter func(*config.Ledger) any
	defaultSetter func(*config.Ledger) any
	defaultValue  string
	keys          *schema.Keys
	optional      bool
	param         string
	short         string
	usages        []string
	validator     config.Validates
	validated     bool
}

func (ob *optionBuilder) buildOption() *config.Option {
	var optionValidator config.Validates
	var optionKey, optionEnv string

	if ob.keys != nil {
		optionKey = ob.keys.Doc
		optionEnv = ob.keys.Env
	}

	if ob.validated && ob.validator != nil {
		optionValidator = ob.validator

		if ob.optional {
			optionValidator = config.Optionally(optionValidator)
		}
	}

	return &config.Option{
		Key:           optionKey,
		Env:           optionEnv,
		Param:         ob.param,
		Short:         ob.short,
		Usage:         strings.Join(ob.usages, ", "),
		Validator:     optionValidator,
		DefaultGetter: ob.defaultGetter,
		DefaultSetter: ob.defaultSetter,
		DefaultValue:  ob.defaultValue,
	}
}

func (ob *optionBuilder) BuildBoolOption() *config.BoolOption {
	return &config.BoolOption{
		Option: *ob.buildOption(),
	}
}

func (ob *optionBuilder) BuildListOption() *config.ListOption {
	return &config.ListOption{
		Option: *ob.buildOption(),
	}
}

func (ob *optionBuilder) BuildStringOption() *config.StringOption {
	return &config.StringOption{
		Option: *ob.buildOption(),
	}
}

func (ob *optionBuilder) Optional(optional bool) OptionBuilder {
	ob.optional = optional
	return ob
}

func (ob *optionBuilder) Validated(validated bool) OptionBuilder {
	ob.validated = validated
	return ob
}

func (ob *optionBuilder) WithDefaultGetter(defaultGetter func(*config.Ledger) any) OptionBuilder {
	ob.defaultGetter = defaultGetter
	return ob
}

func (ob *optionBuilder) WithDefaultSetter(defaultSetter func(*config.Ledger) any) OptionBuilder {
	ob.defaultSetter = defaultSetter
	return ob
}

func (ob *optionBuilder) WithDefaultValue(defaultValue string) OptionBuilder {
	ob.defaultValue = defaultValue
	return ob
}

func (ob *optionBuilder) WithKeys(keys *schema.Keys) OptionBuilder {
	ob.keys = keys
	return ob
}

func (ob *optionBuilder) WithParam(param string) OptionBuilder {
	ob.param = param
	return ob
}

func (ob *optionBuilder) WithShort(short string) OptionBuilder {
	ob.short = short
	return ob
}

func (ob *optionBuilder) WithUsage(usage string) OptionBuilder {
	ob.usages = append(ob.usages, usage)
	return ob
}

func (ob *optionBuilder) WithValidator(validator config.Validates) OptionBuilder {
	ob.validator = validator
	ob.validated = true
	return ob
}

func NewOptionBuilder() OptionBuilder {
	return &optionBuilder{}
}
