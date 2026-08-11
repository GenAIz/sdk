package ws

import (
	"context"
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/node"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type workspaceNodeListTaskFactory func() *task.Task[broker.WorkspaceNodeListParams]

type NodeListExecutor struct {
	BaseExecutor
	*node.ListOptions

	workspaceArg string
	flowArg      string

	accountParams                config.AccountParametric
	printerParams                cli.PrinterParametric
	workspaceNodeFacadeProvider  func() mgmt.UserWorkspaceNodesFacade
	workspaceNodeListTaskFactory workspaceNodeListTaskFactory
}

func (nle *NodeListExecutor) Display() {
	var optionMap = map[string]string{}

	if name := nle.Ledger.GetString(nle.OptionAccount); name != "" {
		optionMap["account"] = name
	}

	if nle.workspaceArg != "" {
		optionMap["workspace"] = nle.workspaceArg
	}

	if nle.flowArg != "" {
		optionMap["flow"] = nle.flowArg
	}

	nle.Ledger.DisplayOptionsWithMap(
		&optionMap,
		&nle.OptionJsonPrinter.Option,
		&nle.OptionReadyOnly.Option,
	)
}

func (nle *NodeListExecutor) List(workspaceArg, flowArg string) error {
	nle.workspaceArg = workspaceArg
	nle.flowArg = flowArg
	nle.Cli.Exec(nle.Ledger, nle)
	return nil
}

func (nle *NodeListExecutor) Pretend() {
	var brokerParams = nle.accountParams.BrokerParams()
	var nodeParams = nle.newNodeListParams(brokerParams)

	nle.workspaceNodeListTaskFactory().Pretend(nodeParams, nle.Ledger.Logger)
}

func (nle *NodeListExecutor) Proceed() {
	var brokerParams = nle.accountParams.BrokerParams()
	var nodeParams = nle.newNodeListParams(brokerParams)
	var printer = nle.printerParams.Printer()
	var workspaceNodes, err = nle.workspaceNodeFacadeProvider().
		WithLogger(nle.Ledger.Logger).
		WithParams(nodeParams).
		Provider().
		Get()

	if err == nil {
		lang.HandleExit(printer.Print(workspaceNodes))
		return
	}

	lang.HandleExit(printer.Error(err))
}

func (nle *NodeListExecutor) newNodeListParams(brokerParams *broker.Broker) *broker.WorkspaceNodeListParams {
	var workspaceId, flowId *int64
	var workspaceName, workflowHandle string
	var err error
	var id int64

	if id, err = strconv.ParseInt(nle.workspaceArg, 10, 64); err == nil {
		workspaceId = new(id)
	} else {
		workspaceName = nle.workspaceArg
	}

	if id, err = strconv.ParseInt(nle.flowArg, 10, 64); err == nil {
		flowId = new(id)
	} else {
		workflowHandle = nle.flowArg
	}

	return &broker.WorkspaceNodeListParams{
		Broker:          brokerParams,
		WorkspaceId:     workspaceId,
		WorkspaceName:   workspaceName,
		WorkspaceFlowId: flowId,
		WorkflowHandle:  workflowHandle,
	}
}

func NewNode(ledger *config.Ledger, wsCli *Cli) *cobra.Command {
	var listOptions = node.NewListOptions()
	var listFactory = newNodeListExecutorFactory(ledger, wsCli, listOptions)
	var flowCmd = &cobra.Command{
		Use:     "node",
		Aliases: []string{"nd"},
		Short:   "Manages nodes under workspace flows",
	}

	flowCmd.AddCommand(node.NewList(ledger, listOptions, listFactory))
	return flowCmd
}

func NewNodeListExecutor(ctx context.Context, ledger *config.Ledger, wsCli *Cli, options *node.ListOptions) *NodeListExecutor {
	return &NodeListExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wsCli,
			Context: ctx,
			Ledger:  ledger,
		},
		ListOptions: options,

		accountParams: config.NewAccountParams(ledger, options.OptionAccount),
		printerParams: cli.NewPrinterParam(ledger, options.OptionJsonPrinter),

		workspaceNodeFacadeProvider:  mgmt.NewUserWorkspaceNodesFacade,
		workspaceNodeListTaskFactory: broker.NewWorkspaceNodeListTask,
	}
}

func newNodeListExecutorFactory(ledger *config.Ledger, wsCli *Cli, options *node.ListOptions) node.ListExecutorFactory {
	return func(cmd *cobra.Command) node.ListExecutor {
		return NewNodeListExecutor(cmd.Context(), ledger, wsCli, options)
	}
}
