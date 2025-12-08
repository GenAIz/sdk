package wf

import (
	"context"
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/wf/nodes"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorIncompleteSfSpec = errors.New("incomplete smart function specification, minimum required: OEM/HANDLE:VERSION")
)

type NodesExecutor struct {
	BaseExecutor
	*NodesOptions

	addNode     *broker.WorkflowNode
	rmNodes     []string
	workflowArg string

	workflowTaskFactory   WorkflowTaskFactory
	workflowWriterFactory workflowWriterFactory
}

func (ne *NodesExecutor) Display() {
	var nodeDetails = map[string]string{
		"workflow": ne.workflowArg,
	}

	if ne.addNode != nil {
		nodeDetails["node.add.handle"] = ne.addNode.Handle
		nodeDetails["node.add.name"] = ne.addNode.Name
		nodeDetails["node.add.description"] = ne.addNode.Description

		if ne.addNode.Sf != nil {
			nodeDetails["node.add.sf.handle"] = ne.addNode.Sf.Handle
			nodeDetails["node.add.sf.oem"] = ne.addNode.Sf.Oem
			nodeDetails["node.add.sf.seq"] = cast.ToString(ne.addNode.Sf.Seq)
			nodeDetails["node.add.sf.version"] = ne.addNode.Sf.Version
		}
	}

	if len(ne.rmNodes) > 0 {
		for i, node := range ne.rmNodes {
			nodeDetails[fmt.Sprintf("node.remove[%d].handle", i)] = node
		}
	}

	ne.Ledger.DisplayOptionsWithMap(
		&nodeDetails,
		&ne.NodesOptions.optionConfigType.Option,
	)
}

func (ne *NodesExecutor) Pretend() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = ne.makeConfigParams(ne.optionConfigType); err == nil {
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = ne.makeWorkflowParams(configParams); err == nil {
			var writer = ne.workflowWriterFactory(ne.Ledger, configParams.GetConfigFile())

			ne.workflowTaskFactory(writer).Pretend(workflowParams, ne.Ledger.Logger)
		}
	}

	lang.HandleExit(err)
}

func (ne *NodesExecutor) Proceed() {
	var configParams *shared.ConfigParams
	var err error

	if configParams, err = ne.makeConfigParams(ne.optionConfigType); err == nil {
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = ne.makeWorkflowParams(configParams); err == nil {
			var writer = ne.workflowWriterFactory(ne.Ledger, configParams.GetConfigFile())
			var plan = task.NewPlan("Workflow", ne.Ledger.Logger)

			plan.PrintReportsOnly = true
			task.Single(plan, workflowParams, ne.workflowTaskFactory(writer))
		}
	}

	lang.HandleExit(err)
}

func (ne *NodesExecutor) Add(workflowArg string, nodeHandle string) error {
	var nodeName = ne.Ledger.GetString(ne.optionName)
	var sfHandle = ne.Ledger.GetString(ne.optionSfHandle)
	var sfOem = ne.Ledger.GetString(ne.optionSfOem)
	var sfVersion = ne.Ledger.GetString(ne.optionSfVersion)

	if nodeName == "" {
		nodeName = nodeHandle
	}

	ne.addNode = &broker.WorkflowNode{
		Description: ne.Ledger.GetString(ne.optionDescription),
		Handle:      nodeHandle,
		Name:        nodeName,
	}

	if sfHandle != "" && sfOem != "" && sfVersion != "" {
		ne.addNode.Sf = &broker.WorkflowNodeFunction{
			Handle:  sfHandle,
			Oem:     sfOem,
			Seq:     cast.ToInt(ne.Ledger.GetString(ne.optionSfSeq)),
			Version: sfVersion,
		}
	} else if strings.Join([]string{sfHandle, sfOem, sfVersion}, "") != "" {
		return errorIncompleteSfSpec
	}

	ne.workflowArg = workflowArg
	ne.Cli.Exec(ne.Ledger, ne)
	return nil
}

func (ne *NodesExecutor) Find(path string) (string, error) {
	var vp *viper.Viper
	var err error

	if vp, err = ne.Ledger.FindPathConfig(path); err == nil {
		var handle = schema.Genaiz.Function.Publish.Handle.GetString(vp)

		if handle != "" {
			return handle + "-node", nil
		}
	} else if !errorz.IsPathError(err) {
		return "", err
	}

	return path, nil
}

