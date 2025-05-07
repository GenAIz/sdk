// Package cmd provides commands for the Genaiz command line toolkit.
//
// See: genaiz --help
package cmd

import (
	"context"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz/cmd/sf"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/version"

	easy "github.com/t-tomalak/logrus-easy-formatter"
)

const (
	keyGenaizLogFormat   = "SmartFunction.LogFormat" // Configuration key used in config files for specifying the log format
	keyGenaizLogLevel    = "SmartFunction.LogLevel"  // Configuration key used in config files for specifying the log level
	paramGenaizLogFormat = "logFormat"               // Parameter label used from the shell for specifying the log format
	paramGenaizLogLevel  = "logLevel"                // Parameter label used from the shell for specifying the log level
)

var (
	configPath string // Path of the configuration file specified on the shell

	// Global root command for initiating sub-commands sf and other toolkit utilities
	rootCmd = &cobra.Command{
		Use:     "genaiz",
		Short:   "Genaiz SDK Toolkits",
		Version: version.Version,
	}

	// Default values for the rootCmd parameters
	rootDefaults = map[string]func() string{
		keyGenaizLogLevel: func() string {
			return "info"
		},
	}

	// Mappings between parametric log levels and their associated Logrus levels
	rootLogLevels = map[string]logrus.Level{
		"debug":   logrus.DebugLevel,
		"error":   logrus.ErrorLevel,
		"info":    logrus.InfoLevel,
		"quiet":   logrus.FatalLevel,
		"trace":   logrus.TraceLevel,
		"warning": logrus.WarnLevel,
	}

	// rootCmd mappings between config files and shell
	rootMappings = map[string]string{
		keyGenaizLogFormat: paramGenaizLogFormat,
		keyGenaizLogLevel:  paramGenaizLogLevel,
	}
)

// Execute calls the rootCmd with the specified shell context
func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config path (default is "+config.DefaultPath+")")
	rootCmd.PersistentFlags().String(paramGenaizLogFormat, "", "log format. Valid values are \"json\" or a custom format string. By default a custom format string with timestamp and level is used")
	rootCmd.PersistentFlags().String(paramGenaizLogLevel, "", "log level. Valid values are trace, debug, error, warning, info and quiet. Defaults to info")
	rootCmd.AddCommand(sf.Cmd())
}

// initConfig initializes the configuration path, the root parameter/configuration mappings and defaults and the logger for the process
func initConfig() {
	var defaultErr = config.Default(configPath)

	if defaultErr != nil && configPath != "" {
		cobra.CheckErr(defaultErr)
	}

	config.BindCmd(rootCmd, rootMappings)
	config.BindDefaults(rootDefaults)
	config.Logger = &logrus.Logger{
		Out:       os.Stdout,
		Level:     getLevel(),
		Formatter: getFormatter(),
	}

	if defaultErr != nil {
		config.Logger.Debugf("Could not locate default config path [%s]", config.DefaultPath)
	}

	config.Logger.Debugf("Using config file [%s]", viper.ConfigFileUsed())
}

// getFormatter determines the log format used by the logger for the process
func getFormatter() logrus.Formatter {
	var format = viper.GetString(keyGenaizLogFormat)

	if format == "json" {
		return &logrus.JSONFormatter{
			TimestampFormat: time.DateTime,
		}
	} else if strings.Contains(format, "%") {
		return &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       format + "\n",
		}
	} else {
		return &easy.Formatter{
			TimestampFormat: time.DateTime,
			LogFormat:       "[%time%|%lvl%] %msg%\n",
		}
	}
}

// getLevel determines the log level to set on the logger for the process
func getLevel() logrus.Level {
	if level, ok := rootLogLevels[viper.GetString(keyGenaizLogLevel)]; ok {
		return level
	}

	return logrus.InfoLevel
}
