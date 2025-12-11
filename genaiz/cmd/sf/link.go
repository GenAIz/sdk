package sf

import (
	"strings"

	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

type ListLinksTaskFactory func() *task.Task[broker.DataLinkParams]

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
		oem, handle, version = parseFqdnv(fqdnv)
		dle.Ledger.InitValue(dle.optionHandle, handle)
		dle.Ledger.InitValue(dle.optionOem, oem)
		dle.Ledger.InitValue(dle.optionVersion, version)
	}

	return &broker.DataLinkParams{
		Broker: broker.Broker{
			AuthFile: dle.Ledger.AuthFile,
		},
		Handle:       dle.Ledger.GetString(dle.optionHandle),
		Oem:          dle.Ledger.GetString(dle.optionOem),
		Version:      dle.Ledger.GetString(dle.optionVersion),
		NoValidation: validate,
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

func parseFqdnv(fqdnv string) (string, string, string) {
	var parts = strings.SplitN(fqdnv, "/", 2)
	var oem, handle, version string

	if len(parts) == 2 {
		oem, parts = parts[0], parts[1:]
	}

	if len(parts) == 1 {
		parts = strings.SplitN(parts[0], ":", 2)
		handle = parts[0]

		if len(parts) == 2 {
			version = parts[1]
		}
	}

	return oem, handle, version
}
