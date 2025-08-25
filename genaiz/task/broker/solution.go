package broker

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorWorkflowConflict     = errors.New("workflow already exists")
	errorWorkflowFileInvalid  = errors.New("workflow config is invalid")
	errorWorkflowFileExists   = errors.New("workflow config file exists")
	errorWorkflowFileNotFound = errors.New("workflow config file not found")
	errorWorkflowNotFound     = errors.New("workflow not found")
)

type WorkflowWriter interface {
	BuildWorkflows() (string, []Workflow)

	GetWorkflowByHandle(string) (*Workflow, error)

	GetWorkflows() []Workflow

	WithWorkflow(*Workflow) WorkflowWriter

	WithWorkflows([]Workflow) WorkflowWriter

	WithWorkflowLinks(string, []WorkflowLink) WorkflowWriter

	WithWorkflowNodes(string, []WorkflowNode) WorkflowWriter

	Write(string) error
}

type WorkflowParams struct {
	shared.ConfigParams
	*Workflow
	WorkflowFolder string
	WorkflowUpdate bool
}

func (wp WorkflowParams) workflowPredicate() func(Workflow) bool {
	if wp.Handle != "" {
		return WorkflowHandlePredicate(wp.Handle)
	} else if wp.Name != " " {
		return WorkflowNamePredicate(wp.Name)
	}

	return func(Workflow) bool {
		return false
	}
}

func NewWorkflowDeleteTask(writer WorkflowWriter) *task.Task[WorkflowParams] {
	return &task.Task[WorkflowParams]{
		Name:       "solution-delete-workflow",
		OnPrepare:  handleWorkflowDeleteContext,
		OnComplete: lang.Assists(writer, handleWorkflowDeleteConfig),
		OnPretend:  handleWorkflowDeletePretend,
	}
}

func NewWorkflowUpdateTask(writer WorkflowWriter) *task.Task[WorkflowParams] {
	return &task.Task[WorkflowParams]{
		Name:         "solution-update-workflow",
		OnPrepare:    handleWorkflowCreateContext,
		OnIncomplete: lang.Assists(writer, handleWorkflowUpdateConfig),
		OnComplete:   lang.Assists(writer, handleWorkflowCreateConfig),
		OnPretend:    lang.Assists(writer, handleWorkflowUpdatePretend),
	}
}

func handleWorkflowCreateContext(params *WorkflowParams, state *task.State) error {
	if state.Output == "" {
		state.Logger.Debugf("Finding a workflow configuration file for writing")

		if params.IsConfigTypeNone() {
			var reset func()
			var err error

			if reset, err = dirz.ChangeWorkingDir(params.WorkflowFolder); err == nil {
				defer reset()
				var file string

				if file, err = filez.FirstNamedFile(params.GetConfigFile()); err == nil {
					state.Output = file
					return errorWorkflowFileExists
				}
			}

			return err
		} else {
			state.Output = params.GetConfigFile(params.WorkflowFolder)

			if info, err := os.Stat(state.Output); err == nil {
				if info.IsDir() {
					return errorWorkflowFileInvalid
				}

				return errorWorkflowFileExists
			}
		}
	}

	return nil
}

func handleWorkflowDeleteContext(params *WorkflowParams, state *task.State) error {
	var err = handleWorkflowCreateContext(params, state)

	if errors.Is(err, errorWorkflowFileExists) {
		return nil
	}

	if err == nil {
		return errorWorkflowFileNotFound
	}

	return err
}

func handleWorkflowCreateConfig(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		state.Logger.Debugf("Workflow writing to [%s]", state.Output)
		return writer.WithWorkflow(params.Workflow).
			Write(state.Output)
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowDeleteConfig(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var comparison = params.workflowPredicate()
		var updated []Workflow

		state.Logger.Debugf("Workflow deleting from [%s]", state.Output)

		for _, wf := range writer.GetWorkflows() {
			if !comparison(wf) {
				updated = append(updated, wf)
			}
		}

		return writer.WithWorkflows(updated).
			Write(state.Output)
	}

	return nil
}

