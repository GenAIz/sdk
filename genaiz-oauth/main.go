package main

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-oauth/cmd"
)

func main() {
	var root = cmd.New()

	if err := root.Execute(); err != nil {
		cobra.CheckErr(err)
	}
}
