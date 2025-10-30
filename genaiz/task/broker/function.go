package broker

import (
	"errors"
	"fmt"
	"strings"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorDuplicatePublishing = errors.New("the smart function was already published")
	errorNoBuildToProvision  = errors.New("could not find an image to provision, call build first")
	errorNoProvision         = errors.New("can not publish non-provisioned functions")
	errorNoRepoIdentity      = errors.New("could not identify repository hash")
	errorNoRepoProvisioning  = errors.New("can not publish without provisioning rights")
)

type ProvisionParams struct {
	Broker
	Arches      []string
	Description string
	Extras      map[string]any
	Handle      string
	Name        string
	Oem         string
	Type        string
	Version     string
}

func (pp ProvisionParams) getExtras() map[string]any {
	if pp.Extras == nil {
		pp.Extras = make(map[string]any)
	}

	return pp.Extras
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

type PublishParams struct {
	Broker
	Handle      string
	Oem         string
	SkipUnknown bool
}

func NewFunctionProvisionTask() *task.Task[ProvisionParams] {
	return &task.Task[ProvisionParams]{
		Name:       "function-provision",
		OnPrepare:  handleFunctionProvisionContext,
		OnComplete: handleFunctionProvisionComplete,
		OnPretend:  handleFunctionProvisionPretend,
	}
}

func NewFunctionPublishTask() *task.Task[PublishParams] {
	return &task.Task[PublishParams]{
		Name:         "function-publish",
		OnPrepare:    handleFunctionPublishContext,
		OnComplete:   handleFunctionPublishComplete,
		OnIncomplete: handleFunctionPublishIncomplete,
		OnPretend:    handleFunctionPublishPretend,
	}
}

func handleFunctionProvisionContext(params *ProvisionParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if !strings.HasPrefix(current.Id, "sha256") {
			return errors.New("can not provision smart function without a known image digest")
		}

		if params.Oem != "" && params.Handle != "" {
			state.Logger.Debugf("Validated smart function provisioning for [%s/%s]", params.Oem, params.Handle)

			if current.HasRepoIdentifier() {
				state.Logger.Debugf("Local function [%s/%s] found with remote hash [%s]", params.Oem, params.Handle, current.Hash)
			}

			return nil
		}

		return errors.New("unknown provisioning oem and/or handle")
	}

	return errorNoBuildToProvision
}

func handleFunctionProvisionComplete(params *ProvisionParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var provisioned = params.asFunction()
		var current = state.Internal.(*shared.Identity)
		var identity *shared.Identity

		state.Logger.Debugf("Provisioning function on url [%s]", brokerClient.ProvisionFunctionUrl())
		provisioned.ImgDigest = current.Id
		state.Logger.Debugf("Provisioning image digest [%s]", provisioned.ImgDigest)

		if identity, err = brokerClient.ProvisionFunction(provisioned, params.getExtras()); err == nil {
			if identity.Hash != "" && identity.Hash != current.Hash {
				state.Logger.Debugf("Overwriting function [%s/%s]?", params.Oem, params.Handle)
				identity.Hash = ""
			} else if current.Hash != "" {
				// Edge case: Hashes out of sync, skip push, try publish
				identity.Hash = current.Hash
			}

			state.Internal = identity
			return nil
		}
	}

	return err
}

func handleFunctionProvisionPretend(params *ProvisionParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
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
			remote.Path = params.Oem + "/" + params.Handle
			return nil
		}

		return err
	}

	state.Logger.Errorf("Provisioning failed with error: [%s]", state.Error)
	return state.Error
}

func handleFunctionPublishContext(params *PublishParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		state.Logger.Debugf("Validating function publishing for [%s/%s]", params.Oem, params.Handle)

		if !current.IsFlagSet(FunctionFlags.Provisioning) {
			if current.Hash == "" {
				return errorNoRepoProvisioning
			}

			state.Logger.Debugf("Smart Function [%s/%s] is already published", params.Oem, params.Handle)
			return errorDuplicatePublishing
		}

		if current.HasRepoIdentifier() {
			return nil
		}
	}

	return errorNoRepoIdentity
}

func handleFunctionPublishComplete(params *PublishParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasRepoIdentifier() {
			var brokerClient Client
			var err error

			if brokerClient, err = params.GetClient(); err == nil {
				var fn *Function

				state.Logger.Debugf("Publishing smart function v%s [%s], %s", current.Version, current.Id, current.Hash)

				if fn, err = brokerClient.PublishFunction(current); err == nil {
					state.Report(fmt.Sprintf("Published smart function %s/%s, version %s to %s",
						fn.Oem, fn.Handle, fn.Version, brokerClient.GetHostAddr()))
				}

				state.Internal = nil
				state.Output = ""
			}

			return err
		}
	}

	return errorNoProvision
}

func handleFunctionPublishIncomplete(params *PublishParams, state *task.State) error {
	state.Completed = true

	if params.SkipUnknown {
		if errors.Is(state.Error, errorNoRepoIdentity) {
			state.Logger.Warnf("Function publish will be skipped because no repository hash is known")
		}

		if errors.Is(state.Error, errorNoRepoProvisioning) {
			state.Logger.Warnf("Function publish will be skipped because provisioning rights do not allow it")
		}

		state.Error = nil
		state.Internal = nil
		state.Output = ""
	} else if errors.Is(state.Error, errorDuplicatePublishing) {
		var current = state.Internal.(*shared.Identity)

		state.Report(fmt.Sprintf("Smart Function was found under path %s", current.Path))
		state.Abort = true
		state.Error = nil
		state.Internal = nil
		state.Output = ""
	}

	return state.Error
}

func handleFunctionPublishPretend(params *PublishParams, state *task.State) error {
	if state.Internal != nil {
		var current = state.Internal.(*shared.Identity)

		if current.HasIdentifier() {
			var brokerClient Client
			var err error

			if brokerClient, err = params.GetClient(); err == nil {
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
