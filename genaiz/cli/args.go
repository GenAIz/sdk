package cli

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
)

func ArgsFolderValidator(typeName string, validates config.Validates) cobra.PositionalArgs {
	return func(cmd *cobra.Command, args []string) error {
		if len(args) >= 1 {
			if info, err := os.Stat(args[0]); err == nil {
				if !info.IsDir() {
					return fmt.Errorf("%s is not a folder", args[0])
				}
			} else if !validates(filepath.Base(args[0])) {
				return fmt.Errorf("%s is not a valid %s name", typeName, args[0])
			}
		}

		return nil
	}
}
