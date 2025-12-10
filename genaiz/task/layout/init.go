package layout

import (
	"errors"
	"fmt"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
	"genaiz.com/genaiz/task/shared"
)

var (
	errorNeedsConfigFile = errors.New("file should exists")
	errorNoConfigFile    = errors.New("no config file found")
)

type ConfigWriter interface {
	Write(string) error

	BuildArches() (string, []string)

	BuildHandle() (string, string)

	BuildInput() (string, string)

	BuildInputPorts() (string, []broker.DataPort)

	BuildInputPortRemoved() (string, *broker.DataPort)

	BuildName() (string, string)

	BuildOem() (string, string)

	BuildOutput() map[string]string

	BuildOutputPorts() (string, []broker.DataPort)

	BuildOutputPortRemoved() (string, *broker.DataPort)

	BuildPropSpecs() (string, []broker.PropSpec)

	BuildPropSpecRemoved() (string, *broker.PropSpec)

	BuildSources() (string, []string)

	BuildType() (string, string)

	BuildVersion() (string, string)

	WithConfigFile(string) ConfigWriter

	WithArches([]string) ConfigWriter

	WithHandle(string) ConfigWriter

	WithInput(string) ConfigWriter

	WithInputPorts([]broker.DataPort) ConfigWriter

	WithInputPortRemoved(*broker.DataPort) ConfigWriter

	WithLog(string) ConfigWriter

	WithName(string) ConfigWriter

	WithOem(string) ConfigWriter

	WithOutput(string) ConfigWriter

	WithOutputPorts([]broker.DataPort) ConfigWriter

	WithOutputPortRemoved(*broker.DataPort) ConfigWriter

	WithPropSpecs([]broker.PropSpec) ConfigWriter

	WithPropSpecRemoved(*broker.PropSpec) ConfigWriter

	WithSources([]string) ConfigWriter

	WithType(string) ConfigWriter

	WithVar(string) ConfigWriter

	WithVersion(string) ConfigWriter
}

type InitParams struct {
	CreateParams
	Arches      []ArchType
	DataSources []string
	Handle      string
	InputPorts  []broker.DataPort
	Name        string
	Type        FunctionType
	MountInput  string
	MountOutput string
	OEM         string
	OutputPorts []broker.DataPort
	PropSpecs   []broker.PropSpec
	Version     string
}

func NewInitTask(writer ConfigWriter) *task.Task[InitParams] {
	return &task.Task[InitParams]{
		Name:         "layout-init",
		OnPrepare:    handleLayoutInitContext,
		OnIncomplete: lang.Assists(writer, handleLayoutInitUpdate),
		OnComplete:   lang.Assists(writer, handleLayoutInitCreate),
		OnPretend:    lang.Assists(writer, handleLayoutInitPretend),
	}
}

func handleLayoutInitContext(params *InitParams, state *task.State) error {
	var initState = NewInitState(state)

	state.Output = stringz.FirstNonEmpty(initState.GetConfigFile(), state.Output)

	if state.Output == "" {
		state.Logger.Debugf("Init finding a configuration file for writing")

		if !params.IsConfigTypeNone() {
			state.Output = params.GetConfigFile()
		} else {
			var file, _ = filez.FirstNamedFile(params.ConfigName)

			if file == "" {
				state.Logger.Errorf("could not find a configuration file for [%s]", params.ConfigName)
			}

			state.Output = file
			return errorNeedsConfigFile
		}
	}

	return nil
}

func handleLayoutInitCreate(writer ConfigWriter, params *InitParams, state *task.State) error {
	if state.Output != "" {
		var initState = NewInitState(state)
		var err error

		state.Logger.Debugf("Init writing to [%s]", state.Output)

		if err = writer.WithArches(params.Arches).
			WithHandle(params.Handle).
			WithName(params.Name).
			WithType(params.Type).
			WithInput(initState.DefaultInput(params.MountInput)).
			WithInputPorts(params.InputPorts).
			WithLog(initState.DefaultLog(params.MountOutput)).
			WithOem(params.OEM).
			WithOutput(initState.DefaultOutput(params.MountOutput)).
			WithOutputPorts(params.OutputPorts).
			WithPropSpecs(params.PropSpecs).
			WithSources(params.DataSources).
			WithVar(initState.DefaultVar(params.MountOutput)).
			WithVersion(params.Version).
			Write(state.Output); err == nil {
			var _, oem = writer.BuildOem()
			var _, handle = writer.BuildHandle()
			var _, version = writer.BuildVersion()

			state.Report(fmt.Sprintf("Initialized function %s/%s, version %s", oem, handle, version))
			return nil
		}

		return err
	}

	return errorNoConfigFile
}

