package auto

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task/broker"
)

type DataSourceAutoBridge struct {
	ledger *config.Ledger

	dataSourceFacadeProvider func() mgmt.UserDataSourceFacade
}

func (sab DataSourceAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var params = &broker.DataInstanceListParams{
		Broker: broker.Broker{
			AuthFile: sab.ledger.AuthFile,
		},
	}
	var facade = sab.dataSourceFacadeProvider().
		WithParams(params).
		WithLogger(sab.ledger.Logger).
		Filtering(toComplete)

	if dataSources, err := facade.Get(); err == nil {
		if len(dataSources) > 0 {
			var results []cobra.Completion

			for _, ds := range dataSources {
				results = append(results, ds.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func NewDataSourceAutoBridge(ledger *config.Ledger) *DataSourceAutoBridge {
	return &DataSourceAutoBridge{
		ledger: ledger,

		dataSourceFacadeProvider: mgmt.NewUserDataSourceFacade,
	}
}
