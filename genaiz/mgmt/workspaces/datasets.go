package workspaces

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/intz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserWorkspaceDataSetsFactory interface {
	mgmt.Factory[broker.WorkspaceNodeResolveParams, []UserWorkspaceDataSet]
}

type UserWorkspaceDataSetsProvider = mgmt.Provider[[]UserWorkspaceDataSet]
type UserWorkspaceDataSetsProviderFactory = mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []UserWorkspaceDataSet]

type UserWorkspaceDataSet struct {
	Id                  *int64 `cli:"Id"`
	Name                string
	Handle              string
	Description         string
	SmartFunctionHandle string `cli:"Function"`
	InputPort           string `cli:"Input Port"`
	TotalFiles          int    `cli:"Files"`
	TotalSize           int64
	Digest              string
	RunId               *int64
	RunJobId            *int64
	Created             string
	Modified            string `cli:"Modified"`
	Active              bool
	Sealed              string `cli:"Sealed?" redGreen:"Sealed"`
	Flags               *int

	matched string
}

func (ud UserWorkspaceDataSet) MarshalJSON() ([]byte, error) {
	return json.Marshal(&struct {
		Id          *int64 `json:"id,omitempty"`
		Name        string `json:"name,omitempty"`
		Handle      string `json:"handle,omitempty"`
		Description string `json:"description,omitempty"`
		InputPort   string `json:"inputPort"`
		TotalSize   int64  `json:"totalSize,omitempty"`
		TotalFiles  int    `json:"totalFiles,omitempty"`
		Digest      string `json:"digest,omitempty"`
		RunId       *int64 `json:"runId,omitempty"`
		RunJobId    *int64 `json:"runJobId,omitempty"`
		Created     string `json:"created,omitempty"`
		Modified    string `json:"modified,omitempty"`
		Active      bool   `json:"active"`
		Flags       *int   `json:"flags,omitempty"`
	}{
		Id:          ud.Id,
		Name:        ud.Name,
		Handle:      ud.Handle,
		Description: ud.Description,
		InputPort:   ud.InputPort,
		TotalSize:   ud.TotalSize,
		TotalFiles:  ud.TotalFiles,
		Digest:      ud.Digest,
		RunId:       ud.RunId,
		RunJobId:    ud.RunJobId,
		Created:     ud.Created,
		Modified:    ud.Modified,
		Active:      ud.Active,
		Flags:       ud.Flags,
	})
}

func (ud UserWorkspaceDataSet) MarshalSlice() ([]string, error) {
	var idString string

	if ud.Id == nil {
		idString = "-"
	} else {
		idString = cast.ToString(ud.Id)
	}

	return []string{
		idString,
		ud.SmartFunctionHandle,
		ud.InputPort,
		cast.ToString(ud.TotalFiles),
		ud.Modified,
		ud.Sealed,
	}, nil
}

func (ud UserWorkspaceDataSet) Match(filter string) *UserWorkspaceDataSet {
	var matched string

	if ud.Id != nil && strings.HasPrefix(cast.ToString(*ud.Id), filter) {
		matched = cast.ToString(*ud.Id)
	} else if ud.Name != "" &&
		// Name is a DataSet field, if the node port returned is not assigned it will be filtered out
		// It is normal as we are ultimately listing existing datasets on filter
		strings.EqualFold(ud.Name, filter) ||
		strings.HasPrefix(ud.Name, filter) {
		matched = ud.Name
	} else if strings.HasPrefix(ud.InputPort, filter) {
		matched = ud.InputPort
	}

	if matched == "" {
		return nil
	}

	return &UserWorkspaceDataSet{
		Id:                  ud.Id,
		Name:                ud.Name,
		Handle:              ud.Handle,
		Description:         ud.Description,
		SmartFunctionHandle: ud.SmartFunctionHandle,
		InputPort:           ud.InputPort,
		TotalSize:           ud.TotalSize,
		TotalFiles:          ud.TotalFiles,
		Digest:              ud.Digest,
		RunId:               ud.RunId,
		RunJobId:            ud.RunJobId,
		Created:             ud.Created,
		Modified:            ud.Modified,
		Sealed:              ud.Sealed,
		Flags:               ud.Flags,

		matched: matched,
	}
}

