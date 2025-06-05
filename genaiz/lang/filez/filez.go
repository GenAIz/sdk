// Package filez is meant to wrap common file
package filez

import (
	"os"
	"path/filepath"
	"strings"
)

// CloseSilently ignores any error on calling Close on an os.File
func CloseSilently(file *os.File) {
	if file != nil {
		_ = file.Close()
	}
}

// DoIfPathExist will invoke the provided call if the path provided exist
func DoIfPathExist(path string, call func() error) error {
	if stat, _ := os.Stat(path); stat != nil {
		return call()
	}

	return nil
}

// FirstNamedFile returns the first file with the provided name under the current work dir non-recursively, that is the part excluding an extension matching
func FirstNamedFile(name string) (string, error) {
	var dir, _ = os.Getwd()

	return FirstNamedFileUnder(dir, name)
}

// FirstNamedFileUnder returns the first file with the provided name under the provided path non-recursively, that is the part excluding an extension matching
func FirstNamedFileUnder(path string, name string) (string, error) {
	var entries, err = os.ReadDir(path)

	if err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				var file = filepath.Base(entry.Name())
				var i = strings.LastIndex(file, ".")

				if i > 0 && strings.EqualFold(name, file[0:i]) {
					return file, nil
				}
			}
		}
	}

	return "", err
}

// FromWorkDir will make a path relative starting with './' if it shares the same parent as the work dir
func FromWorkDir(path string) string {
	var cwd, _ = os.Getwd()

	if i := strings.Index(path, cwd); i == 0 {
		return "./" + path[len(cwd)+1:]
	}

	return path
}
