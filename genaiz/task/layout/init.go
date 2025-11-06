package layout

import (
	"errors"
	"fmt"

	"genaiz.com/genaiz-lib/lang/filez"
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

	BuildName() (string, string)

	BuildOem() (string, string)

	BuildOutput() map[string]string

	BuildPropSpecs() (string, []broker.PropSpec)

	BuildRemovedPropSpec() (string, *broker.PropSpec)

	BuildType() (string, string)

	BuildVersion() (string, string)

	WithConfigFile(string) ConfigWriter

	WithArches([]string) ConfigWriter

	WithHandle(string) ConfigWriter

	WithInput(string) ConfigWriter

	WithName(string) ConfigWriter

	WithOem(string) ConfigWriter

	WithOutput(string) ConfigWriter

	WithPropSpecs([]broker.PropSpec) ConfigWriter

	WithPropSpecRemoved(spec *broker.PropSpec) ConfigWriter

	WithType(string) ConfigWriter

	WithVersion(string) ConfigWriter
}

type InitParams struct {
	CreateParams
	Arches      []ArchType
	Handle      string
	Name        string
	Type        FunctionType
	MountInput  string
	MountOutput string
	OEM         string
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
		var err error

		state.Logger.Debugf("Init writing to [%s]", state.Output)

		if err = writer.WithArches(params.Arches).
			WithHandle(params.Handle).
			WithName(params.Name).
			WithType(params.Type).
			WithInput(params.MountInput).
			WithOutput(params.MountOutput).
			WithOem(params.OEM).
			WithPropSpecs(params.PropSpecs).
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
		var removedKey, rmPropSpec = writer.BuildRemovedPropSpec()
		var rootKey, propSpecs = writer.WithPropSpecs(params.PropSpecs).BuildPropSpecs()

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

		if rmPropSpec != nil {
			shared.PretendDeleteByField(pretender, func() (string, string, string) {
				return fmt.Sprintf("%s[]", removedKey), "key", rmPropSpec.Key
			})
		} else if len(propSpecs) > 0 {
			pretender.PretendDelete(rootKey)

			for i, spec := range params.PropSpecs {
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].key", rootKey, i), spec.Key
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].type", rootKey, i), spec.Type
				})
				shared.PretendValue(pretender, func() (string, string) {
					return fmt.Sprintf("%s[%d].name", rootKey, i), spec.Name
				})

				if spec.Description != "" {
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].description", rootKey, i), spec.Description
					})
				}

				if spec.Value != "" {
					shared.PretendValue(pretender, func() (string, string) {
						return fmt.Sprintf("%s[%d].value", rootKey, i), spec.Value
					})
				} else if len(spec.Values) > 0 {
					shared.PretendSlice(pretender, func() (string, []string) {
						return fmt.Sprintf("%s[%d].values", rootKey, i), spec.Values
					})
				}
			}
		}

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
