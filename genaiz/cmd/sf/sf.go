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

type Decisive func(*config.Repo) bool

type Interactive func(*config.Repo, ...func()) bool

type Executor interface {
	Display()

	Pretend()

	Proceed()
}

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Repo    *config.Repo
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

func (c *Cli) Exec(repo *config.Repo, executor Executor) {
	if !c.isDry(repo, executor.Display) {
		if c.isPretend(repo) {
			executor.Pretend()
		} else {
			c.execConfirm(repo, executor.Proceed, executor.Display)
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

func (c *Cli) execConfirm(repo *config.Repo, exec func(), display ...func()) {
	if c.Confirm != nil && c.Confirm(repo, display...) {
		exec()
	} else {
		fmt.Println("Cancelled, exiting")
		os.Exit(0)
	}
}

func (c *Cli) isDecisive(repo *config.Repo, decisive Decisive, display ...func()) bool {
	if decisive != nil && decisive(repo) {
		for _, d := range display {
			d()
		}

		return true
	}

	return false
}

func (c *Cli) isDry(repo *config.Repo, display ...func()) bool {
	return c.isDecisive(repo, c.Dry, display...)
}

func (c *Cli) isPretend(repo *config.Repo) bool {
	return c.isDecisive(repo, c.Pretend)
}

func NewSf(repo *config.Repo, confirm Interactive, dry, pretend Decisive) *cobra.Command {
	var cli = NewCli(confirm, dry, pretend)
	var sf = &cobra.Command{
		Use:     "function",
		Aliases: []string{"sf"},
		Short:   "Genaiz Smart Function Toolkit",
		PersistentPreRun: func(cmd *cobra.Command, args []string) {
			var flags = cmd.Flags()

			repo.ToWorkDir(cli.optionDockerContext, flags)
			repo.FromWorkDir(cli.optionDockerFile, flags)
		},
	}

	repo.Register(sf, cli.allDefiners()...)
	sf.AddCommand(
		NewBuild(repo, cli),
		NewRun(repo, cli),
		NewDebug(repo, cli),
		NewTest(repo, cli),
		NewStop(repo, cli),
		NewStart(repo, cli),
	)
	// The sf command captures context modifications and those are needed before the repo sets defaults
	cobra.OnInitialize(func() { repo.ChangeWorkDir(cli.optionDockerContext) })
	return sf
}

func NewCli(confirm Interactive, dry, pretend Decisive) *Cli {
	return &Cli{
		Confirm:             confirm,
		Dry:                 dry,
		Pretend:             pretend,
		optionDockerContext: NewOptionDockerContext(),
		optionDockerFile:    NewOptionDockerFile(),
		optionDockerTag:     NewOptionDockerTag(),
		optionDockerVersion: NewOptionDockerVersion(),
	}
}

func NewOptionDockerContext() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.DockerContext",
			Param:        "context",
			Short:        "c",
			Usage:        "Docker build context path",
			DefaultValue: "$PWD",
			DefaultGetter: func(repo *config.Repo) any {
				return repo.WorkDir
			},
			Validator: config.ValidateDir,
		},
	}
}

func NewOptionDockerFile() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.Dockerfile",
			Param:        "file",
			Short:        "f",
			Usage:        "Dockerfile path",
			DefaultValue: "$PWD/Dockerfile",
			DefaultGetter: func(repo *config.Repo) any {
				return filepath.Join(repo.WorkDir, "Dockerfile")
			},
			Validator: config.ValidateFile,
		},
	}
}

func NewOptionDockerTag() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "SF.Build.Tag",
			Param: "tag",
			Usage: "tag the smart function image, defaults to the context dir name",
			DefaultSetter: func(repo *config.Repo) any {
				return filepath.Base(repo.WorkDir)
			},
		},
	}
}

func NewOptionDockerVersion() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "SF.Build.Version",
			Param:        "version",
			Usage:        "version of the smart image",
			DefaultValue: "latest",
		},
	}
}
