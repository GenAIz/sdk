package task

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

func NewErrorBuilder() ErrorBuilder {
	return &buildError{}
}
