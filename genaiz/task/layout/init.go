package layout

import (
	"errors"

	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
)

type ConfigWriter interface {
	Write(string) error

	BuildArches() (string, []string)

	BuildFqdn() (string, string)

	BuildInput() (string, string)

	BuildName() (string, string)

	BuildOem() (string, string)

	BuildOutput() map[string]string

	BuildType() (string, string)

	BuildVersion() (string, string)

	WithConfigFile(string) ConfigWriter

	WithArches([]string) ConfigWriter

	WithInput(string) ConfigWriter

	WithFqdn(string) ConfigWriter

	WithName(string) ConfigWriter

	WithOem(string) ConfigWriter

	WithOutput(string) ConfigWriter

	WithType(string) ConfigWriter

	WithVersion(string) ConfigWriter
}

type InitParams struct {
	CreateParams
	Arches       []ArchType
	FQDN         string
	FunctionName string
	FunctionType FunctionType
	MountInput   string
	MountOutput  string
	OEM          string
	Version      string
}

func NewInitTask(writer ConfigWriter) *task.Task[InitParams] {
	return &task.Task[InitParams]{
		Name:      "layout-init",
		OnPrepare: handleLayoutInitContext,
		OnIncomplete: func(params *InitParams, state *task.State) error {
			return handleLayoutInitUpdate(params, writer, state)
		},
		OnComplete: func(params *InitParams, state *task.State) error {
			return handleLayoutInitCreate(params, writer, state)
		},
		OnPretend: func(params *InitParams, state *task.State) error {
			return handleLayoutInitPretend(params, writer, state)
		},
	}
}

func handleLayoutInitCreate(params *InitParams, writer ConfigWriter, state *task.State) error {
	if state.Output != "" {
		state.Logger.Debugf("Init writing to [%s]", state.Output)
		return writer.WithArches(params.Arches).
			WithFqdn(params.FQDN).
			WithName(params.FunctionName).
			WithType(params.FunctionType).
			WithInput(params.MountInput).
			WithOutput(params.MountOutput).
			WithOem(params.OEM).
			WithVersion(params.Version).
			Write(state.Output)
	}

	return nil
}

func handleLayoutInitContext(params *InitParams, state *task.State) error {
	if state.Output == "" {
		state.Logger.Debugf("Init finding a configuration file for writing")

		if params.NeedsConfigFile() {
			state.Output = filez.FromWorkDir(params.ConfigName + "." + *params.ConfigType)
		} else {
			var file, _ = filez.FirstNamedFile(params.ConfigName)

			if file == "" {
				state.Logger.Errorf("could not find a configuration file for [%s]", params.ConfigName)
			}

			state.Output = file
			return errors.New("file should exists")
		}
	}

	return nil
}

func handleLayoutInitPretend(params *InitParams, writer ConfigWriter, state *task.State) error {
	if state.Output != "" {
		var pretender = newConfigPretender(state.Output)

		state.Logger.Debugf("Pretending to initialize [%s]", state.Output)

		writer.WithFqdn(params.FQDN)
		writer.WithName(params.FunctionName)
		pretendSlice(pretender, writer.WithArches(params.Arches).BuildArches)
		pretendValue(pretender, writer.BuildFqdn)
		pretendValue(pretender, writer.BuildName)
		pretendValue(pretender, writer.WithType(params.FunctionType).BuildType)
		pretendValue(pretender, writer.WithInput(params.MountInput).BuildInput)
		pretendMap(pretender, writer.WithOutput(params.MountOutput).BuildOutput)
		pretendValue(pretender, writer.WithOem(params.OEM).BuildOem)
		pretendValue(pretender, writer.WithVersion(params.Version).BuildVersion)
		return nil
	}

	return errors.New("no configuration file to write")
}

func handleLayoutInitUpdate(params *InitParams, writer ConfigWriter, state *task.State) error {
	if state.Output != "" {
		state.Logger.Debugf("Init updating existing [%s]", state.Output)
		return handleLayoutInitCreate(params, writer.WithConfigFile(state.Output), state)
	}

	return errors.New("no configuration file to update")
}
