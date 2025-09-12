package wf

import (
	"context"
	"fmt"
	"slices"
	"strings"

	"github.com/spf13/cast"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/cmd/wf/nodes"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
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

			task.Single(plan, workflowParams, ne.workflowTaskFactory(writer))
		}
	}

	lang.HandleExit(err)
}

func (ne *NodesExecutor) Add(workflowArg string, nodeHandle string) {
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
	}

	ne.workflowArg = workflowArg
	ne.Cli.Exec(ne.Ledger, ne)
}

func (ne *NodesExecutor) Remove(workflowArg string, nodeHandles ...string) {
	ne.rmNodes = nodeHandles
	ne.workflowArg = workflowArg
	ne.Cli.Exec(ne.Ledger, ne)
}

func (ne *NodesExecutor) makeWorkflowParams(configParams *shared.ConfigParams) (*broker.WorkflowParams, error) {
	var writer = ne.workflowWriterFactory(ne.Ledger, configParams.GetConfigFile())
	var handle = ne.workflowArg
	var edited *broker.Workflow
	var err error

	if ne.addNode != nil {
		writer.addNodes(handle, ne.addNode)
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
	var cmd = "Nodes.Add"
	var serializedOption = newOptionSfSerialized(cmd)
	var deserializedOption = newOptionSfDeserialized(serializedOption, cmd)
	var nameOption = newOptionNodeName(cmd)

	return &NodesOptions{
		optionConfigType:   newOptionConfigType(cmd),
		optionDescription:  newOptionNodeDescription(cmd, &nameOption.Option),
		optionName:         nameOption,
		optionSfHandle:     newOptionSfHandle(deserializedOption, cmd),
		optionSfOem:        newOptionSfOem(deserializedOption, cmd),
		optionSfSeq:        newOptionSfSeq(deserializedOption, cmd),
		optionSfSerialized: serializedOption,
		optionSfVersion:    newOptionSfVersion(deserializedOption, cmd),
	}
}

func NewRemoveNodesOptions() *NodesOptions {
	var cmd = "Nodes.Remove"

	return &NodesOptions{
		optionConfigType: newOptionConfigType(cmd),
	}
}

func makeEnvKey(key string) string {
	return strings.ReplaceAll(strings.ToUpper(key), ".", "_")
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

func newOptionNodeDescription(cmd string, defaultOption *config.Option) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Description",
			Env:   makeEnvKey("WF_" + cmd + "_DESCRIPTION"),
			Param: "description",
			Usage: "description of the workflow node",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.Get(defaultOption)
			},
			Validator: config.Validation.Blob,
		},
	}
}

func newOptionNodeName(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "Workflow." + cmd + ".Name",
			Env:       makeEnvKey("WF_" + cmd + "_NAME"),
			Param:     "name",
			Usage:     "name of the workflow node",
			Validator: config.Validation.Name,
		},
	}
}

func newOptionSfDeserialized(serializedOption *config.StringOption, cmd string) *config.Option {
	var deserialized *broker.WorkflowNodeFunction

	return &config.Option{
		Key: "Workflow." + cmd + ".Function.Deserialized",
		DefaultGetter: func(ledger *config.Ledger) any {
			if deserialized == nil {
				var serialized = ledger.GetString(serializedOption)

				deserialized = &broker.WorkflowNodeFunction{}

				if serialized != "" {
					var i int

					if i = strings.Index(serialized, "/"); i >= 0 {
						deserialized.Oem = serialized[0:i]
						serialized = stringz.SubstrFrom(serialized, i+1)
					}

					if i = strings.Index(serialized, ":"); i >= 0 {
						deserialized.Handle = serialized[0:i]
						serialized = stringz.SubstrFrom(serialized, i+1)
					} else {
						deserialized.Handle = serialized
					}

					if i = strings.Index(serialized, "-rc"); i >= 0 {
						var j = strings.Index(serialized, "-rc.")

						if j < 0 {
							j = strings.Index(serialized, "-rc-")
						}

						if j >= 0 {
							deserialized.Version = serialized[0:j]
							deserialized.Seq = cast.ToInt(serialized[j+4:])
						} else {
							deserialized.Version = serialized[0:i]
							deserialized.Seq = cast.ToInt(serialized[i+3:])
						}
					} else {
						deserialized.Version = serialized
					}
				}
			}

			return deserialized
		},
	}
}

func newOptionSfHandle(defaultOption *config.Option, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Function.Handle",
			Env:   makeEnvKey("WF_" + cmd + "_FUNCTION_HANDLE"),
			Param: "sf.handle",
			Usage: "handle of the node smart function",
			DefaultGetter: func(ledger *config.Ledger) any {
				var dAny = ledger.Get(defaultOption)

				return dAny.(*broker.WorkflowNodeFunction).Handle
			},
			Validator: config.Optionally(config.Validation.Handle),
		},
	}
}

func newOptionSfOem(defaultOption *config.Option, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Function.Oem",
			Env:   makeEnvKey("WF_" + cmd + "_FUNCTION_OEM"),
			Param: "sf.oem",
			Usage: "oem of the node smart function",
			DefaultGetter: func(ledger *config.Ledger) any {
				var dAny = ledger.Get(defaultOption)

				return dAny.(*broker.WorkflowNodeFunction).Oem
			},
			Validator: config.Optionally(config.Validation.Oem),
		},
	}
}

func newOptionSfSeq(defaultOption *config.Option, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Function.Seq",
			Env:   makeEnvKey("WF_" + cmd + "_FUNCTION_SEQ"),
			Param: "sf.seq",
			Usage: "sequence number of the node smart function",
			DefaultGetter: func(ledger *config.Ledger) any {
				var dAny = ledger.Get(defaultOption)

				return dAny.(*broker.WorkflowNodeFunction).Seq
			},
			Validator: config.Optionally(config.Validation.VersionNumber),
		},
	}
}

func newOptionSfSerialized(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Function.Serialized",
			Env:   makeEnvKey("WF_" + cmd + "_FUNCTION_SERIALIZED"),
			Param: "sf",
			Usage: "serialized string of the smart function, the individual options have precedence",
		},
	}
}

func newOptionSfVersion(defaultOption *config.Option, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow." + cmd + ".Function.Version",
			Env:   makeEnvKey("WF_" + cmd + "_FUNCTION_VERSION"),
			Param: "sf.version",
			Usage: "version of the node smart function",
			DefaultGetter: func(ledger *config.Ledger) any {
				var dAny = ledger.Get(defaultOption)

				return dAny.(*broker.WorkflowNodeFunction).Version
			},
			Validator: config.Optionally(config.Validation.Version),
		},
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
