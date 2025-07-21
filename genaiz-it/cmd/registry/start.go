package registry

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-it/compose"
	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
)

func NewStart() *cobra.Command {
	var startCmd = &cobra.Command{
		Use:   "start [WORKDIR]",
		Short: "Starts the CNCF Distribution registry",
		Long:  "Starts the CNCF Distribution registry if it is not running",
		Run: func(cmd *cobra.Command, args []string) {
			var services = compose.NewServices(cmd.Context())
			var reset func()
			var err error

			if reset, err = dirz.CreateWorkingDir(args...); err == nil {
				defer errorz.DeferOnExit(&err, reset)()
				var service = services.Registry()

				if err = service.Init(); err == nil {
					err = service.Start()
				}
			}
		},
	}

	return startCmd
}
