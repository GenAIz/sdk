package broker

import (
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"genaiz.com/genaiz/task"
	ver "genaiz.com/genaiz/version"
)

func NewDataStoreListTask() *task.Task[DataInstanceListParams] {
	return &task.Task[DataInstanceListParams]{
		Name:       "data-store-list",
		OnPrepare:  handleDataStoreListContext,
		OnComplete: handleDataStoreListComplete,
		OnPretend:  handleDataStoreListPretend,
	}
}

func handleDataStoreListComplete(params *DataInstanceListParams, state *task.State) error {
	var err error
	var brokerClient Client

	if brokerClient, err = params.GetClient(); err == nil {
		var stores []DataLinkInstance

		handleDataStoreFilterDebugging(params, state.Logger)

		if stores, err = brokerClient.ListDataStores(); err == nil {
			if params.hasDataLinkFilter() {
				state.Internal = params.filter(stores)
			} else {
				state.Internal = stores
			}

			return nil
		}
	}

	return err
}

func handleDataStoreListContext(params *DataInstanceListParams, state *task.State) error {
	if state.Output == "" {
		if err := params.validateDataLinkFilter(); err != nil {
			return err
		}

		if id := params.getLinkId(); id != nil {
			state.Logger.Debugf("Listing data stores for data link [%d]", *id)
			state.Logger.Warnf("Filtering by id is not supported in version [%s]", ver.GetVersion())
			state.Output = cast.ToString(*id)
		}
	}

	return nil
}

func handleDataStoreListPretend(params *DataInstanceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		state.Logger.Debugf("Pretending to list data store for account [%s]", brokerClient.GetHostAddr())
		handleDataStoreFilterDebugging(params, state.Logger)
		fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())

		if id := params.getLinkId(); id != nil {
			fmt.Printf("  -G -d id=%d\\\n", *id)
		}

		fmt.Printf("%s\n", brokerClient.ListDataStoresUrl())
	}

	return err
}

func handleDataStoreFilterDebugging(params *DataInstanceListParams, logger *logrus.Logger) {
	if params.hasDataLinkFilter() {
		logger.Debugf("Filtering data stores for oem [%s]", params.Oem)

		if params.Handle != "" {
			logger.Debugf("With handle [%s]", params.Handle)
		}

		if params.Version != "" {
			logger.Debugf("And version [%s]", params.DataLink.GetVersion())
		}
	}
}
