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

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
	"golang.org/x/term"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz/lang/mapz"
	"genaiz.com/genaiz/lang/panicz"
)

const (
	defaultConfigName = "Genaiz"
)

// Definer provides methods for defining a pflag.Flag in a pflag.FlagSet, and its associated default value in a config.Repo
type Definer interface {
	// Default provides a Repo with the default value of a Definer
	Default(repo *Repo)

	// Defined provides a strategy for determining the value of a Definer with respect to the pflag.FlagSet providing its value
	Defined(repo *Repo, set *pflag.FlagSet)
}

// Registrar defines an abstraction of a config.Repo for registering and initializing Option(s) and initializing and subcomponents needed to handle the Option(s)
type Registrar interface {
	// InitDefaults should trigger all DefaultSetters of all Option(s) held by the Registrar
	InitDefaults()

	// Init initializes all subcomponents of the Registrar
	Init()

	// InitLogging initializes logging with the set of Option(s) handled by the Registrar
	InitLogging()

	// Register registers cobra.Command(s) with a variable amount of Definer(s)
	Register(cmd *cobra.Command, definers ...Definer)
}

// Repo defines a Mediator which pilots a series of cobra.Command(s), mediating configuration retrieval, work directory and logging services
type Repo struct {
	AuthFile      string                     // AuthFile, set to the current authentification file to query broker accounts
	ConfigName    string                     // ConfigName is used to read the config files related to the repo
	EnvPrefix     string                     // EnvPrefix is applied to environment variables read for this repo
	Logger        *logrus.Logger             // Logger, set to the current logger configurations
	LoggerFactory func(*Repo) *logrus.Logger // LoggerFactory is called OnLogging when the Logger is initialized
	UserPath      string                     // UserPath is where to find the user's general configuration for genaiz toolkits
	TemplatePaths []string                   // TemplatePaths is a list of local filesystem paths to inspect to find recipes
	WorkDir       string                     // WorkDir is by default the context dir, unless a change was recorded

	configurers       []func(*Repo)                 // configurers are a set of functions which modify the configuration paths to read before setting default values
	defaulters        []func(*Repo)                 // defaulters containers all default resolution functions registered
	input             io.Reader                     // os.Stdin by default, swapped to other writers when testing
	loggers           []func(logger *logrus.Logger) // loggers is a list of delayed logging instructions for the repo to call OnLogging
	output            io.Writer                     // os.Stdout by default, swapped to other writers when testing
	originalDir       string                        // originalDir is set to the dir the genaiz command was launched from
	validationHandler func(interface{})             // validationHandler is invoked when an option is not valid
	viper             *viper.Viper                  // viper internal reference
	workspace         *StringOption                 // workspace refers to an owning classification which may enter naming conventions by default
}

// backupConfigs moves any configuration files from the Repo.UserPath to a .back name to avoid losing configurations on write operations
func (r *Repo) backupConfigs() error {
	var files, err = os.ReadDir(r.UserPath)

	if err == nil {
		for _, f := range files {
			if strings.Contains(f.Name(), r.ConfigName) &&
				!strings.Contains(f.Name(), ".back") {
				panicz.PanicIfError(os.Rename(
					filepath.Join(r.UserPath, f.Name()),
					filepath.Join(r.UserPath, f.Name()+".back")))
			}
		}
	}

	return err
}

