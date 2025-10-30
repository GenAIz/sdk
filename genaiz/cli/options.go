package cli

import (
	"fmt"
	"path/filepath"
	"strings"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

var (
	Options = &cliOptions{
		Accounts: accountOptions{
			Host: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("host")
			},
			Password: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Account.Login.Password)
			},
			Refresh: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Account.Login.Refresh).
					WithParam("refresh").
					WithShort("r").
					WithDefaultValue("false")
			},
			Username: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("username").
					WithShort("u")
			},
		},
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
		Docker: dockerOptions{
			ContainerName: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithShort("n").
					WithUsage("name of the container")
			},
			ContainerPrefix: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("prefix").
					WithShort("p").
					WithUsage("prefix used in container naming").
					WithUsage("container names will be suffixed with an increasing counter, based on the existing container list").
					WithValidator(config.Validation.Component).
					Optional(true)
			},
			ContextPath: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Context).
					WithParam("context").
					WithShort("c").
					WithUsage("Docker build context path, defaults to the work dir if not specified").
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return ledger.WorkDir
					}).
					WithValidator(config.Validation.DirExists)
			},
			EnvFile: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("env-file").
					WithUsage("path of an environment file which will be supplied to the container when running a Smart Function").
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return filepath.Join(ledger.WorkDir, ".env")
					})
			},
			EnvVar: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("env").
					WithShort("e").
					WithUsage("environment variables can be specified with a KEY=VALUE list or repeated options").
					WithValidator(config.Validation.EnvKeyValue)
			},
			FilePath: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.File).
					WithParam("file").
					WithShort("f").
					WithUsage("Docker file used to build, defaults to Dockerfile in the current work dir").
					WithDefaultGetter(func(ledger *config.Ledger) any {
						return filepath.Join(ledger.WorkDir, "Dockerfile")
					}).
					WithValidator(config.Validation.FileExists)
			},
			Image: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("image").
					WithUsage("reference to an image by repository, tag or no tag").
					WithUsage("by default the smart function built will be used, and there is no need to specify the image")
			},
			Label: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Label).
					WithParam("label").
					WithUsage("enables labelling, creating a supplementary image layer to hold metadata used with --prune to remove dangling images").
					WithDefaultValue("false")
			},
			Legacy: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Legacy).
					WithParam("legacy-builder").
					WithUsage("enables legacy building using the Moby library instead of the Docker Buildx plugin").
					WithDefaultValue("false")
			},
			NoCache: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.NoCache).
					WithParam("no-cache").
					WithUsage("disables caching of some build layers").
					WithDefaultValue("false")
			},
			Preserve: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("preserve").
					WithUsage("preserves the container after it exits").
					WithDefaultValue("false")
			},
			Prune: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Prune).
					WithParam("prune").
					WithUsage("enables pruning, removing dangling images for the same repository built, this requires --label to work").
					WithDefaultValue("false")
			},
			Replace: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Start.Replace).
					WithParam("replace").
					WithShort("r").
					WithUsage("removes any previous containers before creating a new one").
					WithDefaultValue("false")
			},
			Tag: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Tag).
					WithParam("tag").
					WithUsage("tag the smart function image locally, defaults to the work dir name").
					WithValidator(config.Validation.Repository).
					WithDefaultSetter(func(ledger *config.Ledger) any {
						var parent = filepath.Base(filepath.Dir(ledger.WorkDir))
						var fn = filepath.Base(ledger.WorkDir)

						return fmt.Sprintf("%s/%s", parent, fn)
					})
			},
			Version: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Version).
					WithParam("version").
					WithShort("v").
					WithUsage("build version of the smart image").
					WithDefaultValue("latest")
			},
		},
		Functions: functionOptions{
			Arches: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("arch").
					WithUsage("a list of architectures supported by the function. Supported: x86, x86_64, arm and arm64").
					WithValidator(config.AllFromEnumerated(layout.ArchTypes))
			},
			Extras: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Extras)
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
			MountLog: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("mount-log").
					WithUsage("path of the log files folder").
					WithValidator(config.Validation.DirCreated).
					Optional(true)
			},
			MountOutput: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("mount-out").
					WithUsage("path of the output files folder").
					WithValidator(config.Validation.DirCreated).
					Optional(true)
			},
			MountVar: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("mount-var").
					WithUsage("path of the var files folder").
					WithValidator(config.Validation.DirCreated).
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
		Solutions: solutionOptions{
			Broker: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("broker").
					WithUsage("a publishing broker url")
			},
			Description: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("description").
					WithUsage("description of the solution").
					WithValidator(config.Validation.Blob)
			},
			FunctionArches: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Arches).
					WithValidator(config.AllFromEnumerated(layout.ArchTypes))
			},
			FunctionDesc: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Description).
					WithValidator(config.Validation.Blob)
			},
			FunctionHandle: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Handle).
					WithValidator(config.Validation.Handle)
			},
			FunctionName: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Name).
					WithValidator(config.Validation.Name)
			},
			FunctionOem: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Oem).
					WithValidator(config.Validation.Oem)
			},
			FunctionType: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Type).
					WithDefaultValue(layout.FunctionTypeFunction).
					WithValidator(config.AnyOfEnumerated(layout.FunctionTypes))
			},
			FunctionVersion: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Publish.Version).
					WithValidator(config.Validation.Version)
			},
			Handle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("handle").
					WithUsage("handle of the solution").
					WithValidator(config.Validation.Handle)
			},
			LogFormat: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Solution.Log.Format).
					WithParam("log-format").
					WithUsage("log format as supported by Logrus. Also supports \"json\" for structured logging").
					WithDefaultValue("[%time%|%lvl%] %msg%")
			},
			LogLevel: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Solution.Log.Level).
					WithParam("log-level").
					WithUsage("log level for controlling logging details").
					WithUsage("Supported case insensitive values: debug, d, error e, info, i, quiet q, trace t, warning and w").
					WithDefaultValue("quiet")
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithShort("n").
					WithUsage("name of the solution").
					WithValidator(config.Validation.RequiredName)
			},
			Oem: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("oem").
					WithUsage("oem of the solution")
			},
			Version: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("version").
					WithUsage("version of the solution").
					WithDefaultValue("0.1.0")
			},
			WorkflowDesc: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("workflow-desc").
					WithUsage("description of the default workflow").
					WithDefaultValue("default workflow").
					WithValidator(config.Validation.Blob)
			},
			WorkflowHandle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("workflow-handle").
					WithUsage("handle of the default workflow").
					WithDefaultValue("default").
					WithValidator(config.Validation.Handle)
			},
			WorkflowName: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("workflow-name").
					WithUsage("name of the default workflow").
					WithDefaultValue("Default Workflow").
					WithValidator(config.Validation.RequiredName)
			},
		},
	}
)

