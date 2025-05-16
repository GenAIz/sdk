package docker

import "github.com/docker/docker/client"

type Output struct {
	Stream string
}

var (
	dockerClient *client.Client
)

func init() {
	var err error
	dockerClient, err = client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())

	if err != nil {
		panic(err)
	}
}
