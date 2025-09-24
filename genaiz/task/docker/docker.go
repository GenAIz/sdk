package docker

import (
	"github.com/docker/docker/client"

	"genaiz.com/genaiz-lib/lang/panicz"
)

type Output struct {
	Stream string
}

var (
	dockerClient *client.Client
)

func init() {
	var err error

	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	panicz.PanicIfError(err)
}
