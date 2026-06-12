package proxy

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, string, int) error
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveProxy(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm [OEM/]HANDLE[:VERSION] HOST:PORT",
		Short:   "Removes an outbound proxy from a Data Link",
		Long:    "Removes an outbound proxy by host and port from a Data Link",
		Example: "genaiz dk proxy rm com.genaiz/datalink-1:1.0.0 192.168.1.2:22",
		Args:    cobra.MaximumNArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var host string
			var port int
			var err error

			if host, port, err = cli.ArgsHostAndPort(args[1]); err == nil {
				err = exec.Remove(args[0], host, port)
			}

			lang.HandleExit(err)
		},
	}

	return rmCmd
}
