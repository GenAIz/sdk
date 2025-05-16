package main

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd"
	"genaiz.com/genaiz/config"
)

func main() {
	var repo = config.NewRepo()

	cobra.OnInitialize(repo.Init)
	root := cmd.New(repo)
	cobra.OnInitialize(repo.InitDefaults)
	cobra.OnInitialize(repo.InitLogging)

	if err := root.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}
