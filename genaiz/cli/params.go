package cli

import (
	"io"
	"os"

	"genaiz.com/genaiz/config"
)

type PrinterParametric interface {
	Printer() Printer

	IsDefault() bool
}

type printerParametric struct {
	ledger     *config.Ledger
	jsonOption *config.BoolOption

	consoleFactory func(io.Writer, io.Writer) Printer
	jsonFactory    func(io.Writer) Printer
}

func (pp *printerParametric) Printer() Printer {
	if pp.IsDefault() {
		return pp.consoleFactory(os.Stderr, os.Stdout)
	}

	return pp.jsonFactory(os.Stdout)
}

func (pp *printerParametric) IsDefault() bool {
	return !pp.ledger.GetBool(pp.jsonOption)
}

func NewPrinterParam(ledger *config.Ledger, option *config.BoolOption) PrinterParametric {
	return &printerParametric{
		consoleFactory: NewConsolePrinter,
		ledger:         ledger,
		jsonOption:     option,
		jsonFactory:    NewJsonPrinter,
	}
}
