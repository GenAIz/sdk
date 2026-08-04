package wf

import (
	"context"
	"errors"
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/errorz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/cmd/wf/prop"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type taskWorkerSyncFactory func([]string, *config.Ledger, *config.BoolOption) ([]task.Worker, error)
type taskWorkerPropFactory func(*broker.WorkflowPropParams, *task.Task[broker.WorkflowPropParams]) task.Worker

type PropExecutor struct {
	BaseExecutor
	dk.SyncBridge
	*PropOptions

	addProps map[string]string
	rmProps  []string

	external     *broker.Function
	function     *broker.Function
	solution     *broker.Solution
	workflow     *broker.Workflow
	workflowNode *broker.WorkflowNode

	workflowPropTaskFactory WorkflowPropTaskFactory
	workflowTaskFactory     WorkflowTaskFactory
	workflowWriterFactory   workflowWriterFactory
}

func (pe *PropExecutor) Add(workflowArg string, nodeArg string, key string, value string) error {
	var err error

	if err = pe.findWorkflowAndNode(workflowArg, nodeArg); errors.Is(err, errorNoFunction) &&
		pe.workflowNode.Sf != nil {
		// We may be adding props for an external function
		pe.external = &broker.Function{
			Oem:     pe.workflowNode.Sf.Oem,
			Handle:  pe.workflowNode.Sf.Handle,
			Version: pe.workflowNode.Sf.Version,
		}
	} else if err != nil {
		return err
	}

	pe.addProps[key] = value
	pe.Cli.Exec(pe.Ledger, pe)
	return nil
}

func (pe *PropExecutor) Display() {
	var propDetails []map[string]string

	if len(pe.rmProps) > 0 {
		for _, k := range pe.rmProps {
			propDetails = append(propDetails, map[string]string{
				"prop.rm.key": k,
			})
		}
	}

	if len(pe.addProps) > 0 {
		for k, v := range pe.addProps {
			propDetails = append(propDetails, map[string]string{
				"prop.add.key":   k,
				"prop.add.value": v,
			})
		}
	}

	pe.Ledger.DisplayOptionsWithMap(&map[string]string{
		"prop.workflow":      pe.workflow.Handle,
		"prop.workflow.node": pe.workflowNode.Handle,
	})

	for _, props := range propDetails {
		pe.Ledger.DisplayOptionsWithMap(&props)
	}
}

func (pe *PropExecutor) Edit(workflowArg string, nodeArg string, key string, value string) error {
	var err error

	if err = pe.findWorkflowAndNode(workflowArg, nodeArg); errors.Is(err, errorNoFunction) &&
		pe.workflowNode.Sf != nil {
		// We may be adding props for an external function
		pe.external = &broker.Function{
			Oem:     pe.workflowNode.Sf.Oem,
			Handle:  pe.workflowNode.Sf.Handle,
			Version: pe.workflowNode.Sf.Version,
		}
	} else if err != nil {
		return err
	}

	if !pe.workflowNode.HasProp(key) {
		return fmt.Errorf("the key [%s] could not be found under node [%s]", key, pe.workflowNode.Handle)
	}

	pe.rmProps = append(pe.rmProps, key)
	pe.addProps[key] = value
	pe.Cli.Exec(pe.Ledger, pe)
	return nil
}

func (pe *PropExecutor) Init(nodeArg string) (string, error) {
	var fn *broker.Function
	var err error

	if fn, err = pe.getFunction(nodeArg); err == nil {
		var sn *broker.Solution

		if sn, err = pe.getSolution(); err == nil {
			var initHandle string

			for _, wf := range sn.Workflows {
				if initHandle, err = wf.FindNodeHandleBySf(fn.Oem, fn.Handle, fn.Version); err == nil {
					pe.function = fn
					return initHandle, nil
				}
			}

			return "", errorNoWorkflowNode
		}
	}

	if !errorz.IsPathError(err) {
		return "", err
	}

	return nodeArg, nil
}

func (pe *PropExecutor) Pretend() {
	var configParams = pe.newConfigParams()
	var input string
	var err error

	if input, err = configParams.EnsureConfigPath(); err == nil {
		var writer = pe.workflowWriterFactory(pe.Ledger, input)
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = pe.newWorkflowParams(writer, configParams); err == nil {
			var workers []task.Worker

			if pe.validateProp(pe.Ledger) {
				var propWorkers []task.Worker

				if propWorkers, err = pe.makePropsPretenders(workflowParams.Workflow); err == nil {
					workers = append(workers, propWorkers...)
				}
			}

			if err == nil {
				var plan = task.NewPlan("WorkflowProp", pe.Ledger.Logger)

				workers = append(workers, task.NewPretender(workflowParams, pe.workflowTaskFactory(writer)))
				plan.Sequence(workers...)
				return
			}
		}
	}

	lang.HandleExit(err)
}

func (pe *PropExecutor) Proceed() {
	var configParams = pe.newConfigParams()
	var input string
	var err error

	if input, err = configParams.EnsureConfigPath(); err == nil {
		var writer = pe.workflowWriterFactory(pe.Ledger, input)
		var workflowParams *broker.WorkflowParams

		if workflowParams, err = pe.newWorkflowParams(writer, configParams); err == nil {
			var workers []task.Worker

			if pe.validateProp(pe.Ledger) {
				var propWorkers []task.Worker

				if propWorkers, err = pe.makePropsWorkers(workflowParams.Workflow); err == nil {
					workers = append(workers, propWorkers...)
				}
			}

			if err == nil {
				var plan = task.NewPlan("WorkflowProp", pe.Ledger.Logger)

				workers = append(workers, task.NewWorker(workflowParams, pe.workflowTaskFactory(writer)))
				plan.PrintReportsOnly = true
				plan.Sequence(workers...)
				return
			}
		}
	}

	lang.HandleExit(err)
}

func (pe *PropExecutor) Remove(workflowArg string, nodeArg string, key string) error {
	var err error

	if err = pe.findWorkflowAndNode(workflowArg, nodeArg); err == nil {
		pe.rmProps = append(pe.rmProps, key)
		pe.Cli.Exec(pe.Ledger, pe)
		return nil
	}

	return err
}

func (pe *PropExecutor) findWorkflowAndNode(workflowArg, nodeArg string) error {
	var sn *broker.Solution
	var err error

	if sn, err = pe.getSolution(); err == nil {
		var fn *broker.Function

		if pe.workflow, err = sn.FindWorkflowByHandle(workflowArg); err != nil {
			return fmt.Errorf("workflow handle [%s] not found", workflowArg)
		}

		if fn, err = pe.getFunction(nodeArg); err == nil {
			if pe.workflowNode, err = pe.workflow.FindNodeBySf(fn); err == nil {
				return nil
			}

			return fmt.Errorf("function [%s/%s:%s] is not a member of workflow [%s]", fn.Oem, fn.Handle, fn.Version, workflowArg)
		} else if pe.workflowNode, err = pe.workflow.FindNodeByHandle(nodeArg); err == nil && pe.workflowNode.Sf != nil {
			if pe.function, err = pe.findFunctionByOemHandle(pe.Ledger.WorkDir, pe.workflowNode.Sf.Oem, pe.workflowNode.Sf.Handle); err != nil {
				return err
			}

			return nil
		} else if err == nil && pe.workflowNode.Sf == nil {
			return errorNoFunction
		}

		return fmt.Errorf("node [%s] is not a member of workflow [%s]", nodeArg, workflowArg)
	}

	return err
}

func (pe *PropExecutor) makePropsPretenders(workflow *broker.Workflow) ([]task.Worker, error) {
	return pe.makePropTaskWorkers(workflow, pe.MakeSyncPretenders, task.NewPretender)
}

func (pe *PropExecutor) makePropsWorkers(workflow *broker.Workflow) ([]task.Worker, error) {
	return pe.makePropTaskWorkers(workflow, pe.MakeSyncWorkers, task.NewWorker)
}

func (pe *PropExecutor) makePropTaskWorkers(workflow *broker.Workflow,
	syncFactory taskWorkerSyncFactory, propFactory taskWorkerPropFactory) ([]task.Worker, error) {
	var result []task.Worker
	var err error

	if pe.external != nil {
		return nil, errors.New("function is unknown, and external entities are not supported yet")
	} else if pe.function != nil {
		var datalinks = append(pe.function.DataSources, pe.function.DataStores...)

		if len(datalinks) > 0 {
			var datalinkWorkers []task.Worker

			if datalinkWorkers, err = syncFactory(datalinks, pe.Ledger, pe.optionNoSync); err == nil {
				result = append(result, datalinkWorkers...)
			}
		}

		if err == nil {
			var workflowPropParams = pe.newWorkflowPropParams(workflow)

			result = append(result, propFactory(workflowPropParams, pe.workflowPropTaskFactory()))
		} else {
			return nil, err
		}
	} else {
		return nil, errors.New("no function could be found for the specified node")
	}

	return result, nil
}

func (pe *PropExecutor) newWorkflowParams(writer *workflowWriter, configParams *shared.ConfigParams) (*broker.WorkflowParams, error) {
	var edited *broker.Workflow
	var err error

	for _, k := range pe.rmProps {
		writer.removeProp(pe.workflow.Handle, pe.workflowNode.Handle, k)
	}

	for k, v := range pe.addProps {
		if _, err = writer.addProp(pe.workflow.Handle, pe.workflowNode.Handle, k, v); err != nil {
			return nil, err
		}
	}

	if edited, err = writer.GetWorkflowByHandle(pe.workflow.Handle); err == nil {
		return &broker.WorkflowParams{
			ConfigParams:   *configParams,
			Workflow:       edited,
			WorkflowUpdate: true,
		}, nil
	}

	return nil, err
}

func (pe *PropExecutor) newWorkflowPropParams(workflow *broker.Workflow) *broker.WorkflowPropParams {
	var varSpecs []shared.VarSpec

	for _, propSpec := range pe.function.PropSpecs {
		varSpecs = append(varSpecs, propSpec.VarSpec())
	}

	return &broker.WorkflowPropParams{
		Workflow: workflow,
		VarSpecs: varSpecs,
	}
}

type PropOptions struct {
	optionNoSync       *config.BoolOption
	optionNoValidation *config.BoolOption
}

func (po PropOptions) allDefiners() []config.Definer {
	return []config.Definer{
		po.optionNoSync,
		po.optionNoValidation,
	}
}

func (po PropOptions) validateProp(ledger *config.Ledger) bool {
	if po.optionNoValidation != nil {
		return !ledger.GetBool(po.optionNoValidation)
	}

	return false
}

func NewProp(ledger *config.Ledger, wfCli *Cli) *cobra.Command {
	var propAddOptions = NewAddPropOptions()
	var propEditOptions = NewEditPropOptions()
	var propAddCmd = prop.NewAddProp(newPropAddExecutorFactory(ledger, wfCli, propAddOptions))
	var propEditCmd = prop.NewEditProp(newPropEditExecutorFactory(ledger, wfCli, propEditOptions))
	var propRmCmd = prop.NewRemoveProp(newPropRemoveExecutorFactory(ledger, wfCli, &PropOptions{}))
	var propCmd = &cobra.Command{
		Use:     "prop",
		Aliases: []string{"pr"},
		Short:   "Manages property values for Workflow Nodes",
	}

	propCmd.AddCommand(propAddCmd)
	propCmd.AddCommand(propEditCmd)
	propCmd.AddCommand(propRmCmd)
	ledger.Register(propAddCmd, propAddOptions.allDefiners()...)
	ledger.Register(propEditCmd, propEditOptions.allDefiners()...)
	return propCmd
}

func (pe *PropExecutor) getFunction(nodeArg string) (*broker.Function, error) {
	if pe.function == nil {
		var err error

		if pe.function, err = pe.findFunctionInPath(nodeArg); err != nil {
			return nil, err
		}
	}

	return pe.function, nil
}

func (pe *PropExecutor) getSolution() (*broker.Solution, error) {
	if pe.solution == nil {
		var snReader = config.NewSolutionReader(pe.Ledger).WithConfigPath(pe.Ledger.WorkDir)
		var err error

		if pe.solution, err = snReader.ReadName(pe.Ledger.ConfigName); err != nil {
			return nil, err
		} else if pe.solution == nil {
			return nil, errorNoSolution
		}
	}

	return pe.solution, nil
}

func NewPropExecutor(ctx context.Context, ledger *config.Ledger, wfCli *Cli, options *PropOptions) *PropExecutor {
	return &PropExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		SyncBridge:  dk.NewSyncBridgeBuilder().Build(),
		PropOptions: options,
		addProps:    make(map[string]string),

		workflowPropTaskFactory: broker.NewWorkflowPropTask,
		workflowTaskFactory:     broker.NewWorkflowUpdateTask,
		workflowWriterFactory:   newWorkflowWriter,
	}
}