type cliOptions struct {
	Accounts  accountOptions
	Configs   configOptions
	Docker    dockerOptions
	Functions functionOptions
	Modes     modeOptions
	Solutions solutionOptions
}

type accountOptions struct {
	Host     func() OptionBuilder
	Password func() OptionBuilder
	Refresh  func() OptionBuilder
	Username func() OptionBuilder
}

type configOptions struct {
	NoUpdate     func() OptionBuilder
	SolutionPath func() OptionBuilder
	Type         func() OptionBuilder
}

type dockerOptions struct {
	ContainerName   func() OptionBuilder
	ContainerPrefix func() OptionBuilder
	ContextPath     func() OptionBuilder
	EnvFile         func() OptionBuilder
	EnvVar          func() OptionBuilder
	FilePath        func() OptionBuilder
	Image           func() OptionBuilder
	Label           func() OptionBuilder
	Legacy          func() OptionBuilder
	NoCache         func() OptionBuilder
	Replace         func() OptionBuilder
	Preserve        func() OptionBuilder
	Prune           func() OptionBuilder
	Tag             func() OptionBuilder
	Version         func() OptionBuilder
}

type functionOptions struct {
	Arches      func() OptionBuilder
	Extras      func() OptionBuilder
	Handle      func() OptionBuilder
	MountInput  func() OptionBuilder
	MountLog    func() OptionBuilder
	MountOutput func() OptionBuilder
	MountVar    func() OptionBuilder
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

type solutionOptions struct {
	Broker          func() OptionBuilder
	Description     func() OptionBuilder
	FunctionArches  func() OptionBuilder
	FunctionDesc    func() OptionBuilder
	FunctionHandle  func() OptionBuilder
	FunctionName    func() OptionBuilder
	FunctionOem     func() OptionBuilder
	FunctionType    func() OptionBuilder
	FunctionVersion func() OptionBuilder
	Handle          func() OptionBuilder
	LogFormat       func() OptionBuilder
	LogLevel        func() OptionBuilder
	Name            func() OptionBuilder
	Oem             func() OptionBuilder
	Version         func() OptionBuilder
	WorkflowDesc    func() OptionBuilder
	WorkflowHandle  func() OptionBuilder
	WorkflowName    func() OptionBuilder
}

type OptionBuilder interface {
	BuildOption() *config.Option

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
	validated     bool
	validator     config.Validates
}

func (ob *optionBuilder) BuildOption() *config.Option {
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
		Option: *ob.BuildOption(),
	}
}

func (ob *optionBuilder) BuildListOption() *config.ListOption {
	return &config.ListOption{
		Option: *ob.BuildOption(),
	}
}

func (ob *optionBuilder) BuildStringOption() *config.StringOption {
	return &config.StringOption{
		Option: *ob.BuildOption(),
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
