package sc

import (
	"github.com/spf13/cobra"
)

func NewSc() *cobra.Command {
	var schema = &cobra.Command{
		Use:     "schema",
		Aliases: []string{"sc"},
		Short:   "Genaiz Schema Utility Toolkit",
	}

	schema.AddCommand(NewPrint())
	return schema
}
