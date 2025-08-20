// Package sf provides commands for managing Genaiz Smart Functions.
// Smart Functions commands include build, create, debug, init, list, run, start, stop and test.
//
// See: genaiz sf --help
package sf

import (
	"context"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

type Cli struct {
	cli.BaseCli
	optionDockerContext *config.StringOption
	optionDockerFile    *config.StringOption
	optionDockerTag     *config.StringOption
	optionDockerVersion *config.StringOption
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

func NewSf(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var sfCli = NewSfCli(confirm, dry, pretend)
	var sfCmd = &cobra.Command{
		Use:     "function",
		Aliases: []string{"sf"},
		Short:   "Genaiz Smart Function Toolkit",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var flags = cmd.Flags()

			ledger.ToWorkDir(sfCli.optionDockerContext, flags)
			ledger.FromWorkDir(sfCli.optionDockerFile, flags)
		},
	}

	ledger.Register(sfCmd, sfCli.allDefiners()...)
	sfCmd.AddCommand(
		NewBuild(ledger, sfCli),
		NewCreate(ledger, sfCli),
		NewInit(ledger, sfCli),
		NewList(ledger, sfCli),
		NewRun(ledger, sfCli),
		NewDebug(ledger, sfCli),
		NewTest(ledger, sfCli),
		NewStop(ledger, sfCli),
		NewStart(ledger, sfCli),
		NewPublish(ledger, sfCli),
	)
	// The sf command captures context modifications and those are needed before the ledger sets defaults
	cobra.OnInitialize(func() { ledger.ChangeWorkDir(sfCli.optionDockerContext) })
	return sfCmd
}

func NewSfCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},

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
