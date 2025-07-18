// Package filez is meant to wrap common file
package filez

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"genaiz.com/genaiz-lib/lang/panicz"
)

// CloseSilently ignores any error on calling Close on an os.File
func CloseSilently(toClose io.ReadCloser) {
	if toClose != nil {
		_ = toClose.Close()
	}
}

// CreateRecursive creates a file under the specified dir, creating the path first if necessary
func CreateRecursive(dir string, name string) (*os.File, error) {
	panicz.PanicIfError(os.MkdirAll(dir, 0750))
	return os.Create(filepath.Join(dir, name))
}

// CreateRecursiveTemp creates a temporary file under the specified dir, creating the path if necessary
func CreateRecursiveTemp(dir string, pattern string) (*os.File, error) {
	panicz.PanicIfError(os.MkdirAll(dir, 0750))
	return os.CreateTemp(dir, pattern)
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

// GetFileType returns the non-opinionated extension of the provided filename, that's excluding the leading dot character, taking everything afterward.
func GetFileType(path string) string {
	var filename = filepath.Base(path)
	var dotIndex = strings.Index(filename, ".")

	if dotIndex > 0 {
		return filename[dotIndex+1:]
	}

	return ""
}

func IsReadable(file string) error {
	var path = file
	var info os.FileInfo
	var err error

	if info, err = os.Stat(path); err == nil {
		if info.Mode().Perm()&0400 != 0400 {
			return fmt.Errorf("%s can not be read", path)
		}
	}

	return err
}

// RemoveSilently performs a rm -rf of the provided dir
func RemoveSilently(dir string) {
	_ = os.RemoveAll(dir)
}
