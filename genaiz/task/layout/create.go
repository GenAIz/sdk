package layout

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"

	"genaiz.com/genaiz-lib/lang/filez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/shared"
)

type CreateParams struct {
	shared.ConfigParams
	FolderPath string
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
		if !params.IsConfigTypeNone() || state.Output != "" {
			state.Output, err = handleLayoutCreateFile(path, params, state)
		} else {
			err = os.Chdir(path)
		}
	}

	return err
}

func handleLayoutCreateContext(params *CreateParams, state *task.State) error {
	var path, _ = filepath.Abs(params.FolderPath)
	var dir, _ = os.Stat(path)

	state.Logger.Debugf("Inspecting path [%s]", path)

	if !params.IsConfigTypeNone() {
		var configFilePath = filepath.Join(path, params.ConfigName+"."+*params.ConfigType)

		if file, _ := os.Stat(configFilePath); file != nil {
			return errors.New("context already exist")
		} else if dir == nil {
			return errors.New("context is not writeable")
		}

		state.Output = configFilePath
	}

	return nil
}

func handleLayoutCreateFile(path string, params *CreateParams, state *task.State) (string, error) {
	var err error
	var filePath string

	if !params.IsConfigTypeNone() {
		filePath = params.GetConfigFile(path)
	} else if state.Output != "" {
		filePath = state.Output
	}

	state.Logger.Debugf("Configuration file [%s]", filePath)

	if err = os.Chdir(path); err == nil {
		var absPath, _ = filepath.Abs(filePath)
		var fd *os.File

		if fd, err = os.OpenFile(absPath, os.O_RDWR|os.O_CREATE|os.O_EXCL, 0660); fd != nil {
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

func handleLayoutCreatePretend(params *CreateParams, state *task.State) error {
	fmt.Printf("mkdir -p %s && cd %s\n", params.FolderPath, params.FolderPath)

	if state.Output != "" {
		fmt.Printf("touch %s\n", state.Output)
	}

	return nil
}
