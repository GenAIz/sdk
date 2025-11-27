package mock

import (
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/spf13/cast"
	"github.com/undefinedlabs/go-mpatch"

	"genaiz.com/genaiz-lib/browser"
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

func (p Patches) BrowserOpenUrl(impl func(url string, errOut io.Writer, out io.Writer) error) *Patched {
	var patchedOpenUrl = &Patched{Called: false}
	var patchFunc, err = mpatch.PatchMethod(browser.OpenUrl, func(url string, errOut io.Writer, out io.Writer) error {
		patchedOpenUrl.Called = true
		patchedOpenUrl.CalledWith = []string{url}
		return impl(url, errOut, out)
	})

	if err != nil {
		p.T.Errorf("Failed to patch browser.OpenUrl due to an error: %v", err)
		return nil
	}

	patchedOpenUrl.PatchFunc = patchFunc
	return patchedOpenUrl
}

func (p Patches) FmtPrintf(impl func(format string, a ...any)) *Patched {
	var patchedPrintf = &Patched{Called: false}
	var patchFunc, err = mpatch.PatchMethod(fmt.Printf, func(format string, a ...any) (int, error) {
		var calledWith = []string{format}
		var result = 0

		calledWith = append(calledWith, cast.ToStringSlice(a)...)

		for _, token := range calledWith {
			result += len(token)
		}

		patchedPrintf.Called = true
		patchedPrintf.CalledWith = calledWith
		impl(format, a...)
		return result, nil
	})

	if err != nil {
		p.T.Errorf("Failed to patch fmt.Printf due to an error: %v", err)
		return nil
	}

	patchedPrintf.PatchFunc = patchFunc
	return patchedPrintf
}

func (p Patches) OsExit(impl func(int)) *Patched {
	var patchedExit = &Patched{Called: false}
	var patchFunc, err = mpatch.PatchMethod(os.Exit, func(code int) {
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

func (p Patches) OsGetgid(gid int) *Patched {
	var patchedExit = &Patched{Called: false}
	var patchFunc, err = mpatch.PatchMethod(os.Getgid, func() int {
		patchedExit.Called = true
		return gid
	})

	if err != nil {
		p.T.Errorf("Failed to patch os.Getgid due to an error: %v", err)
		return nil
	}

	patchedExit.PatchFunc = patchFunc
	return patchedExit
}

func (p Patches) OsGetuid(uid int) *Patched {
	var patchedExit = &Patched{Called: false}
	var patchFunc, err = mpatch.PatchMethod(os.Getuid, func() int {
		patchedExit.Called = true
		return uid
	})

	if err != nil {
		p.T.Errorf("Failed to patch os.Getuid due to an error: %v", err)
		return nil
	}

	patchedExit.PatchFunc = patchFunc
	return patchedExit
}
