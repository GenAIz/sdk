package data

import (
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
)

type baseOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (bo baseOptions) AccountParams(ledger *config.Ledger) config.AccountParametric {
	return config.NewAccountParams(ledger, bo.optionAccount)
}

func (bo baseOptions) Definers() []config.Definer {
	return []config.Definer{
		bo.optionAccount,
		bo.optionJsonPrinter,
	}
}

func (bo baseOptions) Options() []*config.Option {
	return []*config.Option{
		&bo.optionAccount.Option,
		&bo.optionJsonPrinter.Option,
	}
}

func (bo baseOptions) PrinterParams(ledger *config.Ledger) cli.PrinterParametric {
	return cli.NewPrinterParam(ledger, bo.optionJsonPrinter)
}
