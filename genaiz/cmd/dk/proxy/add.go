package proxy

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, string, int) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddProxy(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add [OEM/]HANDLE[:VERSION] HOST:PORT",
		Short:   "Adds an outbound proxy to a Data Link",
		Long:    "Adds an outbound proxy by host and port to a Data Link",
		Example: "genaiz dk proxy add com.genaiz/datalink-1:1.0.0 dev.genaiz.com:8081 --tcp --udp",
		Args:    cobra.ExactArgs(2),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var host string
			var port int
			var err error

			if host, port, err = cli.ArgsHostAndPort(args[1]); err == nil {
				err = exec.Add(args[0], host, port)
			}

			lang.HandleExit(err)
		},
	}

	return addCmd
}
