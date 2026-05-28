package mgmt

import (
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/task"
)

type Facade[T any, P any] interface {
	Filtering(filter string) Provider[T]

	Provider() Provider[T]

	WithLogger(*logrus.Logger) Facade[T, P]

	WithParams(*P) Facade[T, P]
}

type Provider[T any] interface {
	Get() (T, task.Error)
}

type baseLoggingFacade struct {
	logger *logrus.Logger
}
