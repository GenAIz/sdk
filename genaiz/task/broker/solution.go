package broker

import (
	"encoding/json"
	"errors"
	"fmt"
	"maps"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorSolutionFileInvalid    = task.NewError("solution config is invalid")
	errorSolutionInvalid        = task.NewError("solution is invalid")
	errorSolutionOemRequired    = task.NewError("solution oem is required")
	errorWorkflowConflict       = task.NewError("workflow already exists")
	errorWorkflowFileInvalid    = task.NewError("workflow config is invalid")
	errorWorkflowFileNotFound   = task.NewError("workflow config file not found")
	errorWorkflowNotFound       = task.NewError("workflow not found")
	ErrorWorkflowPropIncomplete = task.NewError("workflow prop specs are empty")

	errorInvalidNodeProp = func(key, handle string) error {
		return fmt.Errorf("the key [%s] is invalid for node [%s]", strings.ToUpper(key), handle)
	}
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

type SolutionListParams struct {
	Broker
	AccountOnly bool
	Local       []Solution
	Oem         string
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

type WorkflowPropParams struct {
	*Workflow
	VarSpecs []shared.VarSpec
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

func NewSolutionListTask() *task.Task[SolutionListParams] {
	return &task.Task[SolutionListParams]{
		Name:         "solution-list",
		OnPrepare:    handleSolutionListContext,
		OnComplete:   handleSolutionListComplete,
		OnIncomplete: handleSolutionListIncomplete,
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

func NewSolutionReduceTask() *task.Task[SolutionListParams] {
	return &task.Task[SolutionListParams]{
		Name:       "solution-reduce",
		OnPrepare:  handleSolutionReduceContext,
		OnComplete: handleSolutionReduceComplete,
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

func NewWorkflowPropTask() *task.Task[WorkflowPropParams] {
	return &task.Task[WorkflowPropParams]{
		Name:         "solution-prop-workflow",
		OnPrepare:    handleWorkflowPropContext,
		OnComplete:   handleWorkflowPropComplete,
		OnIncomplete: handleWorkflowPropIncomplete,
		OnPretend:    handleWorkflowPropPretend,
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
		var err error

		state.Logger.Debugf("Solution writing to [%s]", state.Output)

		if err = writer.WithSolution(params.Solution).Write(state.Output); err == nil {
			var configDir = filepath.Dir(params.GetConfigPath())

			state.Report(fmt.Sprintf("Created solution %s under folder %s", params.Handle, configDir))
			return nil
		}

		return err
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

func handleSolutionListComplete(params *SolutionListParams, state *task.State) error {
	if state.Internal == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var results []Solution

			state.Logger.Debugf("List solutions from [%s] for oem [%s]", brokerClient.ListSolutionsUrl(), params.Oem)

			if results, err = brokerClient.ListSolutions(params.Oem); err == nil {
				state.Logger.Debugf("Found [%d] solutions", len(results))
				state.Internal = results
				return nil
			}
		}

		return err
	}

	return nil
}

func handleSolutionListContext(params *SolutionListParams, state *task.State) error {
	if state.Output == "" {
		if params.Oem == "" {
			return errorSolutionOemRequired
		}

		if brokerClient, err := params.GetClient(); err == nil {
			if params.AccountOnly {
				state.Logger.Debugf("Account only solution will be listed")

				if len(params.Local) > 0 {
					state.Logger.Warn("Local solutions provided will be ignored")
				}
			}

			state.Output = brokerClient.GetHostAddr()
		} else {
			return err
		}
	}

	return nil
}

func handleSolutionListIncomplete(params *SolutionListParams, state *task.State) error {
	if state.Error != nil {
		if params.AccountOnly {
			state.Logger.Errorf("Could not list solution for account only")
			return state.Error
		}
	}

	state.Completed = true
	return nil
}

func handleSolutionPublishComplete(params *SolutionPublishParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var solution = params.Solution
		var identity *shared.Identity

		state.Logger.Infof("Publishing solution [%s], version [%s]", params.Handle, params.Version)

		if identity, err = brokerClient.PublishSolution(solution); err == nil {
			state.Report(fmt.Sprintf("Published solution %s, version %s to %s", identity.Path, identity.Version, brokerClient.GetHostAddr()))
			return nil
		}

		return err
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

			for _, l := range w.Links {
				if l.RhsNodePort == "" {
					return fmt.Errorf("the workflow link %s requires a right data port", l.String())
				}
			}
		}

		for _, sf := range allNodeFunctions {
			if !params.HasProvision(sf.Oem, sf.Handle) {
				// May have to extend this eventually, but for now this check is purely for logging purposes
				state.Logger.Warnf("the function %s/%s could not be found within the solution", sf.Oem, sf.Handle)
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

		if brokerClient, err = params.GetClient(); err == nil {
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

func handleSolutionReduceComplete(params *SolutionListParams, state *task.State) error {
	var branchMap = map[string]Solution{}
	var result []Solution

	if remotes, hasInternal := state.Internal.([]Solution); hasInternal {
		state.Logger.Debugf("Reducing [%d] account solutions", len(remotes))

		for _, sol := range remotes {
			if sol.IsReleased() {
				result = append(result, sol)
				branchMap[sol.GetBranch()] = sol
			} else if latest, ok := branchMap[sol.GetBranch()]; !ok || sol.IsAfter(latest) {
				branchMap[sol.GetBranch()] = sol
			}
		}

		for _, sol := range slices.Collect(maps.Values(branchMap)) {
			if !sol.IsReleased() {
				if _, ok := branchMap[sol.GetBranch()]; ok {
					result = append(result, sol)
				}
			}
		}
	}

	for _, sol := range params.Local {
		if _, ok := branchMap[sol.GetBranch()]; !ok {
			result = append(result, sol)
		}
	}

	state.Internal = result
	return nil
}

func handleSolutionReduceContext(params *SolutionListParams, state *task.State) error {
	var count = len(params.Local)

	if count > 0 {
		state.Logger.Debugf("Reducing [%d] local solutions", count)
	}

	return nil
}

func handleSolutionUpdateConfig(writer SolutionWriter, params *SolutionParams, state *task.State) error {
	if state.Output != "" {
		var err error

		if params.HasWorkflows() {
			// Remove existing workflows from createConfig, we can not overwrite them with solution update
			slices.DeleteFunc(params.Workflows, func(wf Workflow) bool {
				var _, err2 = writer.GetWorkflowByHandle(wf.Handle)

				return err2 == nil
			})
		}

		state.Logger.Debugf("Solution updating to [%s]", state.Output)
		state.Completed = true

		if err = writer.WithSolution(params.Solution).Write(state.Output); err == nil {
			var configDir = filepath.Dir(params.GetConfigPath())

			state.Report(fmt.Sprintf("Updated solution %s under folder %s", params.Handle, configDir))
			return nil
		}

		return err
	}

	return errorSolutionFileInvalid
}

func handleSolutionUpdatePretend(writer SolutionWriter, params *SolutionParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)
		var rootKey, solution = writer.WithSolution(params.Solution).
			BuildSolution()

		if state.Error == nil {
			var confPath = filepath.Dir(state.Output)

			fmt.Printf("mkdir -p %s && cd %s\n", confPath, confPath)
			fmt.Printf("touch %s\n", state.Output)
		}

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
		var err error

		state.Logger.Debugf("Workflow writing to [%s]", state.Output)

		if err = writer.WithWorkflow(params.Workflow).Write(state.Output); err == nil {
			var configDir = filepath.Dir(params.GetConfigPath())

			state.Report(fmt.Sprintf("Created workflow %s under folder %s", params.Handle, configDir))
			return nil
		}

		return err
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

		if err := writer.WithWorkflows(updated).Write(state.Output); err == nil {
			state.Report(fmt.Sprintf("Removed workflow %s under folder %s", params.Handle, state.Output))
			return nil
		} else {
			return err
		}
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

func handleWorkflowPropContext(params *WorkflowPropParams, state *task.State) error {
	if params.Workflow != nil {
		var varSpecClient = shared.NewVarSpecState(state)

		state.Logger.Debugf("Validating workflow properties for workflow [%s]", params.Workflow.Handle)
		varSpecClient.AddSpecs(params.VarSpecs)

		if params.Workflow.HasNodeProps() && len(varSpecClient.VarSpecs) == 0 {
			// On an incomplete set, we only need to validate if there are any props on any nodes
			return ErrorWorkflowPropIncomplete
		}

		return nil
	}

	return ErrorWorkflowNotFound
}

func handleWorkflowPropComplete(params *WorkflowPropParams, state *task.State) error {
	if params.Workflow != nil {
		var varSpecClient = shared.NewVarSpecState(state)
		var err error

		for _, node := range params.Workflow.Nodes {
			state.Logger.Debugf("Validating props for [%s]", node.Handle)

			if err = node.ValidateProps(varSpecClient.VarSpecs); err != nil {
				return err
			}
		}

		return nil
	}

	return ErrorWorkflowNotFound
}

func handleWorkflowPropIncomplete(params *WorkflowPropParams, state *task.State) error {
	if params.Workflow != nil {
		state.Logger.Debugf("Properties for workflow [%s] are incomplete", params.Workflow.Handle)

		for _, node := range params.Workflow.Nodes {
			if len(node.Props) > 0 {
				for k := range node.Props {
					state.Completed = true
					return errorInvalidNodeProp(k, node.Handle)
				}
			}
		}

		return nil
	}

	return ErrorWorkflowNotFound
}

func handleWorkflowPropPretend(params *WorkflowPropParams, state *task.State) error {
	if params.Workflow != nil {
		state.Logger.Debugf("Pretending validation on Workflow Node props")

		for _, node := range params.Workflow.Nodes {
			if len(node.Props) > 0 {
				for k := range node.Props {
					state.Logger.Debugf("Validating prop [%s] on node [%s]", k, node.Handle)
				}
			}
		}

		return nil
	}

	return ErrorWorkflowNotFound
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

					if err = writer.WithWorkflows(writer.GetWorkflows()).
						WithWorkflowLinks(update.Handle, params.Links).
						WithWorkflowNodes(update.Handle, params.Nodes).
						Write(state.Output); err == nil {
						var configDir = filepath.Dir(params.GetConfigPath())

						state.Report(fmt.Sprintf("Updated workflow %s under folder %s", params.Handle, configDir))
						return nil
					}

					return err
				}

				state.Logger.Errorf("Workflow [%s] already exist", params.Name)
				return errorWorkflowConflict
			}

			return handleWorkflowCreateConfig(writer.WithWorkflows(writer.GetWorkflows()), params, state)
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

			pretendWorkflowLinksUpdate(&wf, pretender, rootKey, i, params.WorkflowUpdate)
			pretendWorkflowNodesUpdate(&wf, pretender, rootKey, i, params.WorkflowUpdate)
		} else {
			if params.Workflow.Handle != "" {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].handle", rootKey, 0), params.Handle
				})
			}

			if params.Workflow.Name != "" {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", rootKey, 0), params.Name
				})
			}

			if params.Workflow.Description != "" {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].description", rootKey, 0), params.Description
				})
			}

			pretendWorkflowLinksUpdate(params.Workflow, pretender, rootKey, 0, false)
			pretendWorkflowNodesUpdate(params.Workflow, pretender, rootKey, 0, false)
		}

		state.Output = ""
		return nil
	}

	return errorWorkflowFileInvalid
}

