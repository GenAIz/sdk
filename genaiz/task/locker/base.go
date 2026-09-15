package locker

import (
	"errors"
	"fmt"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorLockerDataLinkEmpty   = task.NewError("datalink property set is empty, is it incomplete?")
	errorLockerDataLinkInvalid = task.NewError("locker data link invalid")
	errorLockerPassFailed      = task.NewError("the passphrase failed decryption")
)

type BaseParams struct {
	LockerHandle string
	LockerPath   string
	Passphrase   Enclave
}

type LinkParams struct {
	Oem     string
	Handle  string
	Version string
}

type PropertyParams struct {
	Key    string
	Value  string
	Secret Enclave
}

func handleBaseAddContext(params *BaseParams, state *task.State) error {
	if state.Output == "" {
		var varSpecState = shared.NewVarSpecState(state)
		var err error

		if len(varSpecState.VarSpecs) == 0 {
			// We expect to have all the data on the dataLink we need collected or exported and to
			// have specs available for the add to be valid
			return errorLockerDataLinkEmpty
		}

		if err = filez.IsReadable(params.LockerPath); errorz.IsPathError(err) {
			return fmt.Errorf("locker [%s] can not be read", params.LockerPath)
		}

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Output = params.LockerPath
	}

	return nil
}

func handleBaseFindContext(params *BaseParams, linkParams *broker.DataLinkParams, state *task.State) error {
	if state.Output == "" {
		state.Logger.Debugf("Need datalink fqdn for handle [%s]", params.LockerHandle)

		if linkParams != nil && linkParams.IsValid() {
			state.Output = linkParams.ToPublished()
			return errorLockerDataLinkFound
		}

		if params.Passphrase == nil {
			return errorLockerPassFailed
		}

		state.Output = params.LockerHandle
	}

	return nil
}

func handleBaseFindIncomplete(params *broker.DataLinkParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorLockerDataLinkFound) {
		state.Reportf("Already know datalink [%s]", params.ToPublished())
		return nil
	}

	return state.Error
}

func handleBaseSyncContext(params *broker.DataLinkParams, state *task.State) error {
	if state.Output == "" {
		if params == nil || !params.IsValid() {
			return errorLockerDataLinkInvalid
		}

		state.Output = params.ToPublished()
	}

	return nil
}
