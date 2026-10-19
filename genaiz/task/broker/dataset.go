package broker

import (
	"errors"
	"fmt"
	"path/filepath"
	"slices"
	"strings"

	"github.com/spf13/cast"

	"genaiz.com/genaiz-lib/lang"
	"genaiz.com/genaiz-lib/lang/stringz"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/dataset"
	"genaiz.com/genaiz/task/broker/workspace"
)

var (
	DataSetTasks struct {
		CreateTaskFactory    DataSetCreateTaskFactory
		ListFilesTaskFactory DataSetListFilesTaskFactory
		UploadTaskFactory    DataSetUploadTaskFactory
	}
)

type DataSetCreateTask = *task.Task[DataSetCreateParams]
type DataSetCreateTaskFactory = func() DataSetCreateTask
type DataSetListFilesTask = *task.Task[DataSetListFilesParams]
type DataSetListFilesTaskFactory = func() DataSetListFilesTask
type DataSetUploadTask = *task.Task[DataSetUploadParams]
type DataSetUploadTaskFactory = func() DataSetUploadTask

type DataSetCreateParams struct {
	Broker
	dataset.CreateFields
	*workspace.NodeResolveParams
	*workspace.NodeSetParams
}

func (p DataSetCreateParams) getDataSetId() *int64 {
	if p.NodeSetParams == nil {
		return nil
	}

	return p.NodeSetParams.DataSetId
}

func (p DataSetCreateParams) getDataSetHandle() string {
	if p.Handle == "" {
		var first, second string

		if p.NodeResolveParams != nil {
			if id := lang.FirstNonNil(p.NodeResolveParams.NodeId, p.NodeResolveParams.FlowId); id != nil {
				first = cast.ToString(id)
			}
		}

		if first != "" && p.NodeSetParams != nil {
			second = "_" + p.NodeSetParams.PortHandle
		}

		return fmt.Sprintf("%s%s", first, second)
	}

	return p.Handle
}

func (p DataSetCreateParams) getWorkspaceId() *int64 {
	if p.NodeResolveParams == nil {
		return p.CreateFields.WorkspaceId
	}

	return lang.FirstNonNil(p.CreateFields.WorkspaceId, p.NodeResolveParams.WorkspaceId)
}

func (p DataSetCreateParams) setDataSetId(id *int64) {
	if p.NodeSetParams != nil {
		p.DataSetId = id
	}
}

type DataSetListFilesParams struct {
	Broker
	*dataset.ListFields
	*workspace.NodeSetParams
}

func (p DataSetListFilesParams) getDataSetId() *int64 {
	if p.ListFields != nil && p.NodeSetParams != nil {
		return lang.FirstNonNil(p.ListFields.DataSetId, p.NodeSetParams.DataSetId)
	} else if p.ListFields != nil {
		return p.ListFields.DataSetId
	} else if p.NodeSetParams != nil {
		return p.NodeSetParams.DataSetId
	}

	return nil
}

func (p DataSetListFilesParams) getPath() string {
	if p.ListFields == nil {
		return ""
	}

	return p.ListFields.Path
}

type DataSetUploadParams struct {
	Broker
	dataset.UploadFields
	*workspace.NodeSetParams
}

func (p DataSetUploadParams) getDataSetId() *int64 {
	if p.NodeSetParams == nil {
		return nil
	}

	return p.NodeSetParams.DataSetId
}

func (p DataSetUploadParams) getEffectivePath() string {
	var name, dest string

	if p.FileName == "" {
		name = filepath.Base(p.FilePath)
	} else {
		name = p.FileName
	}

	if p.DestinationPath == "" {
		dest = "/"
	} else {
		if !strings.HasPrefix(p.DestinationPath, "/") {
			dest = "/" + p.DestinationPath
		}

		if !strings.HasSuffix(dest, "/") &&
			!strings.HasPrefix(name, "/") {
			dest += "/"
		}
	}

	return dest + name
}

