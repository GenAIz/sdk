package sn

import (
	"context"
	"path/filepath"

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

type SolutionCreateTaskFactory func(broker.SolutionWriter) *task.Task[broker.SolutionParams]

type WorkflowTaskFactory func(broker.WorkflowWriter) *task.Task[broker.WorkflowParams]

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions

	solutionTaskFactory SolutionCreateTaskFactory
	workflowTaskFactory WorkflowTaskFactory
}

func (ce *CreateExecutor) Display() {
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.folderPath,
	},
		&ce.CreateOptions.optionConfigType.Option,
		&ce.optionDescription.Option,
		&ce.optionHandle.Option,
		&ce.optionName.Option,
		&ce.optionOem.Option,
		&ce.optionVersion.Option,
		&ce.optionWorkflowDesc.Option,
		&ce.optionWorkflowHandle.Option,
		&ce.optionWorkflowName.Option,
	)
}

func (ce *CreateExecutor) Pretend() {
	var configParams = ce.makeConfigParams(ce.optionConfigType)
	var snParams = ce.makeSolutionParams(configParams)
	var snWriter = config.NewSolutionWriter().Read(ce.Ledger, snParams.GetConfigPath())
	var wfParams = ce.makeWorkflowParams(configParams)
	var wfWriter = config.NewWorkflowWriter().Read(ce.Ledger, snParams.GetConfigPath())
	var plan = task.NewPlan("Create", ce.Ledger.Logger)

	plan.ContinueOnFailure = true
	plan.Sequence(
		task.NewPretender(snParams, ce.solutionTaskFactory(snWriter)),
		task.NewPretender(wfParams, ce.workflowTaskFactory(wfWriter)),
	)
}

func (ce *CreateExecutor) Proceed() {
	var configParams = ce.makeConfigParams(ce.optionConfigType)
	var snParams = ce.makeSolutionParams(configParams)
	var snWriter = config.NewSolutionWriter().Read(ce.Ledger, snParams.GetConfigPath())
	var wfParams = ce.makeWorkflowParams(configParams)
	var wfWriter = config.NewWorkflowWriter().Read(ce.Ledger, snParams.GetConfigPath())
	var plan = task.NewPlan("Create", ce.Ledger.Logger)

	plan.ContinueOnFailure = true
	plan.PrintReportsOnly = true
	plan.Sequence(
		task.NewWorker(snParams, ce.solutionTaskFactory(snWriter)),
		task.NewWorker(wfParams, ce.workflowTaskFactory(wfWriter)),
	)
}

func (ce *CreateExecutor) makeSolutionParams(configParams *shared.ConfigParams) *broker.SolutionParams {
	return &broker.SolutionParams{
		ConfigParams: *configParams,
		Solution: &broker.Solution{
			Description: ce.Ledger.GetString(ce.optionDescription),
			Handle:      ce.Ledger.GetString(ce.optionHandle),
			Name:        ce.Ledger.GetString(ce.optionName),
			Oem:         ce.Ledger.GetString(ce.optionOem),
			Version:     ce.Ledger.GetString(ce.optionVersion),
		},
	}
}

func (ce *CreateExecutor) makeWorkflowParams(configParams *shared.ConfigParams) *broker.WorkflowParams {
	return &broker.WorkflowParams{
		ConfigParams: *configParams,
		Workflow: &broker.Workflow{
			Description: ce.Ledger.GetString(ce.optionWorkflowDesc),
			Handle:      ce.Ledger.GetString(ce.optionWorkflowHandle),
			Name:        ce.Ledger.GetString(ce.optionWorkflowName),
		},
	}
}

type CreateOptions struct {
	PublishOptions
	optionWorkflowDesc   *config.StringOption
	optionWorkflowHandle *config.StringOption
	optionWorkflowName   *config.StringOption
}

func (co CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionConfigType,
		co.optionDescription,
		co.optionHandle,
		co.optionName,
		co.optionOem,
		co.optionVersion,
		co.optionWorkflowDesc,
		co.optionWorkflowHandle,
		co.optionWorkflowName,
	}
}

func NewCreate(ledger *config.Ledger, snCli *Cli) *cobra.Command {
	var createOptions = NewCreateOptions()
	var createCmd = &cobra.Command{
		Use:     "create [SOLUTION_PATH]",
		Short:   "Creates a Solution from scratch",
		Long:    "Creates a Solution from scratch, adding a default workflow, optionally from a selected template",
		Example: "genaiz sn create solution-1 --oem=com.genaiz",
		Args: cobra.MatchAll(cobra.MaximumNArgs(1),
			cli.ArgsOptionalFolder("solution", 1, config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp = dirz.OptionalWorkingDir(args...)

			if folder, err := wdp(); err == nil {
				ledger.InitValue(createOptions.optionHandle, filepath.Base(folder))
				snCli.Exec(ledger, NewCreateExecutor(cmd.Context(), ledger, snCli, createOptions, folder))
			} else {
				lang.HandleExit(err)
			}
		},
	}

	ledger.Register(createCmd, createOptions.allDefiners()...)
	return createCmd
}

func NewCreateExecutor(ctx context.Context, ledger *config.Ledger, cli *Cli, options *CreateOptions, folderPath string) *CreateExecutor {
	return &CreateExecutor{
		BaseExecutor: BaseExecutor{
			Cli:        cli,
			Context:    ctx,
			Ledger:     ledger,
			folderPath: folderPath,
		},
		CreateOptions: options,

		solutionTaskFactory: broker.NewSolutionUpdateTask,
		workflowTaskFactory: broker.NewWorkflowUpdateTask,
	}
}

func NewCreateOptions() *CreateOptions {
	var handleOption = cli.Options.Solutions.Handle().
		WithKeys(&schema.Genaiz.Solution.Create.Handle).
		BuildStringOption()
	var nameOption = cli.Options.Solutions.Name().
		WithKeys(&schema.Genaiz.Solution.Create.Name).
		WithDefaultGetter(func(ledger *config.Ledger) any {
			return ledger.GetString(handleOption)
		}).
		BuildStringOption()

	return &CreateOptions{
		PublishOptions: PublishOptions{
			optionConfigType: cli.Options.Configs.Type().
				WithKeys(&schema.Genaiz.Solution.Create.ConfigType).
				WithDefaultValue("yaml").
				BuildStringOption(),
			optionDescription: cli.Options.Solutions.Description().
				WithKeys(&schema.Genaiz.Solution.Create.Description).
				WithDefaultGetter(func(ledger *config.Ledger) any {
					return ledger.GetString(nameOption)
				}).
				BuildStringOption(),
			optionHandle: handleOption,
			optionName:   nameOption,
			optionOem: cli.Options.Solutions.Oem().
				WithKeys(&schema.Genaiz.Solution.Create.Oem).
				WithValidator(config.Optionally(config.Validation.Oem)).
				BuildStringOption(),
			optionVersion: cli.Options.Solutions.Version().
				WithKeys(&schema.Genaiz.Solution.Create.Version).
				WithValidator(config.Optionally(config.Validation.Version)).
				BuildStringOption(),
		},
		optionWorkflowDesc: cli.Options.Solutions.WorkflowDesc().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Description).
			BuildStringOption(),
		optionWorkflowHandle: cli.Options.Solutions.WorkflowHandle().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Handle).
			BuildStringOption(),
		optionWorkflowName: cli.Options.Solutions.WorkflowName().
			WithKeys(&schema.Genaiz.Solution.Create.Workflow.Name).
			BuildStringOption(),
	}
}
