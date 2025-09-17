package docker

import (
	"bufio"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorSkipPush = errors.New("skipping push, due to existing repository hash")
)

type PushStatus struct {
	Id             string                    `json:"id"`
	Status         string                    `json:"status"`
	Progress       string                    `json:"progress"`
	ProgressDetail *PushStatusProgressDetail `json:"progressDetail"`
	Aux            *PushStatusAux            `json:"aux"`
}

type PushStatusAux struct {
	Tag    string `json:"Tag"`
	Digest string `json:"Digest"`
	Size   int64  `json:"Size"`
}

type PushStatusProgressDetail struct {
	Current int64 `json:"current"`
	Total   int64 `json:"total"`
}

type PushParams struct {
	task.Env
	*BuildParams
}

func NewPushTask() *task.Task[PushParams] {
	return &task.Task[PushParams]{
		Name:         "docker-push",
		OnPrepare:    handlePushContext,
		OnComplete:   handlePushComplete,
		OnIncomplete: handlePushIncomplete,
		OnPretend:    handlePushPretend,
	}
}

func handlePushContext(params *PushParams, state *task.State) error {
	if state.Internal == nil {
		return errors.New("push called without provisioning")
	} else {
		var current = state.Internal.(*shared.Identity)

		if !current.IsFlagSet(broker.FunctionFlags.Provisioning) {
			state.Logger.Debugf("Function [%s] can not be pushed at this time", current.Hash)
			return errorSkipPush
		}
	}

	if state.Output == "" {
		return errors.New("push called without a built image")
	}

	return nil
}

func handlePushComplete(params *PushParams, state *task.State) error {
	if state.Internal != nil && state.Output != "" {
		var remote = state.Internal.(*shared.Identity)
		var jsonAuth, _ = registry.EncodeAuthConfig(registry.AuthConfig{
			RegistryToken: remote.Auth,
		})
		var pushOptions = image.PushOptions{
			RegistryAuth: jsonAuth,
		}
		var rd io.ReadCloser
		var err error

		state.Logger.Debugf("Tagging docker image [%s]", remote.Path)

		if err = dockerClient.ImageTag(params.Context, state.Output, remote.Path); err == nil {
			state.Logger.Debugf("Pushing smart function id [%s]", remote.Id)

			if rd, err = dockerClient.ImagePush(params.Context, remote.Path, pushOptions); err == nil {
				var output PushStatus
				var scanner = bufio.NewScanner(rd)
				defer filez.CloseSilently(rd)

				for scanner.Scan() {
					if err = json.Unmarshal(scanner.Bytes(), &output); err == nil {
						if !strings.Contains(output.Status, remote.Version) {
							if output.ProgressDetail != nil {
								state.Logger.Debugf("%s: %s [%d:%d] %s", output.Id, output.Status,
									output.ProgressDetail.Current, output.ProgressDetail.Total, output.Progress)
							} else if output.Id != "" {
								state.Logger.Debugf("%s: %s", output.Id, output.Status)
							} else if output.Status != "" {
								state.Logger.Debugf("%s", output.Status)
							}
						}
					} else {
						state.Logger.Warningf("Could not parse json with error: %s", err)
						state.Logger.Debugf("String: %s", scanner.Text())
					}
				}

				if output.Aux != nil {
					remote.Hash = output.Aux.Digest
					state.Logger.Infof("Provisioned smart function v%s [%s], %s, size: %d", remote.Version, remote.Id, remote.Hash, output.Aux.Size)
				}
			}
		}

		state.Output = ""
		return err
	}

	return errors.New("no provisioned image to push")
}

func handlePushIncomplete(params *PushParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorSkipPush) && state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		if remote.Hash != "" {
			return nil
		}
	}

	return state.Error
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
