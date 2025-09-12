package sn

import (
	"context"
	"path/filepath"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
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
			Handle: ce.Ledger.GetString(ce.optionWorkflowHandle),
			Name:   ce.Ledger.GetString(ce.optionWorkflowName),
		},
	}
}

type CreateOptions struct {
	PublishOptions
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
		Example: "genaiz sn create solution-1 --workflow=workflow-1",
		Args: cobra.MatchAll(cobra.MaximumNArgs(1),
			cli.ArgsFolderValidator("solution", config.Validation.Handle)),
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
	var cmd = "Create"
	var handleOption = newOptionHandle(cmd)
	var nameOption = newOptionName(handleOption, cmd)
	var wfHandleOption = newOptionWorkflowHandle()

	return &CreateOptions{
		PublishOptions: PublishOptions{
			optionConfigType:  newOptionConfigType(cmd),
			optionDescription: newOptionDescription(nameOption, cmd),
			optionHandle:      handleOption,
			optionName:        nameOption,
			optionOem:         newOptionOem(cmd),
			optionVersion:     newOptionVersion(cmd),
		},
		optionWorkflowHandle: wfHandleOption,
		optionWorkflowName:   newOptionWorkflowName(wfHandleOption),
	}
}

func newOptionConfigType(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Solution." + cmd + ".ConfigType",
			Env:          "SN_" + strings.ToUpper(cmd) + "_CONFIG_TYPE",
			Param:        "configType",
			Usage:        "sets the format of the configuration file to modify. Supported values are \"yaml\", \"toml\", \"json\" or \"none\"",
			DefaultValue: "yaml",
			Validator:    config.Optionally(config.AnyOfEnumerated(shared.ConfigTypes)),
		},
	}
}

func newOptionDescription(defaultOption *config.StringOption, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Solution." + cmd + ".Description",
			Env:   "SN_" + strings.ToUpper(cmd) + "_DESCRIPTION",
			Param: "description",
			Usage: "description of the solution created",
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
			Key:       "Solution." + cmd + ".Handle",
			Env:       "SN_" + strings.ToUpper(cmd) + "_HANDLE",
			Param:     "handle",
			Usage:     "handle of the solution to " + strings.ToLower(cmd),
			Validator: config.Validation.Handle,
		},
	}
}

func newOptionName(defaultOption *config.StringOption, cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Solution." + cmd + ".Name",
			Env:   "SN_" + strings.ToUpper(cmd) + "_NAME",
			Param: "name",
			Short: "n",
			Usage: "name of the solution to create",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Validation.RequiredName,
		},
	}
}

func newOptionOem(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:       "Solution." + cmd + ".Oem",
			Env:       "SN_" + cmd + "_OEM",
			Param:     "oem",
			Usage:     "oem of the solution",
			Validator: config.Optionally(config.Validation.Oem),
		},
	}
}

func newOptionVersion(cmd string) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Solution." + cmd + ".Version",
			Env:          "SOLUTION_" + cmd + "_VERSION",
			Param:        "version",
			Usage:        "version of the solution",
			DefaultValue: "0.1.0",
			Validator:    config.Optionally(config.Validation.Version),
		},
	}
}

func newOptionWorkflowHandle() *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:          "Solution.Create.Workflow.Handle",
			Env:          "SOLUTION_CREATE_WORKFLOW_HANDLE",
			Param:        "wf.handle",
			Usage:        "handle of the default workflow created with the solution",
			DefaultValue: "default",
			Validator:    config.Optionally(config.Validation.Handle),
		},
	}
}

func newOptionWorkflowName(defaultOption *config.StringOption) *config.StringOption {
	return &config.StringOption{
		Option: config.Option{
			Key:   "Solution.Create.Workflow.Name",
			Env:   "SOLUTION_CREATE_WORKFLOW_NAME",
			Param: "wf.name",
			Usage: "name of the default workflow created with the solution",
			DefaultGetter: func(ledger *config.Ledger) any {
				return ledger.GetString(defaultOption)
			},
			Validator: config.Optionally(config.Validation.RequiredName),
		},
	}
}
