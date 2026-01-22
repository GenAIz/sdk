// Package config defines facilities for managing configurations in multiple files, the command line and the environment
package config

import (
	"bufio"
	"errors"
	"fmt"
	"io"
	"maps"
	"os"
	"path/filepath"
	"strings"
	"syscall"
	"text/tabwriter"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

const (
	defaultConfigName = "Genaiz"
)

// Definer provides methods for defining a pflag.Flag in a pflag.FlagSet, and its associated default value in a config.Ledger
type Definer interface {
	// Default provides a Ledger with the default value of a Definer
	Default(*Ledger)

	// Defined provides a strategy for determining the value of a Definer with respect to the pflag.FlagSet providing its value
	Defined(*Ledger, *pflag.FlagSet)
}

// Registrar defines an abstraction of a config.Ledger for registering and initializing Option(s) and initializing and subcomponents needed to handle the Option(s)
type Registrar interface {
	// InitDefaults should trigger all DefaultSetters of all Option(s) held by the Registrar
	InitDefaults()

	// Init initializes all subcomponents of the Registrar
	Init()

	// InitLogging initializes logging with the set of Option(s) handled by the Registrar
	InitLogging()

	// Register registers cobra.Command(s) with a variable amount of Definer(s)
	Register(*cobra.Command, ...Definer)
}

// Ledger defines a Mediator which pilots a series of cobra.Command(s), mediating configuration retrieval, work directory and logging services
type Ledger struct {
	AuthFile      string                       // AuthFile, set to the current authentification file to query broker accounts
	ConfigName    string                       // ConfigName is used to read the config files related to the Ledger
	EnvPrefix     string                       // EnvPrefix is applied to environment variables read for this Ledger
	Logger        *logrus.Logger               // Logger, set to the current logger configurations
	LoggerFactory func(*Ledger) *logrus.Logger // LoggerFactory is called OnLogging when the Logger is initialized
	UserPath      string                       // UserPath is where to find the user's general configuration for genaiz toolkits
	TemplatePaths []string                     // TemplatePaths is a list of local filesystem paths to inspect to find recipes
	Timestamp     time.Time                    // timestamp is used to mark command execution is a single stamp per invocation
	WorkDir       string                       // WorkDir is by default the context dir, unless a change was recorded

	configurers       []func(*Ledger)        // configurers are a set of functions which modify the configuration paths to read before setting default values
	defaulters        []func(*Ledger)        // defaulters containers all default resolution functions registered
	input             io.Reader              // os.Stdin by default, swapped to other writers when testing
	loggers           []func(*logrus.Logger) // loggers is a list of delayed logging instructions for the Ledger to call OnLogging
	output            io.Writer              // os.Stdout by default, swapped to other writers when testing
	originalDir       string                 // originalDir is set to the dir the genaiz command was launched from
	validationHandler func(interface{})      // validationHandler is invoked when an option is not valid
	viper             *viper.Viper           // viper internal reference
	workspace         *StringOption          // workspace refers to an owning classification which may enter naming conventions by default
}

// backupConfigs moves any configuration files from the Ledger.UserPath to a .back name to avoid losing configurations on write operations
func (lr *Ledger) backupConfigs() error {
	var files, err = os.ReadDir(lr.UserPath)

	if err == nil {
		for _, f := range files {
			if strings.Contains(f.Name(), lr.ConfigName) &&
				!strings.Contains(f.Name(), ".back") {
				panicz.PanicIfError(os.Rename(
					filepath.Join(lr.UserPath, f.Name()),
					filepath.Join(lr.UserPath, f.Name()+".back")))
			}
		}
	}

	return err
}

// makeConfigs will write a new file under Ledger.UserPath with the Ledger.ConfigName in yaml format and write the provided config interface to the file
func (lr *Ledger) makeConfigs(configs interface{}) error {
	var file string

	panicz.RequiresNotNil("configs", configs)
	file = filepath.Join(lr.UserPath, lr.ConfigName+".yaml")

	if stat, _ := os.Stat(file); stat == nil {
		var data []byte
		var err error

		if data, err = yaml.Marshal(configs); err == nil {
			err = os.WriteFile(file, data, 0660)
		}

		return err
	}

	return errors.New("file exists")
}

// rollbackConfigs rolls back saved configuration files fom the Ledger.UserPath, removing the .back suffix
func (lr *Ledger) rollbackConfigs() error {
	var files, err = os.ReadDir(lr.UserPath)

	if err == nil {
		for _, f := range files {
			if strings.Contains(f.Name(), lr.ConfigName) &&
				strings.HasSuffix(f.Name(), ".back") {
				var i = strings.LastIndex(f.Name(), ".back")

				panicz.PanicIfError(os.Rename(
					filepath.Join(lr.UserPath, f.Name()),
					filepath.Join(lr.UserPath, f.Name()[:i])))
			}
		}
	}

	return err
}

// AddConfigOption adds a query with a StringOption for obtaining supplementary configuration paths on InitDefaults
//
// This is done before defaulters are executed, any Option.DefaultSetter or Option.DefaultValue on the option will be ignored in the order of execution.
//
// The Option.DefaultGetter will be invoked if there are no values queried from the previous configuration files in the chain and any bound parameter from cobra.
func (lr *Ledger) AddConfigOption(option *StringOption) {
	lr.configurers = append(lr.configurers, func(ledger *Ledger) {
		if path := ledger.GetString(option); path != "" {
			var fileUsed = lr.viper.ConfigFileUsed()

			if fileUsed != "" {
				lr.viper.SetConfigFile(filepath.Join(path, lr.ConfigName+"."+filez.GetFileType(fileUsed)))
			} else {
				lr.viper.AddConfigPath(path)
			}

			if err := lr.viper.MergeInConfig(); err != nil {
				lr.LogDebug("could not merge config [%s]", path)
				lr.LogDebug("%s", err.Error())
			}
		}
	})
}

// ChangeWorkDir changes the work directory under the program's execution, tracking the change for all Option(s) value resolution
func (lr *Ledger) ChangeWorkDir(option *StringOption) string {
	panicz.RequiresNotNil("option", option)

	if path := lr.GetString(option); path != "" {
		var cleanPath = filepath.Clean(path)
		var absPath, _ = filepath.Abs(cleanPath)

		if lr.WorkDir != absPath {
			lr.LogDebug("Changing working dir %s", absPath)
			panicz.PanicIfError(os.Chdir(absPath))
			lr.WorkDir = absPath
		}
	}

	return lr.WorkDir
}

// DisplayChangeDir displays the command that would be executed if the program's execution work directory was changed
func (lr *Ledger) DisplayChangeDir() {
	if lr.originalDir != lr.WorkDir {
		_, _ = fmt.Fprintf(lr.output, "cd %s\n", lr.WorkDir)
	}
}

// DisplayOptions displays the provided Option(s) param string and their associated values in the Ledger
func (lr *Ledger) DisplayOptions(options ...*Option) {
	lr.DisplayOptionsWithMap(nil, options...)
}

// DisplayOptionsWithMap will display the provided Option(s) with the provided kay value map
func (lr *Ledger) DisplayOptionsWithMap(keyValues *map[string]string, options ...*Option) {
	var writer = tabwriter.NewWriter(lr.output, 1, 1, 1, ' ', 0)
	var mapped = MapOptionsByParam(lr, options...)

	if keyValues != nil {
		maps.Copy(mapped, *keyValues)
	}

	mapz.Sorted(mapped, func(key string) {
		var _, err = fmt.Fprintf(writer, "%s:\t%s\n", key, mapped[key])

		panicz.PanicIfError(err)
	})
	panicz.PanicIfError(writer.Flush())
}

// FindPathConfig returns a viper.Viper configuration for a given path if the Ledgers' ConfigName can be resolved under the path
func (lr *Ledger) FindPathConfig(path string) (*viper.Viper, error) {
	var workingConfig string
	var err error

	if workingConfig, err = filez.FirstNamedFileUnder(path, lr.ConfigName); err == nil {
		var vp = viper.New()

		vp.SetConfigFile(filepath.Join(path, workingConfig))

		if err = vp.ReadInConfig(); err == nil {
			return vp, nil
		}
	}

	return nil, err
}

// FromWorkDir updates a pflag.Flag of pflag.FlagSet corresponding to a StringOption with a relative path to a value using the current Ledger.WorkDir
func (lr *Ledger) FromWorkDir(option *StringOption, flags *pflag.FlagSet) {
	if flag := flags.Lookup(option.Param); flag != nil {
		var flagValue = flag.Value.String()

		if filepath.IsLocal(flagValue) {
			panicz.PanicIfError(flag.Value.Set(filepath.Join(lr.WorkDir, flagValue)))
		} else if strings.HasPrefix(flagValue, ".") {
			var value, _ = filepath.Abs(flagValue)

			panicz.PanicIfError(flag.Value.Set(value))
		}
	}
}

// Get returns the value of the provided Option if any value can be resolved through cobra, viper and the Option.DefaultGetter
func (lr *Ledger) Get(option *Option) any {
	var result any

	if option.Key != "" {
		result = lr.viper.Get(option.Key)
	} else if option.Param != "" {
		result = lr.viper.Get(option.Param)
	}

	if option.IsEmpty(result) {
		for _, pseudo := range option.Pseudonyms {
			if value := lr.viper.Get(pseudo); !option.IsEmpty(value) {
				result = value
				break
			}
		}

		if option.IsEmpty(result) && option.DefaultGetter != nil {
			result = option.DefaultGetter(lr)
		}
	}

	if i := strings.Index(cast.ToString(result), "$"); i == 0 {
		result = os.Getenv(result.(string)[i+1:])
	}

	return result
}

// GetBool returns the value of a BoolOption from Get as a bool
func (lr *Ledger) GetBool(option *BoolOption) bool {
	var result = lr.Get(&option.Option)

	if result == nil {
		return false
	}

	return cast.ToBool(result)
}

// GetConfigType returns a shared.ConfigType if the provided option is set to a supported configuration type
func (lr *Ledger) GetConfigType(option *StringOption) (*shared.ConfigType, error) {
	var configTypeString = lr.GetString(option)

	if configTypeString == "none" {
		return lang.Ref(shared.ConfigTypeNone), nil
	}

	return shared.ConfigTypes.FromString(configTypeString)
}

// GetList returns the list value of a ListOption from Get as a slice of strings
func (lr *Ledger) GetList(option *ListOption) []string {
	var result []string

	if option.Key != "" {
		result = lr.viper.GetStringSlice(option.Key)
	} else if option.Param != "" {
		result = lr.viper.GetStringSlice(option.Param)
	}

	if len(result) == 0 && option.DefaultGetter != nil {
		result = cast.ToStringSlice(option.DefaultGetter(lr))
	}

	for i, r := range result {
		if option.Validator != nil && !option.Validator(r) {
			lr.validationHandler(fmt.Errorf("value [%d:%s] for option [%s] is invalid",
				i, r, strings.ToLower(option.Key)))
		}
	}

	return result
}

// GetPath returns the value of a StringOption relative to the Ledger's working directory if it's not absolute
func (lr *Ledger) GetPath(option *StringOption) string {
	var path = lr.StampString(option)

	if filepath.IsLocal(path) {
		path = filepath.Join(lr.WorkDir, path)
	}

	if option.Validator != nil && !option.Validator(path) {
		lr.validationHandler(fmt.Errorf("path [%s] for option [%s] is invalid",
			path, strings.ToLower(option.Key)))
	}

	return path
}

// GetString returns the value of a StringOption from Get as a string
func (lr *Ledger) GetString(option *StringOption) string {
	var result = cast.ToString(lr.Get(&option.Option))

	if option.Validator != nil && !option.Validator(result) {
		var value = result
		var length = len(value)

		if length > 32 {
			if i := strings.LastIndex(value, "/"); i > 0 {
				// Have a prefix ... for file paths
				value = "..." + value[length-29:]
			} else {
				value = value[0:29] + "..."
			}
		}

		lr.validationHandler(fmt.Errorf("value [%s] for option [%s] is invalid",
			value, strings.ToLower(option.Key)))
	}

	return result
}

// GetValue returns the value resolved by viper as bound with the provided key
func (lr *Ledger) GetValue(key string) string {
	return lr.viper.GetString(key)
}

// GetWorkspace returns the workspace path which should own the definition of the Ledger
func (lr *Ledger) GetWorkspace() string {
	if lr.workspace != nil {
		return lr.viper.GetString(lr.workspace.Param)
	}

	return ""
}

// Init initializes the Ledger primary service facades
func (lr *Ledger) Init() {
	lr.viper.AddConfigPath(lr.UserPath)
	lr.viper.SetConfigName(lr.ConfigName)
	lr.viper.AutomaticEnv()

	if err := lr.viper.ReadInConfig(); err == nil {
		lr.LogDebug("Using config file [%s]", lr.viper.ConfigFileUsed())
	} else {
		lr.LogDebug("Could not read user configurations under [%s]", lr.UserPath)
	}
}

// InitDefaults initializes default values of all Ledger options following each Option(s) defaults definition
func (lr *Ledger) InitDefaults() {
	for _, configurer := range lr.configurers {
		configurer(lr)
	}

	for _, defaulter := range lr.defaulters {
		defaulter(lr)
	}
}

// InitLogging initializes the logging subcomponent of the Ledger with the LoggerFactory of the Ledger, if any. Without a LoggerFactory, this method will instantiate a logrus.Logger without default values
func (lr *Ledger) InitLogging() {
	if lr.LoggerFactory == nil {
		lr.Logger = logrus.New()
	} else {
		lr.Logger = lr.LoggerFactory(lr)
	}

	for _, logFn := range lr.loggers {
		logFn(lr.Logger)
	}

	lr.loggers = nil
}

// InitValue initializes the string value of a configuration on the Ledger, if the option does not resolve to a value or resolved to its default value.
func (lr *Ledger) InitValue(option *StringOption, value string) {
	var currentValue = lr.viper.GetString(option.Key)

	if option.IsEmpty(currentValue) {
		lr.viper.Set(option.Key, value)
	}
}

// InitWorkspace initializes workspace resolution from the provided StringOption
func (lr *Ledger) InitWorkspace(option *StringOption) {
	lr.workspace = option
}

// LogDebug will log a Debug message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (lr *Ledger) LogDebug(message string, args ...interface{}) {
	if lr.Logger == nil {
		lr.loggers = append(lr.loggers, func(l *logrus.Logger) {
			l.Debugf(message, args...)
		})
	} else {
		lr.Logger.Debugf(message, args...)
	}
}

