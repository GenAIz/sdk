// Package wf provides commands for managing Genaiz Solution Workflows.
// Workflow commands include create, delete, links and nodes.
//
// See: genaiz wf --help
package wf

import (
	"context"
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
	"gopkg.in/yaml.v3"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type WorkflowTaskFactory func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams]
type workflowWriterFactory func(*config.Ledger, string) *workflowWriter

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

func (be BaseExecutor) makeConfigParams(typeOption *config.StringOption) (*shared.ConfigParams, error) {
	var configType *shared.ConfigType
	var err error

	if configType, err = be.Ledger.GetConfigType(typeOption); err == nil {
		return &shared.ConfigParams{
			ConfigName: be.Ledger.ConfigName,
			ConfigType: configType,
		}, nil
	}

	return nil, err
}

type Cli struct {
	cli.BaseCli
}

func NewWf(ledger *config.Ledger, confirm cli.Interactive, dry, pretend cli.Decisive) *cobra.Command {
	var wfCli = NewWfCli(confirm, dry, pretend)
	var wfCmd = &cobra.Command{
		Use:     "workflow",
		Aliases: []string{"wf"},
		Short:   "Genaiz Workflow Toolkit",
	}

	wfCmd.AddCommand(NewCreate(ledger, wfCli))
	wfCmd.AddCommand(NewDelete(ledger, wfCli))
	wfCmd.AddCommand(NewLinks(ledger, wfCli))
	wfCmd.AddCommand(NewNodes(ledger, wfCli))
	return wfCmd
}

func NewWfCli(confirm cli.Interactive, dry, pretend cli.Decisive) *Cli {
	return &Cli{
		BaseCli: cli.BaseCli{
			Confirm: confirm,
			Dry:     dry,
			Pretend: pretend,
		},
	}
}

type workflowDocument struct {
	Solution *broker.Solution
}

type workflowWriter struct {
	current *broker.Solution
	logger  *logrus.Logger
	root    string

	links     map[string][]broker.WorkflowLink
	nodes     map[string][]broker.WorkflowNode
	order     []string
	workflows map[string]broker.Workflow
}

func (w *workflowWriter) BuildWorkflows() (string, []broker.Workflow) {
	var result []broker.Workflow

	for _, handle := range w.order {
		var workflow = w.workflows[handle]

		if links, ok := w.links[handle]; ok {
			workflow.Links = links
		}

		if nodes, ok := w.nodes[handle]; ok {
			workflow.Nodes = nodes
		}

		result = append(result, workflow)
	}

	return w.root, result
}

func (w *workflowWriter) GetWorkflowByHandle(handle string) (*broker.Workflow, error) {
	var workflows = w.GetWorkflows()

	if i := slices.IndexFunc(workflows, broker.WorkflowHandlePredicate(handle)); i >= 0 {
		return &workflows[i], nil
	}

	return nil, fmt.Errorf("workflow handle [%s] not found", handle)
}

func (w *workflowWriter) GetWorkflows() []broker.Workflow {
	if w.current == nil {
		return nil
	}

	return w.current.Workflows
}

func (w *workflowWriter) WithWorkflow(workflow *broker.Workflow) broker.WorkflowWriter {
	w.workflows[workflow.Handle] = *workflow

	if !slices.Contains(w.order, workflow.Handle) {
		w.order = append(w.order, workflow.Handle)
	}

	return w
}

func (w *workflowWriter) WithWorkflows(workflows []broker.Workflow) broker.WorkflowWriter {
	for _, workflow := range workflows {
		w.WithWorkflow(&workflow)
	}

	return w
}

func (w *workflowWriter) WithWorkflowLinks(handle string, links []broker.WorkflowLink) broker.WorkflowWriter {
	w.links[handle] = links
	return w
}

func (w *workflowWriter) WithWorkflowNodes(handle string, nodes []broker.WorkflowNode) broker.WorkflowWriter {
	w.nodes[handle] = nodes
	return w
}

