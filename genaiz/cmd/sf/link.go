package sf

import (
	"genaiz.com/genaiz/cmd/dk"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

type ListLinksTaskFactory func() *task.Task[broker.DataLinkParams]
type SyncLinksTaskFactory func(broker.DataLinkWriter) *task.Task[broker.DataLinkParams]

type DataLinkExecutor struct {
	BaseExecutor
	*DataLinkOptions
}

func (dle *DataLinkExecutor) makeDataLinkParams(fqdnv string) *broker.DataLinkParams {
	var handle, oem, version string
	var validate = false

	if dle.optionNoValidation != nil {
		validate = dle.Ledger.GetBool(dle.optionNoValidation)
	}

	if fqdnv != "" {
		oem, handle, version = dk.ParseDataLinkArgument(fqdnv)
		dle.Ledger.OverrideString(dle.optionHandle, handle)
		dle.Ledger.OverrideString(dle.optionOem, oem)
		dle.Ledger.OverrideString(dle.optionVersion, version)
	}

	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: dle.Ledger.AuthFile,
		},
		DataLink: &broker.DataLink{
			Handle:  dle.Ledger.GetString(dle.optionHandle),
			Oem:     dle.Ledger.GetString(dle.optionOem),
			Version: dle.Ledger.GetString(dle.optionVersion),
		},
		NoValidation: validate,
	}
}

func (dle *DataLinkExecutor) makeSyncParams(dataLinkParams *broker.DataLinkParams) *broker.DataLinkParams {
	return &broker.DataLinkParams{
		Broker: dataLinkParams.Broker,
		ConfigParams: shared.ConfigParams{
			ConfigName:   dle.Ledger.ConfigName,
			ConfigFolder: dle.Ledger.UserPath,
		},
		DataLink: dataLinkParams.DataLink,
	}
}

type DataLinkOptions struct {
	optionHandle       *config.StringOption
	optionNoValidation *config.BoolOption
	optionOem          *config.StringOption
	optionVersion      *config.StringOption
}

func (dlo DataLinkOptions) addDefiners() []config.Definer {
	return []config.Definer{
		dlo.optionHandle,
		dlo.optionNoValidation,
		dlo.optionOem,
		dlo.optionVersion,
	}
}

func (dlo DataLinkOptions) removeDefiners() []config.Definer {
	return []config.Definer{
		dlo.optionHandle,
		dlo.optionOem,
		dlo.optionVersion,
	}
}
