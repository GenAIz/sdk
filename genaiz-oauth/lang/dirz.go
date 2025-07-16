package lang

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

func Chdir(args ...string) (func(), error) {
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

func LocalDir(file string) string {
	var cwd, _ = os.Getwd()
	var dotIndex = strings.Index(file, ".")
	var nonRelative string

	if dotIndex == 0 {
		var parentIndex = strings.Index(file, "..")

		if parentIndex == 0 {
			nonRelative = file[2:]
		} else {
			nonRelative = file[1:]
		}
	} else {
		nonRelative = file
	}

	if strings.HasSuffix(cwd, filepath.Dir(nonRelative)) {
		return filepath.Base(file)
	}

	return file
}

func IsReadable(file string) error {
	var path = LocalDir(file)
	var info os.FileInfo
	var err error

	if info, err = os.Stat(path); err == nil {
		if info.Mode().Perm()&0400 != 0400 {
			return fmt.Errorf("%s can not be read", path)
		}
	}

	return err
}
