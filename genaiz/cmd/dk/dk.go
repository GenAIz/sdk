package dk

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

func NewDk(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var dkCli = NewDkCli(confirm, dry, pretend)
	var dkCmd = &cobra.Command{
		Use:     "datalink",
		Aliases: []string{"dk"},
		Short:   "Genaiz Data Link Toolkit",
	}

	dkCmd.AddCommand(NewCreate(ledger, dkCli))
	return dkCmd
}

func NewDkCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}
