package wf

import (
	"context"
	"strings"

	"github.com/iancoleman/strcase"
	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type CreateExecutor struct {
	BaseExecutor
	*CreateOptions
	FolderPath string

	workflowTaskFactory WorkflowTaskFactory
}

func (ce *CreateExecutor) Display() {
	ce.Ledger.DisplayOptionsWithMap(&map[string]string{
		"folder": ce.FolderPath,
	},
		&ce.CreateOptions.optionConfigType.Option,
		&ce.optionDescription.Option,
		&ce.optionHandle.Option,
		&ce.optionName.Option,
	)
}

func (ce *CreateExecutor) Pretend() {
	var params = ce.makeWorkflowParams()
	var writer = newWorkflowWriter(ce.Ledger, params.GetConfigPath())

	ce.workflowTaskFactory(writer).Pretend(params, ce.Ledger.Logger)
}

func (ce *CreateExecutor) Proceed() {
	var params = ce.makeWorkflowParams()
	var writer = newWorkflowWriter(ce.Ledger, params.GetConfigPath())
	var plan = task.NewPlan("Workflow", ce.Ledger.Logger)

	task.Single(plan, params, ce.workflowTaskFactory(writer))
}

func (ce *CreateExecutor) makeWorkflowParams() *broker.WorkflowParams {
	var configType, err = ce.Ledger.GetConfigType(ce.optionConfigType)

	lang.HandleExit(err)
	return &broker.WorkflowParams{
		ConfigParams: shared.ConfigParams{
			ConfigName:   ce.Ledger.ConfigName,
			ConfigType:   configType,
			ConfigFolder: ce.FolderPath,
		},
		Workflow: &broker.Workflow{
			Description: ce.Ledger.GetString(ce.optionDescription),
			Handle:      ce.Ledger.GetString(ce.optionHandle),
			Name:        ce.Ledger.GetString(ce.optionName),
		},
	}
}

type CreateOptions struct {
	optionConfigType  *config.StringOption
	optionDescription *config.StringOption
	optionHandle      *config.StringOption
	optionName        *config.StringOption
}

func (co *CreateOptions) allDefiners() []config.Definer {
	return []config.Definer{
		co.optionConfigType,
		co.optionDescription,
		co.optionHandle,
		co.optionName,
	}
}

func NewCreate(ledger *config.Ledger, wfCli *Cli) *cobra.Command {
	var createOptions = NewCreateOptions()
	var createCmd = &cobra.Command{
		Use:     "create [SOLUTION_PATH]",
		Short:   "Creates a Workflow from scratch",
		Long:    "Creates a Workflow from scratch, optionally using a selected template",
		Example: "genaiz wf create solution-1 --handle=workflow-1",
		Args: cobra.MatchAll(cobra.MaximumNArgs(1),
			cli.ArgsFolderValidator("solution", config.Validation.Handle)),
		Run: func(cmd *cobra.Command, args []string) {
			var wdp = dirz.OptionalWorkingDir(args...)

			if folder, err := wdp(); err == nil {
				wfCli.Exec(ledger, NewCreateExecutor(cmd.Context(), ledger, wfCli, createOptions, folder))
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
			Cli:     cli,
			Context: ctx,
			Ledger:  ledger,
		},
		CreateOptions: options,
		FolderPath:    folderPath,

		workflowTaskFactory: broker.NewWorkflowUpdateTask,
	}
}

func NewCreateOptions() *CreateOptions {
	var cmd = "create"
	var handleOption = newOptionHandle(cmd)
	var nameOption = newOptionCreateName(handleOption)

	return &CreateOptions{
		optionConfigType:  newOptionConfigType(cmd),
		optionDescription: newOptionDescription(nameOption),
		optionHandle:      handleOption,
		optionName:        nameOption,
	}
}

func newOptionConfigType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Workflow." + strcase.ToCamel(cmd) + ".ConfigType",
			Env:          "WF_" + strings.ToUpper(cmd) + "_CONFIG_TYPE",
			Param:        "configType",
			Usage:        "sets the format of the configuration file to modify. Supported values are \"yaml\", \"toml\", \"json\" or \"none\"",
			DefaultValue: "yaml",
			Validator:    config.Optionally(config.AnyOfEnumerated(shared.ConfigTypes)),
		},
	}
}

func newOptionDescription(defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow.Create.Description",
			Env:   "WF_CREATE_DESCRIPTION",
			Param: "description",
			Usage: "description of the workflow created",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Validation.Blob,
		},
	}
}

func newOptionHandle(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "Workflow." + strcase.ToCamel(cmd) + ".Handle",
			Env:       "WF_" + strings.ToUpper(cmd) + "_HANDLE",
			Param:     "handle",
			Usage:     "handle of the workflow to " + strings.ToLower(cmd),
			Validator: config.Validation.Handle,
		},
	}
}

func newOptionCreateName(defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Workflow.Create.Name",
			Env:   "WF_CREATE_NAME",
			Param: "name",
			Short: "n",
			Usage: "name of the workflow to create",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Validation.RequiredName,
		},
	}
}
