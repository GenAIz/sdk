package dk

import (
	"errors"
	"fmt"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorDataLinkNotSynchronized = errors.New("datalinks specified were not synchronized and not found locally")
)

type CollectLinkTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type SyncBridge struct {
	collectTaskFactory     CollectLinkTaskFactory
	exportTaskFactory      ExportLinkTaskFactory
	dataLinksWriterFactory DataLinksWriterFactory
}

func (sb SyncBridge) MakeSyncPretenders(datalinks []string, ledger *config.Ledger, noSyncOption *config.BoolOption) ([]task.Worker, error) {
	if len(datalinks) > 0 {
		var workerFactory func(*broker.DataLinkParams) task.Worker
		var brokerParams = sb.newBrokerParams(ledger)
		var configParams = sb.newUserParams(ledger)

		if ledger.GetBool(noSyncOption) {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = sb.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewPretender(params, sb.collectTaskFactory(dataLinkWriter))
			}
		} else {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = sb.dataLinksWriterFactory(ledger, params.GetConfigFile())

				return task.NewPretender(params, sb.exportTaskFactory(dataLinkWriter))
			}
		}

		return sb.makeDataLinkWorkers(brokerParams, configParams, datalinks, workerFactory)
	}

	return nil, nil
}

func (sb SyncBridge) MakeSyncWorkers(datalinks []string, ledger *config.Ledger, noSyncOption *config.BoolOption) ([]task.Worker, error) {
	if len(datalinks) > 0 {
		var workerFactory func(*broker.DataLinkParams) task.Worker
		var brokerParams = sb.newBrokerParams(ledger)
		var configParams = sb.newUserParams(ledger)
		var userConfig, err = configParams.EnsureConfigPath()

		if ledger.GetBool(noSyncOption) {
			if err != nil {
				// Can not use local repo if it doesn't exist
				return nil, errorDataLinkNotSynchronized
			}

			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = sb.dataLinksWriterFactory(ledger, userConfig)

				return task.NewWorker(params, sb.collectTaskFactory(dataLinkWriter))
			}
		} else {
			workerFactory = func(params *broker.DataLinkParams) task.Worker {
				var dataLinkWriter = sb.dataLinksWriterFactory(ledger, userConfig)

				return task.NewWorker(params, sb.exportTaskFactory(dataLinkWriter))
			}
		}

		return sb.makeDataLinkWorkers(brokerParams, configParams, datalinks, workerFactory)

	}

	return nil, nil
}

func (sb SyncBridge) makeDataLinkWorkers(brokerParams *broker.Broker, configParams *shared.ConfigParams, dataLinks []string, workFactory func(*broker.DataLinkParams) task.Worker) ([]task.Worker, error) {
	var result []task.Worker

	for _, link := range dataLinks {
		if o, h, v := ParseDataLinkArgument(link); o != "" && h != "" && v != "" {
			var params = sb.newDataLinkParams(o, h, v, brokerParams, configParams)

			result = append(result, workFactory(params))
		} else {
			return nil, fmt.Errorf("invalid datalink found [%s]", link)
		}
	}

	return result, nil
}

func (sb SyncBridge) newBrokerParams(ledger *config.Ledger) *broker.Broker {
	return &broker.Broker{
		AuthFile: ledger.AuthFile,
	}
}

func (sb SyncBridge) newUserParams(ledger *config.Ledger) *shared.ConfigParams {
	return &shared.ConfigParams{
		ConfigName:   ledger.ConfigName,
		ConfigFolder: ledger.UserPath,
	}
}

func (sb SyncBridge) newDataLinkParams(oem, handle, version string, brokerParams *broker.Broker, configParams *shared.ConfigParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker:       *brokerParams,
		ConfigParams: *configParams,
		DataLink: &broker.DataLink{
			Oem:     oem,
			Handle:  handle,
			Version: version,
		},
	}
}

type SyncBridgeBuilder struct {
	collectLinkTaskFactory CollectLinkTaskFactory
	exportLinkTaskFactory  ExportLinkTaskFactory
	dataLinksWriterFactory DataLinksWriterFactory
}

func (sb *SyncBridgeBuilder) Build() SyncBridge {
	return SyncBridge{
		collectTaskFactory:     sb.collectLinkTaskFactory,
		exportTaskFactory:      sb.exportLinkTaskFactory,
		dataLinksWriterFactory: sb.dataLinksWriterFactory,
	}
}

func (sb *SyncBridgeBuilder) WithCollectLinkTaskFactory(factory CollectLinkTaskFactory) *SyncBridgeBuilder {
	sb.collectLinkTaskFactory = factory
	return sb
}

func (sb *SyncBridgeBuilder) WithDataLinksWriterFactory(factory DataLinksWriterFactory) *SyncBridgeBuilder {
	sb.dataLinksWriterFactory = factory
	return sb
}

func (sb *SyncBridgeBuilder) WithExportLinkTaskFactory(factory ExportLinkTaskFactory) *SyncBridgeBuilder {
	sb.exportLinkTaskFactory = factory
	return sb
}

func NewSyncBridgeBuilder() *SyncBridgeBuilder {
	return &SyncBridgeBuilder{
		collectLinkTaskFactory: broker.NewDataLinkCollectTask,
		exportLinkTaskFactory:  broker.NewDataLinkExportTask,
		dataLinksWriterFactory: NewDataLinksWriter,
	}
}