// makeConfigs will write a new file under Repo.UserPath with the Repo.ConfigName in yaml format and write the provided config interface to the file
func (r *Repo) makeConfigs(configs interface{}) error {
	var file string

	panicz.RequiresNotNil("configs", configs)
	file = filepath.Join(r.UserPath, r.ConfigName+".yaml")

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

// rollbackConfigs rolls back saved configuration files fom the Repo.UserPath, removing the .back suffix
func (r *Repo) rollbackConfigs() error {
	var files, err = os.ReadDir(r.UserPath)

	if err == nil {
		for _, f := range files {
			if strings.Contains(f.Name(), r.ConfigName) &&
				strings.HasSuffix(f.Name(), ".back") {
				var i = strings.LastIndex(f.Name(), ".back")

				panicz.PanicIfError(os.Rename(
					filepath.Join(r.UserPath, f.Name()),
					filepath.Join(r.UserPath, f.Name()[:i])))
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
func (r *Repo) AddConfigOption(option *StringOption) {
	r.configurers = append(r.configurers, func(repo *Repo) {
		if path := repo.GetString(option); path != "" {
			r.viper.AddConfigPath(path)

			if err := r.viper.MergeInConfig(); err != nil {
				r.LogDebug("could not merge config [%s]", path)
			}
		}
	})
}

// ChangeWorkDir changes the work directory under the program's execution, tracking the change for all Option(s) value resolution
func (r *Repo) ChangeWorkDir(option *StringOption) string {
	panicz.RequiresNotNil("option", option)

	if path := r.GetString(option); path != "" {
		var cleanPath = filepath.Clean(path)
		var absPath, _ = filepath.Abs(cleanPath)

		if r.WorkDir != absPath {
			r.LogDebug("Changing working dir %s", absPath)
			panicz.PanicIfError(os.Chdir(absPath))
			r.WorkDir = absPath
		}
	}

	return r.WorkDir
}

// DisplayChangeDir displays the command that would be executed if the program's execution work directory was changed
func (r *Repo) DisplayChangeDir() {
	if r.originalDir != r.WorkDir {
		_, _ = fmt.Fprintf(r.output, "cd %s\n", r.WorkDir)
	}
}

// DisplayOptions displays the provided Option(s) param string and their associated values in the Repo
func (r *Repo) DisplayOptions(options ...*Option) {
	r.DisplayOptionsWithMap(nil, options...)
}

// DisplayOptionsWithMap will display the provided Option(s) with the provided kay value map
func (r *Repo) DisplayOptionsWithMap(keyValues *map[string]string, options ...*Option) {
	var writer = tabwriter.NewWriter(r.output, 1, 1, 1, ' ', 0)
	var mapped = MapOptionsByParam(r, options...)

	if keyValues != nil {
		maps.Copy(mapped, *keyValues)
	}

	mapz.Sorted(mapped, func(key string) {
		var _, err = fmt.Fprintf(writer, "%s:\t%s\n", key, mapped[key])

		panicz.PanicIfError(err)
	})
	panicz.PanicIfError(writer.Flush())
}

// FromWorkDir updates a pflag.Flag of pflag.FlagSet corresponding to a StringOption with a relative path to a value using the current Repo.WorkDir
func (r *Repo) FromWorkDir(option *StringOption, flags *pflag.FlagSet) {
	if flag := flags.Lookup(option.Param); flag != nil {
		var flagValue = flag.Value.String()

		if filepath.IsLocal(flagValue) {
			panicz.PanicIfError(flag.Value.Set(filepath.Join(r.WorkDir, flagValue)))
		} else if strings.HasPrefix(flagValue, ".") {
			var value, _ = filepath.Abs(flagValue)

			panicz.PanicIfError(flag.Value.Set(value))
		}
	}
}

// Get returns the value of the provided Option if any value can be resolved through cobra, viper and the Option.DefaultGetter
func (r *Repo) Get(option *Option) any {
	var result any

	if option.Key != "" {
		result = r.viper.Get(option.Key)
	} else if option.Param != "" {
		result = r.viper.Get(option.Param)
	}

	if (result == nil || result == "" || result == option.DefaultValue) && option.DefaultGetter != nil {
		result = option.DefaultGetter(r)
	}

	if i := strings.Index(cast.ToString(result), "$"); i == 0 {
		result = os.Getenv(result.(string)[i+1:])
	}

	return result
}

// GetBool returns the value of a BoolOption from Get as a bool
func (r *Repo) GetBool(option *BoolOption) bool {
	var result = r.Get(&option.Option)

	if result == nil {
		return false
	}

	return cast.ToBool(result)
}

// GetList returns the list value of a ListOption from Get as a slice of strings
func (r *Repo) GetList(option *ListOption) []string {
	var value = r.Get(&option.Option)
	var result []string
	var err error

	if value == "" || value == nil {
		return []string{}
	}

	result, err = cast.ToStringSliceE(value)
	panicz.PanicIfError(err)
	return result
}

// GetString returns the value of a StringOption from Get as a string
func (r *Repo) GetString(option *StringOption) string {
	var result = cast.ToString(r.Get(&option.Option))

	if option.Validator != nil && !option.Validator(result) {
		r.validationHandler(fmt.Errorf("option %s is invalid", option.Key))
	}

	return result
}

// GetValue returns the value resolved by viper as bound with the provided key
func (r *Repo) GetValue(key string) string {
	return r.viper.GetString(key)
}

// GetWorkspace returns the workspace path which should own the definition of the Repo
func (r *Repo) GetWorkspace() string {
	if r.workspace != nil {
		return r.viper.GetString(r.workspace.Param)
	}

	return ""
}

// Init initializes the repository primary service facades
func (r *Repo) Init() {
	r.viper.AddConfigPath(r.UserPath)
	r.viper.SetConfigName(r.ConfigName)
	r.viper.AutomaticEnv()

	if err := r.viper.ReadInConfig(); err == nil {
		r.LogDebug("Using config file [%s]", r.viper.ConfigFileUsed())
	} else {
		r.LogDebug("Could not read user configurations under [%s]", r.UserPath)
	}
}

// InitDefaults initializes default values of all repository options following each Option(s) defaults definition
func (r *Repo) InitDefaults() {
	for _, configurer := range r.configurers {
		configurer(r)
	}

	for _, defaulter := range r.defaulters {
		defaulter(r)
	}
}

// InitLogging initializes the logging subcomponent of the repository with the LoggerFactory of the Repo, if any. Without a LoggerFactory, this method will instantiate a logrus.Logger without default values
func (r *Repo) InitLogging() {
	if r.LoggerFactory == nil {
		r.Logger = logrus.New()
	} else {
		r.Logger = r.LoggerFactory(r)
	}

	for _, logFn := range r.loggers {
		logFn(r.Logger)
	}

	r.loggers = nil
}

// InitValue initializes the string value of a configuration on the Repo, if the option does not resolve to a value already.
func (r *Repo) InitValue(option *StringOption, value string) {
	if r.viper.GetString(option.Key) == "" {
		r.viper.Set(option.Key, value)
	}
}

// InitWorkspace initializes workspace resolution from the provided StringOption
func (r *Repo) InitWorkspace(option *StringOption) {
	r.workspace = option
}

// LogDebug will log a Debug message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (r *Repo) LogDebug(message string, args ...interface{}) {
	if r.Logger == nil {
		r.loggers = append(r.loggers, func(l *logrus.Logger) {
			l.Debugf(message, args...)
		})
	} else {
		r.Logger.Debugf(message, args...)
	}
}

// LogError will log an Error message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (r *Repo) LogError(message string, args ...interface{}) {
	if r.Logger == nil {
		r.loggers = append(r.loggers, func(l *logrus.Logger) {
			l.Errorf(message, args...)
		})
	} else {
		r.Logger.Errorf(message, args...)
	}
}

// LogInfo will log an Info message with the logger if it has been initialized, otherwise will defer it to logging initialization
func (r *Repo) LogInfo(message string, args ...interface{}) {
	if r.Logger == nil {
		r.loggers = append(r.loggers, func(l *logrus.Logger) {
			l.Infof(message, args...)
		})
	} else {
		r.Logger.Infof(message, args...)
	}
}

// QueryMandatory queries the user for input using the provided message and will keep on asking until the input is not the empty string
func (r *Repo) QueryMandatory(message string) string {
	var buff = bufio.NewReader(r.input)
	var result string

	for {
		if _, err := fmt.Fprint(r.output, message); err == nil {
			result, _ = buff.ReadString('\n')
		}

		if result != "" {
			return result
		}
	}
}

// QuerySecret queries the user for a secret input and will take whatever is passed, returning it as a byte array
func (r *Repo) QuerySecret(message string) *[]byte {
	var result []byte

	if _, err := fmt.Fprint(r.output, message); err == nil {
		result, _ = term.ReadPassword(syscall.Stdin)
	}

	_, _ = fmt.Fprintln(r.output)
	return &result
}

// Register configures Option Definer(s) with the provided cobra.Command and add their Definer.Default method as deferred initialization to be called on InitDefaults
func (r *Repo) Register(cmd *cobra.Command, definers ...Definer) {
	var flagSet = cmd.PersistentFlags()

	for _, def := range definers {
		def.Defined(r, flagSet)
		r.defaulters = append(r.defaulters, def.Default)
	}
}

// ToWorkDir sets the value of the provided StringOption's pflag.Flag from a pflag.FlagSet to the current Repo.WorkDir
func (r *Repo) ToWorkDir(option *StringOption, flags *pflag.FlagSet) {
	if flag := flags.Lookup(option.Param); flag != nil {
		panicz.PanicIfError(flag.Value.Set(r.WorkDir))
	}
}

// Validate returns the validity of an Option according its associated viper value. It does not query command flags
func (r *Repo) Validate(option *Option) bool {
	if option.Validator != nil {
		var value = r.viper.Get(option.Key)

		return option.Validator(value)
	}

	return true
}

// NewRepo returns a pointer to a new Repo instance, initialized with Repo.UserPath, Repo.WorkDir and an instance of viper.Viper
func NewRepo() *Repo {
	return NewBuilder().Build()
}

// Builder is an instance of the builder pattern for deferring construction of Repo to runtime
type Builder struct {
	Viper         func() *viper.Viper
	Input         func() io.Reader
	Output        func() io.Writer
	TemplatePaths []string
}

// Build builds a Repo with the recorded Viper, Input and Output factory methods
func (b *Builder) Build() *Repo {
	var templatePaths = b.TemplatePaths

	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	cwd, err := os.Getwd()
	cobra.CheckErr(err)
	templatePaths = append(templatePaths, filepath.Join(home, "/.local/genaiz/recipes"))

	return &Repo{
		AuthFile:      filepath.Join(home, "/.cache/genaiz/.auth"),
		ConfigName:    defaultConfigName,
		UserPath:      filepath.Join(home, "/.config/genaiz"),
		TemplatePaths: templatePaths,
		WorkDir:       cwd,

		input:             b.Input(),
		output:            b.Output(),
		originalDir:       cwd,
		validationHandler: cobra.CheckErr,
		viper:             b.Viper(),
	}
}

// WithInput replaces the factory method providing input to the Repo to build
func (b *Builder) WithInput(i io.Reader) *Builder {
	b.Input = func() io.Reader {
		return i
	}
	return b
}

// WithOutput replaces the factory method providing output to the Repo to build
func (b *Builder) WithOutput(o io.Writer) *Builder {
	b.Output = func() io.Writer {
		return o
	}
	return b
}

// WithTemplates will build the repository adding the specified paths to Repo.TemplatePaths
func (b *Builder) WithTemplates(paths ...string) *Builder {
	b.TemplatePaths = paths
	return b
}

// WithViper replaces the way Viper is constructed when building a new Repo
func (b *Builder) WithViper(v *viper.Viper) *Builder {
	b.Viper = func() *viper.Viper {
		return v
	}
	return b
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
