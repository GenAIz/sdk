package data

import (
	"strconv"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/ws/auto"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/broker/dataset"
	"genaiz.com/genaiz/task/broker/workspace"
)

type SetAutoBridge struct {
	dataSets       auto.DataSetAutoBridge
	workspaces     auto.Bridge
	workspaceFlows auto.WorkspaceBridge
	workspaceNodes auto.WorkspaceFlowBridge
}

func (sab SetAutoBridge) bridgeArguments(cmd *cobra.Command, args []string, toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var results []cobra.Completion
	var directive cobra.ShellCompDirective
	var argsCount = len(args)

	_ = cmd

	if argsCount == 0 {
		results, directive = sab.workspaces.Bridge(toComplete)
	} else if argsCount == 1 {
		results, directive = sab.workspaceFlows.WithWorkspace(args[0]).Bridge(toComplete)
	} else if argsCount == 2 {
		results, directive = sab.workspaceNodes.WithWorkspaceFlow(args[0], args[1]).Bridge(toComplete)
	} else if argsCount == 3 {
		results, directive = sab.dataSets.WithWorkspaceNode(args[0], args[1], args[2]).Bridge(toComplete)
	} else if argsCount == 4 {
		directive = cobra.ShellCompDirectiveDefault
	} else {
		directive = cobra.ShellCompDirectiveNoFileComp
	}

	return results, directive
}

type SetExecutor interface {
	Set(string, string, string, string, string) error
}

type SetExecutorFactory func(command *cobra.Command) SetExecutor

type SetOptions struct {
	baseOptions

	optionDestination *config.StringOption
	optionFilename    *config.StringOption
	optionOverwrite   *config.BoolOption
	optionSetDesc     *config.StringOption
	optionSetHandle   *config.StringOption
	optionSetName     *config.StringOption
}

func (so SetOptions) CreateSetParams(ledger *config.Ledger, brokerParams *broker.Broker,
	nodeResolveParams *broker.WorkspaceNodeResolveParams,
	setResolveParams *broker.WorkspaceNodeSetResolveParams) *broker.DataSetCreateParams {
	var name = ledger.GetString(so.optionSetName)
	var handle = ledger.GetString(so.optionSetHandle)
	var description = ledger.GetString(so.optionSetDesc)

	return &broker.DataSetCreateParams{
		Broker:            *brokerParams,
		NodeResolveParams: &nodeResolveParams.NodeResolveParams,
		NodeSetParams:     &setResolveParams.NodeSetParams,
		CreateFields: dataset.CreateFields{
			Name:        name,
			Handle:      handle,
			Description: description,
		},
	}
}

func (so SetOptions) IsOverwrite(ledger *config.Ledger) bool {
	return ledger.GetBool(so.optionOverwrite)
}

func (so SetOptions) ListFilesParams(brokerParams *broker.Broker,
	setResolveParams *broker.WorkspaceNodeSetResolveParams) *broker.DataSetListFilesParams {
	return &broker.DataSetListFilesParams{
		Broker:        *brokerParams,
		NodeSetParams: &setResolveParams.NodeSetParams,
	}
}

func (so SetOptions) ResolveSetParams(dataSet string, brokerParams *broker.Broker,
	nodeResolveParams *broker.WorkspaceNodeResolveParams) *broker.WorkspaceNodeSetResolveParams {
	var nodeSetParams workspace.NodeSetParams

	if setId, err := strconv.ParseInt(dataSet, 10, 64); err == nil {
		nodeSetParams = workspace.NodeSetParams{
			DataSetId: &setId,
		}
	} else {
		nodeSetParams = workspace.NodeSetParams{
			PortHandle: dataSet,
		}
	}

	return &broker.WorkspaceNodeSetResolveParams{
		Broker:            *brokerParams,
		NodeSetParams:     nodeSetParams,
		NodeResolveParams: &nodeResolveParams.NodeResolveParams,
	}
}

func (so SetOptions) UploadFileParams(ledger *config.Ledger, path string, brokerParams *broker.Broker,
	setResolveParams *broker.WorkspaceNodeSetResolveParams) *broker.DataSetUploadParams {
	return &broker.DataSetUploadParams{
		Broker:        *brokerParams,
		NodeSetParams: &setResolveParams.NodeSetParams,
		UploadFields: dataset.UploadFields{
			FileName:        ledger.GetString(so.optionFilename),
			FilePath:        path,
			DestinationPath: ledger.GetString(so.optionDestination),
		},
	}
}

func (so SetOptions) allDefiners() []config.Definer {
	return []config.Definer{
		so.optionAccount,
		so.optionDestination,
		so.optionJsonPrinter,
		so.optionOverwrite,
		so.optionSetDesc,
		so.optionSetHandle,
		so.optionSetName,
	}
}

func NewDataSet(ledger *config.Ledger, options *SetOptions, factory SetExecutorFactory) *cobra.Command {
	var setAutoBridge = NewDataSetAutoBridge(ledger)
	var setCmd = &cobra.Command{
		Use:     "set [WORKSPACE_NAME]|WORKSPACE_ID [WORKFLOW_HANDLE]|FLOW_ID [SF_HANDLE]|NODE_ID [PORT_HANDLE]|DATASET_ID FILE_PATH",
		Short:   "Uploads a file to a data set",
		Long:    "Uploads a file to a data set, assigning it to the port of a node, using workspace, flow handle and function handle or their ids",
		Example: "genaiz ws node data set my-workspace my-flow my-node portHandle myFile.json -d=mySubFolder",
		Args:    cobra.ExactArgs(5),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = factory(cmd)

			lang.HandleExit(exec.Set(args[0], args[1], args[2], args[3], args[4]))
		},
	}

	ledger.Register(setCmd, options.allDefiners()...)
	setCmd.ValidArgsFunction = setAutoBridge.bridgeArguments
	cli.AutoBridge.Accounts().Option(setCmd, ledger, options.optionAccount)
	return setCmd
}

func NewDataSetAutoBridge(ledger *config.Ledger) *SetAutoBridge {
	return &SetAutoBridge{
		dataSets:       auto.NewDataSetAutoBridge(ledger),
		workspaces:     auto.NewWorkspaceBridge(ledger),
		workspaceFlows: auto.NewWorkspaceFlowBridge(ledger),
		workspaceNodes: auto.NewWorkspaceNodeAutoBridge(ledger),
	}
}

func NewDataSetOptions() *SetOptions {
	var handleOption = cli.Options.DataSets.Handle().
		WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.SetHandle).
		Optional(true).
		BuildStringOption()

	return &SetOptions{
		baseOptions: baseOptions{
			optionAccount: cli.Options.Workspaces.Account().
				WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.Account).
				BuildStringOption(),
			optionJsonPrinter: cli.Options.Printer.JsonPrinter().
				WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.Printer).
				BuildBoolOption(),
		},
		optionDestination: cli.Options.DataSets.Destination().
			WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.SetDesc).
			Optional(true).
			BuildStringOption(),
		optionFilename: cli.Options.DataSets.Filename().
			WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.Filename).
			BuildStringOption(),
		optionOverwrite: cli.Options.DataSets.Overwrite().
			WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.Overwrite).
			BuildBoolOption(),
		optionSetDesc: cli.Options.DataSets.Description().
			WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.SetDesc).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(handleOption)
			}).BuildStringOption(),
		optionSetHandle: handleOption,
		optionSetName: cli.Options.DataSets.Name().
			WithKeys(&schema.Genaiz.Workspace.Node.Data.Set.SetName).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(handleOption)
			}).BuildStringOption(),
	}
}
