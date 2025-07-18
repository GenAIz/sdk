// Package panicz is made to handle situations where we want to panic or handle a panic. We assume that a panic in Go is indicative of a Bug
package panicz

import (
	"errors"
	"fmt"
	"reflect"
)

var (
	ErrCanNotBeNil = errors.New("can not be nil")
)

// PanicIfError will panic if the error is not nil, this is particular to when the error is a Bug, otherwise the code needs to handle it
func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}

func PanicIfNoReturn[P any](result P, err error) P {
	PanicIfError(err)
	return result
}

// RequiresNotNil will panic if an object, typically a pointer evaluates to nil. This is for catching bugs before they happen, narrowing the stack trace to as soon as we are aware nil is not valid
func RequiresNotNil(vName string, obj any) {
	if obj == nil || reflect.ValueOf(obj).IsNil() {
		panic(fmt.Errorf("%s %s", vName, ErrCanNotBeNil))
	}
}
