// Package filez is meant to wrap common file
package filez

import (
	"io"
	"os"
	"path/filepath"
	"strings"

	"genaiz.com/genaiz-lib/lang/errorz"
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
	var err error

	if err = os.MkdirAll(dir, 0750); err == nil {
		return os.OpenFile(filepath.Join(dir, name), os.O_RDWR|os.O_TRUNC|os.O_CREATE, 0660)
	}

	return nil, err
}

// CreateRecursiveTemp creates a temporary file under the specified dir, creating the path if necessary
func CreateRecursiveTemp(dir string, pattern string) (*os.File, error) {
	panicz.PanicIfError(os.MkdirAll(dir, 0750))
	return os.CreateTemp(dir, pattern)
}

// FindNamedFilesRecursively inspects a paths and all its subfolders attempting to find files with the provided name. That is the portion before the extension starting with the first "." encountered in the filename.
func FindNamedFilesRecursively(path string, name string) ([]string, error) {
	var result []string
	var entries []os.DirEntry
	var err error

	if entries, err = os.ReadDir(path); err == nil {
		for _, entry := range entries {
			var entryPath = filepath.Join(path, entry.Name())

			if entry.IsDir() {
				var subResults []string

				if subResults, err = FindNamedFilesRecursively(entryPath, name); err == nil {
					result = append(result, subResults...)
				} else {
					return nil, err
				}
			} else if strings.EqualFold(name, GetFileName(entryPath)) {
				result = append(result, entryPath)
			}
		}

		return result, nil
	}

	return nil, err
}

// FirstNamedFile returns the first file with the provided name under the current work dir non-recursively, that is the part excluding an extension matching
func FirstNamedFile(name string) (string, error) {
	var dir, _ = os.Getwd()

	return FirstNamedFileUnder(dir, name)
}

// FirstNamedFileUnder returns the first file with the provided name under the provided path non-recursively, that is the part excluding an extension matching
func FirstNamedFileUnder(path string, name string) (string, error) {
	var entries []os.DirEntry
	var err error

	if entries, err = os.ReadDir(path); err == nil {
		for _, entry := range entries {
			if !entry.IsDir() {
				var file = filepath.Base(entry.Name())
				var i = strings.LastIndex(file, ".")

				if i > 0 && strings.EqualFold(name, file[0:i]) {
					return file, nil
				}
			}
		}

		return "", errorz.LocalPathError
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

// GetFileName extracts the name of the file after its path before its extension. Hidden files are returned as-is, as well as files without any extensions
func GetFileName(path string) string {
	var file = filepath.Base(path)

	if i := strings.Index(file, "."); i > 0 {
		return file[0:i]
	}

	return file
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

// IsReadable returns an error if the file path provided can not be read
func IsReadable(file string) error {
	var fd *os.File
	var err error

	// That's the only way to test readability in Go, os.Stat doesn't cover ownership
	if fd, err = os.OpenFile(file, os.O_RDONLY, 666); fd != nil {
		return fd.Close()
	}

	return err
}

// RemoveSilently performs a rm -rf of the provided dir
func RemoveSilently(dir string) {
	_ = os.RemoveAll(dir)
}
