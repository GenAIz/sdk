package wf

import (
	"context"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
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
	var params *broker.WorkflowParams
	var err error

	if params, err = de.makeWorkflowParams(); err == nil {
		var writer = newWorkflowWriter(de.Ledger, params.GetConfigFile())

		de.workflowTaskFactory(writer).Pretend(params, de.Ledger.Logger)
	}

	lang.HandleExit(err)
}

func (de *DeleteExecutor) Proceed() {
	var params *broker.WorkflowParams
	var err error

	if params, err = de.makeWorkflowParams(); err == nil {
		var writer = newWorkflowWriter(de.Ledger, params.GetConfigFile())
		var plan = task.NewPlan("Workflow", de.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, de.workflowTaskFactory(writer))
	}

	lang.HandleExit(err)
}

func (de *DeleteExecutor) makeWorkflowParams() (*broker.WorkflowParams, error) {
	var configType *shared.ConfigType
	var err error

	if configType, err = de.Ledger.GetConfigType(de.optionConfigType); err == nil {
		return &broker.WorkflowParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: de.Ledger.ConfigName,
				ConfigType: configType,
			},
			Workflow: &broker.Workflow{
				Handle: de.workflowArg,
			},
		}, nil
	}

	return nil, err
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
	var deleteOptions = NewDeleteOptions(cli)
	var deleteCmd = &cobra.Command{
		Use:     "delete WORKFLOW_HANDLE",
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

func NewDeleteOptions(wfCli *Cli) *DeleteOptions {
	return &DeleteOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Create.ConfigType).
			WithDefaultGetter(wfCli.WorkingConfigType()).
			BuildStringOption(),
	}
}
