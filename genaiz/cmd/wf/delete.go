package wf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type DeleteExecutor struct {
	BaseExecutor
	*DeleteOptions

	workflowArg string

	workflowTaskFactory WorkflowTaskFactory
}

func (de *DeleteExecutor) Display() {
	de.Ledger.DisplayOptionsWithMap(&map[string]string{
		"workflow": de.workflowArg,
	},
		&de.optionConfigType.Option)
}

func (de *DeleteExecutor) Pretend() {
	var params = de.makeWorkflowParams()
	var writer = newWorkflowWriter(de.Ledger, params.GetConfigFile())

	de.workflowTaskFactory(writer).Pretend(params, de.Ledger.Logger)
}

func (de *DeleteExecutor) Proceed() {
	var params = de.makeWorkflowParams()
	var writer = newWorkflowWriter(de.Ledger, params.GetConfigFile())
	var plan = task.NewPlan("Workflow", de.Ledger.Logger)

	task.Single(plan, params, de.workflowTaskFactory(writer))
}

func (de *DeleteExecutor) makeWorkflowParams() *broker.WorkflowParams {
	var configType, err = de.Ledger.GetConfigType(de.optionConfigType)

	lang.HandleExit(err)
	return &broker.WorkflowParams{
		ConfigParams: shared.ConfigParams{
			ConfigName: de.Ledger.ConfigName,
			ConfigType: configType,
		},
		Workflow: &broker.Workflow{
			Handle: de.workflowArg,
		},
	}
}

type DeleteOptions struct {
	optionConfigType *config.StringOption
}

func (co DeleteOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionConfigType,
	}
}

func NewDelete(ledger *config.Ledger, cli *Cli) *cobra.Command {
	var deleteOptions = NewDeleteOptions()
	var deleteCmd = &cobra.Command{
		Use:     "delete",
		Short:   "Deletes a Workflow from the local config",
		Long:    "Deletes a Workflow from the local config, providing there is under the current workdir",
		Example: "genaiz wf delete workflow-1",
		Args:    cobra.ExactArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			cli.Exec(ledger, NewDeleteExecutor(cmd.Context(), ledger, cli, deleteOptions, args[0]))
		},
	}

	ledger.Register(deleteCmd, deleteOptions.allDefiners()...)
	return deleteCmd
}

func NewDeleteExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *DeleteOptions, workflowArg string) *DeleteExecutor {
	return &DeleteExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		DeleteOptions: options,

		workflowArg: workflowArg,

		workflowTaskFactory: broker.NewWorkflowDeleteTask,
	}
}

func NewDeleteOptions() *DeleteOptions {
	var cmd = "delete"

	return &DeleteOptions{
		optionConfigType: newOptionConfigType(cmd),
	}
}
