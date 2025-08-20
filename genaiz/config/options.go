package config

import (
	"fmt"
	"strings"

	"github.com/spf13/pflag"
	"github.com/spf13/viper"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/stringz"
)

// Option is a struct describing all the data and facilities needed to manage a configuration value in a config.Ledger
type Option struct {
	Key           string            // Key is used in configuration files
	Param         string            // Param is used from the shell
	Short         string            // Short is the shortened one letter alias for param
	Env           string            // Env is the environment string used to retrieve the config
	Usage         string            // Usage is the help string display for the parameter on the shell
	DefaultGetter func(*Ledger) any // DefaultGetter should be called when GetString, GetInt and GetBool are called, it has priority over DefaultValue
	DefaultSetter func(*Ledger) any // DefaultSetter should be called on cobra.OnInitialize with a reference to the Ledger
	DefaultValue  any               // DefaultValue is configured on viper.SetDefault when Registrar.Register is called
	Validator     Validates         // Validator should be called manually and liberally, not validated does not imply failed command
}

// Equals test whether the option matches the scalar data of the provided other Option
func (o *Option) Equals(other *Option) bool {
	return o.Key == other.Key &&
		o.Param == other.Param &&
		o.Short == other.Short &&
		o.Env == other.Env &&
		o.Usage == other.Usage &&
		o.DefaultValue == other.DefaultValue
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

// Default provides the default value of the Option when a Ledger.Get is processed or when Ledger.InitDefaults is invoked
func (o *Option) Default(ledger *Ledger) {
	var value any

	panicz.RequiresNotNil("ledger", ledger)

	if o.DefaultSetter == nil {
		value = o.DefaultValue
	} else {
		value = o.DefaultSetter(ledger)
	}

	if value != nil {
		if o.Key != "" {
			ledger.viper.SetDefault(o.Key, value)
		} else if o.Param != "" {
			ledger.viper.SetDefault(o.Param, value)
		}
	}
}

// Defined provides a strategy for defining how to retrieve the Option from a pflag.FlagSet, viper, its DefaultSetter, DefaultGetter and DefaultValue
func (o *Option) Defined(ledger *Ledger, set *pflag.FlagSet) {
	panicz.RequiresNotNil("ledger", ledger)

	if flag := set.Lookup(o.Param); flag != nil {
		o.bindKeyFlag(ledger.viper, flag)
		o.bindDefValue(flag)
	}

	o.bindKeyEnv(ledger.viper)
}

// BoolOption treats the value of an option as a boolean when defining its pflag.Flag and processing its Option.DefaultValue
type BoolOption struct {
	Option
}

// Defined of a BoolOption defined its pflag.Flag under a pflag.FlagSet as a Bool or BoolP with the Option.Param and Option.Key
func (bo *BoolOption) Defined(ledger *Ledger, flags *pflag.FlagSet) {
	panicz.RequiresNotNil("ledger", ledger)
	panicz.RequiresNotNil("flags", flags)

	if bo.Param != "" {
		flags.BoolP(bo.Param, bo.Short, false, bo.Usage)
	}

	bo.Option.Defined(ledger, flags)
}

// ListOption treats the value of an option as a list of strings when defining its pflag.Flag and processing its Option.DefaultValue
type ListOption struct {
	Option
}

// Defined of a ListOption defined its pflag.Flag under pflag.FlagSet as a ListValue with the Option.Param and Option.Key
func (lo *ListOption) Defined(ledger *Ledger, flags *pflag.FlagSet) {
	panicz.RequiresNotNil("ledger", ledger)
	panicz.RequiresNotNil("flags", flags)

	if lo.Param != "" {
		flags.VarP(newListValue(), lo.Param, lo.Short, "")
	}

	lo.Option.Defined(ledger, flags)
}

// ListValue is a value adapter for converting multiple pflag.Flag instances on a command line into a list of configurations
type ListValue struct {
	values *[]string // values corresponds to the set of string entered on the command line
}

// GetSlice returns a pointer to the values slice untouched
func (lv *ListValue) GetSlice() []string {
	return *lv.values
}

// Set adds the provided value to the values of this ListValue
func (lv *ListValue) Set(value string) error {
	if value != "" {
		*lv.values = append(*lv.values, value)
	}

	return nil
}

// String converts the slice of values into a readable string
func (lv *ListValue) String() string {
	if len(*lv.values) == 0 {
		return ""
	}

	return strings.Join(*lv.values, " ")
}

// Type returns the type string for this Value
func (lv *ListValue) Type() string {
	return "list"
}

// StringOption treats the value of an option as a string when defining its pflag.Flag and processing its Option.DefaultValue
type StringOption struct {
	Option
}

// Defined of a StringOption defines its pflag.Flag under a pflag.FlagSet as a String or StringP with the Option.Param and Option.Key
func (so *StringOption) Defined(ledger *Ledger, flags *pflag.FlagSet) {
	panicz.RequiresNotNil("ledger", ledger)
	panicz.RequiresNotNil("flags", flags)

	if so.Param != "" {
		flags.StringP(so.Param, so.Short, "", so.Usage)
	}

	so.Option.Defined(ledger, flags)
}

// MapOptionsByEnvKey returns a map of Option.Env keys to their resolved Ledger value
func MapOptionsByEnvKey(ledger *Ledger, options ...*Option) map[string]string {
	return mapOptionsByKeyFunc(ledger, func(opt *Option) string {
		return opt.GetEnvKey()
	}, options...)
}

// MapOptionsByParam returns a map of Option.Param keys to their resolved Ledger value
func MapOptionsByParam(ledger *Ledger, options ...*Option) map[string]string {
	return mapOptionsByKeyFunc(ledger, func(opt *Option) string {
		return opt.Param
	}, options...)
}

// mapOptionsByKeyFunc returns a map of Option keys, according to the provided toKey function, to their resolved Ledger value
func mapOptionsByKeyFunc(ledger *Ledger, toKey func(*Option) string, options ...*Option) map[string]string {
	var result = make(map[string]string, len(options))

	for _, opt := range options {
		var key = toKey(opt)

		if key != "" {
			var value = ledger.Get(opt)

			if value == nil {
				result[key] = ""
			} else {
				result[key] = fmt.Sprintf("%v", value)
			}
		}
	}

	return result
}

// newListValues builds a new ListValue reference with values set to an empty slice
func newListValue() *ListValue {
	return &ListValue{
		values: &[]string{},
	}
}
