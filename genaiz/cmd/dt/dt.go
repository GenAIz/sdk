package dt

import (
	"context"
	"strconv"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

var (
	errorLinkSeqInvalid = task.NewError("sequence number is invalid")
)

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

type BaseListExecutor struct {
	ledger        *config.Ledger
	accountParams config.AccountParametric
	printerParams cli.PrinterParametric
}

func (ble BaseListExecutor) newDataInstanceListParams(arg string) (*broker.DataInstanceListParams, error) {
	var oem, handle, vn = dk.ParseDataLinkArgument(arg)
	var result = &broker.DataInstanceListParams{
		Broker: *ble.accountParams.BrokerParams(),
	}
	var sequence *int

	if vn != "" {
		if ts := strings.SplitN(vn, "-rc-", 2); len(ts) > 1 {
			var err error
			var ti int

			if ti, err = strconv.Atoi(ts[1]); err != nil {
				return nil, errorLinkSeqInvalid
			}

			vn = ts[0]
			sequence = &ti
		}
	}

	if oem != "" {
		result.DataLink = &broker.DataLink{
			Oem:     oem,
			Handle:  handle,
			Version: vn,
			Seq:     sequence,
		}
	}

	return result, nil
}

type Cli struct {
	cli.BaseCli
}

func NewDt(ledger *config.Ledger) *cobra.Command {
	var dtCmd = &cobra.Command{
		Use:     "data",
		Aliases: []string{"dt"},
		Short:   "GenAIz Data Instance Toolkit",
	}

	dtCmd.AddCommand(NewSource(ledger))
	dtCmd.AddCommand(NewStore(ledger))
	return dtCmd
}
