package auto

import (
	"github.com/spf13/cobra"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task/broker"
)

type DataStoreAutoBridge struct {
	ledger *config.Ledger

	dataStoreFacadeProvider func() mgmt.UserDataStoreFacade
}

func (sab DataStoreAutoBridge) Bridge(toComplete string) ([]cobra.Completion, cobra.ShellCompDirective) {
	var params = &broker.DataInstanceListParams{
		Broker: broker.Broker{
			AuthFile: sab.ledger.AuthFile,
		},
	}
	var facade = sab.dataStoreFacadeProvider().
		WithParams(params).
		WithLogger(sab.ledger.Logger).
		Filtering(toComplete)

	if dataStores, err := facade.Get(); err == nil {
		if len(dataStores) > 0 {
			var results []cobra.Completion

			for _, ds := range dataStores {
				results = append(results, ds.Matched())
			}

			return results, cobra.ShellCompDirectiveKeepOrder
		}

		return nil, cobra.ShellCompDirectiveNoFileComp
	}

	return nil, cobra.ShellCompDirectiveError
}

func NewDataStoreAutoBridge(ledger *config.Ledger) *DataStoreAutoBridge {
	return &DataStoreAutoBridge{
		ledger: ledger,

		dataStoreFacadeProvider: mgmt.NewUserDataStoreFacade,
	}
}
