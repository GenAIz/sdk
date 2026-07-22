package mgmt

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserWorkspacesFacade Facade[[]UserWorkspace, broker.WorkspaceListParams]
type WorkspaceListTaskFactory func() *task.Task[broker.WorkspaceListParams]

type UserWorkspace struct {
	Id          int64  `cli:"Id"`
	Name        string `cli:"Name"`
	Description string
	Created     int64 `cli:"Created"`
	Modified    int64
	OwnerAppId  int
	OwnerUserId int
	Visibility  string
	RcEnabled   bool `cli:"RC?"`
	Active      bool `cli:"Active"`
	Flags       *int

	matched string
}

func (uw UserWorkspace) MarshalJSON() ([]byte, error) {
	var created, modified string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	}

	if uw.Modified > 0 {
		modified = createdFormatter.FormatMillis(uw.Modified)
	}

	return json.Marshal(&struct {
		Id          int64  `json:"id,omitempty"`
		Name        string `json:"name,omitempty"`
		Description string `json:"description,omitempty"`
		OwnerAppId  int    `json:"ownerAppId,omitempty"`
		OwnerUserId int    `json:"ownerUserId"`
		Created     string `json:"created,omitempty"`
		Modified    string `json:"modified,omitempty"`
		Visibility  string `json:"visibility"`
		Flags       *int   `json:"flags,omitempty"`
	}{
		Id:          uw.Id,
		Name:        uw.Name,
		Description: uw.Description,
		OwnerAppId:  uw.OwnerAppId,
		OwnerUserId: uw.OwnerUserId,
		Created:     created,
		Modified:    modified,
		Visibility:  uw.Visibility,
		Flags:       uw.Flags,
	})
}

func (uw UserWorkspace) MarshalSlice() ([]string, error) {
	var created string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	} else {
		created = "-"
	}

	return []string{
		cast.ToString(uw.Id),
		uw.Name,
		created,
		stringz.YesOrNo(uw.RcEnabled),
		stringz.YesOrNo(uw.Active),
	}, nil
}

func (uw UserWorkspace) Match(filter string) *UserWorkspace {
	var idString = cast.ToString(uw.Id)
	var matched string

	if strings.EqualFold(idString, filter) ||
		strings.HasPrefix(idString, filter) {
		matched = idString
	} else if strings.EqualFold(uw.Name, filter) ||
		strings.HasPrefix(uw.Name, filter) {
		matched = uw.Name
	}

	if matched == "" {
		return nil
	}

	return &UserWorkspace{
		Id:          uw.Id,
		Name:        uw.Name,
		Description: uw.Description,
		Created:     uw.Created,
		Modified:    uw.Modified,
		OwnerAppId:  uw.OwnerAppId,
		OwnerUserId: uw.OwnerUserId,
		Visibility:  uw.Visibility,
		RcEnabled:   uw.RcEnabled,
		Active:      uw.Active,
		Flags:       uw.Flags,
		matched:     matched,
	}
}

func (uw UserWorkspace) Matched() cobra.Completion {
	if uw.matched == "" {
		return cobra.CompletionWithDesc(cast.ToString(uw.Id), uw.Name)
	}

	if _, err := strconv.Atoi(uw.matched); err == nil {
		return cobra.CompletionWithDesc(uw.matched, uw.Name)
	}

	return cobra.CompletionWithDesc(uw.matched, cast.ToString(uw.Id))
}

type userWorkspacesFacade struct {
	baseLoggingFacade
	params *broker.WorkspaceListParams
}

func (uwf *userWorkspacesFacade) Filtering(filter string) Provider[[]UserWorkspace] {
	return &userWorkspacesProvider{
		Plan: task.Plan{
			Logger: uwf.logger,
		},
		filter:                   filter,
		params:                   uwf.params,
		workspaceListTaskFactory: broker.NewWorkspaceListTask,
	}
}

func (uwf *userWorkspacesFacade) Provider() Provider[[]UserWorkspace] {
	return uwf.Filtering("")
}

func (uwf *userWorkspacesFacade) WithLogger(logger *logrus.Logger) Facade[[]UserWorkspace, broker.WorkspaceListParams] {
	uwf.logger = logger
	return uwf
}

func (uwf *userWorkspacesFacade) WithParams(params *broker.WorkspaceListParams) Facade[[]UserWorkspace, broker.WorkspaceListParams] {
	uwf.params = params
	return uwf
}

type userWorkspacesProvider struct {
	task.Plan
	filter                   string
	params                   *broker.WorkspaceListParams
	workspaceListTaskFactory WorkspaceListTaskFactory
}

func (uwp userWorkspacesProvider) Get() ([]UserWorkspace, task.Error) {
	var workspaces []broker.Workspace
	var failure interface{}

	uwp.OnReturn = func(i interface{}) { workspaces = i.([]broker.Workspace) }
	uwp.OnFailure = func(i interface{}) { failure = i }
	uwp.Sequence(task.NewWorker(uwp.params, uwp.workspaceListTaskFactory()))

	if failure == nil {
		var result = make([]UserWorkspace, 0)

		for _, ws := range workspaces {
			var userWorkspace = &UserWorkspace{
				Id:          ws.Id,
				Name:        ws.Name,
				Description: ws.Description,
				OwnerAppId:  ws.OwnerAppId,
				OwnerUserId: ws.OwnerUserId,
				Visibility:  ws.Visibility,
				RcEnabled:   ws.IsRcEnabled(),
				Created:     ws.Created,
				Modified:    ws.Modified,
				Flags:       ws.Flags,
				Active:      ws.IsActive(),
			}

			if uwp.filter == "" {
				result = append(result, *userWorkspace)
			} else if matched := userWorkspace.Match(uwp.filter); matched != nil {
				result = append(result, *matched)
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserWorkspacesFacade() UserWorkspacesFacade {
	return &userWorkspacesFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