func handleDataSetCreateComplete(params *DataSetCreateParams, state *task.State) error {
	if state.Output != "" {
		var name = stringz.FirstNonEmpty(params.Name, state.Output)
		var brokerClient Client
		var err error

		state.Logger.Debugf("Creating DataSet handle [%s] with name [%s]", state.Output, name)

		if brokerClient, err = params.GetClient(); err == nil {
			var dataSet = &DataSet{
				WorkspaceId: *params.getWorkspaceId(),
				Name:        name,
				Description: params.Description,
				Handle:      state.Output,
			}
			var created *DataSet

			if created, err = brokerClient.CreateDataSet(dataSet); err == nil {
				state.Logger.Debugf("Created DataSet [%d] on workspace [%d]", *created.Id, created.WorkspaceId)
				params.setDataSetId(created.Id)
				state.Internal = *created
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return dataset.Errors.DataSetHandleRequired()
}

func handleDataSetCreateContext(params *DataSetCreateParams, state *task.State) error {
	if state.Output == "" {
		if params.getDataSetId() != nil {
			return dataset.Errors.DataSetIdKnown()
		}

		if params.getWorkspaceId() == nil {
			return dataset.Errors.DataSetWorkspaceRequired()
		}

		if handle := params.getDataSetHandle(); handle != "" {
			state.Output = handle
			return nil
		}

		return dataset.Errors.DataSetHandleRequired()
	}

	return nil
}

func handleDataSetCreateIncomplete(params *DataSetCreateParams, state *task.State) error {
	state.Completed = true

	if errors.Is(state.Error, dataset.Errors.DataSetIdKnown()) {
		state.Logger.Debugf("DataSet id provided [%d], skipping creation", *params.getDataSetId())

		if params.Handle != "" {
			state.Logger.Warnf("Dataset handle [%s] ignored, dataset already exists", params.Handle)
		}

		if params.Name != "" {
			state.Logger.Warnf("Dataset name [%s] ignored, dataset already exists", params.Name)
		}

		if params.Description != "" {
			state.Logger.Warnf("Dataset description ignored, dataset already exists")
		}

		return nil
	}

	return state.Error
}

func handleDataSetCreatePretend(params *DataSetCreateParams, state *task.State) error {
	if state.Error == nil {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			state.Logger.Debugf("Pretending to create a dataset with handle: [%s]", params.Handle)
			fmt.Printf("genaiz raw post %s \\\n", brokerClient.CreateDataSetUrl())
			fmt.Printf("  -f workspaceId=%d \\\n", *params.getWorkspaceId())
			fmt.Printf("  -f name='%s' \\\n", params.Name)
			fmt.Printf("  -f description='%s'\\\n", params.Description)
			fmt.Printf("  -f handle='%s' \\\n", params.getDataSetHandle())
			fmt.Printf("  -r .id\n")
			return nil
		}

		return err
	}

	if errors.Is(state.Error, dataset.Errors.DataSetIdKnown()) {
		state.Logger.Debugf("DataSet id provided [%d], skipping creation", *params.getDataSetId())
		return nil
	}

	return state.Error
}

func handleDataSetListFilesComplete(params *DataSetListFilesParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		state.Logger.Debugf("Listing files under data set [%s]", state.Output)

		if brokerClient, err = params.GetClient(); err == nil {
			var dataSetId = *params.getDataSetId()
			var filesPath = params.getPath()
			var files []DataSetFile

			if filesPath != "" {
				state.Logger.Debugf("Listing from path [%s]", filesPath)
			}

			if files, err = brokerClient.ListDataSetFiles(dataSetId, filesPath); err == nil {
				state.Output = ""

				if ds, ok := state.Internal.(DataSet); ok {
					// If we have the dataSet use it
					var count = len(files)

					if len(ds.Mimes) == count && len(ds.Formats) == count {
						// Details will be out of sync if the dataSet is modified in between requests
						var detailedFiles []DataSetFile

						for i, file := range files {
							detailedFiles = append(detailedFiles, *file.With(ds.Formats[i], ds.Mimes[i]))
						}

						state.Internal = detailedFiles
						return nil
					}
				}

				state.Internal = files
				return nil
			}
		}

		return err
	}

	return dataset.Errors.DataSetIdUnknown()
}

func handleDataSetListFilesContext(params *DataSetListFilesParams, state *task.State) error {
	if state.Output == "" {
		var id = params.getDataSetId()

		if id == nil {
			return dataset.Errors.DataSetIdUnknown()
		}

		state.Output = cast.ToString(id)
	}

	return nil
}

func handleDataSetListFilesPretend(params *DataSetListFilesParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var dataSetId = *params.getDataSetId()
		var filePath = params.getPath()

		state.Logger.Debugf("Pretending to list dataset files by id: [%d]", dataSetId)
		fmt.Printf("genaiz raw get %s \\\n", brokerClient.ListDataSetFilesUrl())
		fmt.Printf("  -f dataSetId=%d", dataSetId)

		if filePath != "" {
			fmt.Printf(" \\\n  -f path=%s\n", filePath)
		} else {
			fmt.Printf("\n")
		}

		return nil
	}

	return err
}

