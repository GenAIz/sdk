package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorSolutionFileInvalid  = errors.New("solution config is invalid")
	errorSolutionInvalid      = errors.New("solution is invalid")
	errorWorkflowConflict     = errors.New("workflow already exists")
	errorWorkflowFileInvalid  = errors.New("workflow config is invalid")
	errorWorkflowFileNotFound = errors.New("workflow config file not found")
	errorWorkflowNotFound     = errors.New("workflow not found")
)

type SolutionWriter interface {
	BuildSolution() (string, Solution)

	GetWorkflowByHandle(string) (*Workflow, error)

	WithSolution(*Solution) SolutionWriter

	WithWorkflow(*Workflow) SolutionWriter

	Write(string) error
}

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

type SolutionParams struct {
	shared.ConfigParams
	*Solution
}

func (sp SolutionParams) HasWorkflows() bool {
	return sp.Solution != nil && len(sp.Workflows) > 0
}

type SolutionPublishParams struct {
	Broker
	*Solution
	Provisions []ProvisionParams
}

func (spp SolutionPublishParams) HasProvision(oem, handle string) bool {
	return slices.ContainsFunc(spp.Provisions, func(fn ProvisionParams) bool {
		return strings.EqualFold(fn.Oem, oem) && strings.EqualFold(fn.Handle, handle)
	})
}

type WorkflowParams struct {
	shared.ConfigParams
	*Workflow
	WorkflowUpdate bool
}

func (wp WorkflowParams) GetHandle() string {
	if wp.Workflow != nil {
		return wp.Workflow.Handle
	}

	return ""
}

func (wp WorkflowParams) GetName() string {
	if wp.Workflow != nil {
		return wp.Workflow.Name
	}

	return ""
}

func (wp WorkflowParams) workflowPredicate() func(Workflow) bool {
	if wp.Handle != "" {
		return WorkflowHandlePredicate(wp.Handle)
	} else if wp.Name != "" {
		return WorkflowNamePredicate(wp.Name)
	}

	return func(Workflow) bool {
		return false
	}
}

func NewSolutionPublishTask() *task.Task[SolutionPublishParams] {
	return &task.Task[SolutionPublishParams]{
		Name:       "solution-publish",
		OnPrepare:  handleSolutionPublishContext,
		OnComplete: handleSolutionPublishComplete,
		OnPretend:  handleSolutionPublishPretend,
	}
}

