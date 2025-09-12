package sn

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task/shared"
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger

	folderPath string
}

func (be BaseExecutor) makeConfigParams(option *config.StringOption) *shared.ConfigParams {
	var configType, err = be.Ledger.GetConfigType(option)

	lang.HandleExit(err)
	return &shared.ConfigParams{
		ConfigName:   be.Ledger.ConfigName,
		ConfigType:   configType,
		ConfigFolder: be.folderPath,
	}
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
	wfCmd.AddCommand(NewPublish(ledger, snCli))
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
