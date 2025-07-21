package docker

import (
	"errors"
	"fmt"
	"io"

	"github.com/docker/docker/api/types/image"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type PushParams struct {
	task.Env
	*BuildParams
}

func NewPushTask() *task.Task[PushParams] {
	return &task.Task[PushParams]{
		Name:       "docker-push",
		OnPrepare:  handlePushContext,
		OnComplete: handlePushComplete,
		OnPretend:  handlePushPretend,
	}
}

func handlePushContext(params *PushParams, state *task.State) error {
	if state.Internal == nil {
		return errors.New("push called without provisioning")
	}

	if state.Output == "" {
		return errors.New("push called without a built image")
	}

	return nil
}

func handlePushComplete(params *PushParams, state *task.State) error {
	if state.Internal != nil && state.Output != "" {
		var remote = state.Internal.(*shared.Identity)
		var pushOptions = image.PushOptions{RegistryAuth: remote.Auth}
		var rd io.ReadCloser
		var err error

		state.Logger.Debugf("Tagging docker image [%s]", remote.Path)

		if err = dockerClient.ImageTag(params.Context, state.Output, remote.Path); err == nil {
			state.Logger.Debugf("Pushing docker image [%s]", remote.Id)

			if rd, err = dockerClient.ImagePush(params.Context, remote.Path, pushOptions); err == nil {
				defer filez.CloseSilently(rd)
			}
		}

		state.Output = ""
		return err
	}

	return errors.New("no provisioned image to push")
}

func handlePushPretend(params *PushParams, state *task.State) error {
	if state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Pretending to tag docker image [%s]", remote.Path)
		fmt.Printf("docker tag %s %s\n", remote.Hash, remote.Path)
		state.Logger.Debugf("Pretending to push docker image [%s]", remote.Hash)
		fmt.Printf("docker push %s\n", remote.Path)
		state.Output = ""
		return nil
	}

	state.Logger.Errorf("Could not pretend to push docker image for file [%s] under context [%s]", params.Dockerfile, params.DockerContext)
	return errors.New("no provisioned image to push")
}
