package config

import (
	"fmt"
	"maps"
	"os"
	"sort"
	"strings"
	"text/tabwriter"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"github.com/spf13/viper"
)

var (
	DefaultPath string
	Logger      *logrus.Logger
)

func BindCmd(cmd *cobra.Command, keys map[string]string) {
	for key, value := range keys {
		var envKey = strings.ToUpper(key)

		cobra.CheckErr(viper.BindPFlag(key, cmd.PersistentFlags().Lookup(value)))
		envKey = strings.ReplaceAll(envKey, ".", "_")
		cobra.CheckErr(viper.BindEnv(key, envKey))
	}
}

func BindDefaults(defaults map[string]func() string) {
	for key, resolver := range defaults {
		viper.SetDefault(key, resolver())
	}
}

func Default(customPath string) error {
	if customPath == "" {
		viper.AddConfigPath(DefaultPath)
		viper.SetConfigName("genaiz")
	} else {
		viper.SetConfigFile(customPath)
	}

	viper.SetEnvPrefix("genaiz")
	viper.AutomaticEnv()

	if err := viper.ReadInConfig(); err != nil {
		return err
	}

	return nil
}

func Display(bindings ...map[string]string) {
	if len(bindings) > 0 {
		var writer = tabwriter.NewWriter(os.Stdout, 1, 1, 1, ' ', 0)
		var merged = merge(bindings)
		var labels = reverse(merged)
		var ordering = sortKeys(labels)

		for _, key := range ordering {
			_, err := fmt.Fprintf(writer, "%s:\t%s\n", key, viper.GetString(labels[key]))
			cobra.CheckErr(err)
		}

		cobra.CheckErr(writer.Flush())
	}
}

func FollowUp(flags *pflag.FlagSet, flag string, on func()) bool {
	var follow, err = flags.GetBool(flag)

	cobra.CheckErr(err)

	if follow {
		on()
		return true
	}

	return false
}

func init() {
	home, err := os.UserHomeDir()
	cobra.CheckErr(err)
	DefaultPath = home + "/.config/genaiz"
}

func merge(merging []map[string]string) map[string]string {
	var result = map[string]string{}

	for _, m := range merging {
		maps.Copy(result, m)
	}

	return result
}

func reverse(reversing map[string]string) map[string]string {
	var result = map[string]string{}

	for k, v := range reversing {
		result[v] = k
	}

	return result
}

func sortKeys(sorting map[string]string) []string {
	var result = make([]string, 0, len(sorting))

	for key := range sorting {
		result = append(result, key)
	}

	sort.Strings(result)
	return result
}
