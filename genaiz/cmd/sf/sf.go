// Package sf provides commands for managing Genaiz Smart Functions.
// Smart Functions commands include build, create, debug, init, list, run, start, stop and test.
//
// See: genaiz sf --help
package sf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

type Decisive func(*config.Ledger) bool

type Interactive func(*config.Ledger, ...func()) bool

type Executor interface {
	Display()

	Pretend()

	Proceed()
}

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

type Cli struct {
	Confirm Interactive
	Dry     Decisive
	Pretend Decisive

	optionDockerContext *config.StringOption
	optionDockerFile    *config.StringOption
	optionDockerTag     *config.StringOption
	optionDockerVersion *config.StringOption
}

func (c *Cli) Exec(ledger *config.Ledger, executor Executor) {
	if !c.isDry(ledger, executor.Display) {
		if c.isPretend(ledger) {
			executor.Pretend()
		} else {
			c.execConfirm(ledger, executor.Proceed, executor.Display)
		}
	}
}

func (c *Cli) SfOptions() []*config.Option {
	return []*config.Option{
		&c.optionDockerContext.Option,
		&c.optionDockerFile.Option,
		&c.optionDockerTag.Option,
		&c.optionDockerVersion.Option,
	}
}

func (c *Cli) allDefiners() []config.Definer {
	return []config.Definer{
		c.optionDockerContext,
		c.optionDockerFile,
		c.optionDockerTag,
		c.optionDockerVersion,
	}
}

func (c *Cli) execConfirm(ledger *config.Ledger, exec func(), display ...func()) {
	if c.Confirm != nil && c.Confirm(ledger, display...) {
		exec()
	} else {
		fmt.Println("Cancelled, exiting")
		os.Exit(0)
	}
}

func (c *Cli) isDecisive(ledger *config.Ledger, decisive Decisive, display ...func()) bool {
	if decisive != nil && decisive(ledger) {
		for _, d := range display {
			d()
		}

		return true
	}

	return false
}

func (c *Cli) isDry(ledger *config.Ledger, display ...func()) bool {
	return c.isDecisive(ledger, c.Dry, display...)
}

func (c *Cli) isPretend(ledger *config.Ledger) bool {
	return c.isDecisive(ledger, c.Pretend)
}

func NewSf(ledger *config.Ledger, confirm Interactive, dry, pretend Decisive) *cobra.Command {
	var cli = NewSfCli(confirm, dry, pretend)
	var sf = &cobra.Command{
		Use:     "function",
		Aliases: []string{"sf"},
		Short:   "Genaiz Smart Function Toolkit",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var flags = cmd.Flags()

			ledger.ToWorkDir(cli.optionDockerContext, flags)
			ledger.FromWorkDir(cli.optionDockerFile, flags)
		},
	}

	ledger.Register(sf, cli.allDefiners()...)
	sf.AddCommand(
		NewBuild(ledger, cli),
		NewCreate(ledger, cli),
		NewInit(ledger, cli),
		NewList(ledger, cli),
		NewRun(ledger, cli),
		NewDebug(ledger, cli),
		NewTest(ledger, cli),
		NewStop(ledger, cli),
		NewStart(ledger, cli),
		NewPublish(ledger, cli),
	)
	// The sf command captures context modifications and those are needed before the ledger sets defaults
	cobra.OnInitialize(func() { ledger.ChangeWorkDir(cli.optionDockerContext) })
	return sf
}

func NewSfCli(confirm Interactive, dry, pretend Decisive) *Cli {
	return &Cli{
		Confirm: confirm,
		Dry:     dry,
		Pretend: pretend,

		optionDockerContext: newOptionDockerContext(),
		optionDockerFile:    newOptionDockerFile(),
		optionDockerTag:     newOptionDockerTag(),
		optionDockerVersion: newOptionDockerVersion(),
	}
}

func newOptionDockerContext() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.DockerContext",
			Param:        "context",
			Short:        "c",
			Usage:        "Docker build context path",
			DefaultValue: "$PWD",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.WorkDir
			},
			Validator: config.Validation.DirExists,
		},
	}
}

func newOptionDockerFile() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.Dockerfile",
			Param:        "file",
			Short:        "f",
			Usage:        "Dockerfile path",
			DefaultValue: "$PWD/Dockerfile",
			DefaultGetter: func(ledger *config.Ledger) any {
				return filepath.Join(ledger.WorkDir, "Dockerfile")
			},
			Validator: config.Validation.FileExists,
		},
	}
}

func newOptionDockerTag() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Build.Tag",
			Param: "tag",
			Usage: "tag the smart function image, defaults to the context dir name",
			DefaultSetter: func(ledger *config.Ledger) any {
				return filepath.Base(ledger.WorkDir)
			},
		},
	}
}

func newOptionDockerVersion() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.Build.Version",
			Param:        "version",
			Usage:        "version of the smart image",
			DefaultValue: "latest",
		},
	}
}
