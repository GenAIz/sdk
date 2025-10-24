package docker

import (
	"encoding/base64"
	"errors"
	"fmt"
	"os"

	"github.com/docker/docker/api/types/registry"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz/task"
)

type config struct {
	Auths map[string]*auth
}

type auth struct {
	Auth string
}

type LoginParams struct {
	task.Env
	AuthFile string
	Host     string
	Password *[]byte
	Username string
}

func (lp LoginParams) basicAuth() string {
	var basic = lp.Username + ":" + string(*lp.Password)

	return base64.StdEncoding.EncodeToString([]byte(basic))
}

// NewLoginTask constructs a task.Task with the objective of requesting that containerd logs in with a registry.
func NewLoginTask() *task.Task[LoginParams] {
	return &task.Task[LoginParams]{
		Name:       "docker-login",
		OnPrepare:  handleLoginContext,
		OnComplete: handleLoginComplete,
		OnPretend:  handleLoginPretend,
	}
}

func handleLoginContext(params *LoginParams, state *task.State) error {
	var configBytes []byte
	var err error

	if params.AuthFile != "" {
		if configBytes, err = os.ReadFile(params.AuthFile); err == nil {
			var dockerConfig config

			state.Logger.Debugf("Finding a docker login token for host [%s]", params.Host)

			if err = yaml.Unmarshal(configBytes, &dockerConfig); err == nil {
				for host, dockerAuth := range dockerConfig.Auths {
					if host == params.Host {
						state.Output = dockerAuth.Auth
						return nil
					}
				}
			}

			return err
		}

		if errors.Is(err, os.ErrNotExist) {
			return nil
		}
	}

	return err
}

func handleLoginComplete(params *LoginParams, state *task.State) error {
	if state.Output == "" {
		var dockerClient = dockerFactory.Get()
		var registryAuth = registry.AuthConfig{
			Username:      params.Username,
			Password:      string(*params.Password),
			ServerAddress: params.Host,
			// RegistryLogin will not encode username/password for you and docker registry most likely still wants them,
			// It's a mess: https://github.com/moby/moby/issues/42381
			IdentityToken: params.basicAuth(),
		}

		state.Logger.Debugf("Docker login on host [%s] with username [%s]", params.Host, params.Username)

		if resp, err := dockerClient.RegistryLogin(params.Context, registryAuth); err == nil {
			state.Output = resp.IdentityToken
		} else {
			return err
		}
	}

	return nil
}

func handleLoginPretend(params *LoginParams, state *task.State) error {
	if state.Output == "" {
		fmt.Printf("docker login %s --username %s --password XXXXXXXXX\n", params.Host, params.Username)
	}

	state.Logger.Debugf("Login token found")
	return nil
}
