package dirz

import (
	"os"
	"path/filepath"
	"strings"

	"genaiz.com/genaiz-lib/lang/panicz"
)

// AnchorWorkingFile modifies a path if it starts with . or .., anchoring the rest of the path to the current work dir
func AnchorWorkingFile(path string) string {
	var cwd, _ = os.Getwd()
	var dotIndex = strings.Index(path, ".")
	var nonRelative string

	if dotIndex == 0 {
		var parentIndex = strings.Index(path, "..")

		if parentIndex == 0 {
			nonRelative = path[3:]
		} else {
			nonRelative = path[2:]
		}
	} else {
		nonRelative = path
	}

	if strings.HasSuffix(cwd, filepath.Dir(nonRelative)) {
		return filepath.Base(path)
	}

	return nonRelative
}

// ChangeWorkingDir changes a working directory if it exists, returning a reset function to reposition the context to its original path
func ChangeWorkingDir(args ...string) (func(), error) {
	var reset = func() {}
	var err error

	if len(args) > 0 && args[0] != "." {
		var cwd string

		cwd, err = os.Getwd()
		panicz.PanicIfError(err)

		if err = os.Chdir(args[0]); err == nil {
			reset = func() {
				_ = os.Chdir(cwd)
			}
		}
	}

	return reset, err
}

// CreateWorkingDir creates a working directory, if it doesn't exist, changes the context working dir and returns a reset function to reposition the context to its original path
func CreateWorkingDir(args ...string) (func(), error) {
	var reset = func() {}
	var err error

	if len(args) > 0 && args[0] != "." {
		if err = os.MkdirAll(args[0], 0750); err == nil {
			var cwd string

			cwd, err = os.Getwd()
			panicz.PanicIfError(err)
			panicz.PanicIfError(os.Chdir(args[0]))
			reset = func() {
				_ = os.Chdir(cwd)
			}
		}
	}

	return reset, err
}

// DoIfPathExist will invoke the provided call if the path provided exist
func DoIfPathExist(path string, call func() error) error {
	if stat, _ := os.Stat(path); stat != nil {
		return call()
	}

	return nil
}

// OptionalWorkingDir returns a provider returning the absolute path of the provided arguments or the current working dir if no arguments are provided
func OptionalWorkingDir(args ...string) func() (string, error) {
	var result func() (string, error)

	if len(args) > 0 {
		result = func() (string, error) {
			return filepath.Abs(filepath.Join(args...))
		}
	} else {
		result = os.Getwd
	}

	return result
}