func (ne *NodesExecutor) Init(path string) (string, error) {
	var vp *viper.Viper
	var err error

	if vp, err = ne.Ledger.FindPathConfig(path); err == nil {
		var handle = schema.Genaiz.Function.Publish.Handle.GetString(vp)
		var oem = schema.Genaiz.Function.Publish.Oem.GetString(vp)
		var version = schema.Genaiz.Function.Publish.Version.GetString(vp)

		if !config.Validation.Handle(handle) {
			return "", fmt.Errorf("function under path [%s] has no valid handle", path)
		}

		if !config.Validation.Oem(oem) {
			return "", fmt.Errorf("function under path [%s] has no valid oem", path)
		}

		if !config.Validation.Version(version) {
			return "", fmt.Errorf("function under path [%s] has no valid version", path)
		}

		if err = ne.initPathOption(path, oem, ne.optionSfOem); err == nil {
			if err = ne.initPathOption(path, handle, ne.optionSfHandle); err == nil {
				if err = ne.initPathOption(path, version, ne.optionSfVersion); err == nil {
					return filepath.Base(path) + "-node", nil
				}
			}
		}
	}

	if !errorz.IsPathError(err) {
		return "", err
	}

	return path, nil
}

func (ne *NodesExecutor) Remove(workflowArg string, nodeHandles ...string) {
	ne.rmNodes = nodeHandles
	ne.workflowArg = workflowArg
	ne.Cli.Exec(ne.Ledger, ne)
}

func (ne *NodesExecutor) initPathOption(path, value string, option *config.StringOption) error {
	if value != "" {
		var oldValue = ne.Ledger.GetString(option)

		ne.Ledger.InitValue(option, value)

		if oldValue != "" && value != oldValue {
			return fmt.Errorf("value [%s] for option [%s] conflicts with [%s] under [%s]",
				oldValue, option.Key, value, path)
		}
	}

	return nil
}

func (ne *NodesExecutor) makeWorkflowParams(configParams *shared.ConfigParams) (*broker.WorkflowParams, error) {
	var writer = ne.workflowWriterFactory(ne.Ledger, configParams.GetConfigFile())
	var handle = ne.workflowArg
	var edited *broker.Workflow
	var err error

	if ne.addNode != nil {
		if _, err = writer.addNodes(handle, ne.addNode); err != nil {
			return nil, err
		}
	}

	writer.removeNodes(handle, ne.rmNodes...)

	if edited, err = writer.GetWorkflowByHandle(handle); err == nil {
		return &broker.WorkflowParams{
			ConfigParams:   *configParams,
			Workflow:       edited,
			WorkflowUpdate: true,
		}, nil
	}

	return nil, err
}

type NodesOptions struct {
	optionConfigType   *config.StringOption
	optionDescription  *config.StringOption
	optionName         *config.StringOption
	optionSfHandle     *config.StringOption
	optionSfOem        *config.StringOption
	optionSfSeq        *config.StringOption
	optionSfSerialized *config.StringOption
	optionSfVersion    *config.StringOption
}

func (no NodesOptions) addDefiners() []config.Definer {
	return []config.Definer{
		no.optionConfigType,
		no.optionDescription,
		no.optionName,
		no.optionSfHandle,
		no.optionSfOem,
		no.optionSfSeq,
		no.optionSfSerialized,
		no.optionSfVersion,
	}
}

func (no NodesOptions) removeDefiners() []config.Definer {
	return []config.Definer{
		no.optionConfigType,
	}
}

type SerializedOptions struct {
	node *broker.WorkflowNodeFunction

	optionSerialized   *config.StringOption
	optionDeserialized *config.Option
}

func (so *SerializedOptions) GetHandle(ledger *config.Ledger) any {
	return so.getNode(ledger).Handle
}

func (so *SerializedOptions) GetOem(ledger *config.Ledger) any {
	return so.getNode(ledger).Oem
}

func (so *SerializedOptions) GetSeq(ledger *config.Ledger) any {
	return so.getNode(ledger).Seq
}

func (so *SerializedOptions) GetVersion(ledger *config.Ledger) any {
	return so.getNode(ledger).Version
}

func (so *SerializedOptions) getDefault(ledger *config.Ledger) any {
	var serialized = ledger.GetString(so.optionSerialized)
	var result = &broker.WorkflowNodeFunction{}

	if serialized != "" {
		var i int

		if i = strings.Index(serialized, "/"); i >= 0 {
			result.Oem = serialized[0:i]
			serialized = stringz.SubstrFrom(serialized, i+1)
		}

		if i = strings.Index(serialized, ":"); i >= 0 {
			result.Handle = serialized[0:i]
			serialized = stringz.SubstrFrom(serialized, i+1)
		} else {
			result.Handle = serialized
		}

		if i = strings.Index(serialized, "-rc"); i >= 0 {
			var j = strings.Index(serialized, "-rc.")

			if j < 0 {
				j = strings.Index(serialized, "-rc-")
			}

			if j >= 0 {
				result.Version = serialized[0:j]
				result.Seq = cast.ToInt(serialized[j+4:])
			} else {
				result.Version = serialized[0:i]
				result.Seq = cast.ToInt(serialized[i+3:])
			}
		} else {
			result.Version = serialized
		}
	}

	return result
}