func handleLayoutInitPretend(writer ConfigWriter, params *InitParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)
		var rmInputPortKey, rmInputPort = writer.BuildInputPortRemoved()
		var inputPortKey, inputPorts = writer.WithInputPorts(params.InputPorts).BuildInputPorts()
		var rmOutputPortKey, rmOutputPort = writer.BuildOutputPortRemoved()
		var outputPortKey, outputPorts = writer.WithOutputPorts(params.OutputPorts).BuildOutputPorts()
		var rmPropSpecKey, rmPropSpec = writer.BuildPropSpecRemoved()
		var propSpecKey, propSpecs = writer.WithPropSpecs(params.PropSpecs).BuildPropSpecs()
		var dataSourcesKey, dataSources = writer.BuildSources()

		state.Logger.Debugf("Pretending to initialize [%s]", state.Output)
		writer.WithHandle(params.Handle)
		shared.PretendSlice(pretender, writer.WithArches(params.Arches).BuildArches)
		shared.PretendValue(pretender, writer.WithName(params.Name).BuildName)
		shared.PretendValue(pretender, writer.BuildHandle)
		shared.PretendValue(pretender, writer.WithType(params.Type).BuildType)
		shared.PretendValue(pretender, writer.WithInput(params.MountInput).BuildInput)
		shared.PretendMap(pretender, writer.WithOutput(params.MountOutput).BuildOutput)
		shared.PretendValue(pretender, writer.WithOem(params.OEM).BuildOem)
		shared.PretendValue(pretender, writer.WithVersion(params.Version).BuildVersion)

		if rmInputPort != nil {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return fmt.Sprintf("%s[]", rmInputPortKey), "handle", rmInputPort.Handle
			})
		} else if len(inputPorts) > 0 {
			pretender.PretendDelete(inputPortKey)

			for i, port := range inputPorts {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].handle", inputPortKey, i), port.Handle
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", inputPortKey, i), port.Name
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].description", inputPortKey, i), port.Description
				})
			}
		}

		if rmOutputPort != nil {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return fmt.Sprintf("%s[]", rmOutputPortKey), "handle", rmOutputPort.Handle
			})
		} else if len(outputPorts) > 0 {
			pretender.PretendDelete(outputPortKey)

			for i, port := range outputPorts {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].handle", outputPortKey, i), port.Handle
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", outputPortKey, i), port.Name
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].description", outputPortKey, i), port.Description
				})
			}
		}

		if rmPropSpec != nil {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return fmt.Sprintf("%s[]", rmPropSpecKey), "key", rmPropSpec.Key
			})
		} else if len(propSpecs) > 0 {
			pretender.PretendDelete(propSpecKey)

			for i, spec := range propSpecs {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].key", propSpecKey, i), spec.Key
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].type", propSpecKey, i), spec.Type
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", propSpecKey, i), spec.Name
				})

				if spec.Description != "" {
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].description", propSpecKey, i), spec.Description
					})
				}

				if spec.Value != "" {
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].value", propSpecKey, i), spec.Value
					})
				} else if len(spec.Values) > 0 {
					shared.PretendSlice(pretender, func() (string, []string) {
						return fmt.Sprintf("%s[%d].values", propSpecKey, i), spec.Values
					})
				}
			}
		}

		if len(dataSources) > 0 {
			shared.PretendSlice(pretender, func() (string, []string) {
				return dataSourcesKey, dataSources
			})
		}

		state.Output = ""
		return nil
	}

	return errorNoConfigFile
}

func handleLayoutInitUpdate(writer ConfigWriter, params *InitParams, state *task.State) error {
	if state.Output != "" {
		var err error

		state.Logger.Debugf("Init updating existing [%s]", state.Output)

		if err = handleLayoutInitCreate(writer.WithConfigFile(state.Output), params, state); err == nil {
			state.Report(fmt.Sprintf("Updated %s successfully", state.Output))
			state.Completed = true
			state.Output = ""
			return nil
		}

		return err
	}

	return nil
}
