package mgmt

import (
	"github.com/sirupsen/logrus"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserDataStoreFacade Facade[[]UserLinkInstance, broker.DataInstanceListParams]

type userDataStoresFacade struct {
	baseLoggingFacade
	params *broker.DataInstanceListParams
}

func (udf userDataStoresFacade) Filtering(filter string) Provider[[]UserLinkInstance] {
	return &userLinkInstancesProvider{
		Plan: task.Plan{
			Logger: udf.logger,
		},
		filter:                      filter,
		params:                      udf.params,
		dataInstanceListTaskFactory: broker.NewDataStoreListTask,
	}
}

func (udf userDataStoresFacade) Provider() Provider[[]UserLinkInstance] {
	return udf.Filtering("")
}

func (udf userDataStoresFacade) WithLogger(logger *logrus.Logger) Facade[[]UserLinkInstance, broker.DataInstanceListParams] {
	udf.logger = logger
	return udf
}

func (udf userDataStoresFacade) WithParams(params *broker.DataInstanceListParams) Facade[[]UserLinkInstance, broker.DataInstanceListParams] {
	udf.params = params
	return udf
}

func NewUserDataStoreFacade() UserDataStoreFacade {
	return &userDataStoresFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
