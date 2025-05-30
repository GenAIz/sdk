package mock

import (
	"os"
	"testing"

	"github.com/undefinedlabs/go-mpatch"
)

type Patched struct {
	Called     bool
	CalledWith any
	PatchFunc  *mpatch.Patch
}

func (p *Patched) Unpatch() {
	_ = p.PatchFunc.Unpatch()
}

type Patches struct {
	T *testing.T
}

func (p Patches) OsExit(impl func(int)) *Patched {
	patchedExit := &Patched{Called: false}

	patchFunc, err := mpatch.PatchMethod(os.Exit, func(code int) {
		patchedExit.Called = true
		patchedExit.CalledWith = code
		impl(code)
	})

	if err != nil {
		p.T.Errorf("Failed to patch os.Exit due to an error: %v", err)
		return nil
	}

	patchedExit.PatchFunc = patchFunc
	return patchedExit
}
