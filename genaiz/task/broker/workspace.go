package broker

import (
	"fmt"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
)

var (
	ErrorWorkspaceEmpty      = task.NewError("no workspace definition provided")
	ErrorWorkspaceVisibility = task.NewError("workspace visibility is required")
)

type WorkspaceCreateParams struct {
	Broker
	Workspace *Workspace
}

func NewWorkspaceCreateTask() *task.Task[WorkspaceCreateParams] {
	return &task.Task[WorkspaceCreateParams]{
		Name:       "workspace-create",
		OnPrepare:  handleWorkspaceCreateContext,
		OnComplete: handleWorkspaceCreateComplete,
		OnPretend:  handleWorkspaceCreatePretend,
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
