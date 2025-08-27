package sn

import (
	"context"

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
}

func NewSn(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var snCli = NewSnCli(confirm, dry, pretend)
	var wfCmd = &cobra.Command{
		Use:     "solution",
		Aliases: []string{"sn"},
		Short:   "Genaiz Solution Toolkit",
	}

	wfCmd.AddCommand(NewCreate(ledger, snCli))
	return wfCmd
}

func NewSnCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}
