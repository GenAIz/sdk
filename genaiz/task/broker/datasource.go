package broker

import (
	"errors"
	"fmt"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"

	"genaiz.com/genaiz/task"
	ver "genaiz.com/genaiz/version"
)

var (
	errorDataSourceKnown    = task.NewError("data source has been resolved")
	errorDataSourceRequired = task.NewError("data source id or name is required")
)

type DataSourceResolveParams struct {
	Broker
	DataSourceId   *int64
	DataSourceName string
}

func NewDataSourceListTask() *task.Task[DataInstanceListParams] {
	return &task.Task[DataInstanceListParams]{
		Name:       "data-source-list",
		OnPrepare:  handleDataSourceListContext,
		OnComplete: handleDataSourceListComplete,
		OnPretend:  handleDataSourceListPretend,
	}
}

func NewDataSourceResolveTask() *task.Task[DataSourceResolveParams] {
	return &task.Task[DataSourceResolveParams]{
		Name:         "resolve-data-source",
		OnPrepare:    handleDataSourceResolveContext,
		OnComplete:   handleDataSourceResolveComplete,
		OnIncomplete: handleDataSourceResolveIncomplete,
		OnPretend:    handleDataSourceResolvePretend,
	}
}

func handleDataSourceListComplete(params *DataInstanceListParams, state *task.State) error {
	var err error
	var brokerClient Client

	if brokerClient, err = params.GetClient(); err == nil {
		var sources []DataLinkInstance

		handleDataSourceFilterDebugging(params, state.Logger)

		if sources, err = brokerClient.ListDataSources(); err == nil {
			if params.hasDataLinkFilter() {
				state.Internal = params.filter(sources)
			} else {
				state.Internal = sources
			}

			return nil
		}
	}

	return err
}

func handleDataSourceListContext(params *DataInstanceListParams, state *task.State) error {
	if state.Output == "" {
		if err := params.validateDataLinkFilter(); err != nil {
			return err
		}

		if id := params.getLinkId(); id != nil {
			state.Logger.Debugf("Listing data sources for data link [%d]", *id)
			state.Logger.Warnf("Filtering by id is not supported in version [%s]", ver.GetVersion())
			state.Output = cast.ToString(*id)
		}
	}

	return nil
}

func handleDataSourceListPretend(params *DataInstanceListParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		state.Logger.Debugf("Pretending to list data source for account [%s]", brokerClient.GetHostAddr())
		handleDataSourceFilterDebugging(params, state.Logger)
		fmt.Printf("curl -X GET -H \"Content-Type: application/x-www-form-urlencoded\" \\\n")
		fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())

		if id := params.getLinkId(); id != nil {
			fmt.Printf("  -G -d id=%d\\\n", *id)
		}

		fmt.Printf("%s\n", brokerClient.ListDataSourcesUrl())
	}

	return err
}

func handleDataSourceFilterDebugging(params *DataInstanceListParams, logger *logrus.Logger) {
	if params.hasDataLinkFilter() {
		logger.Debugf("Filtering data sources for oem [%s]", params.Oem)

		if params.Handle != "" {
			logger.Debugf("With handle [%s]", params.Handle)
		}

		if params.Version != "" {
			logger.Debugf("And version [%s]", params.DataLink.GetVersion())
		}
	}
}

func handleDataSourceResolveComplete(params *DataSourceResolveParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var sources []DataLinkInstance

			state.Logger.Debugf("Listing data sources for account [%s]", brokerClient.GetHostAddr())

			if sources, err = brokerClient.ListDataSources(); err == nil {
				for _, ds := range sources {
					if ds.Name == params.DataSourceName {
						params.DataSourceId = ds.Id
						break
					}
				}

				state.Output = ""
				return nil
			}
		}

		return err
	}

	return errorDataSourceRequired
}

func handleDataSourceResolveContext(params *DataSourceResolveParams, state *task.State) error {
	if state.Output == "" {
		if params.DataSourceId != nil {
			return errorDataSourceKnown
		}

		if params.DataSourceName == "" {
			return errorDataSourceRequired
		}

		state.Output = params.DataSourceName
	}

	return nil
}

func handleDataSourceResolveIncomplete(params *DataSourceResolveParams, state *task.State) error {
	if errors.Is(state.Error, errorDataSourceKnown) {
		state.Logger.Debugf("Data source id provided [%d], skipping resolution", *params.DataSourceId)
		state.Completed = true
		return nil
	}

	return state.Error
}

func handleDataSourceResolvePretend(params *DataSourceResolveParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to find a data source named: [%s]", params.DataSourceName)
			fmt.Printf("curl -X GET \\\n")
			fmt.Printf("  --cookie=\"s=%s\"\\\n", brokerClient.GetAuthToken())
			fmt.Printf("%s\n", brokerClient.ListDataSourcesUrl())
			return nil
		}

		return err
	}

	if errors.Is(state.Error, errorDataSourceKnown) {
		state.Logger.Debugf("Data source id provided [%d], skipping resolution", *params.DataSourceId)
		return nil
	}

	return state.Error
}
