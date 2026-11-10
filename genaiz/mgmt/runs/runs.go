package runs

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/intz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserRunsFactory interface {
	mgmt.Factory[broker.RunListParams, []UserRun]
}

type UserRunsProvider = mgmt.Provider[[]UserRun]
type UserRunsProviderFactory = mgmt.ProviderFactory[broker.RunListParams, []UserRun]

type UserRun struct {
	Id              int64 `cli:"Id"`
	SolutionId      int64
	Solution        string `cli:"Solution"`
	WorkspaceId     int64
	WorkspaceFlowId int64
	WorkflowId      int64
	WorkflowHandle  string `cli:"Workflow"`
	Flags           int
	Name            string `cli:"Name"`
	Created         string `cli:"Created"`
	Modified        string
	Progress        int
	Status          string `cli:"Status"`
}

func (ur UserRun) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Id              int64  `json:"id"`
		Created         string `json:"created"`
		Modified        string `json:"modified"`
		SolutionId      int64  `json:"solutionId"`
		WorkspaceId     int64  `json:"workspaceId"`
		WorkspaceFlowId int64  `json:"workspaceFlowId"`
		WorkflowId      int64  `json:"workflowId"`
		Flags           int    `json:"flags"`
		Progress        int    `json:"progress"`
		Status          string `json:"status"`
	}{
		Id:              ur.Id,
		Created:         stringz.FirstNonEmpty(ur.Created, "-"),
		Modified:        stringz.FirstNonEmpty(ur.Modified, "-"),
		SolutionId:      ur.SolutionId,
		WorkspaceId:     ur.WorkspaceId,
		WorkspaceFlowId: ur.WorkspaceFlowId,
		WorkflowId:      ur.WorkflowId,
		Flags:           ur.Flags,
		Progress:        ur.Progress,
		Status:          ur.Status,
	})
}

func (ur UserRun) MarshalSlice() ([]string, error) {
	return []string{
		cast.ToString(ur.Id),
		stringz.FirstNonEmpty(ur.Solution, "???"),
		stringz.FirstNonEmpty(ur.WorkflowHandle, "???"),
		ur.Name,
		ur.Created,
		ur.Status,
	}, nil
}

type userRunsFactory struct {
	mgmt.ProviderFactory[broker.RunListParams, []UserRun]
}

func (uf userRunsFactory) provider(factory mgmt.ProviderFactory[broker.RunListParams, []UserRun]) mgmt.Provider[[]UserRun] {
	return &userRunsProvider{
		Plan: task.Plan{
			Logger: factory.Logger,
		},

		filter: factory.Filter,
		format: factory.Format,
		params: factory.Params,

		runListTaskFactory: broker.RunTasks.ListRunTaskFactory,
	}
}

type userRunsProvider struct {
	task.Plan

	filter    *string
	format    timez.Formatter
	params    *broker.RunListParams
	workspace string

	runListTaskFactory          broker.RunListTaskFactory
	workspaceResolveTaskFactory broker.WorkspaceResolveTaskFactory
}

func (up userRunsProvider) Get() ([]UserRun, task.Error) {
	var workspaceRuns []broker.WorkspaceRun
	var failure interface{}
	var workers []task.Worker

	up.OnReturn = func(i interface{}) { workspaceRuns = i.([]broker.WorkspaceRun) }
	up.OnFailure = func(i interface{}) { failure = i }

	if up.params.GetWorkspaceId() == nil {
		var workspaceParams = &broker.WorkspaceFlowResolveParams{
			Broker:         up.params.Broker,
			FlowReadParams: up.params.FlowReadParams,
			WorkspaceName:  up.workspace,
		}

		workers = append(workers, task.NewWorker(workspaceParams, up.workspaceResolveTaskFactory()))
	}

	workers = append(workers, task.NewWorker(up.params, up.runListTaskFactory()))
	up.Sequence(workers...)

	if failure == nil {
		var result = make([]UserRun, 0)

		for _, wr := range workspaceRuns {
			var userRun = &UserRun{
				Id:              intz.Int64ToDefault(wr.Id, -1),
				SolutionId:      intz.Int64ToDefault(wr.SolutionId, -1),
				WorkspaceId:     wr.WorkspaceId,
				WorkspaceFlowId: wr.WorkspaceFlowId,
				WorkflowId:      intz.Int64ToDefault(wr.WorkflowId, -1),
				Name:            wr.Name,
				Flags:           intz.IntToDefault(wr.Flags, 0),
				Status:          wr.Status,
				Progress:        wr.Progress,
			}

			if wr.Solution != nil {
				userRun.Solution = fmt.Sprintf("%s:%s", wr.Solution.GetFqdn(), wr.Solution.GetVersion())
			}

			if wr.Created != nil {
				userRun.Created = up.format.FormatMillis(*wr.Created)
			}

			if wr.Modified != nil {
				userRun.Modified = up.format.FormatMillis(*wr.Modified)
			}

			result = append(result, *userRun)
		}

		return result, nil
	}

	return nil, failure.(task.Error)
}

func NewUserRunsFactory() UserRunsFactory {
	var factory = &userRunsFactory{
		ProviderFactory: mgmt.ProviderFactory[broker.RunListParams, []UserRun]{
			Format: mgmt.DefaultTimeFormat,
		},
	}

	return factory.WithProvider(factory.provider)
}
