package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func ArgsOptionalFolder(typeName string, maxSize int, validates config.Validates) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) >= maxSize {
			var lastArg = maxSize - 1

			if info, err := os.Stat(args[lastArg]); err == nil {
				if !info.IsDir() {
					return fmt.Errorf("%s is not a folder", args[lastArg])
				}
			} else if !validates(filepath.Base(args[lastArg])) {
				return fmt.Errorf("%s is not a valid %s name", typeName, args[lastArg])
			}
		}

		return nil
	}
}
