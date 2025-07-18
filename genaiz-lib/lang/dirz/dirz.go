package dirz

import (
	"os"
	"path/filepath"
	"strings"
)

// AnchorWorkingFile modifies a path if it starts with . or .., anchoring the rest of the path to the current work dir
func AnchorWorkingFile(path string) string {
	var cwd, _ = os.Getwd()
	var dotIndex = strings.Index(path, ".")
	var nonRelative string

	if dotIndex == 0 {
		var parentIndex = strings.Index(path, "..")

		if parentIndex == 0 {
			nonRelative = path[2:]
		} else {
			nonRelative = path[1:]
		}
	} else {
		nonRelative = path
	}

	if strings.HasSuffix(cwd, filepath.Dir(nonRelative)) {
		return filepath.Base(path)
	}

	return path
}

// CreateWorkingDir creates a working directory, if ir doesn't exist, changes the context working dir and returns a reset function to reposition the context to its original path
func CreateWorkingDir(args ...string) (func(), error) {
	var reset func()
	var err error

	if len(args) > 0 && args[0] != "." {
		if err = os.MkdirAll(args[0], 0750); err == nil {
			var cwd, _ = os.Getwd()

			if err = os.Chdir(args[0]); err == nil {
				reset = func() {
					_ = os.Chdir(cwd)
				}
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
