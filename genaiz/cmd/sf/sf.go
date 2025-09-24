// Package sf provides commands for managing Genaiz Smart Functions.
// Smart Functions commands include build, create, debug, init, list, run, start, stop and test.
//
// See: genaiz sf --help
package sf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
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

	parentConfigType *shared.ConfigType
	parentSolution   *broker.Solution
}

func (c *Cli) ParentConfigType(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentConfigType == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		return c.parentConfigType
	}
}

func (c *Cli) ParentOem(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentSolution == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		return c.parentSolution.Oem
	}
}

func (c *Cli) ParentVersion(parentOption *config.StringOption) func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		if c.parentSolution == nil {
			c.parentSolution = c.parseParentSolution(ledger, parentOption)
		}

		return c.parentSolution.Version
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

func (c *Cli) parseParentSolution(ledger *config.Ledger, parentOption *config.StringOption) *broker.Solution {
	var solutionPath = ledger.GetString(parentOption)
	var solutionReader = config.NewSolutionReader(ledger).
		WithConfigPath(solutionPath)

	if result, err := solutionReader.ReadName(ledger.ConfigName); err == nil {
		c.parentConfigType = lang.Ref(solutionReader.GetConfigType())
		return result
	} else {
		ledger.LogDebug("could not read a solution under path [%s]", solutionPath)
		c.parentConfigType = lang.Ref(shared.ConfigTypeNone)
	}

	return &broker.Solution{Version: "0.1.0"}
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

		optionDockerContext: cli.Options.Docker.ContextPath().BuildStringOption(),
		optionDockerFile:    cli.Options.Docker.FilePath().BuildStringOption(),
		optionDockerTag:     cli.Options.Docker.Tag().BuildStringOption(),
		optionDockerVersion: cli.Options.Docker.Version().BuildStringOption(),
	}
}
