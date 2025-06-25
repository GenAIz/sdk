package broker

import (
	"errors"
	"fmt"
	"strings"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorNoBuildToProvision  = errors.New("could not find an image to provision, call build first")
	errorNoProvision         = errors.New("can not publish non-provisioned functions")
	errorNoProvisionIdentity = errors.New("could not identify provisioned image")
)

type ProvisionParams struct {
	Broker
	Arches      []string
	Description string
	Fqdn        string
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
		Fqdn:        pp.Fqdn,
		Handle:      pp.Handle,
		Name:        pp.Name,
		Oem:         pp.Oem,
		Type:        strings.ToUpper(pp.Type),
		Version:     pp.Version,
	}
}

func NewProvisionTask() *task.Task[ProvisionParams] {
	return &task.Task[ProvisionParams]{
		Name:       "broker-provision",
		OnPrepare:  handleProvisionContext,
		OnComplete: handleProvisionComplete,
		OnPretend:  handleProvisionPretend,
	}
}

func NewPublishTask() *task.Task[ProvisionParams] {
	return &task.Task[ProvisionParams]{
		Name:       "broker-publish",
		OnPrepare:  handlePublishContext,
		OnComplete: handlePublishComplete,
		OnPretend:  handlePublishPretend,
	}
}

func handleProvisionContext(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)
		var client *Client
		var err error

		if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
			var sf *Function

			state.Logger.Debugf("Finding existing provisions for [%s/%s]", params.Oem, params.Handle)

			if sf, err = client.GetFunction(params.Oem, params.Handle); err == nil {
				state.Internal, err = sf.asIdentity().Next(current)
			} else if errors.Is(err, errorNotFound) {
				err = nil
			}
		}

		return err
	}

	return errorNoBuildToProvision
}

func handleProvisionComplete(params *ProvisionParams, state *task.State) error {
	var client *Client
	var err error

	if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
		state.Logger.Debugf("Provisioning function on url [%s]", client.provisionUrl())
		state.Internal, err = client.ProvisionFunction(params.asFunction())
	}

	return err
}

func handleProvisionPretend(params *ProvisionParams, state *task.State) error {
	if state.Error == nil {
		var client *Client
		var err error

		if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
			var remote = state.Internal.(*shared.Identity)

			state.Logger.Debugf("Pretending to provision to [%s]", params.HostAddr)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", client.AuthToken)
			fmt.Printf("  -d arches=%s\\\n", params.Arches)
			fmt.Printf("  -d name=%s\\\n", params.Name)
			fmt.Printf("  -d description=%s\\\n", params.Description)
			fmt.Printf("  -d fqdn=%s\\\n", params.Fqdn)
			fmt.Printf("  -d oem=%s\\\n", params.Oem)
			fmt.Printf("  -d handle=%s\\\n", params.Handle)
			fmt.Printf("  -d version=%s\\\n", params.Version)
			fmt.Printf("  -d type=%s\\\n", params.Type)
			fmt.Printf("%s\n", client.provisionUrl())
			remote.Id = "$ID"
			remote.Path = params.HostAddr + "/" + params.Handle
			return nil
		}

		return err
	}

	state.Logger.Errorf("Provisioning failed with error: [%s]", state.Error)
	return state.Error
}

func handlePublishContext(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)
		var client *Client

		if current.HasIdentifiers() {
			var err error

			if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
				var published *Function

				if published, err = client.GetFunctionById(current.Id); err == nil {
					if published.Digest == current.Hash {
						return nil
					}

					err = fmt.Errorf("provisioning conflicts:\npublished: [%s]\nlocal: [%s]", published.Digest, current.Hash)
				}
			}

			return err
		}

		return errorNoProvisionIdentity
	}

	return errorNoProvision
}

func handlePublishComplete(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasIdentifiers() {
			var client *Client
			var err error

			if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
				_, err = client.PublishFunction(current)
				state.Output = ""
			}

			return err
		}
	}

	return errorNoProvision
}

func handlePublishPretend(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasIdentifiers() {
			var client *Client
			var err error

			if client, err = GetClient(params.AuthFile, params.HostAddr); err == nil {
				state.Logger.Debugf("Pretending to provision to [%s]", params.HostAddr)
				fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
				fmt.Printf("  --cookie=\"s=%s\"\\\n", client.AuthToken)
				fmt.Printf("  -d id=%s\\\n", current.Id)
				fmt.Printf("%s\n", client.publishUrl())
				return nil
			}

			return err
		}
	}

	return errorNoProvision
}
