package task

import "fmt"

// EagerErrors eagerly builds "final" error references which can be used, although not encouraged, in runtime flow control.
// The primary use of this pattern is running unit tests which will check the error types propagated out of specific units.
type EagerErrors struct {
}

func (se EagerErrors) SetOrReturn(err *Error, provider ErrorProvider) Error {
	if err != nil {
		if *err == nil {
			*err = provider()
		}

		return *err
	}

	return nil
}

// Error is a custom error interface used to frame the printing of hierarchical error information in various formats. To note: JSON
type Error interface {
	error

	HasAllowedValues() bool

	IsFieldError() bool

	IsRequestError() bool
}

type baseError struct {
	Allowed []string `json:"allowed,omitempty"`
	Field   *string  `json:"field,omitempty"`
	Message string   `json:"message"`
	Status  *int     `json:"status,omitempty"`

	wrapped error
}

func (fe baseError) Error() string {
	return fe.Message
}

func (fe baseError) HasAllowedValues() bool {
	return len(fe.Allowed) > 0
}

func (fe baseError) IsFieldError() bool {
	return fe.Field != nil
}

func (fe baseError) IsRequestError() bool {
	return fe.Status != nil
}

func (fe baseError) Unwrap() error {
	return fe.wrapped
}

func NewError(msg string) Error {
	return &baseError{Message: msg}
}

type ErrorBuilder interface {
	Allowed(...string) ErrorBuilder

	Field(string) ErrorBuilder

	Status(int) ErrorBuilder

	Build(string) Error
}

type buildError struct {
	allowed []string
	field   *string
	status  *int
}

func (be *buildError) Allowed(values ...string) ErrorBuilder {
	be.allowed = append(be.allowed, values...)
	return be
}

func (be *buildError) Field(field string) ErrorBuilder {
	be.field = &field
	return be
}

func (be *buildError) Status(status int) ErrorBuilder {
	be.status = &status
	return be
}

func (be *buildError) Build(message string) Error {
	return &baseError{
		Message: message,
		Allowed: be.allowed,
		Field:   be.field,
		Status:  be.status,
	}
}

type ErrorProvider func() Error

func MakeErrorProvider(msg string) ErrorProvider {
	return func() Error {
		return NewError(msg)
	}
}

func NewErrorBuilder() ErrorBuilder {
	return &buildError{}
}

func NewFailure(i interface{}) Error {
	//it's just plain wrong, Error is an interface not an implementation of error
	//goland:noinspection ALL
	if taskError, ok := i.(Error); ok {
		return taskError
	} else if err, _ := i.(error); err != nil {
		return NewError(err.Error())
	}

	return NewError("unknown failure object")
}

func NewRequestError(msg string, status int) Error {
	return &baseError{
		Message: msg,
		Status:  new(status),
	}
}

func NewFormattedError(parent Error, format string, params ...any) Error {
	return &baseError{
		Message: fmt.Sprintf(format, params...),
		wrapped: parent,
	}
}
