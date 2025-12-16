package sf

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cobra"

	"genaiz.com/genaiz-lib/lang/dirz"
	"genaiz.com/genaiz-lib/lang/mapz"
	"genaiz.com/genaiz/cli"
	"genaiz.com/genaiz/cmd/sf/data"
	"genaiz.com/genaiz/config"
	"genaiz.com/genaiz/schema"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/layout"
	"genaiz.com/genaiz/task/shared"
)

type dataPortType = string

const (
	inputPortType  dataPortType = "input"
	outputPortType dataPortType = "output"
)

var (
	inputPortsWriter  = makeInputPortsWriter()
	outputPortsWriter = makeOutputPortsWriter()
	dataPortWriters   = mapz.Mapped([]dataPortWriter{inputPortsWriter, outputPortsWriter},
		func(writer dataPortWriter) string {
			return writer.portType
		})
)

type dataPortWriter struct {
	runOption  *config.StringOption
	portOption *config.Option
	portType   dataPortType
}

func (dpw dataPortWriter) GetRunFolder(ledger *config.Ledger) string {
	return ledger.GetString(dpw.runOption)
}

func (dpw dataPortWriter) HasPort(ports []broker.DataPort, handle string) bool {
	return slices.ContainsFunc(ports, func(port broker.DataPort) bool {
		return strings.EqualFold(port.Handle, handle)
	})
}

func (dpw dataPortWriter) ListPorts(ledger *config.Ledger) []broker.DataPort {
	return broker.ListDataPorts(ledger.Get(dpw.portOption))
}

func makeInputPortsWriter() dataPortWriter {
	return dataPortWriter{
		portOption: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.InputPorts).
			BuildOption(),
		portType: inputPortType,
		runOption: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Run.MountInput).
			BuildStringOption(),
	}
}

func makeOutputPortsWriter() dataPortWriter {
	return dataPortWriter{
		portOption: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Publish.OutputPorts).
			BuildOption(),
		portType: outputPortType,
		runOption: cli.NewOptionBuilder().
			WithKeys(&schema.Genaiz.Function.Run.MountOutput).
			BuildStringOption(),
	}
}

type DataExecutor struct {
	BaseExecutor

	inputOptions  *DataOptions
	outputOptions *DataOptions

	addedPort    *broker.DataPort
	removedPort  *broker.DataPort
	updatedPorts map[string][]broker.DataPort

	initTaskFactory InitTaskFactory
}

func (de *DataExecutor) Add(dataType, handle string) error {
	if writer, ok := dataPortWriters[dataType]; ok {
		var current = writer.ListPorts(de.Ledger)
		var options = de.getDataOptions(dataType)

		if !writer.HasPort(current, handle) {
			var addedPort *broker.DataPort
			var err error

			if addedPort, err = de.makeDataPort(handle, options); err == nil {
				current = append(current, *addedPort)
				de.updatedPorts[writer.portType] = current
				de.addedPort = addedPort
				de.Cli.Exec(de.Ledger, de)
				return nil
			} else {
				return err
			}
		}

		return fmt.Errorf("[%s] is already configured", handle)
	}

	panic("unknown data type " + dataType)
}

func (de *DataExecutor) Display() {
	var detailsMap = map[string]string{}

	if de.addedPort != nil {
		_, _ = fmt.Printf("Adding the following data port configuration:\n")
		detailsMap["handle"] = de.addedPort.Handle
		detailsMap["name"] = de.addedPort.Name
		detailsMap["description"] = de.addedPort.Description
	} else if de.removedPort != nil {
		_, _ = fmt.Printf("Removing the following data port configuration:\n")
		detailsMap["handle"] = de.removedPort.Handle

	}

	if len(detailsMap) != 0 {
		de.Ledger.DisplayOptionsWithMap(&detailsMap)
	}
}

func (de *DataExecutor) Init(dataType, pathOrHandle string) (string, error) {
	if writer, ok := dataPortWriters[dataType]; ok {
		var runFolder = writer.GetRunFolder(de.Ledger)
		var runRoot = dirz.FirstParentPath(runFolder)
		var runDir = filepath.Base(runFolder)
		var pathDir = filepath.Base(filepath.Dir(pathOrHandle))
		var dir os.FileInfo
		var err error

		if runFolder != "" && strings.HasPrefix(pathOrHandle, runRoot) && runDir == pathDir {
			if dir, err = os.Stat(pathOrHandle); err == nil {
				if dir.IsDir() {
					return filepath.Base(pathOrHandle), nil
				} else {
					return "", fmt.Errorf("port handle [%s] maps to a file, should be a directory", pathOrHandle)
				}
			} else if err = os.MkdirAll(pathOrHandle, 0750); err == nil {
				return filepath.Base(pathOrHandle), nil
			}

			return "", err
		}

		return pathOrHandle, nil
	}

	panic("unknown data type " + dataType)
}

func (de *DataExecutor) Pretend() {
	var params = de.makeInitParams()
	var builder = makeInitBuilder(de.Ledger, de.Cli)

	for k := range de.updatedPorts {
		if k == inputPortType {
			builder.WithInputPortRemoved(de.removedPort)
		} else if k == outputPortType {
			builder.WithOutputPortRemoved(de.removedPort)
		}
	}

	de.Ledger.DisplayChangeDir()
	de.initTaskFactory(builder).Pretend(params, de.Ledger.Logger)
}