// LogError will log an Error message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (lr *Ledger) LogError(message string, args ...interface{}) {
	if lr.Logger == nil {
		lr.loggers = append(lr.loggers, func(l *logrus.Logger) {
			l.Errorf(message, args...)
		})
	} else {
		lr.Logger.Errorf(message, args...)
	}
}

// LogInfo will log an Info message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (lr *Ledger) LogInfo(message string, args ...interface{}) {
	if lr.Logger == nil {
		lr.loggers = append(lr.loggers, func(l *logrus.Logger) {
			l.Infof(message, args...)
		})
	} else {
		lr.Logger.Infof(message, args...)
	}
}

// OverrideBool overrides the value of a BoolOption
func (lr *Ledger) OverrideBool(option *BoolOption, value bool) {
	lr.viper.Set(option.Key, value)
}

// OverrideString overrides the value of a StringOption, if the value is not empty
func (lr *Ledger) OverrideString(option *StringOption, value string) {
	if !option.IsEmpty(value) {
		lr.viper.Set(option.Key, value)
	}
}

// QueryMandatory queries the user for input using the provided message and will keep on asking until the input is not the empty string
func (lr *Ledger) QueryMandatory(message string) string {
	var buff = bufio.NewReader(lr.input)
	var result string

	for {
		if _, err := fmt.Fprint(lr.output, message); err == nil {
			result, _ = buff.ReadString('\n')
		}

		if result != "" {
			return result
		}
	}
}

