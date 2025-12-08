// Package cmd provides commands for the Genaiz SDK Toolkits
//
// See: genaiz --help
package cmd

import (
	"bufio"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ac"
	"genaiz.com/genaiz/cmd/sc"
	"genaiz.com/genaiz/cmd/sf"
	"genaiz.com/genaiz/cmd/sn"
	"genaiz.com/genaiz/cmd/wf"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/version"

	easy "github.com/t-tomalak/logrus-easy-formatter"
)

type Runner interface {
	Confirm(*config.Ledger, ...func()) bool

	Dry(*config.Ledger) bool

	Pretend(*config.Ledger) bool
}

type RunnerOptions struct {
	logFormat  *config.StringOption
	logLevel   *config.StringOption
	runConfig  *config.StringOption
	runConfirm *config.BoolOption
	runDry     *config.BoolOption
	runPretend *config.BoolOption
	stdIn      io.Reader
	stdOut     io.Writer
}

func (ro *RunnerOptions) Confirm(ledger *config.Ledger, display ...func()) bool {
	if ledger.GetBool(ro.runConfirm) {
		var r = bufio.NewReader(ro.stdIn)
		var message = "Proceed?"

		if len(display) > 0 {
			for _, d := range display {
				d()
			}

			message = "Confirm all options?"
		}

		for {
			if _, err := fmt.Fprintf(ro.stdOut, "%s (%s) ", message, "[y/n]"); err == nil {
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

func (ro *RunnerOptions) Dry(ledger *config.Ledger) bool {
	return ledger.GetBool(ro.runDry)
}

func (ro *RunnerOptions) Pretend(ledger *config.Ledger) bool {
	return ledger.GetBool(ro.runPretend)
}

func (ro *RunnerOptions) allDefiners() []config.Definer {
	return []config.Definer{
		ro.runConfig,
		ro.logFormat,
		ro.logLevel,
		ro.runConfirm,
		ro.runDry,
		ro.runPretend,
	}
}

func New(ledger *config.Ledger) *cobra.Command {
	var options = NewRunnerOptions()
	var root = &cobra.Command{
		Use:     "genaiz",
		Short:   "Genaiz SmartFunction Toolkit",
		Version: version.GetVersion(),
	}

	ledger.LoggerFactory = func(ledger *config.Ledger) *logrus.Logger {
		var level, err = getLevel(ledger.GetString(options.logLevel))

		if err != nil {
			ledger.LogError("Could not set log level with error %s", err)
		}

		return &logrus.Logger{
			Out:       os.Stdout,
			Level:     level,
			Formatter: getFormatter(ledger.GetString(options.logFormat)),
		}
	}

	ledger.Register(root, options.allDefiners()...)
	ledger.AddConfigOption(options.runConfig)
	root.AddCommand(wf.NewWf(ledger, options.Confirm, options.Dry, options.Pretend))
	root.AddCommand(sn.NewSn(ledger, options.Confirm, options.Dry, options.Pretend))
	root.AddCommand(sf.NewSf(ledger, options.Confirm, options.Dry, options.Pretend))
	root.AddCommand(ac.NewAc(ledger))
	root.AddCommand(sc.NewSc())
	return root
}

func NewRunnerOptions() *RunnerOptions {
	return &RunnerOptions{
		logFormat:  cli.Options.Solutions.LogFormat().BuildStringOption(),
		logLevel:   cli.Options.Solutions.LogLevel().BuildStringOption(),
		runConfig:  newOptionRunConfig(),
		runConfirm: newOptionRunConfirm(),
		runDry:     newOptionRunDry(),
		runPretend: newOptionRunPretend(),
		stdIn:      os.Stdin,
		stdOut:     os.Stdout,
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

func newOptionRunConfig() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Param: "config",
			Usage: "configuration file path of the smart function or a solution",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.WorkDir
			},
		},
	}
}

func newOptionRunConfirm() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "confirm",
			Usage:        "confirm command options before executing",
			DefaultValue: false,
		},
	}
}

func newOptionRunDry() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "dry",
			Usage:        "dry-run displays command option resolution only",
			DefaultValue: false,
		},
	}
}

func newOptionRunPretend() *config.BoolOption {
	return &config.BoolOption{
		Option: config.Option{
			Param:        "pretend",
			Usage:        "pretending displays shell commands that would be executed to accomplish the toolkit command, if there are any",
			DefaultValue: false,
		},
	}
}
