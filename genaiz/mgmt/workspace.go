package mgmt

import (
	"encoding/json"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

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

func (uw UserWorkspace) Match(filter string) bool {
	var lowFilter = strings.ToLower(filter)
	var idString = cast.ToString(uw.Id)

	return strings.EqualFold(idString, lowFilter) ||
		strings.HasPrefix(idString, lowFilter) ||
		strings.EqualFold(uw.Name, lowFilter) ||
		strings.HasPrefix(uw.Name, lowFilter) ||
		strings.HasSuffix(uw.Name, lowFilter)
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

			if uwp.filter == "" || userWorkspace.Match(uwp.filter) {
				result = append(result, *userWorkspace)
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
