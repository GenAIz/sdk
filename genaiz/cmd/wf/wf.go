// Package wf provides commands for managing Genaiz Solution Workflows.
// Workflow commands include create, delete, links and nodes.
//
// See: genaiz wf --help
package wf

import (
	"context"
	"slices"
	"strings"

	"github.com/spf13/cobra"

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

type workflowWriter struct {
	*config.WorkflowWriter
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
	return &workflowWriter{
		WorkflowWriter: config.NewWorkflowWriter().
			Read(ledger, output),
	}
}