func handleDataSetUploadComplete(params *DataSetUploadParams, state *task.State) error {
	if state.Output != "" {
		var brokerClient Client
		var err error

		if brokerClient, err = params.GetClient(); err == nil {
			var dsId = *params.getDataSetId()
			var effectivePath = params.getEffectivePath()
			var ds *DataSet

			if files, ok := state.Internal.([]DataSetFile); ok {
				state.Logger.Debugf("Overwriting disabled, checking existing files under data set [%s]", state.Output)

				if i := slices.IndexFunc(files, func(file DataSetFile) bool {
					// Sketchy, but DataSetFile.Name should have the full path
					return effectivePath == file.Name
				}); i >= 0 {
					return dataset.Errors.DataSetFileExists()
				}
			}

			state.Logger.Debugf("Uploading file [%s] to data set [%s]", params.FilePath, state.Output)

			if ds, err = brokerClient.UploadDataSetFile(dsId, params.FilePath, params.FileName, params.DestinationPath); err == nil {
				state.Logger.Debugf("Uploaded file [%s] to data set [%s]", params.FilePath, stringz.FirstNonEmpty(ds.Handle, ds.Name))

				if effectivePath != params.FilePath {
					state.Logger.Debugf("File path at destination: [%s]", effectivePath)
				}

				state.Reportf("Uploaded file [%s] to data set [%d]", effectivePath, dsId)
				state.Internal = *ds
				state.Output = ""
				return nil
			}
		}

		return err
	}

	return dataset.Errors.DataSetIdUnknown()
}

func handleDataSetUploadContext(params *DataSetUploadParams, state *task.State) error {
	if state.Output == "" {
		if params.getDataSetId() == nil {
			return dataset.Errors.DataSetIdUnknown()
		}

		if params.FilePath == "" {
			return dataset.Errors.DataSetFileUnknown()
		}

		state.Output = params.FilePath
	}

	return nil
}

func handleDataSetUploadPretend(params *DataSetUploadParams, state *task.State) error {
	var brokerClient Client
	var err error

	if brokerClient, err = params.GetClient(); err == nil {
		var dataSetId = *params.getDataSetId()
		var supplementary []string

		state.Logger.Debugf("Pretending to upload dataset file: [%s]", state.Output)
		fmt.Printf("genaiz raw post %s \\\n", brokerClient.UploadDataSetFileUrl())
		fmt.Printf("  -f dataSetId=%d\\\n", dataSetId)
		fmt.Printf("  -d file=%s", params.FilePath)

		if params.FileName != "" {
			supplementary = append(supplementary, fmt.Sprintf("-f name=%s", params.FileName))
		}

		if params.DestinationPath != "" {
			supplementary = append(supplementary, fmt.Sprintf("-f path=%s", params.DestinationPath))
		}

		if len(supplementary) > 0 {
			var sep = " \\\n  "

			fmt.Printf("%s%s\n", sep, strings.Join(supplementary, sep))
		} else {
			fmt.Printf("\n")
		}

		return nil
	}

	return err
}

func init() {
	DataSetTasks.CreateTaskFactory = newDataSetCreateTask
	DataSetTasks.ListFilesTaskFactory = newDataSetListFilesTask
	DataSetTasks.UploadTaskFactory = newDataSetUploadTask
}

func newDataSetCreateTask() DataSetCreateTask {
	return &task.Task[DataSetCreateParams]{
		Name:         "dataset-create",
		OnPrepare:    handleDataSetCreateContext,
		OnComplete:   handleDataSetCreateComplete,
		OnIncomplete: handleDataSetCreateIncomplete,
		OnPretend:    handleDataSetCreatePretend,
	}
}

func newDataSetListFilesTask() DataSetListFilesTask {
	return &task.Task[DataSetListFilesParams]{
		Name:       "dataset-list-files",
		OnPrepare:  handleDataSetListFilesContext,
		OnComplete: handleDataSetListFilesComplete,
		OnPretend:  handleDataSetListFilesPretend,
	}
}

func newDataSetUploadTask() DataSetUploadTask {
	return &task.Task[DataSetUploadParams]{
		Name:       "dataset-upload",
		OnPrepare:  handleDataSetUploadContext,
		OnComplete: handleDataSetUploadComplete,
		OnPretend:  handleDataSetUploadPretend,
	}
}
