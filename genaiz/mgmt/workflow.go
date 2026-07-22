package mgmt

import (
	"encoding/json"
	"strconv"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/intz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type UserWorkflowFacade interface {
	Facade[[]UserWorkflow, broker.WorkflowListParams]

	WithPathGraphers(string, map[string]broker.SolutionGrapher) UserWorkflowFacade
}

type solutionCollectTaskFactory func() *task.Task[broker.SolutionCollectParams]
type workflowListTaskFactory func() *task.Task[broker.WorkflowListParams]

type UserWorkflow struct {
	Id              *int64 `cli:"Id"`
	SolutionId      *int64
	SolutionFqdn    string `cli:"Solution"`
	SolutionVersion string `cli:"Version"`
	Handle          string `cli:"Handle"`
	Created         int64  `cli:"Created"`
	Modified        int64
	Flags           *int
	Name            string
	Description     string
	Nodes           []broker.WorkflowNode `cli:"Nodes"`
	Links           []broker.WorkflowLink
	Local           bool `cli:"Local?"`

	matched string
}

func (uw UserWorkflow) MarshalJSON() ([]byte, error) {
	var created, modified string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	}

	if uw.Modified > 0 {
		modified = createdFormatter.FormatMillis(uw.Modified)
	}

	return json.Marshal(&struct {
		Id          *int64                `json:"id,omitempty"`
		Handle      string                `json:"handle"`
		Created     string                `json:"created,omitempty"`
		Modified    string                `json:"modified,omitempty"`
		Flags       *int                  `json:"flags,omitempty"`
		Name        string                `json:"name,omitempty"`
		Description string                `json:"description,omitempty"`
		Nodes       []broker.WorkflowNode `json:"nodes,omitempty"`
		Links       []broker.WorkflowLink `json:"links,omitempty"`
	}{
		Id:          uw.Id,
		Handle:      uw.Handle,
		Created:     created,
		Modified:    modified,
		Flags:       uw.Flags,
		Name:        uw.Name,
		Description: uw.Description,
		Nodes:       uw.Nodes,
		Links:       uw.Links,
	})
}

func (uw UserWorkflow) MarshalSlice() ([]string, error) {
	var id, created string

	if uw.Created > 0 {
		created = createdFormatter.FormatMillis(uw.Created)
	} else {
		created = "-"
	}

	if uw.Id != nil {
		id = cast.ToString(*uw.Id)
	}

	return []string{
		id,
		uw.SolutionFqdn,
		uw.SolutionVersion,
		uw.Handle,
		created,
		cast.ToString(len(uw.Nodes)),
		stringz.YesOrNo(uw.Local),
	}, nil
}

func (uw UserWorkflow) Match(filter string) *UserWorkflow {
	var matched string

	if strings.EqualFold(uw.Handle, filter) ||
		strings.HasPrefix(uw.Handle, filter) {
		matched = uw.Handle
	} else if uw.Id != nil &&
		strings.HasPrefix(cast.ToString(*uw.Id), filter) {
		matched = cast.ToString(*uw.Id)
	}

	if matched == "" {
		return nil
	}

	return &UserWorkflow{
		Id:              uw.Id,
		SolutionId:      uw.SolutionId,
		SolutionFqdn:    uw.SolutionFqdn,
		SolutionVersion: uw.SolutionVersion,
		Handle:          uw.Handle,
		Created:         uw.Created,
		Modified:        uw.Modified,
		Flags:           uw.Flags,
		Name:            uw.Name,
		Description:     uw.Description,
		Nodes:           uw.Nodes,
		Links:           uw.Links,
		Local:           uw.Local,
		matched:         matched,
	}
}

func (uw UserWorkflow) Matched() string {
	if uw.matched == "" {
		return cobra.CompletionWithDesc(cast.ToString(*uw.Id), uw.Handle)
	}

	if _, err := strconv.Atoi(uw.matched); err == nil {
		return cobra.CompletionWithDesc(uw.matched, uw.Handle)
	}

	return cobra.CompletionWithDesc(uw.matched, cast.ToString(*uw.Id))
}

