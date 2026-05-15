package cli

import (
	"path/filepath"
	"strings"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	Options = &cliOptions{
		Accounts: accountOptions{
			Host: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("host")
			},
			NoBrowser: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-browser").
					WithUsage("prevents the shell from redirecting login urls to the system default browser").
					WithDefaultValue("false")
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
		DataLinks: dataLinkOptions{
			Description: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("description").
					WithUsage("description of the data link").
					WithValidator(config.Validation.Blob)
			},
			Handle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("handle").
					WithUsage("handle of the data link").
					WithValidator(config.Validation.Handle)
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithUsage("name of the data link").
					WithValidator(config.Validation.RequiredName)
			},
			NoValidation: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-validation").
					WithUsage("if true, the data link used will only be validated on publishing to the broker").
					WithDefaultValue("false")
			},
			Oem: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("oem").
					WithUsage("oem of the data link").
					WithValidator(config.Validation.Oem)
			},
			PublishedVersion: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("new-version").
					WithUsage("assigned a new version to the published datalink").
					WithValidator(config.Validation.Version)
			},
			Sequence: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sequence").
					WithUsage("a sequence integer, usually for non-released datalinks").
					WithValidator(config.Validation.VersionNumber)
			},
			UserDefined: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("user-defined").
					WithUsage("user defined implies writing the configurations under the $HOME/.config/genaiz folder")
			},
			Version: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("version").
					WithUsage("version of the data link").
					WithValidator(config.Validation.Version)
			},
		},
		DataPorts: dataPortOptions{
			Desc: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("description").
					WithShort("d").
					WithUsage("description of the data port").
					WithValidator(config.Validation.Blob)
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithShort("n").
					WithUsage("name of the data port").
					WithValidator(config.Validation.Name)
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
					WithValidator(config.Validation.EnvKeyPairs)
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
					WithUsage("reference to an image by repository with a version tag or no tag").
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
					WithKeys(&schema.Genaiz.Function.Build.LegacyBuilder).
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
			Platform: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Platform).
					WithParam("platform").
					WithUsage("specifies the build platform of the image").
					WithDefaultValue("linux/amd64")
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
			Repository: func() OptionBuilder {
				return NewOptionBuilder().
					WithKeys(&schema.Genaiz.Function.Build.Repository).
					WithParam("repository").
					WithUsage("repository of the smart function image locally, defaults to the work dir name").
					WithValidator(config.Validation.Repository)
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
					WithValidator(config.AllFromEnumerated(shared.ArchTypes))
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
			NoPropSync: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-prop-sync").
					WithUsage("disables property specification sync when creating a container").
					WithDefaultValue("false")
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
					WithValidator(config.AnyOfEnumerated(shared.FunctionTypes))
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
		Proxies: proxyOptions{
			Inactive: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("inactive").
					WithUsage("sets whether the proxy is active or not").
					WithDefaultValue("false")
			},
			Tcp: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("tcp").
					WithUsage("sets the tcp flag for the proxy")
			},
			Udp: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("udp").
					WithUsage("sets the udp flag for the proxy").
					WithDefaultValue("false")
			},
		},
		PropSpecs: propSpecsOptions{
			DefaultValue: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("default-value").
					WithUsage("a default value to use when the property is not provided")
			},
			Description: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("description").
					WithUsage("a description of what the property is").
					WithValidator(config.Validation.Blob)
			},
			EnumAddValue: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("add-enum-value").
					WithUsage("adds an enum value to the existing set of valid values").
					WithValidator(config.Validation.PropEnum)
			},
			EnumRemoveValue: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("rm-enum-value").
					WithUsage("removes an enum value from the existing set of valid values")
				// No-need to validate, unsuccessful removals should be Warnings, regardless of validity
			},
			EnumValue: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("enum-value").
					WithUsage("specifies a value for the set of valid values on enum properties").
					WithValidator(config.Validation.PropEnum)
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithUsage("the name of the property, as human-readable text")
			},
			Secret: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("secret").
					WithUsage("the property will be defined amongst the secret specifications, which can not have a default value").
					WithDefaultValue("false")
			},
			Type: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("type").
					WithUsage("the type of the property").
					WithUsage("only string, int, double, bool and enum are valid").
					WithValidator(config.AnyOfEnumerated(broker.PropSpecTypes))
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
					WithValidator(config.AllFromEnumerated(shared.ArchTypes))
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
					WithDefaultValue(shared.FunctionTypeFunction).
					WithValidator(config.AnyOfEnumerated(shared.FunctionTypes))
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
					WithDefaultValue("1.0.0")
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
		Workflows: workflowOptions{
			Description: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("description").
					WithUsage("description of the workflow node").
					WithValidator(config.Validation.Blob)
			},
			Handle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("handle").
					WithUsage("handle of the workflow").
					WithValidator(config.Validation.Handle)
			},
			Name: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("name").
					WithUsage("name of the workflow node").
					WithValidator(config.Validation.Name)
			},
			NoLinkValidation: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-validation").
					WithUsage("instructs the command to skip validation of link handles and ports").
					WithDefaultValue("false")
			},
			NoPropSync: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-prop-sync").
					WithUsage("disables property specification sync when creating a container").
					WithDefaultValue("false")
			},
			NoPropValidation: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("no-validation").
					WithUsage("instructs the command to skip prop spec validation of workflow properties").
					WithDefaultValue("false")
			},
			SfHandle: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sf-handle").
					WithUsage("handle of the node smart function").
					WithValidator(config.Validation.Handle)
			},
			SfOem: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sf-oem").
					WithUsage("oem of the node smart function").
					WithValidator(config.Validation.Oem)
			},
			SfSequence: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sf-seq").
					WithUsage("sequence number of the node smart function").
					WithValidator(config.Validation.VersionNumber)
			},
			SfSerialized: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sf").
					WithUsage("serialized string of the smart function, the individual options have precedence")
			},
			SfVersion: func() OptionBuilder {
				return NewOptionBuilder().
					WithParam("sf-version").
					WithUsage("version of the node smart function").
					WithValidator(config.Validation.Version)
			},
		},
	}
)

