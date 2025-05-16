// Package filez is meant to wrap common file
package filez

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"genaiz.com/genaiz/lang/panicz"
)

func AbsOrPanic(path string) string {
	var info, err = filepath.Abs(path)

	panicz.PanicIfError(err)
	return info
}

func RelativeIfWithinWorkDir(path string) string {
	var cwd, err = os.Getwd()

	if err == nil {
		if i := strings.Index(path, cwd); i == 0 {
			return "./" + path[len(cwd)+1:]
		}
	}

	return path
}

func DoIfPathExist(path string, todo func() error) error {
	if _, err := os.Stat(path); err == nil {
		return todo()
	} else if !errors.Is(err, os.ErrNotExist) {
		return err
	}

	return nil
}
