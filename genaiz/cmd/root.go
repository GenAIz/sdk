package cmd

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/sf"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/version"
)

var (
	configPath string
	rootCmd    = &cobra.Command{
		Use:     "genaiz",
		Short:   "Genaiz SDK Toolkits",
		Version: version.Version,
	}
)

func Execute(ctx context.Context) error {
	return rootCmd.ExecuteContext(ctx)
}

func init() {
	cobra.OnInitialize(initConfig)
	rootCmd.PersistentFlags().StringVar(&configPath, "config", "", "config path (default is "+config.DefaultPath+")")
	rootCmd.AddCommand(sf.Cmd())
}

func initConfig() {
	config.Default(configPath)
}
