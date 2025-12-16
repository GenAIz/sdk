package proxy

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/lang"
)

type AddExecutor interface {
	Add(string, int) error
}

type AddExecutorFactory func(command *cobra.Command) AddExecutor

func NewAddProxy(factory AddExecutorFactory) *cobra.Command {
	var addCmd = &cobra.Command{
		Use:     "add HOST:PORT",
		Short:   "Adds an outbound proxy to a Smart Function",
		Long:    "Adds an outbound proxy by host and port to a Smart Function",
		Example: "genaiz sf data proxy add dev.genaiz.com:8081 --tcp --udp",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)
			var host string
			var port int
			var err error

			if host, port, err = cli.ArgsHostAndPort(args[0]); err == nil {
				err = exec.Add(host, port)
			}

			lang.HandleExit(err)
		},
	}

	return addCmd
}
