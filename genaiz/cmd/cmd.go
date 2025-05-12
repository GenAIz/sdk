// Package cmd provides commands for the Genaiz SDK Toolkits
//
// See: genaiz --help
package cmd

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/sf"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/version"

	easy "github.com/t-tomalak/logrus-easy-formatter"
)

type Runner interface {
	Confirm(*config.Repo, ...func()) bool

	Dry(repo *config.Repo) bool

	Pretend(repo *config.Repo) bool
}

type RunnerOptions struct {
	logFormat      *config.StringOption
	logLevel       *config.StringOption
	overrideConfig *config.StringOption
	runConfirm     *config.BoolOption
	runDry         *config.BoolOption
	runPretend     *config.BoolOption
	solutionPath   *config.StringOption
}

func (ro *RunnerOptions) Confirm(repo *config.Repo, display ...func()) bool {
	if repo.GetBool(ro.runConfirm) {
		var r = bufio.NewReader(os.Stdin)
		var message = "Proceed?"

		if len(display) > 0 {
			for _, d := range display {
				d()
			}

			message = "Confirm all options?"
		}

		for {
			if _, err := fmt.Fprintf(os.Stdout, "%s (%s) ", message, "[y/n]"); err == nil {
				var s, _ = r.ReadString('\n')

				s = strings.TrimSpace(s)
				s = strings.ToLower(s)

				if s != "" {
					if s == "y" || s == "yes" {
						return true
					}

					if s == "n" || s == "no" {
						return false
					}
				}
			} else {
				return false
			}
		}
	}

	return true
}

func (ro *RunnerOptions) Dry(repo *config.Repo) bool {
	return repo.GetBool(ro.runDry)
}

func (ro *RunnerOptions) Pretend(repo *config.Repo) bool {
	return repo.GetBool(ro.runPretend)
}

func (ro *RunnerOptions) allDefiners() []config.Definer {
	return []config.Definer{
		ro.overrideConfig,
		ro.solutionPath,
		ro.logFormat,
		ro.logLevel,
		ro.runConfirm,
		ro.runDry,
		ro.runPretend,
	}
}

func New(repo *config.Repo) *cobra.Command {
	var options = NewRunnerOptions()
	var root = &cobra.Command{
		Use:     "genaiz",
		Short:   "Genaiz SDK Toolkits",
		Version: version.Version,
	}

	repo.LoggerFactory = func(repo *config.Repo) *logrus.Logger {
		var level, err = getLevel(repo.GetString(options.logLevel))

		if err != nil {
			repo.LogError("Could not set log level with error %s", err)
		}

		return &logrus.Logger{
			Out:       os.Stdout,
			Level:     level,
			Formatter: getFormatter(repo.GetString(options.logFormat)),
		}
	}

	repo.Register(root, options.allDefiners()...)
	root.AddCommand(sf.NewSf(repo, options.Confirm, options.Dry, options.Pretend))
	return root
}

func NewOptionRunConfirm() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "confirm",
			Usage:        "confirm command options before executing",
			DefaultValue: false,
		},
	}
}

func NewOptionRunDry() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "dry",
			Usage:        "dry-run displays command option resolution only",
			DefaultValue: false,
		},
	}
}

func NewOptionRunPretend() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "pretend",
			Usage:        "pretending displays shell commands that would be executed to accomplish the toolkit command, if there are any",
			DefaultValue: false,
		},
	}
}

func NewOptionLogFormat() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.LogFormat",
			Param:        "logFormat",
			Usage:        "log format as supported by Logrus. Also supports \"json\" for structured logging",
			DefaultValue: "[%time%|%lvl%] %msg%",
		},
	}
}

func NewOptionLogLevel() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.LogLevel",
			Param:        "logLevel",
			Usage:        "log level for controlling logging details. Supported case insensitive values: debug, d, error e, info, i, quiet q, trace t, warning and w",
			DefaultValue: "quiet",
		},
	}
}

func NewOptionOverrideConfig() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Config",
			Param: "config",
			Usage: "configuration file path of the smart function or a solution",
			DefaultSetter: func(repo *config.Repo) any {
				return filepath.Join(repo.WorkDir, ".genaiz")
			},
			Validator: func(value any) bool {
				return config.ValidateFile(value.(string))
			},
		},
	}
}

func NewOptionSolutionPath() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "WF.Path",
			Param: "solution",
			Usage: "configuration file path of the smart function solution, if any",
			Validator: func(value any) bool {
				return config.ValidateDir(value.(string))
			},
		},
	}
}

func NewRunnerOptions() *RunnerOptions {
	return &RunnerOptions{
		logFormat:      NewOptionLogFormat(),
		logLevel:       NewOptionLogLevel(),
		overrideConfig: NewOptionOverrideConfig(),
		runConfirm:     NewOptionRunConfirm(),
		runDry:         NewOptionRunDry(),
		runPretend:     NewOptionRunPretend(),
		solutionPath:   NewOptionSolutionPath(),
	}
}

// getFormatter determines the log format used by the logger for the process
func getFormatter(format string) logrus.Formatter {
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

// getLevel returns the logrus.Level corresponding to the lower cased levelString specified, logrus.InfoLevel if not recognized
func getLevel(levelString string) (logrus.Level, error) {
	switch strings.ToLower(levelString) {
	case "d", "v", "debug", "verbose":
		return logrus.DebugLevel, nil
	case "e", "err", "error":
		return logrus.ErrorLevel, nil
	case "i", "nfo", "info":
		return logrus.InfoLevel, nil
	case "q", "qq", "quiet":
		return logrus.FatalLevel, nil
	case "t", "trc", "trace":
		return logrus.TraceLevel, nil
	case "w", "warn", "warning":
		return logrus.WarnLevel, nil
	case "":
		return logrus.InfoLevel, nil
	}

	return logrus.InfoLevel, fmt.Errorf("level [%s] not supported, info will be used", levelString)
}
