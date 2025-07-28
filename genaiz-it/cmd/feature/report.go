package feature

import (
	"fmt"

	"github.com/spf13/cobra"
)

func NewReport() *cobra.Command {
	var reportCmd = &cobra.Command{
		Use:   "report [WORKDIR]",
		Short: "Collects test results from the specified work directory",
		Long:  "Collects test results from the specified work directory, producing a report",
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Println("TODO: See what kind of report we can produce with Cucumber compatibility")
		},
	}

	return reportCmd
}
