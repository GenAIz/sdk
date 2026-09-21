package mgmt

import (
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserDataSourceFacade Facade[[]UserLinkInstance, broker.DataInstanceListParams]

type ListDataSourceTaskFactory func() *task.Task[broker.DataInstanceListParams]

type userDataSourcesFacade struct {
	baseLoggingFacade
	params *broker.DataInstanceListParams
}

func (udf userDataSourcesFacade) Filtering(filter string) Provider[[]UserLinkInstance] {
	return &userLinkInstancesProvider{
		Plan: task.Plan{
			Logger: udf.logger,
		},
		filter:                      filter,
		params:                      udf.params,
		dataInstanceListTaskFactory: broker.NewDataSourceListTask,
	}
}

func (udf userDataSourcesFacade) Provider() Provider[[]UserLinkInstance] {
	return udf.Filtering("")
}

func (udf userDataSourcesFacade) WithLogger(logger *logrus.Logger) Facade[[]UserLinkInstance, broker.DataInstanceListParams] {
	udf.logger = logger
	return udf
}

func (udf userDataSourcesFacade) WithParams(params *broker.DataInstanceListParams) Facade[[]UserLinkInstance, broker.DataInstanceListParams] {
	udf.params = params
	return udf
}

func NewUserDataSourceFacade() UserDataSourceFacade {
	return &userDataSourcesFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
