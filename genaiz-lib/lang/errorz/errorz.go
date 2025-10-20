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

func IsPathError(err error) bool {
	if err != nil {
		var pathError *os.PathError

		if errors.As(err, &pathError) ||
			errors.Is(err, LocalPathError) {
			return true
		}
	}

	return false
}
