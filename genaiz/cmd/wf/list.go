package wf

import (
	"errors"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task/broker"
)

var (
	errorFqdnInvalid = errors.New("invalid fqdn")
)

type ListExecutor struct {
	*ListOptions
	ledger                     *config.Ledger
	accountParams              config.AccountParametric
	printerParams              cli.PrinterParametric
	userWorkflowFacadeProvider func() mgmt.UserWorkflowFacade
}

func (le ListExecutor) List(fqdnOrFolder string) error {
	var workflowList = le.userWorkflowFacadeProvider()
	var printer = le.printerParams.Printer()
	var solutions []mgmt.UserWorkflow
	var params *broker.WorkflowListParams
	var configPaths []string
	var err error

	if strings.Contains(fqdnOrFolder, ":") {
		params, err = le.newListParams(fqdnOrFolder)
	} else {
		var searchPath = fqdnOrFolder

		if searchPath == "" {
			searchPath = le.ledger.WorkDir
		}

		if configPaths, err = filez.FindNamedFilesRecursively(searchPath, le.ledger.ConfigName); err == nil {
			var pathGraphers = le.makePathGraphers(configPaths)

			workflowList = workflowList.WithPathGraphers(searchPath, pathGraphers)
		}
	}

	if err == nil {
		solutions, err = workflowList.
			WithParams(params).
			WithLogger(le.ledger.Logger).
			Provider().
			Get()

		if err == nil {
			return printer.Print(solutions)
		}
	}

	return printer.Error(err)
}

func (le ListExecutor) makePathGraphers(paths []string) map[string]broker.SolutionGrapher {
	var result = make(map[string]broker.SolutionGrapher)

	for _, path := range paths {
		var solutionReader = config.NewSolutionReader(le.ledger)

		if grapher, err := solutionReader.GraphFile(path); err == nil {
			result[path] = grapher
		}
	}

	return result
}

func (le ListExecutor) newListParams(fqdnOrFolder string) (*broker.WorkflowListParams, error) {
	var oem, handle, ver = broker.ParseFqdnVersion(fqdnOrFolder)

	if oem != "" && handle != "" && ver != "" {
		var brokerParams = le.accountParams.BrokerParams()

		return &broker.WorkflowListParams{
			Broker:  *brokerParams,
			Oem:     oem,
			Handle:  handle,
			Version: ver,
		}, nil
	}

	return nil, errorFqdnInvalid
}

type ListOptions struct {
	optionAccount     *config.StringOption
	optionJsonPrinter *config.BoolOption
}

func (lo ListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionAccount,
		lo.optionJsonPrinter,
	}
}

func NewList(ledger *config.Ledger) *cobra.Command {
	var listOptions = NewListOptions()
	var listCmd = &cobra.Command{
		Use:     "list [FOLDER|FQDN]",
		Aliases: []string{"ls"},
		Short:   "Lists sessions, filtering by path or fqdn",
		Long:    "Lists sessions, filtering by path for local solutions or fqdn for remote",
		Example: "genaiz workflow list com.genaiz/my-handle:1.0.0 --account=dev.genaiz.com",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewListExecutor(ledger, listOptions)
			var filter = cli.ArgsOptionalSingle(args)
			var err = exec.List(filter)

			lang.HandleExit(err)
		},
	}

	ledger.Register(listCmd, listOptions.allDefiners()...)
	cli.AutoBridge.Accounts().Option(listCmd, ledger, listOptions.optionAccount)
	cli.AutoBridge.Solutions().Arguments(listCmd, ledger)
	return listCmd
}

func NewListExecutor(ledger *config.Ledger, options *ListOptions) *ListExecutor {
	return &ListExecutor{
		ListOptions: options,
		ledger:      ledger,

		accountParams:              config.NewAccountParams(ledger, options.optionAccount),
		printerParams:              cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		userWorkflowFacadeProvider: mgmt.NewUserWorkflowFacade,
	}
}

func NewListOptions() *ListOptions {
	return &ListOptions{
		optionAccount: cli.Options.Workflows.Account().
			WithKeys(&schema.Genaiz.Workflow.List.Account).
			BuildStringOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Workflow.List.Printer).
			BuildBoolOption(),
	}
}
