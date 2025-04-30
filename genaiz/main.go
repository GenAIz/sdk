package main

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd"
)

func main() {
	cobra.CheckErr(cmd.Execute(context.Background()))
}
