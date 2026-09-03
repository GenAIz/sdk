package lk

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
)

const (
	passphraseEnvKey = "GENAIZ_LK_PASSWORD"
	passphrasePrompt = "enter passphrase: "
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

type Cli struct {
	cli.BaseCli
}

func NewLk(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var lkCli = NewLkCli(confirm, dry, pretend)
	var lkCmd = &cobra.Command{
		Use:     "locker",
		Aliases: []string{"lk"},
		Short:   "GenAIz Locker Toolkit",
	}

	lkCmd.AddCommand(NewInit(ledger, lkCli))
	lkCmd.AddCommand(NewSource(ledger, lkCli))
	return lkCmd
}

func NewLkCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}
