package errorz

import (
	"errors"
	"fmt"
	"io"
	"os"
)

var (
	LocalPathError = errors.New("path not found")
)

func deferOnExit(err *error, errStream io.Writer, fn func()) func() {
	return func() {
		if fn != nil {
			fn()
		}

		if err != nil && *err != nil {
			_, _ = fmt.Fprintf(errStream, "Error: %s\n", *err)
			os.Exit(1)
		}
	}
}

func DeferOnExit(err *error, fn func()) func() {
	return deferOnExit(err, os.Stderr, fn)
}

// IsPathError indicates whether an error is of type "file not found", which in certain cases is not an error. The method will return false for Permission Denied, which most of the time are user errors.
func IsPathError(err error) bool {
	if err != nil {
		var pathError *os.PathError

		if errors.As(err, &pathError) {
			return !os.IsPermission(err)
		} else if errors.Is(err, LocalPathError) {
			return true
		}
	}

	return false
}
