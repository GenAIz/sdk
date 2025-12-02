package wf

import (
	"context"
	"fmt"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions

	workflowArg string

	workflowTaskFactory WorkflowTaskFactory
}

func (ce *CreateExecutor) Display() {
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.Ledger.WorkDir,
		"handle": ce.workflowArg,
	},
		&ce.CreateOptions.optionConfigType.Option,
		&ce.optionDescription.Option,
		&ce.optionName.Option,
	)
}

func (ce *CreateExecutor) Pretend() {
	var params *broker.WorkflowParams
	var err error

	if params, err = ce.makeWorkflowParams(); err == nil {
		var writer = newWorkflowWriter(ce.Ledger, params.GetConfigPath())

		ce.workflowTaskFactory(writer).Pretend(params, ce.Ledger.Logger)
	}

	lang.HandleExit(err)
}

func (ce *CreateExecutor) Proceed() {
	var params *broker.WorkflowParams
	var err error

	if params, err = ce.makeWorkflowParams(); err == nil {
		var writer = newWorkflowWriter(ce.Ledger, params.GetConfigPath())
		var plan = task.NewPlan("Workflow", ce.Ledger.Logger)

		plan.PrintReportsOnly = true
		task.Single(plan, params, ce.workflowTaskFactory(writer))
	}

	lang.HandleExit(err)
}

func (ce *CreateExecutor) makeWorkflowParams() (*broker.WorkflowParams, error) {
	var configType *shared.ConfigType
	var err error

	if configType, err = ce.Ledger.GetConfigType(ce.optionConfigType); err == nil {
		var name = ce.Ledger.GetString(ce.optionName)
		var desc = ce.Ledger.GetString(ce.optionDescription)

		if !config.Validation.Handle(ce.workflowArg) {
			return nil, fmt.Errorf("value [%s] is not a valid handle", ce.workflowArg)
		}

		if name == "" {
			name = ce.workflowArg
		}

		if desc == "" {
			desc = ce.workflowArg
		}

		return &broker.WorkflowParams{
			ConfigParams: shared.ConfigParams{
				ConfigName:   ce.Ledger.ConfigName,
				ConfigType:   configType,
				ConfigFolder: ce.Ledger.WorkDir,
			},
			Workflow: &broker.Workflow{
				Description: desc,
				Handle:      ce.workflowArg,
				Name:        name,
			},
		}, nil
	}

	return nil, err
}

type CreateOptions struct {
	optionConfigType  *config.StringOption
	optionDescription *config.StringOption
	optionName        *config.StringOption
}

func (co *CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionConfigType,
		co.optionDescription,
		co.optionName,
	}
}

func NewCreate(ledger *config.Ledger, wfCli *Cli) *cobra.Command {
	var createOptions = NewCreateOptions(wfCli)
	var createCmd = &cobra.Command{
		Use:     "create WORKFLOW_HANDLE [SOLUTION_PATH]",
		Short:   "Creates a Workflow from scratch",
		Long:    "Creates a Workflow from scratch, optionally using a selected template",
		Example: "genaiz wf create workflow-1 solution-1 --name='Workflow One'",
		Args: cobra.MatchAll(cobra.MinimumNArgs(1), cobra.MaximumNArgs(2),
			cli.ArgsOptionalFolder("solution", 2, config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp func() (string, error)
			var err error

			if len(args) == 2 {
				wdp = dirz.OptionalWorkingDir(args[1:]...)
			} else {
				wdp = dirz.OptionalWorkingDir()
			}

			if ledger.WorkDir, err = wdp(); err == nil {
				wfCli.Exec(ledger, NewCreateExecutor(cmd.Context(), ledger, wfCli, createOptions, args[0]))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(createCmd, createOptions.allDefiners()...)
	return createCmd
}

func NewCreateExecutor(ctx context.Context, ledger *config.Ledger, wfCli *Cli, options *CreateOptions, workflowArg string) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:     wfCli,
			Context: ctx,
			Ledger:  ledger,
		},
		CreateOptions: options,

		workflowArg: workflowArg,

		workflowTaskFactory: broker.NewWorkflowUpdateTask,
	}
}

func NewCreateOptions(wfCli *Cli) *CreateOptions {
	var nameOption = cli.Options.Workflows.Name().
		WithKeys(&schema.Genaiz.Workflow.Create.Name).
		WithValidator(config.Validation.Name).
		BuildStringOption()

	return &CreateOptions{
		optionConfigType: cli.Options.Configs.Type().
			WithKeys(&schema.Genaiz.Workflow.Create.ConfigType).
			WithDefaultGetter(wfCli.WorkingConfigType()).
			BuildStringOption(),
		optionDescription: cli.Options.Workflows.Description().
			WithKeys(&schema.Genaiz.Workflow.Create.Description).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(nameOption)
			}).
			BuildStringOption(),
		optionName: nameOption,
	}
}