func NewSolutionUpdateTask(writer SolutionWriter) *task.Task[SolutionParams] {
	return &task.Task[SolutionParams]{
		Name:         "solution-update",
		OnPrepare:    handleSolutionCreateContext,
		OnIncomplete: lang.Assists(writer, handleSolutionUpdateConfig),
		OnComplete:   lang.Assists(writer, handleSolutionCreateConfig),
		OnPretend:    lang.Assists(writer, handleSolutionUpdatePretend),
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

func handleSolutionCreateConfig(writer SolutionWriter, params *SolutionParams, state *task.State) error {
	if state.Output != "" {
		state.Logger.Debugf("Solution writing to [%s]", state.Output)
		return writer.WithSolution(params.Solution).
			Write(state.Output)
	}

	return errorSolutionFileInvalid
}

func handleSolutionCreateContext(params *SolutionParams, state *task.State) error {
	if state.Output == "" {
		var err error

		state.Logger.Debugf("Finding a solution configuration file for writing")

		if state.Output, err = params.ResolveConfigPath(); err != nil {
			return err
		}
	}

	return nil
}

func handleSolutionPublishComplete(params *SolutionPublishParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
		var solution = params.Solution
		var identity *shared.Identity

		if identity, err = brokerClient.PublishSolution(solution); err == nil {
			state.Logger.Infof("Published solution %s:%s", identity.Path, identity.Version)
			return nil
		}
	}

	return err
}

func handleSolutionPublishContext(params *SolutionPublishParams, state *task.State) error {
	if params.Solution != nil {
		var allNodeFunctions []WorkflowNodeFunction

		state.Logger.Debugf("Validating solution publishing for [%s/%s]", params.Oem, params.Handle)

		for _, w := range params.Workflows {
			if len(w.Nodes) > 0 {
				for _, n := range w.Nodes {
					if n.Sf != nil {
						allNodeFunctions = append(allNodeFunctions, *n.Sf)
					}
				}
			} else {
				return fmt.Errorf("the workflow %s must have at least one node", w.Handle)
			}
		}

		for _, sf := range allNodeFunctions {
			if !params.HasProvision(sf.Oem, sf.Handle) {
				return fmt.Errorf("the function %s/%s could not be found within the solution", sf.Oem, sf.Handle)
			}
		}

		return nil
	}

	return errorSolutionInvalid
}

func handleSolutionPublishPretend(params *SolutionPublishParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = clientFactory.Get(params.AuthFile, params.HostAddr); err == nil {
			var data, _ = json.Marshal(params.Solution)

			state.Logger.Debugf("Pretending to publish solution to [%s]", params.HostAddr)
			fmt.Printf("curl -X POST -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("  -d solution=%s\\\n", string(data))
			fmt.Printf("%s\n", brokerClient.PublishSolutionUrl())
			return nil
		}

		return err
	}

	return state.Error
}

func handleSolutionUpdateConfig(writer SolutionWriter, params *SolutionParams, state *task.State) error {
	if state.Output != "" {
		if params.HasWorkflows() {
			// Remove existing workflows from createConfig, we can not overwrite them with solution update
			slices.DeleteFunc(params.Workflows, func(wf Workflow) bool {
				var _, err = writer.GetWorkflowByHandle(wf.Handle)

				return err == nil
			})
		}

		state.Completed = true
		return handleSolutionCreateConfig(writer, params, state)
	}

	return errorSolutionFileInvalid
}

func handleSolutionUpdatePretend(writer SolutionWriter, params *SolutionParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)
		var rootKey, solution = writer.WithSolution(params.Solution).
			BuildSolution()

		state.Logger.Debugf("Pretending to update solution %s", params.Name)
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s.description", rootKey), solution.Description
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s.handle", rootKey), solution.Handle
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s.name", rootKey), solution.Name
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s.oem", rootKey), solution.Oem
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s.version", rootKey), solution.Version
		})
		return nil
	}

	return errorSolutionFileInvalid
}

func handleWorkflowCreateConfig(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		state.Logger.Debugf("Workflow writing to [%s]", state.Output)
		return writer.WithWorkflow(params.Workflow).
			Write(state.Output)
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowCreateContext(params *WorkflowParams, state *task.State) error {
	if state.Output == "" {
		var err error

		state.Logger.Debugf("Finding a workflow configuration file for writing")

		if state.Output, err = params.ResolveConfigPath(); err != nil {
			return err
		}
	}

	return nil
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

	return errorWorkflowFileInvalid
}

func handleWorkflowDeleteContext(params *WorkflowParams, state *task.State) error {
	var err = handleWorkflowCreateContext(params, state)

	if errors.Is(err, shared.ErrorConfigFileExists) {
		return nil
	}

	if err == nil {
		return errorWorkflowFileNotFound
	}

	return err
}

func handleWorkflowDeletePretend(params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)
		var handle = params.GetHandle()
		var name = params.GetName()

		if handle != "" {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return "solution.workflows[]", "handle", handle
			})
		} else if name != "" {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return "solution.workflows[]", "name", name
			})
		} else {
			return errorWorkflowNotFound
		}

		return nil
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowUpdateConfig(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var update *Workflow
		var err error

		if params.Workflow != nil {
			state.Completed = true

			if update, err = writer.GetWorkflowByHandle(params.Handle); err == nil {
				if params.WorkflowUpdate {
					state.Logger.Debugf("Workflow [%s] updated to [%s]", update.Handle, state.Output)

					if params.Description != "" {
						update.Description = params.Description
					}

					if params.Name != "" {
						update.Name = params.Name
					}

					err = writer.WithWorkflows(writer.GetWorkflows()).
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

		return errors.New("no workflow found in params")
	}

	return errorWorkflowFileInvalid
}

func handleWorkflowUpdatePretend(writer WorkflowWriter, params *WorkflowParams, state *task.State) error {
	if state.Output != "" {
		var comparison = params.workflowPredicate()
		var pretender = shared.NewConfigPretender(state.Output)
		var rootKey, workflows = writer.WithWorkflows(writer.GetWorkflows()).
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
					return fmt.Sprintf("%s[%d].name", rootKey, i), params.Name
				})
			}

			if !strings.EqualFold(wf.Description, params.Description) || !params.WorkflowUpdate {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].description", rootKey, i), params.Description
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
