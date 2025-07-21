package errorz

import (
	"fmt"
	"io"
	"os"
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
