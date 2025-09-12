package broker

import (
	"errors"
	"fmt"
	"strings"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorNoBuildToProvision = errors.New("could not find an image to provision, call build first")
	errorNoProvision        = errors.New("can not publish non-provisioned functions")
	errorNoRepoIdentity     = errors.New("could not identify repository hash")
)

type ProvisionParams struct {
	Broker
	Arches      []string
	Description string
	Handle      string
	Name        string
	Oem         string
	Type        string
	Version     string
}

func (pp ProvisionParams) asFunction() *Function {
	return &Function{
		Arches:      pp.Arches,
		Description: pp.Description,
		Handle:      pp.Handle,
		Name:        pp.Name,
		Oem:         pp.Oem,
		Type:        strings.ToUpper(pp.Type),
		Version:     pp.Version,
	}
}

func NewFunctionProvisionTask() *task.Task[ProvisionParams] {
	return &task.Task[ProvisionParams]{
		Name:       "function-provision",
		OnPrepare:  handleFunctionProvisionContext,
		OnComplete: handleFunctionProvisionComplete,
		OnPretend:  handleFunctionProvisionPretend,
	}
}

func NewFunctionPublishTask() *task.Task[ProvisionParams] {
	return &task.Task[ProvisionParams]{
		Name:       "function-publish",
		OnPrepare:  handleFunctionPublishContext,
		OnComplete: handleFunctionPublishComplete,
		OnPretend:  handleFunctionPublishPretend,
	}
}

func handleFunctionProvisionContext(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Validating smart function provisioning for [%s/%s]", params.Oem, params.Handle)

		if current.HasRepoIdentifier() {
			state.Logger.Debugf("Provisioning for [%s/%s] exists", params.Oem, params.Handle)
		}

		return nil
	}

	return errorNoBuildToProvision
}

func handleFunctionProvisionComplete(params *ProvisionParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
		state.Logger.Debugf("Provisioning function on url [%s]", brokerClient.ProvisionFunctionUrl())
		state.Internal, err = brokerClient.ProvisionFunction(params.asFunction())
	}

	return err
}

func handleFunctionProvisionPretend(params *ProvisionParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
			var remote = state.Internal.(*shared.Identity)

			state.Logger.Debugf("Pretending to provision to [%s]", params.HostAddr)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d arches=%s\\\n", params.Arches)
			fmt.Printf("  -d name=%s\\\n", params.Name)
			fmt.Printf("  -d description=%s\\\n", params.Description)
			fmt.Printf("  -d oem=%s\\\n", params.Oem)
			fmt.Printf("  -d handle=%s\\\n", params.Handle)
			fmt.Printf("  -d version=%s\\\n", params.Version)
			fmt.Printf("  -d type=%s\\\n", params.Type)
			fmt.Printf("%s\n", brokerClient.ProvisionFunctionUrl())
			remote.Id = "$ID"
			remote.Path = params.HostAddr + "/" + params.Handle
			return nil
		}

		return err
	}

	state.Logger.Errorf("Provisioning failed with error: [%s]", state.Error)
	return state.Error
}

func handleFunctionPublishContext(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Validating smart function publishing for [%s/%s]", params.Oem, params.Handle)

		if current.HasRepoIdentifier() {
			return nil
		}
	}

	return errorNoRepoIdentity
}

func handleFunctionPublishComplete(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasRepoIdentifier() {
			var brokerClient Client
			var err error

			if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
				_, err = brokerClient.PublishFunction(current)
				state.Output = ""
				state.Logger.Infof("Publish smart function v%s [%s], %s", current.Version, current.Id, current.Hash)
			}

			return err
		}
	}

	return errorNoProvision
}

func handleFunctionPublishPretend(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasIdentifier() {
			var brokerClient Client
			var err error

			if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
				state.Logger.Debugf("Pretending to provision to [%s]", params.HostAddr)
				fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
				fmt.Printf("  -d id=%s\\\n", current.Id)
				fmt.Printf("%s\n", brokerClient.PublishFunctionUrl())
				return nil
			}

			return err
		}
	}

	return errorNoProvision
}
