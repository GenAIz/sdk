// Package config defines facilities for managing configurations in multiple files, the command line and the environment
package config

import (
	"bufio"
	"fmt"
	"io"
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

	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/lang/mapz"
	"genaiz.com/genaiz/lang/panicz"
	"genaiz.com/genaiz/lang/stringz"
)

// Definer provides methods for defining a pflag.Flag in a pflag.FlagSet, and its associated default value in a config.Repo
type Definer interface {
	// Default provides a Repo with the default value of a Definer
	Default(repo *Repo)

	// Defined provides a strategy for determining the value of a Definer with respect to the pflag.FlagSet providing its value
	Defined(repo *Repo, set *pflag.FlagSet)
}

// Option is a struct describing all the data and facilities needed to manage a configuration value in a config.Repo
type Option struct {
	Key           string               // Key is used in configuration files
	Alias         string               // Alias is also used in configuration files
	Param         string               // Param is used from the shell
	Short         string               // Short is the shortened one letter alias for param
	Env           string               // Env is the environment string used to retrieve the config
	Usage         string               // Usage is the help string display for the parameter on the shell
	DefaultGetter func(repo *Repo) any // DefaultGetter should be called when GetString, GetInt and GetBool are called, it has priority over DefaultValue
	DefaultSetter func(repo *Repo) any // DefaultSetter should be called on cobra.OnInitialize with a reference to the repo
	DefaultValue  any                  // DefaultValue is configured on viper.SetDefault when Registrar.Register is called
	Validator     func(value any) bool // Validator should be called manually and liberally, not validated does not imply failed command
}

// bindDefValue binds a DefValue on a pflag.Flag if the DefaultValue of the Option and the flag are defined
func (o *Option) bindDefValue(flag *pflag.Flag) {
	panicz.RequiresNotNil("flag", flag)

	if o.DefaultValue != nil {
		flag.DefValue = fmt.Sprintf("%v", o.DefaultValue)
	}
}

// bindKeyEnv binds an environment variable lookup under viper if the Key of the Option is defined
func (o *Option) bindKeyEnv(viper *viper.Viper) {
	panicz.RequiresNotNil("viper", viper)

	if o.Key != "" {
		var envKey = o.GetEnvKey()

		if envKey != "" {
			panicz.PanicIfError(viper.BindEnv(o.Key, envKey))
		}
	}
}

// bindKeyFlag binds the Key or the Param of the Option under viper with the pflag.Flag provided. The Key always has priority over the Param
func (o *Option) bindKeyFlag(viper *viper.Viper, flag *pflag.Flag) {
	panicz.RequiresNotNil("flag", flag)
	panicz.PanicIfError(viper.BindPFlag(
		stringz.FirstNonEmpty(o.Key, o.Param),
		flag,
	))
}

// GetEnvKey returns the Env value of the Option if it's defined, otherwise it will replace all '.' with '_' on the Key of the Option, capitalized. If no Key is defined, it returns the empty string
func (o *Option) GetEnvKey() string {
	if o.Env != "" {
		return o.Env
	}

	if o.Key != "" {
		return strings.ReplaceAll(
			strings.ToUpper(o.Key), ".", "_")
	}

	return ""
}

// Default provides the default value of the Option when a Repo.Get is processed or when Repo.InitDefaults is invoked
func (o *Option) Default(repo *Repo) {
	var value any

	panicz.RequiresNotNil("repo", repo)

	if o.DefaultSetter == nil {
		value = o.DefaultValue
	} else {
		value = o.DefaultSetter(repo)
	}

	if value != nil {
		if o.Key != "" {
			repo.viper.SetDefault(o.Key, value)
		} else if o.Param != "" {
			repo.viper.SetDefault(o.Param, value)
		}
	}
}

// Defined provides a strategy for defining how to retrieve the Option from a pflag.FlagSet, viper, its DefaultSetter, DefaultGetter and DefaultValue
func (o *Option) Defined(repo *Repo, set *pflag.FlagSet) {
	panicz.RequiresNotNil("repo", repo)

	if flag := set.Lookup(o.Param); flag != nil {
		o.bindKeyFlag(repo.viper, flag)
		o.bindDefValue(flag)
	}

	o.bindKeyEnv(repo.viper)
}

// BoolOption treats the value of an option as a boolean when defining its pflag.Flag and processing its DefaultValue
type BoolOption struct {
	Option
}

// Defined of a BoolOption defines its pflag.Flag under a pflag.FlagSet as a Bool or BoolP with a short value
func (bo *BoolOption) Defined(repo *Repo, flags *pflag.FlagSet) {
	panicz.RequiresNotNil("repo", repo)
	panicz.RequiresNotNil("flags", flags)

	if bo.Param != "" {
		flags.BoolP(bo.Param, bo.Short, false, bo.Usage)
	}

	bo.Option.Defined(repo, flags)
}

// StringOption treats the value of an option as a string when defining its pflag.Flag and processing its DefaultValue
type StringOption struct {
	Option
}

// Defined of a StringOption defines its pflag.Flag under a pflag.FlagSet as a String or StringP with a short value
func (so *StringOption) Defined(repo *Repo, set *pflag.FlagSet) {
	panicz.RequiresNotNil("repo", repo)
	panicz.RequiresNotNil("set", set)

	if so.Param != "" {
		set.StringP(so.Param, so.Short, "", so.Usage)
	}

	so.Option.Defined(repo, set)
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
	Logger        *logrus.Logger             // Logger, set to the current logger configurations
	LoggerFactory func(*Repo) *logrus.Logger // LoggerFactory is called OnLogging when the Logger is initialized
	UserPath      string                     // UserPath is where to find the user's general configuration for genaiz toolkits
	WorkDir       string                     // WorkDir is by default the context dir, unless a change was recorded

	defaulters        []func(*Repo)                 // defaulters containers all default resolution functions registered
	input             io.Reader                     // os.Stdin by default, swapped to other writers when testing
	loggers           []func(logger *logrus.Logger) // loggers is a list of delayed logging instructions for the repo to call OnLogging
	output            io.Writer                     // os.Stdout by default, swapped to other writers when testing
	originalDir       string                        // originalDir is set to the dir the genaiz command was launched from
	validationHandler func(interface{})             // validationHandler is invoked when an option is not valid
	viper             *viper.Viper                  // viper internal reference
	workspace         *StringOption                 // workspace refers to an owning classification which may enter naming conventions by default
}

