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
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorNodeExists = errors.New("the node specified already exists")
)

type WorkflowTaskFactory func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams]
type workflowWriterFactory func(*config.Ledger, string) *workflowWriter

type BaseExecutor struct {
	Cli     *Cli
	Context context.Context
	Ledger  *config.Ledger
}

func (be BaseExecutor) makeConfigParams(typeOption *config.StringOption) (*shared.ConfigParams, error) {
	var configType, _ = be.Ledger.GetConfigType(typeOption)
	var result = &shared.ConfigParams{
		ConfigName: be.Ledger.ConfigName,
		ConfigType: configType,
	}
	var err error

	if result.IsConfigTypeNone() {
		var workingConfig string

		if workingConfig, err = filez.FirstNamedFile(result.ConfigName); err == nil {
			var fileType = filez.GetFileType(workingConfig)

			result.ConfigType, err = shared.ConfigTypes.FromString(fileType)
		}
	}

	if err == nil {
		return result, nil
	} else {
		var wd, _ = os.Getwd()

		err = fmt.Errorf("could not find local config [%s] under [%s]", result.ConfigName, wd)
	}

	return nil, err
}

type Cli struct {
	cli.BaseCli
}

func (c Cli) WorkingConfigType() func(*config.Ledger) any {
	return func(ledger *config.Ledger) any {
		var solutionReader = config.NewSolutionReader(ledger).
			WithConfigPath(ledger.WorkDir)

		if _, err := solutionReader.ReadName(ledger.ConfigName); err == nil {
			return solutionReader.GetConfigType()
		} else if !errorz.IsPathError(err) {
			lang.HandleExit(err)
			return nil
		}

		return shared.ConfigTypeYaml
	}
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

func (w *workflowWriter) addNodes(handle string, nodes ...*broker.WorkflowNode) (broker.WorkflowWriter, error) {
	if workflow, err := w.GetWorkflowByHandle(handle); err == nil {
		for _, add := range nodes {
			var predicate = func(node broker.WorkflowNode) bool {
				return node.Equals(*add)
			}

			if slices.ContainsFunc(workflow.Nodes, predicate) {
				return w, errorNodeExists
			} else {
				workflow.Nodes = append(workflow.Nodes, *add)
			}
		}
	}

	return w, nil
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