func ToUserWorkflow(solution *broker.Solution, workflow *broker.Workflow) *UserWorkflow {
	return &UserWorkflow{
		SolutionId:      solution.Id,
		SolutionFqdn:    solution.GetFqdn(),
		SolutionVersion: solution.GetVersion(),
		Id:              workflow.Id,
		Created:         intz.Int64ToDefault(workflow.Created, 0),
		Modified:        intz.Int64ToDefault(workflow.Modified, 0),
		Flags:           workflow.Flags,
		Handle:          workflow.Handle,
		Name:            workflow.Name,
		Description:     workflow.Description,
		Nodes:           workflow.Nodes,
		Links:           workflow.Links,
		Local:           workflow.Created == nil,
	}
}

type userWorkflowFacade struct {
	baseLoggingFacade
	graphers map[string]broker.SolutionGrapher
	path     string
	params   *broker.WorkflowListParams
}

func (uwf *userWorkflowFacade) Filtering(filter string) Provider[[]UserWorkflow] {
	return &userWorkflowProvider{
		Plan: task.Plan{
			Logger: uwf.logger,
		},
		filter:   filter,
		graphers: uwf.graphers,
		params:   uwf.params,
		path:     uwf.path,

		solutionCollectTasFactory: broker.NewSolutionCollectTask,
		workflowListTaskFactory:   broker.NewWorkflowListTask,
	}
}

func (uwf *userWorkflowFacade) Provider() Provider[[]UserWorkflow] {
	return uwf.Filtering("")
}

func (uwf *userWorkflowFacade) WithLogger(logger *logrus.Logger) Facade[[]UserWorkflow, broker.WorkflowListParams] {
	uwf.logger = logger
	return uwf
}

func (uwf *userWorkflowFacade) WithParams(params *broker.WorkflowListParams) Facade[[]UserWorkflow, broker.WorkflowListParams] {
	uwf.params = params
	return uwf
}

func (uwf *userWorkflowFacade) WithPathGraphers(path string, graphers map[string]broker.SolutionGrapher) UserWorkflowFacade {
	uwf.path = path
	uwf.graphers = graphers
	return uwf
}

type userWorkflowProvider struct {
	task.Plan
	filter   string
	graphers map[string]broker.SolutionGrapher
	params   *broker.WorkflowListParams
	path     string

	solutionCollectTasFactory solutionCollectTaskFactory
	workflowListTaskFactory   workflowListTaskFactory
}

func (uwp userWorkflowProvider) Get() ([]UserWorkflow, task.Error) {
	var solutions []broker.Solution
	var workers []task.Worker
	var failure interface{}

	uwp.OnReturn = func(i interface{}) { solutions = i.([]broker.Solution) }
	uwp.OnFailure = func(i interface{}) { failure = i }

	if uwp.path == "" {
		workers = append(workers, task.NewWorker(uwp.params, uwp.workflowListTaskFactory()))
	} else {
		var collectParams = &broker.SolutionCollectParams{
			Path:     uwp.path,
			Graphers: uwp.graphers,
		}

		workers = append(workers, task.NewWorker(collectParams, uwp.solutionCollectTasFactory()))
	}

	uwp.Sequence(workers...)

	if failure == nil {
		var result []UserWorkflow

		for _, sol := range solutions {
			for _, wf := range sol.Workflows {
				var userWorkflow = ToUserWorkflow(&sol, &wf)

				if uwp.filter == "" {
					result = append(result, *userWorkflow)
				} else if matchedWorkflow := userWorkflow.Match(uwp.filter); matchedWorkflow != nil {
					result = append(result, *matchedWorkflow)
				}
			}
		}

		return result, nil
	}

	return nil, task.NewFailure(failure)
}

func NewUserWorkflowFacade() UserWorkflowFacade {
	return &userWorkflowFacade{
		baseLoggingFacade: baseLoggingFacade{},
	}
}