func NewAddPropOptions() *PropOptions {
	return &PropOptions{
		optionNoSync: cli.Options.Workflows.NoPropSync().
			WithKeys(&schema.Genaiz.Workflow.Props.Add.NoSync).
			BuildBoolOption(),
		optionNoValidation: cli.Options.Workflows.NoPropValidation().
			WithKeys(&schema.Genaiz.Workflow.Props.Add.NoValidation).
			BuildBoolOption(),
	}
}

func NewEditPropOptions() *PropOptions {
	return &PropOptions{
		optionNoSync: cli.Options.Workflows.NoPropSync().
			WithKeys(&schema.Genaiz.Workflow.Props.Edit.NoSync).
			BuildBoolOption(),
		optionNoValidation: cli.Options.Workflows.NoPropValidation().
			WithKeys(&schema.Genaiz.Workflow.Props.Edit.NoValidation).
			BuildBoolOption(),
	}
}

func newPropAddExecutorFactory(ledger *config.Ledger, wfCli *Cli, options *PropOptions) func(*cobra.Command) prop.AddExecutor {
	return func(cmd *cobra.Command) prop.AddExecutor {
		return NewPropExecutor(cmd.Context(), ledger, wfCli, options)
	}
}

func newPropEditExecutorFactory(ledger *config.Ledger, wfCli *Cli, options *PropOptions) func(*cobra.Command) prop.EditExecutor {
	return func(cmd *cobra.Command) prop.EditExecutor {
		return NewPropExecutor(cmd.Context(), ledger, wfCli, options)
	}
}

func newPropRemoveExecutorFactory(ledger *config.Ledger, wfCli *Cli, options *PropOptions) func(*cobra.Command) prop.RemoveExecutor {
	return func(cmd *cobra.Command) prop.RemoveExecutor {
		return NewPropExecutor(cmd.Context(), ledger, wfCli, options)
	}
}
