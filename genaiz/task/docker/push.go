package docker

import (
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strings"

	"github.com/moby/moby/api/types/registry"
	"github.com/moby/moby/client"
	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/ioz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	PushTasks struct {
		PushTaskFactory PushTaskFactory // PushTaskFactory builds a task which can be used with an instance of PushParams to initiate the push of an image from dockerd to a registry
	}
)

type PushTask = *task.Task[PushParams]
type PushTaskFactory = func() PushTask

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

func NewPushStatusAux(plain string) *PushStatusAux {
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
}

func encodeAuthConfig(authConfig *registry.AuthConfig) (string, error) {
	buf, _ := json.Marshal(authConfig)
	return base64.URLEncoding.EncodeToString(buf), nil
}

func handlePushContext(params *PushParams, state *task.State) error {
	_ = params

	if state.Output == "" {
		return Errors.PushImageNoBuild()
	}

	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.Auth == "" {
			if current.IsFlagSet(broker.FunctionFlags.Active) {
				state.Logger.Debugf("Function [%s] is already provisioned", current.Path)
				return Errors.PushImageNeutral()
			}

			return Errors.PushImageInactive()
		}

		if !current.IsFlagSet(broker.FunctionFlags.Provisioning) {
			state.Logger.Debugf("Function [%s] can not be pushed at this time", current.Path)
			return Errors.PushImageConflict()
		}

		return nil
	}

	return Errors.PushImageNoProvision()
}

func handlePushComplete(params *PushParams, state *task.State) error {
	if state.Internal != nil && state.Output != "" {
		var dockerClient = dockerFactory.Get()
		var remote = state.Internal.(*shared.Identity)
		var jsonAuth, _ = encodeAuthConfig(&registry.AuthConfig{
			RegistryToken: remote.Auth,
		})
		var pushOptions = client.ImagePushOptions{
			RegistryAuth: jsonAuth,
		}
		var tagOptions = client.ImageTagOptions{
			Source: state.Output,
			Target: remote.Path,
		}
		var rd io.ReadCloser
		var err error

		state.Logger.Debugf("Tagging docker image [%s]", remote.Path)

		if _, err = dockerClient.ImageTag(params.Context, tagOptions); err == nil {
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
								return NewPushStatusAux(output.Status), nil
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
					err = Errors.PushImageUnrecognizedResponse()
				}
			}
		}

		state.Output = ""
		return task.NewFailure(err)
	}

	return Errors.PushImageNoProvision()
}

func handlePushIncomplete(params *PushParams, state *task.State) error {
	_ = params
	state.Completed = true

	if errors.Is(state.Error, Errors.PushImageNeutral()) && state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		if remote.Hash == "" {
			return Errors.PushImageUnknownRepoDigest()
		}

		return nil
	}

	return state.Error
}

func handlePushPretend(params *PushParams, state *task.State) error {
	_ = params

	if state.Internal != nil {
		var remote = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Pretending to tag docker image [%s]", state.Output)
		fmt.Printf("docker tag %s %s\n", remote.Hash, remote.Path)
		state.Logger.Debugf("Pretending to push docker image [%s]", remote.Hash)
		fmt.Printf("docker push %s\n", remote.Path)
		state.Output = ""
		return nil
	}

	return Errors.PushImageNoProvision()
}

func init() {
	PushTasks.PushTaskFactory = newPushTask
}

func newPushTask() *task.Task[PushParams] {
	return &task.Task[PushParams]{
		Name:         "docker-push",
		OnPrepare:    handlePushContext,
		OnComplete:   handlePushComplete,
		OnIncomplete: handlePushIncomplete,
		OnPretend:    handlePushPretend,
	}
}
