package config

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"

	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task/broker"
)

type BaseWriter struct {
	current *broker.Solution
}

func (bw *BaseWriter) GetWorkflowByHandle(handle string) (*broker.Workflow, error) {
	var workflows = bw.GetWorkflows()

	if i := slices.IndexFunc(workflows, broker.WorkflowHandlePredicate(handle)); i >= 0 {
		return &workflows[i], nil
	}

	return nil, fmt.Errorf("workflow handle [%s] not found", handle)
}

func (bw *BaseWriter) GetWorkflows() []broker.Workflow {
	if bw.current == nil {
		return nil
	}

	return bw.current.Workflows
}

func (bw *BaseWriter) Read(ledger *Ledger, input string) *BaseWriter {
	var document SolutionDocument

	if bytes, err := os.ReadFile(input); err == nil {
		if err = yaml.Unmarshal(bytes, &document); err != nil {
			ledger.Logger.Error(err)
		}
	}

	bw.current = document.Solution
	return bw
}

func (bw *BaseWriter) UpdatePath(vp *viper.Viper, path string) (*os.File, error) {
	vp.SetConfigFile(path)

	if err := vp.ReadInConfig(); err != nil {
		var pathError *os.PathError

		if !errors.As(err, &pathError) {
			return nil, err
		}
	}

	return filez.CreateRecursive(filepath.Dir(path), filepath.Base(path))
}

type SolutionDocument struct {
	Solution *broker.Solution
}

type SolutionWriter struct {
	BaseWriter
	updated   *broker.Solution
	order     []string
	workflows map[string]broker.Workflow
}

func (sw *SolutionWriter) BuildSolution() (string, broker.Solution) {
	var result *broker.Solution

	if sw.current == nil {
		result = sw.updated
	} else if sw.updated == nil {
		result = sw.current
	} else {
		result = sw.current.Merge(*sw.updated)
	}

	if len(sw.workflows) > 0 {
		var updatedWf []broker.Workflow

		for _, wf := range result.Workflows {
			updatedWf = append(updatedWf, wf)
		}

		for _, key := range sw.order {
			if !slices.ContainsFunc(updatedWf, broker.WorkflowHandlePredicate(key)) {
				updatedWf = append(updatedWf, sw.workflows[key])
			}
		}

		result.Workflows = updatedWf
	}

	return "solution", *result
}

func (sw *SolutionWriter) Read(ledger *Ledger, input string) *SolutionWriter {
	sw.BaseWriter.Read(ledger, input)
	return sw
}

func (sw *SolutionWriter) WithCurrent(current *broker.Solution) *SolutionWriter {
	sw.current = current
	return sw
}

func (sw *SolutionWriter) WithSolution(solution *broker.Solution) broker.SolutionWriter {
	sw.updated = solution
	return sw
}

func (sw *SolutionWriter) WithWorkflow(workflow *broker.Workflow) broker.SolutionWriter {
	sw.workflows[workflow.Handle] = *workflow

	if !slices.Contains(sw.order, workflow.Handle) {
		sw.order = append(sw.order, workflow.Handle)
	}

	return sw
}

func (sw *SolutionWriter) Write(path string) error {
	var vp = viper.New()
	var output *os.File
	var err error

	if output, err = sw.UpdatePath(vp, path); err == nil {
		defer filez.CloseSilently(output)
		var key, solution = sw.BuildSolution()

		vp.Set(key, solution)
		return vp.WriteConfigTo(output)
	}

	return err
}

func NewSolutionWriter() *SolutionWriter {
	return &SolutionWriter{
		workflows: make(map[string]broker.Workflow),
	}
}

type WorkflowWriter struct {
	BaseWriter
	root      string
	links     map[string][]broker.WorkflowLink
	nodes     map[string][]broker.WorkflowNode
	order     []string
	workflows map[string]broker.Workflow
}

func (ww *WorkflowWriter) BuildWorkflows() (string, []broker.Workflow) {
	var result []broker.Workflow

	for _, handle := range ww.order {
		var workflow = ww.workflows[handle]

		if links, ok := ww.links[handle]; ok {
			workflow.Links = links
		}

		if nodes, ok := ww.nodes[handle]; ok {
			workflow.Nodes = nodes
		}

		result = append(result, workflow)
	}

	return ww.root, result
}

func (ww *WorkflowWriter) Read(ledger *Ledger, input string) *WorkflowWriter {
	ww.BaseWriter.Read(ledger, input)
	return ww
}

func (ww *WorkflowWriter) WithCurrent(current *broker.Solution) *WorkflowWriter {
	ww.current = current
	return ww
}

func (ww *WorkflowWriter) WithWorkflow(workflow *broker.Workflow) broker.WorkflowWriter {
	ww.workflows[workflow.Handle] = *workflow

	if !slices.Contains(ww.order, workflow.Handle) {
		ww.order = append(ww.order, workflow.Handle)
	}

	return ww
}

func (ww *WorkflowWriter) WithWorkflows(workflows []broker.Workflow) broker.WorkflowWriter {
	for _, workflow := range workflows {
		ww.WithWorkflow(&workflow)
	}

	return ww
}

func (ww *WorkflowWriter) WithWorkflowLinks(handle string, links []broker.WorkflowLink) broker.WorkflowWriter {
	ww.links[handle] = links
	return ww
}

func (ww *WorkflowWriter) WithWorkflowNodes(handle string, nodes []broker.WorkflowNode) broker.WorkflowWriter {
	ww.nodes[handle] = nodes
	return ww
}

func (ww *WorkflowWriter) Write(path string) error {
	var vp = viper.New()
	var output *os.File
	var err error

	if output, err = ww.UpdatePath(vp, path); err == nil {
		defer filez.CloseSilently(output)
		var key, workflows = ww.BuildWorkflows()

		vp.Set(key, workflows)
		return vp.WriteConfigTo(output)
	}

	return err
}

func NewWorkflowWriter() *WorkflowWriter {
	return &WorkflowWriter{
		root:      "solution.workflows",
		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
}
