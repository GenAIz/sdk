// Package lang is a collection of utility methods to make the Go language easier to handle.
package lang

import (
	"fmt"
	"os"
)

// Assists returns a lambda defined by a functional call containing one more parameter, effectively hiding that parameter from the original lambda signature
func Assists[A any, B any, C any](a A, lambda func(A, B, C) error) func(B, C) error {
	return func(b B, c C) error {
		return lambda(a, b, c)
	}
}

// HandleExit will terminate execution of the program if the provided msg is not nil, writing the content of msg to os.Stderr.
func HandleExit(msg interface{}) {
	if msg != nil {
		_, _ = fmt.Fprintln(os.Stderr, "Error:", msg)
		os.Exit(1)
	}
}

// Supplier returns a function supplying the provided value, useful for testing otherwise should only be used in limited cases. Always pass a real factory when possible.
func Supplier[P any](value *P) func() *P {
	return func() *P {
		return value
	}
}
