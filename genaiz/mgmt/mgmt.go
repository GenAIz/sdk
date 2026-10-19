package mgmt

import (
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz-lib/lang/panicz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
)

var (
	DefaultTimeFormat = timez.NewTodayFormatter()

	// Deprecated: use mgmt.DefaultTimeFormat
	createdFormatter = DefaultTimeFormat
	expiryFormatter  = timez.NewMidnightFormatter()
)

// Facade provides a way to hide the complexity of retrieving a list of user-consumable data, formatting the data with different instances of object listing.
type Facade[T any, P any] interface {
	Filtering(filter string) Provider[T]

	Provider() Provider[T]

	WithLogger(*logrus.Logger) Facade[T, P]

	WithParams(*P) Facade[T, P]
}

// Factory is used to build a Provider around a type T and parameter definition P, allowing listing or filtering of user-consumable records using common configuration tweaks.
type Factory[P any, T any] interface {
	Provider() Provider[T]

	WithLogger(*logrus.Logger) Factory[P, T]

	WithFilter(string) Factory[P, T]

	WithParams(*P) Factory[P, T]

	WithTimeFormat(timez.Formatter) Factory[P, T]
}

// Provider is a generic pattern applicable to an object which can provide an instance of T on-demand.
type Provider[T any] interface {
	Get() (T, task.Error)
}

// Deprecated: Use ProviderFactory where a single parameter type and a logger is needed
type baseLoggingFacade struct {
	logger *logrus.Logger
}

// ProviderFactory implements common Factory methods in terms of a single task parameter type, a logrus.Logger, and potentially a string filter. Composition must be done by aggregating parameters under P, if needed.
type ProviderFactory[P any, T any] struct {
	Filter *string
	Format timez.Formatter
	Logger *logrus.Logger
	Params *P

	provider func(ProviderFactory[P, T]) Provider[T]
}

func (f *ProviderFactory[P, T]) Provider() Provider[T] {
	panicz.RequiresNotNil("provider", f.provider)
	return f.provider(*f)
}

func (f *ProviderFactory[P, T]) WithFilter(filter string) Factory[P, T] {
	if filter != "" {
		f.Filter = &filter
	}

	return f
}

func (f *ProviderFactory[P, T]) WithLogger(logger *logrus.Logger) Factory[P, T] {
	f.Logger = logger
	return f
}

func (f *ProviderFactory[P, T]) WithParams(params *P) Factory[P, T] {
	f.Params = params
	return f
}

func (f *ProviderFactory[P, T]) WithProvider(provider func(ProviderFactory[P, T]) Provider[T]) Factory[P, T] {
	f.provider = provider
	return f
}

func (f *ProviderFactory[P, T]) WithTimeFormat(format timez.Formatter) Factory[P, T] {
	f.Format = format
	return f
}
