package locker

import (
	"bytes"
	"errors"
	"strings"

	"github.com/awnumar/memguard"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
)

var (
	errorLockerPassCapLetter   = task.NewError("the passphrase must contain at least 1 capital letter")
	errorLockerPassDigit       = task.NewError("the passphrase must contain at least 1 digit")
	errorLockerPassFailed      = task.NewError("the passphrase failed decryption")
	errorLockerPassInvalid     = task.NewError("the passphrase must be 8 characters long and contain capital, small letters, digits at least one special character")
	errorLockerPassShort       = task.NewError("the passphrase can not be empty")
	errorLockerPassSmallLetter = task.NewError("the passphrase must contain at least 1 small letter")
	errorLockerPassSpecial     = task.NewError("the passphrase must contain at least 1 special character")
	errorLockerPassUnchanged   = task.NewError("the passphrase must be different")
	errorLockerPathInvalid     = task.NewError("the locker path is not valid")
	errorLockerReadNeeded      = task.NewError("can not access current locker for update")
	errorLockerUpdateNeeded    = task.NewError("the locker will be updated with a new passphrase")
)

type InitParams struct {
	LockerPath    string
	OldPassphrase Enclave
	Passphrase    Enclave
	Update        bool
}

func (ip InitParams) Validate() error {
	if ip.Passphrase != nil {
		var lb *memguard.LockedBuffer
		var err error

		if lb, err = ip.Passphrase.Open(); err == nil {
			var b = lb.Bytes()

			defer lb.Destroy()
			if len(b) < 8 {
				return errorLockerPassShort
			}

			if !bytes.ContainsAny(b, "abcdefghijklmnopqrstuvwxyz") {
				return errorLockerPassSmallLetter
			}

			if !bytes.ContainsAny(b, "ABCDEFGHIJKLMNOPQRSTUVWXYZ") {
				return errorLockerPassCapLetter
			}

			if !bytes.ContainsAny(b, "0123456789") {
				return errorLockerPassDigit
			}

			if !bytes.ContainsAny(b, "!\"#$%&'()*+,-./:;<=>?@[\\]^_\\`{|}~ \t") {
				return errorLockerPassSpecial
			}

			if ip.OldPassphrase != nil {
				var olb *memguard.LockedBuffer

				if olb, err = ip.OldPassphrase.Open(); err == nil {
					defer olb.Destroy()

					if olb.EqualTo(b) {
						err = errorLockerPassUnchanged
					}
				}
			}
		}

		return err
	}

	return errorLockerPassInvalid
}

func NewInitLockerTask() *task.Task[InitParams] {
	return &task.Task[InitParams]{
		Name:         "locker-init",
		OnPrepare:    handleLockerInitContext,
		OnComplete:   handleLockerInitComplete,
		OnIncomplete: handleLockerInitUpdate,
		OnPretend:    handleLockerInitPretend,
	}
}

func handleLockerInitComplete(params *InitParams, state *task.State) error {
	if state.Output != "" {
		var lockerState = NewSecuredLockerState(state)
		var err error

		defer lockerState.Destroy()
		defer func() {
			if err == nil {
				if err = lockerState.Close(params.LockerPath); err != nil {
					state.Logger.Errorf("Could not rename locker file [%s] to [%s]", state.Output, params.LockerPath)
				}
			}
		}()

		state.Logger.Debugf("Writing locker file [%s]", state.Output)

		if err = lockerState.Write(state.Output, params.Passphrase); err == nil {
			if params.Update {
				state.Reportf("Updated locker file %s", params.LockerPath)
			} else {
				state.Reportf("Initialized locker file %s", params.LockerPath)
			}

			return nil
		}

		return err
	}

	return errorLockerPathInvalid
}

func handleLockerInitContext(params *InitParams, state *task.State) error {
	if state.Output == "" {
		var err error

		if err = params.Validate(); err == nil {
			if err = filez.IsReadable(params.LockerPath); errorz.IsPathError(err) {
				state.Logger.Debugf("Creating new locker file [%s]", params.LockerPath)
				state.Output = params.LockerPath
				return nil
			}

			if err == nil {
				// It should be overwritten or updated with a new passphrase
				state.Output = params.LockerPath + ".new"

				if params.Update {
					if params.OldPassphrase == nil {
						return errorLockerReadNeeded
					}

					return errorLockerUpdateNeeded
				}
			}
		}

		return err
	}

	return nil
}

func handleLockerInitPretend(params *InitParams, state *task.State) error {
	if state.Output != "" {
		var using = "using passphrase"

		state.Logger.Debugf("Pretending to initialize a locker file")

		if errors.Is(state.Error, errorLockerUpdateNeeded) {
			state.Logger.Debugf("Reading locker file [%s] using current passphrase", params.LockerPath)
			using = "using new passphrase"
		}

		state.Logger.Debugf("Writing locker file [%s] %s", state.Output, using)

		if params.LockerPath != state.Output {
			state.Logger.Debugf("Moving written file [%s] to [%s]", state.Output, params.LockerPath)
		}

		state.Logger.Debugf("Cleaning up locked buffers")
		return nil
	}

	return errorLockerPathInvalid
}

func handleLockerInitUpdate(params *InitParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, errorLockerUpdateNeeded) {
		var lockerState = NewSecuredLockerState(state)
		var err error

		state.Logger.Debugf("Updating current locker file [%s]", params.LockerPath)

		if err = lockerState.Read(params.LockerPath, params.OldPassphrase); err == nil {
			state.Logger.Debugf("Read locker file [%s]", params.LockerPath)
			state.Completed = false
			return nil
		}

		if strings.HasPrefix(err.Error(), "chacha20poly1305") {
			return errorLockerPassFailed
		}

		return err
	}

	return state.Error
}
