// Package panicz is made to handle situations where we want to panic or handle a panic. We assume that a panic in Go is indicative of a Bug
package panicz

// PanicIfError will panic if the error is not nil, this is particular to when the error is a Bug, otherwise the code needs to handle it
func PanicIfError(err error) {
	if err != nil {
		panic(err)
	}
}