func (de *DataExecutor) Proceed() {
	var builder = makeInitBuilder(de.Ledger, de.Cli)
	var params = de.makeInitParams()
	var plan = task.NewPlan("Init", de.Ledger.Logger)

	for k := range de.updatedPorts {
		if k == inputPortType {
			builder.WithInputPortRemoved(de.removedPort)
		} else if k == outputPortType {
			builder.WithOutputPortRemoved(de.removedPort)
		}
	}

	plan.PrintReportsOnly = true
	task.Single(plan, params, de.initTaskFactory(builder))
}

func (de *DataExecutor) Remove(dataType, handle string) error {
	if writer, ok := dataPortWriters[dataType]; ok {
		var current = writer.ListPorts(de.Ledger)
		var ports []broker.DataPort

		for _, port := range current {
			if !strings.EqualFold(port.Handle, handle) {
				ports = append(ports, port)
			} else {
				de.removedPort = &port
			}
		}

		de.updatedPorts[writer.portType] = ports
		de.Cli.Exec(de.Ledger, de)
		return nil
	}

	panic("unknown data type " + dataType)
}

func (de *DataExecutor) getDataOptions(portType dataPortType) *DataOptions {
	if portType == inputPortType {
		return de.inputOptions
	}

	return de.outputOptions
}

func (de *DataExecutor) makeDataPort(handle string, options *DataOptions) (*broker.DataPort, error) {
	if config.Validation.Handle(handle) {
		var name = de.Ledger.GetString(options.optionName)

		if name == "" {
			name = handle
		}

		return &broker.DataPort{
			Description: de.Ledger.GetString(options.optionDesc),
			Handle:      handle,
			Name:        name,
		}, nil
	}

	return nil, fmt.Errorf("[%s] is not a valid handle value", handle)
}

func (de *DataExecutor) makeInitParams() *layout.InitParams {
	var inputPorts, _ = de.updatedPorts[inputPortType]
	var outputPorts, _ = de.updatedPorts[outputPortType]

	return &layout.InitParams{
		CreateParams: layout.CreateParams{
			ConfigParams: shared.ConfigParams{
				ConfigName: de.Ledger.ConfigName,
			},
		},
		InputPorts:  inputPorts,
		OutputPorts: outputPorts,
	}
}

type DataOptions struct {
	optionDesc *config.StringOption
	optionName *config.StringOption
}

func (do DataOptions) allDefiners() []config.Definer {
	return []config.Definer{
		do.optionDesc,
		do.optionName,
	}
}

func NewData(ledger *config.Ledger, sfCli *Cli) *cobra.Command {
	var dataInputOptions = NewDataOptionsInput()
	var dataOutputOptions = NewDataOptionsOutput()
	var dataExecFactory = newDataExecutorFactory(ledger, sfCli, dataInputOptions, dataOutputOptions)
	var dataCmd = &cobra.Command{
		Use:   "data",
		Short: "Manages data port and source configurations for Smart Functions",
	}

	dataCmd.AddCommand(data.NewDataInput(ledger, dataInputOptions.allDefiners(), dataExecFactory))
	dataCmd.AddCommand(data.NewDataOutput(ledger, dataOutputOptions.allDefiners(), dataExecFactory))
	dataCmd.AddCommand(NewSource(ledger, sfCli))
	dataCmd.AddCommand(NewStore(ledger, sfCli))
	dataCmd.AddCommand(NewProxy(ledger, sfCli))
	return dataCmd
}

func NewDataExecutor(ctx context.Context, ledger *config.Ledger, sfCli *Cli, inputOptions *DataOptions, outputOptions *DataOptions) *DataExecutor {
	return &DataExecutor{
		BaseExecutor: BaseExecutor{
			Context: ctx,
			Ledger:  ledger,
			Cli:     sfCli,
		},

		inputOptions:    inputOptions,
		outputOptions:   outputOptions,
		updatedPorts:    make(map[string][]broker.DataPort),
		initTaskFactory: layout.NewInitTask,
	}
}

func NewDataOptionsInput() *DataOptions {
	return &DataOptions{
		optionDesc: cli.Options.DataPorts.Desc().
			WithKeys(&schema.Genaiz.Function.Publish.DataPortAdd.Input.Desc).
			BuildStringOption(),
		optionName: cli.Options.DataPorts.Name().
			WithKeys(&schema.Genaiz.Function.Publish.DataPortAdd.Input.Name).
			BuildStringOption(),
	}
}

func NewDataOptionsOutput() *DataOptions {
	return &DataOptions{
		optionDesc: cli.Options.DataPorts.Desc().
			WithKeys(&schema.Genaiz.Function.Publish.DataPortAdd.Output.Desc).
			BuildStringOption(),
		optionName: cli.Options.DataPorts.Name().
			WithKeys(&schema.Genaiz.Function.Publish.DataPortAdd.Output.Name).
			BuildStringOption(),
	}
}

func newDataExecutorFactory(ledger *config.Ledger, sfCli *Cli, inputOptions *DataOptions, outputOptions *DataOptions) data.ExecutorFactory {
	return func(cmd *cobra.Command) data.Executor {
		return NewDataExecutor(cmd.Context(), ledger, sfCli, inputOptions, outputOptions)
	}
}