func pretendWorkflowLinksUpdate(wf *Workflow, pretender shared.ConfigPretender, rootKey string, index int, update bool) {
	if update {
		shared.PretendDelete(pretender, func() string {
			return fmt.Sprintf("%s[%d].links", rootKey, index)
		})
	}

	for j, link := range wf.Links {
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].links[%d].lhsNode", rootKey, index, j), link.LhsNode
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].links[%d].lhsNodePort", rootKey, index, j), link.LhsNodePort
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].links[%d].rhsNode", rootKey, index, j), link.RhsNode
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].links[%d].rhsNodePort", rootKey, index, j), link.RhsNodePort
		})
	}
}

func pretendWorkflowNodesUpdate(wf *Workflow, pretender shared.ConfigPretender, rootKey string, index int, update bool) {
	if update {
		shared.PretendDelete(pretender, func() string {
			return fmt.Sprintf("%s[%d].nodes", rootKey, index)
		})
	}

	for j, node := range wf.Nodes {
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].nodes[%d].name", rootKey, index, j), node.Name
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].nodes[%d].description", rootKey, index, j), node.Description
		})
		shared.PretendValue(pretender, func() (string, string) {
			return fmt.Sprintf("%s[%d].nodes[%d].handle", rootKey, index, j), node.Handle
		})

		if node.Sf != nil {
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].nodes[%d].sf.oem", rootKey, index, j), node.Sf.Oem
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].nodes[%d].sf.handle", rootKey, index, j), node.Sf.Handle
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].nodes[%d].sf.version", rootKey, index, j), node.Sf.Version
			})
			shared.PretendValue(pretender, func() (string, string) {
				return fmt.Sprintf("%s[%d].nodes[%d].sf.seq", rootKey, index, j), cast.ToString(node.Sf.Seq)
			})
		}
	}
}