func (so *SerializedOptions) getNode(ledger *config.Ledger) *broker.WorkflowNodeFunction {
	if so.node == nil {
		var dAny = ledger.Get(so.optionDeserialized)
		var ok bool

		if so.node, ok = dAny.(*broker.WorkflowNodeFunction); !ok {
			so.node = &broker.WorkflowNodeFunction{}
		}
	}

	return so.node
}

func NewNodes(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var addNodesOptions = NewAddNodesOptions()
	var rmNodesOptions = NewRemoveNodesOptions()
	var addNodesCmd = nodes.NewAddNodes(newNodesAddExecutorFactory(ledger, cli, addNodesOptions), validateArgNodes)
	var rmNodesCmd = nodes.NewRemoveNodes(newNodesRemoveExecutorFactory(ledger, cli, rmNodesOptions), validateArgNodes)
	var nodesCmd = &cobra.Command{
		Use:     "nodes",
		Aliases: []string{"nd"},
		Short:   "Manages nodes of an existing workflow",
	}

	nodesCmd.AddCommand(addNodesCmd)
	nodesCmd.AddCommand(rmNodesCmd)
	ledger.Register(addNodesCmd, addNodesOptions.addDefiners()...)
	ledger.Register(rmNodesCmd, rmNodesOptions.removeDefiners()...)
	return nodesCmd
}

func NewNodesExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *NodesOptions) *NodesExecutor {
	return &NodesExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		NodesOptions: options,

		workflowTaskFactory:   broker.NewWorkflowUpdateTask,
		workflowWriterFactory: newWorkflowWriter,
	}
}

func NewAddNodesOptions() *NodesOptions {
	var serializedOptions = NewSerializedOptions()
	var nameOption = cli.Options.Workflows.Name().
		WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Name).
		BuildStringOption()

	return &NodesOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.ConfigType).
			BuildStringOption(),
		optionDescription: cli.Options.Workflows.Description().
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Description).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(nameOption)
			}).BuildStringOption(),
		optionName: nameOption,
		optionSfHandle: cli.Options.Workflows.SfHandle().
			Optional(true).
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Handle).
			WithDefaultGetter(serializedOptions.GetHandle).
			BuildStringOption(),
		optionSfOem: cli.Options.Workflows.SfOem().
			Optional(true).
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Oem).
			WithDefaultGetter(serializedOptions.GetOem).
			BuildStringOption(),
		optionSfSeq: cli.Options.Workflows.SfSequence().
			Optional(true).
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Sequence).
			WithDefaultGetter(serializedOptions.GetSeq).
			BuildStringOption(),
		optionSfSerialized: serializedOptions.optionSerialized,
		optionSfVersion: cli.Options.Workflows.SfVersion().
			Optional(true).
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Version).
			WithDefaultGetter(serializedOptions.GetVersion).
			BuildStringOption(),
	}
}

func NewSerializedOptions() *SerializedOptions {
	var result = &SerializedOptions{
		optionSerialized: cli.Options.Workflows.SfSerialized().
			WithKeys(&schema.Genaiz.Workflow.Nodes.Add.Serialized).
			BuildStringOption(),
		optionDeserialized: &config.Option{
			Key: schema.Genaiz.Workflow.Nodes.Add.Deserialized.Doc,
		},
	}

	result.optionDeserialized.DefaultGetter = result.getDefault
	return result
}

func NewRemoveNodesOptions() *NodesOptions {
	return &NodesOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Nodes.Remove.ConfigType).
			BuildStringOption(),
	}
}

func newNodesAddExecutorFactory(ledger *config.Ledger, cli *Cli, options *NodesOptions) func(*cobra.Command) nodes.AddExecutor {
	return func(cmd *cobra.Command) nodes.AddExecutor {
		return NewNodesExecutor(cmd.Context(), ledger, cli, options)
	}
}

func newNodesRemoveExecutorFactory(ledger *config.Ledger, cli *Cli, options *NodesOptions) func(*cobra.Command) nodes.RemoveExecutor {
	return func(cmd *cobra.Command) nodes.RemoveExecutor {
		return NewNodesExecutor(cmd.Context(), ledger, cli, options)
	}
}

func validateArgNodes(arg ...string) error {
	var predicate = func(s string) bool {
		return !config.Validation.Handle(s)
	}

	if i := slices.IndexFunc(arg, predicate); i >= 0 {
		return fmt.Errorf("[%s] is not a valid node handle", arg[i])
	}

	return nil
}