// ChangeWorkDir changes the work directory under the program's execution, tracking the change for all Option(s) value resolution
func (r *Repo) ChangeWorkDir(option *StringOption) string {
	panicz.RequiresNotNil("option", option)

	if path := r.GetString(option); path != "" {
		var cleanPath = filepath.Clean(path)
		var absPath = filez.AbsOrPanic(cleanPath)

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
	if len(options) > 0 {
		var writer = tabwriter.NewWriter(r.output, 1, 1, 1, ' ', 0)
		var mapped = mapOptionsByParam(r, options...)

		mapz.Sorted(mapped, func(key string) {
			var _, err = fmt.Fprintf(writer, "%s:\t%s\n", key, mapped[key])

			panicz.PanicIfError(err)
		})
		panicz.PanicIfError(writer.Flush())
	}
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

// FromWorkDir updates a pflag.Flag of pflag.FlagSet corresponding to a StringOption with a relative path to a value using the current Repo.WorkDir
func (r *Repo) FromWorkDir(option *StringOption, flags *pflag.FlagSet) {
	if flag := flags.Lookup(option.Param); flag != nil {
		var flagValue = flag.Value.String()

		if filepath.IsLocal(flagValue) {
			panicz.PanicIfError(flag.Value.Set(filepath.Join(r.WorkDir, flagValue)))
		} else if strings.HasPrefix(flagValue, ".") {
			var value = filez.AbsOrPanic(flagValue)

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
	r.viper.SetConfigName(".genaiz")
	r.viper.SetEnvPrefix("genaiz")
	r.viper.AutomaticEnv()

	if err := r.viper.ReadInConfig(); err == nil {
		r.LogDebug("Using config file [%s]", r.viper.ConfigFileUsed())
	} else {
		r.LogDebug("Cound not read user configurations under %s", r.UserPath)
	}
}

// InitDefaults initializes default values of all repository options following each Option(s) defaults definition
func (r *Repo) InitDefaults() {
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

// InitWorkspace initializes workspace resolution from the provided StringOption
func (r *Repo) InitWorkspace(option *StringOption) {
	r.workspace = option
}

// QueryMandatory queries the user for input using the provided message and will keep on asking until the input is not the empty string
func (r *Repo) QueryMandatory(message string) string {
	var buff = bufio.NewReader(r.input)
	var result string

	for {
		if _, err := fmt.Fprint(r.output, message); err == nil {
			result, _ = buff.ReadString('\n')
		}

		if result == "" {
			continue
		}

		return result
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
		if flag.Value != nil {
			panicz.PanicIfError(flag.Value.Set(r.WorkDir))
		}
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
	return NewRepoWithViper(viper.GetViper())
}

// NewRepoWithViper returns a pointer to a new Repo instance, initialized with Repo.UserPath, Repo.WorkDir and the provided instance of viper.Viper
func NewRepoWithViper(v *viper.Viper) *Repo {
	return NewRepoWith(v, os.Stdin, os.Stdout)
}

// NewRepoWith returns a pointer to a new Repo instance, using the provided instance of viper.Viper and the input io.Reader and output io.Writer specified for stdin and stdout
func NewRepoWith(v *viper.Viper, i io.Reader, o io.Writer) *Repo {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	cwd, err := os.Getwd()
	cobra.CheckErr(err)

	return &Repo{
		AuthFile: filepath.Join(home, "/.config/genaiz/.genaiz.auth"),
		UserPath: filepath.Join(home, "/.config/genaiz"),
		WorkDir:  cwd,

		input:             i,
		output:            o,
		originalDir:       cwd,
		validationHandler: cobra.CheckErr,
		viper:             v,
	}
}

// ValidateDir validates that a path exists and is a directory
func ValidateDir(path any) bool {
	var s, _ = os.Stat(path.(string))

	return s != nil && s.IsDir()
}

// ValidateDirCreation creates a path if it does not exist and returns true
func ValidateDirCreation(path any) bool {
	var stringPath = path.(string)

	if s, _ := os.Stat(stringPath); s == nil {
		var err = os.MkdirAll(stringPath, 0750)

		return err == nil
	}

	return true
}

// ValidateFile validates that a path exists and is not a directory
func ValidateFile(path any) bool {
	var s, _ = os.Stat(path.(string))

	return s != nil && !s.IsDir()
}

// ValidateOptional only applies the validator function if the value is not nil or the empty string
func ValidateOptional(value any, validator func(any) bool) bool {
	if value != nil && cast.ToString(value) != "" {
		return validator(value)
	}

	return true
}

// mapOptionsByParam returns a map of Option.Param keys to their resolved Repo value
func mapOptionsByParam(repo *Repo, options ...*Option) map[string]string {
	var result = make(map[string]string, len(options))

	for _, opt := range options {
		if opt.Param != "" {
			var value = repo.Get(opt)

			if value == nil {
				result[opt.Param] = ""
			} else {
				result[opt.Param] = fmt.Sprintf("%v", value)
			}
		}
	}

	return result
}
