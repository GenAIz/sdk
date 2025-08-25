package wf

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cmd/wf/links"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type LinksExecutor struct {
	BaseExecutor
	*LinksOptions

	addLinks    []string
	rmLinks     []string
	workflowArg string

	workflowTaskFactory   WorkflowTaskFactory
	workflowWriterFactory workflowWriterFactory
}

func (le *LinksExecutor) Display() {
	var linkDetails = map[string]string{
		"workflow": le.workflowArg,
	}

	if len(le.addLinks) > 0 {
		for i, link := range le.addLinks {
			linkDetails[fmt.Sprintf("add-%d", i)] = link
		}
	} else if len(le.rmLinks) > 0 {
		for i, link := range le.rmLinks {
			linkDetails[fmt.Sprintf("rm-%d", i)] = link
		}
	}

	le.Ledger.DisplayOptionsWithMap(
		&linkDetails,
		&le.optionConfigType.Option,
	)
}

func (le *LinksExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = le.makeConfigParams(); err == nil {
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = le.makeWorkflowParams(configParams); err == nil {
			var writer = le.workflowWriterFactory(le.Ledger, configParams.GetConfigFile())

			le.workflowTaskFactory(writer).Pretend(workflowParams, le.Ledger.Logger)
		}
	}

	lang.HandleExit(err)
}

func (le *LinksExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = le.makeConfigParams(); err == nil {
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = le.makeWorkflowParams(configParams); err == nil {
			var writer = le.workflowWriterFactory(le.Ledger, configParams.GetConfigFile())
			var plan = task.NewPlan("Workflow", le.Ledger.Logger)

			task.Single(plan, workflowParams, le.workflowTaskFactory(writer))
		}
	}

	lang.HandleExit(err)
}

func (le *LinksExecutor) Add(workflowArg string, addLinks []string) {
	le.addLinks = addLinks
	le.workflowArg = workflowArg
	le.Cli.Exec(le.Ledger, le)
}

func (le *LinksExecutor) Remove(workflowArg string, rmLinks []string) {
	le.rmLinks = rmLinks
	le.workflowArg = workflowArg
	le.Cli.Exec(le.Ledger, le)
}

func (le *LinksExecutor) makeConfigParams() (*shared.ConfigParams, error) {
	var configType *shared.ConfigType
	var err error

	if configType, err = le.Ledger.GetConfigType(le.optionConfigType); err == nil {
		return &shared.ConfigParams{
			ConfigName: le.Ledger.ConfigName,
			ConfigType: configType,
		}, nil
	}

	return nil, err
}

func (le *LinksExecutor) makeWorkflowParams(configParams *shared.ConfigParams) (*broker.WorkflowParams, error) {
	var writer = le.workflowWriterFactory(le.Ledger, configParams.GetConfigFile())
	var edited *broker.Workflow
	var err error

	writer.addLinks(le.workflowArg, parseArgsLinks(le.addLinks...))
	writer.removeLinks(le.workflowArg, parseArgsLinks(le.rmLinks...))

	if edited, err = writer.GetWorkflowByHandle(le.workflowArg); err == nil {
		return &broker.WorkflowParams{
			ConfigParams:   *configParams,
			Workflow:       edited,
			WorkflowUpdate: true,
		}, nil
	}

	return nil, err
}

type LinksOptions struct {
	optionConfigType *config.StringOption
}

func (lo LinksOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionConfigType,
	}
}

func NewLinks(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var addLinksOptions = NewLinksOptions("Links.Add")
	var rmLinksOptions = NewLinksOptions("Links.Remove")
	var addLinksCmd = links.NewAddLinks(newLinksAddExecutorFactory(ledger, cli, addLinksOptions), validateArgLinks)
	var rmLinksCmd = links.NewRemoveLinks(newLinksRemoveExecutorFactory(ledger, cli, rmLinksOptions), validateArgLinks)
	var linksCmd = &cobra.Command{
		Use:     "links",
		Aliases: []string{"ln"},
		Short:   "Manages links of an existing workflow",
	}

	linksCmd.AddCommand(addLinksCmd)
	linksCmd.AddCommand(rmLinksCmd)
	ledger.Register(addLinksCmd, addLinksOptions.allDefiners()...)
	ledger.Register(rmLinksCmd, rmLinksOptions.allDefiners()...)
	return linksCmd
}

func NewLinksExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *LinksOptions) *LinksExecutor {
	return &LinksExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		LinksOptions: options,

		workflowTaskFactory:   broker.NewWorkflowUpdateTask,
		workflowWriterFactory: newWorkflowWriter,
	}
}

func NewLinksOptions(cmd string) *LinksOptions {
	return &LinksOptions{
		optionConfigType: newOptionConfigType(cmd),
	}
}

func parseArgsHandlePort(handlePort string) (string, string) {
	if i := strings.Index(handlePort, "["); i > 0 {
		return handlePort[0:i], handlePort[i+1 : len(handlePort)-1]
	} else {
		return handlePort, ""
	}
}

func parseArgsLinks(links ...string) []broker.WorkflowLink {
	var result []broker.WorkflowLink

	for _, arg := range links {
		var pair = strings.SplitN(arg, ":", 2)

		if len(pair) == 2 {
			var handleLeft, portLeft = parseArgsHandlePort(pair[0])
			var handleRight, portRight = parseArgsHandlePort(pair[1])

			result = append(result, broker.WorkflowLink{
				LhsNode:     handleLeft,
				LhsNodePort: portLeft,
				RhsNode:     handleRight,
				RhsNodePort: portRight,
			})
		}
	}

	return result
}

func newLinksAddExecutorFactory(ledger *config.Ledger, cli *Cli, options *LinksOptions) func(*cobra.Command) links.AddExecutor {
	return func(cmd *cobra.Command) links.AddExecutor {
		return NewLinksExecutor(cmd.Context(), ledger, cli, options)
	}
}

func newLinksRemoveExecutorFactory(ledger *config.Ledger, cli *Cli, options *LinksOptions) func(*cobra.Command) links.RemoveExecutor {
	return func(cmd *cobra.Command) links.RemoveExecutor {
		return NewLinksExecutor(cmd.Context(), ledger, cli, options)
	}
}

func validateArgLinks(args []string) error {
	var predicate = func(s string) bool {
		return !strings.Contains(s, ":")
	}

	if i := slices.IndexFunc(args, predicate); i >= 0 {
		return fmt.Errorf("[%s] is not a valid workflow link pair (handle[port]:handle[port])", args[i])
	}

	return nil
}
