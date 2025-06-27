package main

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd"
	"genaiz.com/genaiz/config"
)

func main() {
	var ledger = config.NewLedger()

	cobra.OnInitialize(ledger.Init)
	root := cmd.New(ledger)
	cobra.OnInitialize(ledger.InitDefaults)
	cobra.OnInitialize(ledger.InitLogging)

	if err := root.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}
