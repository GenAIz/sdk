package layout

import (
	"errors"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
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

	BuildType() (string, string)

	BuildVersion() (string, string)

	WithConfigFile(string) ConfigWriter

	WithArches([]string) ConfigWriter

	WithHandle(string) ConfigWriter

	WithInput(string) ConfigWriter

	WithName(string) ConfigWriter

	WithOem(string) ConfigWriter

	WithOutput(string) ConfigWriter

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
		state.Logger.Debugf("Init writing to [%s]", state.Output)
		return writer.WithArches(params.Arches).
			WithHandle(params.Handle).
			WithName(params.Name).
			WithType(params.Type).
			WithInput(params.MountInput).
			WithOutput(params.MountOutput).
			WithOem(params.OEM).
			WithVersion(params.Version).
			Write(state.Output)
	}

	return errorNoConfigFile
}

func handleLayoutInitPretend(writer ConfigWriter, params *InitParams, state *task.State) error {
	if state.Output != "" {
		var pretender = shared.NewConfigPretender(state.Output)

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
		return nil
	}

	return errorNoConfigFile
}

func handleLayoutInitUpdate(writer ConfigWriter, params *InitParams, state *task.State) error {
	if state.Output != "" {
		var err error

		state.Logger.Debugf("Init updating existing [%s]", state.Output)

		if err = handleLayoutInitCreate(writer.WithConfigFile(state.Output), params, state); err == nil {
			state.Completed = true
		}

		return err
	}

	return nil
}
