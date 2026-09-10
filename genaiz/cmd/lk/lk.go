package lk

import (
	"context"
	"os"

	"github.com/awnumar/memguard"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/locker"
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

func (be BaseExecutor) newSourceBaseParams(optionLocker *config.StringOption) *locker.BaseParams {
	var baseParams = &locker.BaseParams{
		LockerPath: be.Ledger.GetString(optionLocker),
	}

	if envPwd := os.Getenv(passphraseEnvKey); envPwd != "" {
		baseParams.Passphrase = memguard.NewEnclave([]byte(envPwd))
	} else if pwdEnclave := be.Ledger.QuerySecret(passphrasePrompt); pwdEnclave != nil {
		baseParams.Passphrase = pwdEnclave
	}

	return baseParams
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
