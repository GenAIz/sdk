package ws

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

type Cli struct {
	cli.BaseCli
}

func NewWs(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var wsCli = NewWsCli(confirm, dry, pretend)
	var wsValidation = NewWsValidation()
	var wsCmd = &cobra.Command{
		Use:     "workspace",
		Aliases: []string{"ws"},
		Short:   "Genaiz Workspace Management Toolkit",
	}

	wsCmd.AddCommand(NewCreate(ledger, wsCli, wsValidation))
	wsCmd.AddCommand(NewFlow(ledger, wsCli))
	wsCmd.AddCommand(NewList(ledger, wsCli))
	return wsCmd
}

func NewWsCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}

type Validation struct {
	workspaceName config.Validates
}

func (v Validation) ArgsWorkspaceName(position int) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if position < len(args) {
			if v.workspaceName(args[position]) {
				return nil
			}

			return task.NewErrorBuilder().
				Field("workspace.name").
				Build("workspace name must not exceed 255 characters")
		}

		panic("positional arg out of range")
	}
}

func NewWsValidation() *Validation {
	return &Validation{
		workspaceName: config.Validation.Name,
	}
}