type accountOptions struct {
	Host      func() OptionBuilder
	NoBrowser func() OptionBuilder
	Password  func() OptionBuilder
	Refresh   func() OptionBuilder
	Username  func() OptionBuilder
}

type cliOptions struct {
	Accounts  accountOptions
	Configs   configOptions
	DataLinks dataLinkOptions
	DataPorts dataPortOptions
	Docker    dockerOptions
	Functions functionOptions
	Modes     modeOptions
	Proxies   proxyOptions
	PropSpecs propSpecsOptions
	Solutions solutionOptions
	Workflows workflowOptions
}

type configOptions struct {
	NoUpdate     func() OptionBuilder
	SolutionPath func() OptionBuilder
	Type         func() OptionBuilder
}

type dataLinkOptions struct {
	Description      func() OptionBuilder
	Handle           func() OptionBuilder
	Name             func() OptionBuilder
	NoValidation     func() OptionBuilder
	Oem              func() OptionBuilder
	PublishedVersion func() OptionBuilder
	Sequence         func() OptionBuilder
	UserDefined      func() OptionBuilder
	Version          func() OptionBuilder
}

type dataPortOptions struct {
	Desc func() OptionBuilder
	Name func() OptionBuilder
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
	Platform        func() OptionBuilder
	Preserve        func() OptionBuilder
	Prune           func() OptionBuilder
	Replace         func() OptionBuilder
	Repository      func() OptionBuilder
	Version         func() OptionBuilder
}

type functionOptions struct {
	Arches      func() OptionBuilder
	Handle      func() OptionBuilder
	MountInput  func() OptionBuilder
	MountLog    func() OptionBuilder
	MountOutput func() OptionBuilder
	MountVar    func() OptionBuilder
	Name        func() OptionBuilder
	NoPropSync  func() OptionBuilder
	Oem         func() OptionBuilder
	Rebuild     func() OptionBuilder
	Recipe      func() OptionBuilder
	Type        func() OptionBuilder
	Version     func() OptionBuilder
}

type modeOptions struct {
	Interactive func() OptionBuilder
}

type propSpecsOptions struct {
	DefaultValue    func() OptionBuilder
	Description     func() OptionBuilder
	EnumAddValue    func() OptionBuilder
	EnumRemoveValue func() OptionBuilder
	EnumValue       func() OptionBuilder
	Name            func() OptionBuilder
	Secret          func() OptionBuilder
	Type            func() OptionBuilder
}

type proxyOptions struct {
	Inactive func() OptionBuilder
	Tcp      func() OptionBuilder
	Udp      func() OptionBuilder
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

type workflowOptions struct {
	Description      func() OptionBuilder
	Handle           func() OptionBuilder
	Name             func() OptionBuilder
	NoLinkValidation func() OptionBuilder
	NoPropValidation func() OptionBuilder
	NoPropSync       func() OptionBuilder
	SfHandle         func() OptionBuilder
	SfOem            func() OptionBuilder
	SfSequence       func() OptionBuilder
	SfSerialized     func() OptionBuilder
	SfVersion        func() OptionBuilder
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
	var optionPseudonyms []string

	if ob.keys != nil {
		optionKey = ob.keys.Doc
		optionEnv = ob.keys.Env
		optionPseudonyms = ob.keys.Pseudonyms
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
		Pseudonyms:    optionPseudonyms,
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
