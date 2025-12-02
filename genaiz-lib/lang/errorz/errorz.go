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

// DeferOnExit returns a function which can exit the program with an error code if err points to an error when the function is evaluated
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

// ResultOrError guarantees either a non-empty reference to an interface{} or an error will be returned if err is not nil
func ResultOrError[T any](result *T, err error) (*T, error) {
	if err == nil {
		return result, nil
	}

	return nil, err
}

// StringSliceOrError guarantees either a string slice or an error will be returned if err is not nil
func StringSliceOrError(slice []string, err error) ([]string, error) {
	if err == nil {
		return slice, nil
	}

	return nil, err
}