func (ud UserWorkspaceDataSet) Matched() cobra.Completion {
	var result, description string

	if id, err := strconv.ParseInt(ud.matched, 10, 64); err == nil {
		if ud.Id != nil && id == *ud.Id {
			return cobra.CompletionWithDesc(ud.matched, ud.InputPort)
		}
	}

	if ud.Id != nil {
		description = cast.ToString(*ud.Id)
	}

	if ud.matched == "" {
		result = ud.InputPort
	} else if ud.matched == ud.Name {
		result = ud.matched
		description = ud.InputPort
	} else {
		result = ud.matched
	}

	if description == "" {
		return result
	}

	return cobra.CompletionWithDesc(result, description)
}

type userWorkspaceDataSetsFactory struct {
	mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []UserWorkspaceDataSet]
}

func (uf userWorkspaceDataSetsFactory) provider(factory mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []UserWorkspaceDataSet]) mgmt.Provider[[]UserWorkspaceDataSet] {
	return &userWorkspaceDataSetsProvider{
		Plan: task.Plan{
			Logger: factory.Logger,
		},

		filter: factory.Filter,
		format: factory.Format,
		params: factory.Params,

		workspaceNodeResolveTaskFactory: broker.WorkspaceTasks.ResolveNodeTaskFactory,
		workspaceNodeSetListTaskFactory: broker.WorkspaceTasks.ListSetNodeTaskFactory,
	}
}

type userWorkspaceDataSetsProvider struct {
	task.Plan

	filter *string
	format timez.Formatter
	params *broker.WorkspaceNodeResolveParams

	workspaceNodeResolveTaskFactory broker.WorkspaceNodeResolveTaskFactory
	workspaceNodeSetListTaskFactory broker.WorkspaceNodeSetListTaskFactory
}

func (up userWorkspaceDataSetsProvider) Get() ([]UserWorkspaceDataSet, task.Error) {
	var listParams = &broker.WorkspaceNodeSetListParams{
		Broker:            up.params.Broker,
		NodeResolveParams: &up.params.NodeResolveParams,
	}
	var dataSets []broker.WorkspaceNodePort
	var failure interface{}

	up.OnReturn = func(i interface{}) { dataSets = i.([]broker.WorkspaceNodePort) }
	up.OnFailure = func(i interface{}) { failure = i }
	up.Sequence(task.NewWorker(up.params, up.workspaceNodeResolveTaskFactory()),
		task.NewWorker(listParams, up.workspaceNodeSetListTaskFactory()))

	if failure == nil {
		var result = make([]UserWorkspaceDataSet, 0)

		for _, set := range dataSets {
			var userSet = &UserWorkspaceDataSet{
				Id:        set.DataSetId,
				InputPort: set.PortHandle,
			}

			if set.DataSet != nil {
				userSet.Name = set.DataSet.Name
				userSet.Handle = set.DataSet.Handle
				userSet.Description = set.DataSet.Description
				userSet.TotalFiles = intz.IntToDefault(set.DataSet.TotalFiles, 0)
				userSet.TotalSize = intz.Int64ToDefault(set.DataSet.TotalSize, 0)
				userSet.Digest = set.DataSet.Digest
				userSet.RunId = set.DataSet.RunId
				userSet.RunJobId = set.DataSet.RunJobId
				userSet.Active = set.DataSet.IsActive()
				userSet.Sealed = stringz.YesOrNo(set.DataSet.IsSealed())
				userSet.Flags = set.DataSet.Flags

				if set.DataSet.Created != nil {
					userSet.Created = up.format.FormatMillis(*set.DataSet.Created)
				}

				if set.DataSet.Modified != nil {
					userSet.Modified = up.format.FormatMillis(*set.DataSet.Modified)
				}
			} else {
				userSet.Sealed = "-"
			}

			if set.SmartFunction != nil {
				userSet.SmartFunctionHandle = set.SmartFunction.Handle
			}

			if up.filter == nil {
				result = append(result, *userSet)
			} else if matched := userSet.Match(*up.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserWorkspaceDataSetsFactory() UserWorkspaceDataSetsFactory {
	var factory = &userWorkspaceDataSetsFactory{
		ProviderFactory: mgmt.ProviderFactory[broker.WorkspaceNodeResolveParams, []UserWorkspaceDataSet]{
			Format: mgmt.DefaultTimeFormat,
		},
	}

	return factory.WithProvider(factory.provider)
}
