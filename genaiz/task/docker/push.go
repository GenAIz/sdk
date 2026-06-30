package docker

import (
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/api/types/registry"
	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/ioz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorConflictPush       = errors.New("smart function can not be provisioned")
	errorIllegalPush        = errors.New("can not push function, because of inactive state")
	errImproperPushResponse = errors.New("docker push failed without a proper response")
	errorNeutralPush        = errors.New("smart function digest is already known")
	errorNoBuild            = errors.New("smart function push called without an image built")
	errorNoProvision        = errors.New("smart function push called without a provision")
	errorSynchronizedPush   = errors.New("smart function repository digest is unknown")
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

func NewPushStatuAux(plain string) *PushStatusAux {
	var parts = strings.Split(plain, " ")

	return &PushStatusAux{
		Tag:    strings.ReplaceAll(parts[0], ":", ""),
		Digest: parts[2],
		Size:   cast.ToInt64(parts[4]),
	}
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
	_ = params

	if state.Output == "" {
		return errorNoBuild
	}

	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.Auth == "" {
			if current.IsFlagSet(broker.FunctionFlags.Active) {
				state.Logger.Debugf("Function [%s] is already provisioned", current.Path)
				return errorNeutralPush
			}

			return errorIllegalPush
		}

		if !current.IsFlagSet(broker.FunctionFlags.Provisioning) {
			state.Logger.Debugf("Function [%s] can not be pushed at this time", current.Path)
			return errorConflictPush
		}

		return nil
	}

	return errorNoProvision
}

func handlePushComplete(params *PushParams, state *task.State) error {
	if state.Internal != nil && state.Output != "" {
		var dockerClient = dockerFactory.Get()
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
				var aux *PushStatusAux
				var probe = ioz.NewProber[PushStatusAux](func(data []byte) (*PushStatusAux, error) {
					var output PushStatus

					state.Logger.Tracef("<- %s", string(data))

					if err = json.Unmarshal(data, &output); err == nil {
						if strings.Contains(output.Status, remote.Version) {
							// Quirks-mode where a request is producing push result as a plain string in the status field
							if strings.Contains(output.Status, "digest") {
								state.Logger.Debugf("Quirks status payload: [%s]", output.Status)
								return NewPushStatuAux(output.Status), nil
							}

							state.Logger.Warningf("Quirks status did not provide a digest")
						} else {
							if output.ProgressDetail != nil && output.Id != "" {
								state.Logger.Debugf("%s: %s [%d:%d] %s", output.Id, output.Status,
									output.ProgressDetail.Current, output.ProgressDetail.Total, output.Progress)
							} else if output.Id != "" {
								state.Logger.Debugf("%s: %s", output.Id, output.Status)
							} else if output.Status != "" {
								state.Logger.Debugf("%s", output.Status)
							} else if output.Aux == nil {
								state.Logger.Debugf("Unknown payload: [%s]", string(data))
							}
						}
					} else {
						state.Logger.Warningf("Could not parse json with error: %s", err)
						state.Logger.Debugf("String: %s", string(data))
						return nil, err
					}

					return output.Aux, nil
				})

				defer filez.CloseSilently(rd)

				if aux, err = probe.Until(rd); err == nil {
					remote.Hash = aux.Digest
					state.Reportf("Pushed smart function v%s [%s], %s, size: %d", remote.Version, remote.Id, remote.Hash, aux.Size)
					return nil
				} else if errors.Is(err, ioz.ErrorUntilNotFound) {
					err = errImproperPushResponse
				}
			}
		}

		state.Output = ""
		return task.NewFailure(err)
	}

	return errorNoProvision
}

func handlePushIncomplete(params *PushParams, state *task.State) error {
	_ = params
	state.Completed = true

	if errors.Is(state.Error, errorNeutralPush) && state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		if remote.Hash == "" {
			return errorSynchronizedPush
		}

		return nil
	}

	return state.Error
}

func handlePushPretend(params *PushParams, state *task.State) error {
	if state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Pretending to tag docker image [%s]", params.DockerRepository)
		fmt.Printf("docker tag %s %s\n", remote.Hash, remote.Path)
		state.Logger.Debugf("Pretending to push docker image [%s]", remote.Hash)
		fmt.Printf("docker push %s\n", remote.Path)
		state.Output = ""
		return nil
	}

	return errorNoProvision
}
