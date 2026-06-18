package sn

import (
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

type ListExecutor struct {
	*ListOptions
	ledger                     *config.Ledger
	accountParams              config.AccountParametric
	printerParams              cli.PrinterParametric
	userSolutionFacadeProvider func() mgmt.UserSolutionFacade
}

func (le ListExecutor) List(filter string) error {
	var solutionList = le.userSolutionFacadeProvider()
	var printer = le.printerParams.Printer()
	var solutions, err = solutionList.WithParams(le.newListParams(filter)).
		WithLogger(le.ledger.Logger).
		Filtering(filter).
		Get()

	if err == nil {
		return printer.Print(solutions)
	}

	return printer.Error(err)
}

func (le ListExecutor) collectLocal(filter string) []broker.Solution {
	var folder = filter

	if folder == "" {
		folder = le.ledger.WorkDir
	}

	// it's the only character we don't support in file paths used in solution addressing
	if !strings.Contains(folder, ":") {
		var result []broker.Solution
		var files []string
		var err error

		if files, err = filez.FindNamedFilesRecursively(folder, le.ledger.ConfigName); err == nil && len(files) > 0 {
			var solutionReader = config.NewSolutionReader(le.ledger)

			for _, file := range files {
				var solution *broker.Solution

				if solution, err = solutionReader.ReadFile(file); err == nil && solution != nil {
					result = append(result, *solution)
				}
			}

			return result
		}

		le.ledger.Logger.Warnf("No local solutions found under [%s]", folder)
	}

	return nil
}

func (le ListExecutor) newListParams(filter string) *broker.SolutionListParams {
	var brokerParams = le.accountParams.BrokerParams()
	var localSolutions []broker.Solution
	var oem string

	if !le.ledger.GetBool(le.optionAccountOnly) {
		localSolutions = le.collectLocal(filter)
	}

	if parts := strings.Split(filter, "/"); len(parts) > 1 {
		oem = parts[0]
	} else {
		oem = filter
	}

	return &broker.SolutionListParams{
		Broker: *brokerParams,
		Local:  localSolutions,
		Oem:    oem,
	}
}

type ListOptions struct {
	optionAccount     *config.StringOption
	optionAccountOnly *config.BoolOption
	optionJsonPrinter *config.BoolOption
}

func (lo ListOptions) allDefiners() []config.Definer {
	return []config.Definer{
		lo.optionAccount,
		lo.optionAccountOnly,
		lo.optionJsonPrinter,
	}
}

func NewList(ledger *config.Ledger) *cobra.Command {
	var listOptions = NewListOptions()
	var listCmd = &cobra.Command{
		Use:     "list FOLDER|OEM[/HANDLE][:VERSION]",
		Aliases: []string{"ls"},
		Short:   "Lists sessions, filtering by path or fqdn",
		Long:    "Lists sessions, filtering by path for local solutions or fqdn for remote",
		Example: "genaiz solution list com.genaiz/my-handle --account=dev.genaiz.com",
		Args:    cobra.MaximumNArgs(1),
		Run: func(cmd *cobra.Command, args []string) {
			var exec = NewListExecutor(ledger, listOptions)
			var filter = cli.ArgsOptionalSingle(args)
			var err = exec.List(filter)

			lang.HandleExit(err)
		},
	}

	ledger.Register(listCmd, listOptions.allDefiners()...)
	return listCmd
}

func NewListExecutor(ledger *config.Ledger, options *ListOptions) *ListExecutor {
	return &ListExecutor{
		ListOptions: options,
		ledger:      ledger,

		accountParams:              config.NewAccountParams(ledger, options.optionAccount),
		printerParams:              cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		userSolutionFacadeProvider: mgmt.NewUserSolutionFacade,
	}
}

func NewListOptions() *ListOptions {
	var account = cli.Options.Solutions.Account().
		WithKeys(&schema.Genaiz.Solution.List.Account).
		BuildStringOption()

	return &ListOptions{
		optionAccount: account,
		optionAccountOnly: cli.Options.Solutions.AccountOnly().
			WithKeys(&schema.Genaiz.Solution.List.AccountOnly).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(account) != ""
			}).
			BuildBoolOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.Solution.List.Printer).
			BuildBoolOption(),
	}
}
