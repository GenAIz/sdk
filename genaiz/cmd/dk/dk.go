package dk

import (
	"context"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task/shared"
)

type dataLinksWriterFactory func(*config.Ledger, string) *dataLinksWriter

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

func (pe BaseExecutor) getConfigPath(userOption *config.BoolOption) string {
	var result string

	if pe.Ledger.GetBool(userOption) {
		result = pe.Ledger.UserPath
	} else {
		result = pe.Ledger.WorkDir
	}

	return result
}

func (pe BaseExecutor) makeConfigParams(typeOption *config.StringOption, userOption *config.BoolOption) (*shared.ConfigParams, error) {
	var configType *shared.ConfigType
	var err error

	if configType, err = pe.Ledger.GetConfigType(typeOption); err == nil {
		var configPath = pe.getConfigPath(userOption)

		return &shared.ConfigParams{
			ConfigName:   pe.Ledger.ConfigName,
			ConfigType:   configType,
			ConfigFolder: configPath,
		}, nil
	}

	return nil, err
}

func (pe BaseExecutor) parseDataLinkArgument(linkArgument string) (string, string, string) {
	var oem, handle, version string

	if linkArgument != "" {
		var handleWithOem = strings.SplitN(linkArgument, "/", 2)
		var handleWithVersion []string
		var handleSuffix string

		if len(handleWithOem) == 2 {
			oem = handleWithOem[0]
			handleSuffix = handleWithOem[1]
		} else {
			handleSuffix = handleWithOem[0]
		}

		handleWithVersion = strings.SplitN(handleSuffix, ":", 2)

		if len(handleWithVersion) == 2 {
			handle = handleWithVersion[0]
			version = handleWithVersion[1]
		} else {
			handle = handleWithVersion[0]
		}
	}

	return oem, handle, version
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
	dkCmd.AddCommand(NewProp(ledger, dkCli))
	dkCmd.AddCommand(NewPublish(ledger, dkCli))
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

type dataLinksWriter struct {
	*config.DataLinksWriter
}

func newDataLinksWriter(ledger *config.Ledger, output string) *dataLinksWriter {
	return &dataLinksWriter{
		DataLinksWriter: config.NewDataLinkWriter().
			Read(ledger, output),
	}
}
