package auto

import (
	"github.com/spf13/cobra"
)

type Bridge interface {
	Bridge(string) ([]cobra.Completion, cobra.ShellCompDirective)
}