func (w *workflowWriter) Write(path string) error {
	var vp = viper.New()
	var output *os.File
	var err error

	vp.SetConfigFile(path)

	if err = vp.ReadInConfig(); err != nil {
		var pathError *os.PathError

		if !errors.As(err, &pathError) {
			return err
		}
	}

	if output, err = filez.CreateRecursive(filepath.Dir(path), filepath.Base(path)); err == nil {
		defer filez.CloseSilently(output)
		var key, workflows = w.BuildWorkflows()

		vp.Set(key, workflows)
		return vp.WriteConfigTo(output)
	}

	return err
}

func (w *workflowWriter) addLinks(handle string, links []broker.WorkflowLink) broker.WorkflowWriter {
	if workflow, err := w.GetWorkflowByHandle(handle); err == nil {
		for _, add := range links {
			var predicate = func(link broker.WorkflowLink) bool {
				return link.Equals(add)
			}

			if !slices.ContainsFunc(workflow.Links, predicate) {
				workflow.Links = append(workflow.Links, add)
			}
		}
	}

	return w
}

func (w *workflowWriter) addNodes(handle string, nodes ...*broker.WorkflowNode) broker.WorkflowWriter {
	if workflow, err := w.GetWorkflowByHandle(handle); err == nil {
		for _, add := range nodes {
			var predicate = func(node broker.WorkflowNode) bool {
				return node.Equals(*add)
			}

			if !slices.ContainsFunc(workflow.Nodes, predicate) {
				workflow.Nodes = append(workflow.Nodes, *add)
			}
		}
	}

	return w
}

func (w *workflowWriter) removeLinks(handle string, links []broker.WorkflowLink) broker.WorkflowWriter {
	if workflow, err := w.GetWorkflowByHandle(handle); err == nil {
		var removedIndices = map[int]struct{}{}

		for _, rm := range links {
			var predicate = func(link broker.WorkflowLink) bool {
				return link.Equals(rm)
			}

			if i := slices.IndexFunc(workflow.Links, predicate); i >= 0 {
				removedIndices[i] = struct{}{}
			}
		}

		if len(removedIndices) > 0 {
			var updated []broker.WorkflowLink

			for i, wfl := range workflow.Links {
				if _, ok := removedIndices[i]; !ok {
					updated = append(updated, wfl)
				}
			}

			workflow.Links = updated
		}
	}

	return w
}

func (w *workflowWriter) removeNodes(handle string, nodes ...string) broker.WorkflowWriter {
	if workflow, err := w.GetWorkflowByHandle(handle); err == nil {
		var removedIndices = map[int]struct{}{}

		for _, rm := range nodes {
			var predicate = func(node broker.WorkflowNode) bool {
				return strings.EqualFold(node.Handle, rm)
			}

			if i := slices.IndexFunc(workflow.Nodes, predicate); i >= 0 {
				removedIndices[i] = struct{}{}
			}
		}

		if len(removedIndices) > 0 {
			var updated []broker.WorkflowNode

			for i, wfn := range workflow.Nodes {
				if _, ok := removedIndices[i]; !ok {
					updated = append(updated, wfn)
				}
			}

			workflow.Nodes = updated
		}
	}

	return w
}

func newWorkflowWriter(ledger *config.Ledger, output string) *workflowWriter {
	var document workflowDocument

	if bytes, err := os.ReadFile(output); err == nil {
		if err = yaml.Unmarshal(bytes, &document); err != nil {
			ledger.Logger.Error(err)
		}
	}

	return &workflowWriter{
		current: document.Solution,
		logger:  ledger.Logger,
		root:    "solution.workflows",

		links:     make(map[string][]broker.WorkflowLink),
		nodes:     make(map[string][]broker.WorkflowNode),
		workflows: make(map[string]broker.Workflow),
	}
}