// QuerySecret queries the user for a secret input and will take whatever is passed, returning it as a byte array
func (lr *Ledger) QuerySecret(message string) *[]byte {
	var result []byte

	if _, err := fmt.Fprint(lr.output, message); err == nil {
		result, _ = term.ReadPassword(syscall.Stdin)
	}

	_, _ = fmt.Fprintln(lr.output)
	return &result
}

// Register configures Option Definer(s) with the provided cobra.Command and add their Definer.Default method as deferred initialization to be called on InitDefaults
func (lr *Ledger) Register(cmd *cobra.Command, definers ...Definer) {
	var flagSet = cmd.PersistentFlags()

	for _, def := range definers {
		def.Defined(lr, flagSet)
		lr.defaulters = append(lr.defaulters, def.Default)
	}
}

// StampString stamps a string if the '{timestamp:...}' placeholder can be found in it
func (lr *Ledger) StampString(option *StringOption) string {
	var raw = cast.ToString(lr.Get(&option.Option))

	return layout.StampString(raw, lr.Timestamp)
}

// ToWorkDir sets the value of the provided StringOption's pflag.Flag from a pflag.FlagSet to the current Ledger.WorkDir
func (lr *Ledger) ToWorkDir(option *StringOption, flags *pflag.FlagSet) {
	if flag := flags.Lookup(option.Param); flag != nil {
		panicz.PanicIfError(flag.Value.Set(lr.WorkDir))
	}
}

