package proxy

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type RemoveExecutor interface {
	Remove(string, int)
}

type RemoveExecutorFactory func(command *cobra.Command) RemoveExecutor

func NewRemoveProxy(factory RemoveExecutorFactory) *cobra.Command {
	var rmCmd = &cobra.Command{
		Use:     "rm HOST:PORT",
		Short:   "Removes an outbound proxy from a Smart Function",
		Long:    "Removes an outbound proxy by host and port from a Smart Function",
		Example: "genaiz sf data proxy rm 192.168.1.2:22",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)

			if host, port, err := cli.ArgsHostAndPort(args[0]); err == nil {
				exec.Remove(host, port)
			} else {
				lang.HandleExit(err)
			}
		},
	}

	return rmCmd
}
