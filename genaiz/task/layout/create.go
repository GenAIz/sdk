package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"genaiz.com/genaiz/task"
)

type CreateParams struct {
	ConfigName string
	ConfigType *ConfigType
	FolderPath string
}

func (cp CreateParams) NeedsConfigFile() bool {
	return cp.ConfigType != nil && *cp.ConfigType != configTypeNone
}

func NewCreateTask() *task.Task[CreateParams] {
	return &task.Task[CreateParams]{
		Name:       "layout-create",
		OnPrepare:  handleLayoutCreateContext,
		OnComplete: handleLayoutCreate,
		OnPretend:  handleLayoutCreatePretend,
	}
}

func handleLayoutCreate(params *CreateParams, state *task.State) error {
	var file, path string
	var err error

	if state.Output == "" {
		path = params.FolderPath
		file, _ = filepath.Abs(filepath.Join(path, params.ConfigName+"."+*params.ConfigType))
	} else {
		path = filepath.Dir(state.Output)
		file, _ = filepath.Abs(state.Output)
	}

	state.Logger.Debugf("Creating path [%s]", path)

	if err = os.MkdirAll(path, 0750); err == nil {
		state.Logger.Debugf("Configuration file [%s]", file)

		if err = os.Chdir(path); err == nil {
			_, err = os.OpenFile(file, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0640)
		}
	}

	return err
}

func handleLayoutCreateContext(params *CreateParams, state *task.State) error {
	var path, _ = filepath.Abs(params.FolderPath)
	var dir, _ = os.Stat(path)

	state.Logger.Debugf("Inspecting path [%s]", path)

	if params.NeedsConfigFile() {
		var configFilePath = filepath.Join(path, params.ConfigName+"."+*params.ConfigType)

		if file, _ := os.Stat(configFilePath); file != nil {
			return errors.New("context already exist")
		} else if dir != nil && dir.Mode()&os.ModePerm == os.ModePerm {
			return errors.New("context is not writeable")
		}

		state.Output = configFilePath
	}

	return nil
}

func handleLayoutCreatePretend(params *CreateParams, state *task.State) error {
	fmt.Printf("mkdir %s && cd %s\n", params.FolderPath, params.FolderPath)
	fmt.Printf("touch %s\n", state.Output)
	return nil
}
