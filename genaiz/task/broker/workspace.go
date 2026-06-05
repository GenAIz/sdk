package broker

import (
	"fmt"
	"time"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
)

var (
	ErrorWorkspaceEmpty        = task.NewError("no workspace definition provided")
	ErrorWorkspaceInvalidNco   = task.NewError("workspace creation date is invalid")
	ErrorWorkspaceInvalidOwner = task.NewError("workspace ownership can not be established with the selected session")
	ErrorWorkspaceVisibility   = task.NewError("workspace visibility is required")
)

type WorkspaceCreateParams struct {
	Broker
	Workspace *Workspace
}

type WorkspaceListParams struct {
	Broker
	FromDate  *time.Time
	OwnerOnly bool
	RcEnabled bool
}

func (wlp WorkspaceListParams) GetMaskFlags() (int, int) {
	if wlp.RcEnabled {
		// see the orchestrator API
		return WorkspaceFlags.Active | WorkspaceFlags.RcEnabled,
			WorkspaceFlags.Active | WorkspaceFlags.RcEnabled
	}

	return WorkspaceFlags.Active | WorkspaceFlags.RcEnabled,
		WorkspaceFlags.Active
}

func NewWorkspaceCreateTask() *task.Task[WorkspaceCreateParams] {
	return &task.Task[WorkspaceCreateParams]{
		Name:       "workspace-create",
		OnPrepare:  handleWorkspaceCreateContext,
		OnComplete: handleWorkspaceCreateComplete,
		OnPretend:  handleWorkspaceCreatePretend,
	}
}

func NewWorkspaceListTask() *task.Task[WorkspaceListParams] {
	return &task.Task[WorkspaceListParams]{
		Name:       "workspace-list",
		OnPrepare:  handleWorkspaceListContext,
		OnComplete: handleWorkspaceListComplete,
		OnPretend:  handleWorkspaceListPretend,
	}
}

func handleWorkspaceCreateContext(params *WorkspaceCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.Workspace != nil {
			var nameSuffix string

			if params.Workspace.Visibility == "" {
				return ErrorWorkspaceVisibility
			}

			if params.Workspace.Name == "" {
				state.Logger.Warn("Nameless workspaces are allowed, but not recommended")
			} else {
				nameSuffix = fmt.Sprintf(" with name [%s]", params.Workspace.Name)
			}

			if params.Workspace.IsRcEnabled() {
				state.Logger.Debugf("Creating a development workspace%s", nameSuffix)
			} else {
				state.Logger.Debugf("Creating a production workspace%s", nameSuffix)
			}

			return nil
		}

		return ErrorWorkspaceEmpty
	}

	return nil
}

func handleWorkspaceCreateComplete(params *WorkspaceCreateParams, state *task.State) error {
	if params.Workspace != nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var created *Workspace

			state.Logger.Debugf("Workspace created on host [%s]", brokerClient.GetHostAddr())

			if created, err = brokerClient.CreateWorkspace(params.Workspace); err == nil {
				var formatter = timez.NewTodayFormatter()

				state.Logger.Debugf("Workspace id [%d] created on [%s]", created.Id, formatter.FormatMillis(created.Created))
				state.Reportf("Created workspace id [%d]", created.Id)
				state.Output = cast.ToString(created.Id)
				state.Internal = created
				return nil
			}
		}

		return err
	}

	return ErrorWorkspaceEmpty
}

func handleWorkspaceCreatePretend(params *WorkspaceCreateParams, state *task.State) error {
	if params.Workspace != nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending creating workspace [%s]", params.Workspace.Name)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d name=%s\\\n", params.Workspace.Name)
			fmt.Printf("  -d description=%s\\\n", params.Workspace.Description)
			fmt.Printf("  -d visibility=%s\\\n", params.Workspace.Visibility)
			fmt.Printf("  -d rcEnabled=%s\\\n", cast.ToString(params.Workspace.RcEnabled))
			fmt.Printf("%s\n", brokerClient.CreateWorkspaceUrl())
			return nil
		}

		return err
	}

	return ErrorWorkspaceEmpty
}

func handleWorkspaceListContext(params *WorkspaceListParams, state *task.State) error {
	if state.Output == "" {
		var brokerClient Client
		var err error

		if params.FromDate != nil && time.Now().Before(*params.FromDate) {
			return ErrorWorkspaceInvalidNco
		}

		if params.OwnerOnly {
			if brokerClient, err = params.GetClient(); err == nil {
				if brokerClient.GetUserId() > 0 {
					return nil
				}

				return ErrorWorkspaceInvalidOwner
			}
		}

		return err
	}

	return nil
}

func handleWorkspaceListComplete(params *WorkspaceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var mask, flag = params.GetMaskFlags()
		var workspaces []Workspace

		if workspaces, err = brokerClient.ListWorkspaces(mask, flag); err == nil {
			var results = workspaces

			if params.OwnerOnly {
				var ownerId = brokerClient.GetUserId()
				var filtered []Workspace

				state.Logger.Debugf("Filtering workspaces by owner [%d]", ownerId)

				for _, ws := range results {
					if ownerId == ws.OwnerUserId {
						filtered = append(filtered, ws)
					}
				}

				results = filtered
			}

			if params.FromDate != nil {
				var filtered []Workspace

				state.Logger.Debugf("Filtering workspaces after date [%s]", params.FromDate.Format(time.DateOnly))

				for _, ws := range results {
					if time.UnixMilli(ws.Created).After(*params.FromDate) {
						filtered = append(filtered, ws)
					}
				}

				results = filtered
			}

			state.Output = cast.ToString(len(results))
			state.Internal = results
			return nil
		}

		return err
	}

	return err
}

func handleWorkspaceListPretend(params *WorkspaceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var mask, flags = params.GetMaskFlags()
		var loggingSuffix string

		if params.FromDate != nil {
			loggingSuffix = fmt.Sprintf(" after date [%s]", params.FromDate.Format(time.DateOnly))
		}

		state.Logger.Debugf("Pretending to list workspaces%s", loggingSuffix)
		fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
		fmt.Printf("%s?mask=%d&flags=%d\n", brokerClient.ListWorkspacesUrl(), mask, flags)
		return nil
	}

	return err
}
