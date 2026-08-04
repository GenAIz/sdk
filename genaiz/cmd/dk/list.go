package dk

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
	ledger *config.Ledger

	accountParams config.AccountParametric
	printerParams cli.PrinterParametric

	userDataLinkFacadeProvider func() mgmt.UserDataLinkFacade
}

func (le ListExecutor) List(filter string) error {
	var dataLinkList = le.userDataLinkFacadeProvider()
	var printer = le.printerParams.Printer()
	var dataLinks, err = dataLinkList.WithParams(le.newListParams(filter)).
		WithLogger(le.ledger.Logger).
		Filtering(filter).
		Get()

	if err == nil {
		return printer.Print(dataLinks)
	}

	return printer.Error(err)
}

func (le ListExecutor) findLocalLinks(folder string) []broker.DataLink {
	var result []broker.DataLink
	var files []string
	var err error

	if files, err = filez.FindNamedFilesRecursively(folder, le.ledger.ConfigName); err == nil {
		var reader = config.NewDataLinkReader()
		var datalinks []broker.DataLink

		for _, file := range files {
			if datalinks, err = reader.ReadFile(file); err == nil {
				result = append(result, datalinks...)
			}
		}
	}

	return result
}

func (le ListExecutor) newListParams(filter string) *broker.DataLinkListParams {
	var brokerParams = le.accountParams.BrokerParams()
	var localLinks []broker.DataLink
	var oem, handle, ver string
	var folder string

	if filter == "" {
		folder = le.ledger.UserPath
	} else {
		if !strings.Contains(filter, ":") {
			// try to find datalinks locally in a folder if it exists
			folder = filter
		}

		oem, handle, ver = broker.ParseFqdnVersion(filter)
	}

	if folder != "" {
		localLinks = le.findLocalLinks(folder)
	}

	return &broker.DataLinkListParams{
		Broker:      *brokerParams,
		AccountOnly: le.ledger.GetBool(le.optionAccountOnly),
		Local:       localLinks,
		Oem:         oem,
		Handle:      handle,
		Version:     ver,
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
		Short:   "Lists datalinks, filtering by path or fqdn",
		Long:    "Lists datalinks, filtering by path for local definitions or fqdn for remote",
		Example: "genaiz datalink list com.genaiz/my-handle --account=dev.genaiz.com",
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
	return listCmd
}

func NewListExecutor(ledger *config.Ledger, options *ListOptions) *ListExecutor {
	return &ListExecutor{
		ListOptions: options,
		ledger:      ledger,

		accountParams:              config.NewAccountParams(ledger, options.optionAccount),
		printerParams:              cli.NewPrinterParam(ledger, options.optionJsonPrinter),
		userDataLinkFacadeProvider: mgmt.NewUserDataLinkFacade,
	}
}

func NewListOptions() *ListOptions {
	var account = cli.Options.DataLinks.Account().
		WithKeys(&schema.Genaiz.DataLink.List.Account).
		BuildStringOption()

	return &ListOptions{
		optionAccount: account,
		optionAccountOnly: cli.Options.DataLinks.AccountOnly().
			WithKeys(&schema.Genaiz.DataLink.List.AccountOnly).
			WithDefaultGetter(func(ledger *config.Ledger) any {
				return ledger.GetString(account) != ""
			}).
			BuildBoolOption(),
		optionJsonPrinter: cli.Options.Printer.JsonPrinter().
			WithKeys(&schema.Genaiz.DataLink.List.Printer).
			BuildBoolOption(),
	}
}
