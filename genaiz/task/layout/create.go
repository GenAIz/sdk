package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"genaiz.com/genaiz/lang/filez"
	"genaiz.com/genaiz/task"
)

type CreateParams struct {
	ConfigName string
	ConfigType *ConfigType
	FolderPath string
}

func (cp CreateParams) GetConfigFile(path string) string {
	return filepath.Join(path, cp.ConfigName+"."+*cp.ConfigType)
}

func (cp CreateParams) NeedsConfigFile() bool {
	return cp.ConfigType != nil && *cp.ConfigType != ConfigTypeNone
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
	var path string
	var err error

	if path, err = handleLayoutCreatePath(params, state); err == nil {
		if params.NeedsConfigFile() || state.Output != "" {
			state.Output, err = handleLayoutCreateFile(path, params, state)
		}
	}

	return err
}

func handleLayoutCreateFile(path string, params *CreateParams, state *task.State) (string, error) {
	var err error
	var filePath string

	if params.NeedsConfigFile() {
		filePath = params.GetConfigFile(path)
	} else if state.Output != "" {
		filePath = state.Output
	}

	state.Logger.Debugf("Configuration file [%s]", filePath)

	if err = os.Chdir(path); err == nil {
		var absPath, _ = filepath.Abs(filePath)
		var fd *os.File

		if fd, err = os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0640); fd != nil {
			filez.CloseSilently(fd)
		}
	}

	return filePath, err
}

func handleLayoutCreatePath(params *CreateParams, state *task.State) (string, error) {
	var path string
	var err error

	if state.Output == "" {
		path = params.FolderPath
	} else {
		path = filepath.Dir(state.Output)
	}

	state.Logger.Debugf("Creating path [%s]", path)
	err = os.MkdirAll(path, 0750)
	return path, err
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

	if state.Output != "" {
		fmt.Printf("touch %s\n", state.Output)
	}

	return nil
}