// Validate returns the validity of an Option according its associated viper value. It does not query command flags
func (lr *Ledger) Validate(option *Option) bool {
	if option.Validator != nil {
		var value = lr.viper.Get(option.Key)

		return option.Validator(value)
	}

	return true
}

// NewLedger returns a pointer to a new Ledger instance, initialized with Ledger.UserPath, Ledger.WorkDir and an instance of viper.Viper
func NewLedger() *Ledger {
	return NewBuilder().Build()
}

// Builder is an instance of the builder pattern for deferring construction of Ledger to runtime
type Builder struct {
	Viper         func() *viper.Viper
	Input         func() io.Reader
	Output        func() io.Writer
	TemplatePaths []string
	UserPath      string
}

// Build builds a Ledger with the recorded Viper, Input and Output factory methods
func (b *Builder) Build() *Ledger {
	var templatePaths = b.TemplatePaths

	home, err := b.resolveUserPath()
	cobra.CheckErr(err)
	cwd, err := os.Getwd()
	cobra.CheckErr(err)
	templatePaths = append(templatePaths, filepath.Join(home, ".local", "genaiz", "recipes"))

	return &Ledger{
		AuthFile:      filepath.Join(home, ".cache", "genaiz", ".auth"),
		ConfigName:    defaultConfigName,
		UserPath:      filepath.Join(home, ".config", "genaiz"),
		TemplatePaths: templatePaths,
		Timestamp:     time.Now(),
		WorkDir:       cwd,

		input:             b.Input(),
		output:            b.Output(),
		originalDir:       cwd,
		validationHandler: cobra.CheckErr,
		viper:             b.Viper(),
	}
}