func handleWorkflowUpdateConfig(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var update *Workflow
		var err error

		state.Completed = true

		if update, err = writer.GetWorkflowByHandle(params.Handle); err == nil {
			if params.WorkflowUpdate {
				var updated = params.Workflow

				state.Logger.Debugf("Workflow [%s] updated to [%s]", update.Name, state.Output)

				if params.Description != "" {
					update.Description = params.Description
				}

				if params.Handle != "" {
					update.Handle = params.Handle
				}

				err = writer.WithWorkflows(writer.GetWorkflows()).
					WithWorkflowLinks(update.Handle, updated.Links).
					WithWorkflowNodes(update.Handle, updated.Nodes).
					Write(state.Output)
				return nil
			} else {
				state.Logger.Errorf("Workflow [%s] already exist", params.Name)
				return errorWorkflowConflict
			}
		} else {
			return handleWorkflowCreateConfig(
				writer.WithWorkflows(writer.GetWorkflows()), params, state)
		}
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowDeletePretend(params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)

		if params.Workflow.Handle != "" {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return "solution.workflows[]", "handle", params.Workflow.Handle
			})
		} else if params.Workflow.Name != "" {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return "solution.workflows[]", "name", params.Workflow.Name
			})
		} else {
			return errorWorkflowNotFound
		}

		return nil
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowUpdatePretend(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var comparison = params.workflowPredicate()
		var pretender = shared.NewConfigPretender(state.Output)
		var rootKey, workflows = writer.WithWorkflows(writer.GetWorkflows()).
			WithWorkflow(params.Workflow).
			BuildWorkflows()

		state.Logger.Debugf("Pretending to writing workflow [%s] to [%s]", params.Name, state.Output)

		if state.Error == nil {
			var confPath = filepath.Dir(state.Output)

			fmt.Printf("mkdir -p %s && cd %s\n", confPath, confPath)
			fmt.Printf("touch %s\n", state.Output)
		}

		if i := slices.IndexFunc(workflows, comparison); i >= 0 {
			var wf = workflows[i]

			if !strings.EqualFold(wf.Name, params.Name) || !params.WorkflowUpdate {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", rootKey, i), wf.Name
				})
			}

			if !strings.EqualFold(wf.Description, params.Description) || !params.WorkflowUpdate {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].description", rootKey, i), wf.Description
				})
			}

			if !strings.EqualFold(wf.Handle, params.Handle) || !params.WorkflowUpdate {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].handle", rootKey, i), wf.Handle
				})
			}

			if params.WorkflowUpdate {
				shared.PretendDelete(pretender, func() string {
					return fmt.Sprintf("%s[%d].links", rootKey, i)
				})
			}

			for j, link := range wf.Links {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].links[%d].lhsNode", rootKey, i, j), link.LhsNode
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].links[%d].lhsNodePort", rootKey, i, j), link.LhsNodePort
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].links[%d].rhsNode", rootKey, i, j), link.RhsNode
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].links[%d].rhsNodePort", rootKey, i, j), link.RhsNodePort
				})
			}

			if params.WorkflowUpdate {
				shared.PretendDelete(pretender, func() string {
					return fmt.Sprintf("%s[%d].nodes", rootKey, i)
				})
			}

			for j, node := range wf.Nodes {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].nodes[%d].name", rootKey, i, j), node.Name
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].nodes[%d].description", rootKey, i, j), node.Description
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].nodes[%d].handle", rootKey, i, j), node.Handle
				})

				if node.Sf != nil {
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].nodes[%d].sf.oem", rootKey, i, j), node.Sf.Oem
					})
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].nodes[%d].sf.handle", rootKey, i, j), node.Sf.Handle
					})
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].nodes[%d].sf.version", rootKey, i, j), node.Sf.Version
					})
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].nodes[%d].sf.seq", rootKey, i, j), cast.ToString(node.Sf.Seq)
					})
				}
			}
		}

		return nil
	}

	return errorWorkflowFileInvalid
}
