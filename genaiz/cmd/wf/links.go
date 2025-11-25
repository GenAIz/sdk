package wf

import (
	"context"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/wf/links"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
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

	if configParams, err = le.makeConfigParams(le.optionConfigType); err == nil {
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

	if configParams, err = le.makeConfigParams(le.optionConfigType); err == nil {
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = le.makeWorkflowParams(configParams); err == nil {
			var writer = le.workflowWriterFactory(le.Ledger, configParams.GetConfigFile())
			var plan = task.NewPlan("Workflow", le.Ledger.Logger)

			plan.PrintReportsOnly = true
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

func (le *LinksExecutor) Init(workflowArg string, links []string) ([]string, error) {
	var configParams *shared.ConfigParams
	var result []string
	var err error

	if configParams, err = le.makeConfigParams(le.optionConfigType); err == nil {
		var writer = le.workflowWriterFactory(le.Ledger, configParams.GetConfigFile())
		var workflowLinks = parseArgsLinks(links...)
		var current *broker.Workflow

		if current, err = writer.GetWorkflowByHandle(workflowArg); err == nil {
			for _, link := range workflowLinks {
				var leftReference, leftPort, rightReference, rightPort string

				if leftReference, leftPort, err = parseNodeRefs(link.LhsNode, link.LhsNodePort); err == nil {
					if rightReference, rightPort, err = parseNodeRefs(link.RhsNode, link.RhsNodePort); err == nil {
						var leftHandle, rightHandle string
						var leftValue, rightValue string

						if current.ContainsNode(leftReference) {
							leftHandle = leftReference
						} else {
							leftHandle, err = le.findFunctionNode(current, leftReference, leftPort)
						}

						if err == nil {
							if current.ContainsNode(rightReference) {
								rightHandle = rightReference
							} else {
								rightHandle, err = le.findFunctionNode(current, rightReference, rightPort)
							}
						}

						if err == nil {
							if leftPort == "" {
								leftValue = leftHandle
							} else {
								leftValue = fmt.Sprintf("%s[%s]", leftHandle, leftPort)
							}

							if rightPort == "" {
								rightValue = rightHandle
							} else {
								rightValue = fmt.Sprintf("%s[%s]", rightHandle, rightPort)
							}

							result = append(result, fmt.Sprintf("%s:%s", leftValue, rightValue))
						}
					}
				}

				if err != nil {
					break
				}
			}
		}
	}

	return result, err
}

func (le *LinksExecutor) Remove(workflowArg string, rmLinks []string) {
	le.rmLinks = rmLinks
	le.workflowArg = workflowArg
	le.Cli.Exec(le.Ledger, le)
}

func (le *LinksExecutor) findFunctionNode(workflow *broker.Workflow, ref, port string) (string, error) {
	var vp *viper.Viper
	var err error

	// We'll eventually have to validate the port
	_ = port

	if vp, err = le.Ledger.FindPathConfig(ref); err == nil {
		var sfOem = vp.GetString(schema.Genaiz.Function.Publish.Oem.Doc)
		var sfHandle = vp.GetString(schema.Genaiz.Function.Publish.Handle.Doc)
		var sfVersion = vp.GetString(schema.Genaiz.Function.Publish.Version.Doc)

		return workflow.FindNodeHandleBySf(sfOem, sfHandle, sfVersion)
	} else if errorz.IsPathError(err) {
		return "", fmt.Errorf("value [%s] could not resolve to a workflow node", ref)
	}

	return "", err
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
	var addLinksOptions = NewAddLinksOptions()
	var rmLinksOptions = NewRemoveLinksOptions()
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

func NewAddLinksOptions() *LinksOptions {
	return &LinksOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Links.Add.ConfigType).
			BuildStringOption(),
	}
}

func NewRemoveLinksOptions() *LinksOptions {
	return &LinksOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Links.Remove.ConfigType).
			BuildStringOption(),
	}
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

// parseNodeRefs tries to interpret the nodes specified on the command line as references to smart function folders relative to the current work dir.
//
// Node handles can be composed of smart functions with ports at the end of a file path. This function makes sure the references are split into handle/port if any are specified in this manner.
// A syntax error is raised if there's a conflicting spec. That is both path and a port string ([port]) are used and different. If there are no paths, the handles and ports are kept as-is.
func parseNodeRefs(nodeValue, nodePort string) (string, string, error) {
	var funcPath = filepath.Clean(nodeValue)
	var funcDir = filepath.Dir(funcPath)
	var funcPort = filepath.Base(funcPath)

	if funcDir == "." {
		funcDir = funcPath
	} else {
		funcDir = dirz.FirstParentName(funcPath)
	}

	if funcDir == funcPort {
		funcPort = nodePort
	}

	if nodePort != "" && funcPort != nodePort {
		return "", "", fmt.Errorf("conflicting port specification: [%s] and [%s] diverge", funcPort, nodePort)
	}

	return funcDir, funcPort, nil
}

func validateArgLinks(args []string) error {
	var predicate = func(s string) bool {
		return !strings.Contains(s, ":")
	}

	if i := slices.IndexFunc(args, predicate); i >= 0 {
		return fmt.Errorf("[%s] is not a valid workflow link pair", args[i])
	}

	return nil
}