// WithInput replaces the factory method providing input to the Ledger to build
func (b *Builder) WithInput(i io.Reader) *Builder {
	b.Input = func() io.Reader {
		return i
	}
	return b
}

// WithOutput replaces the factory method providing output to the Ledger to build
func (b *Builder) WithOutput(o io.Writer) *Builder {
	b.Output = func() io.Writer {
		return o
	}
	return b
}

// WithTemplates will build the Ledger adding the specified paths to Ledger.TemplatePaths
func (b *Builder) WithTemplates(paths ...string) *Builder {
	b.TemplatePaths = paths
	return b
}

// WithUserPath overrides the default user path used by the ledger for the current user
func (b *Builder) WithUserPath(path string) *Builder {
	b.UserPath = path
	return b
}

// WithViper replaces the way Viper is constructed when building a new Ledger
func (b *Builder) WithViper(v *viper.Viper) *Builder {
	b.Viper = func() *viper.Viper {
		return v
	}
	return b
}

func (b *Builder) resolveUserPath() (string, error) {
	if b.UserPath == "" {
		return os.UserHomeDir()
	}

	return b.UserPath, nil
}

// NewBuilder returns a builder with the default viper.Viper static instance, STDIN and STDOUT
func NewBuilder() *Builder {
	return &Builder{
		Viper: viper.GetViper,
		Input: func() io.Reader {
			return os.Stdin
		},
		Output: func() io.Writer {
			return os.Stdout
		},
	}
}
