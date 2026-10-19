package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker/dataset"
	"genaiz.com/genaiz/task/broker/workspace"
)

type stubDataSetClient struct {
	client
	createDataSetArg       *DataSet
	createDataSetError     error
	createDataSet          *DataSet
	listDataSetFilesError  error
	listDataSetFilesId     int64
	listDataSetFilesPath   string
	listDataSetFiles       []DataSetFile
	uploadDataSetError     error
	uploadDataSetFileId    int64
	uploadDataSetLocalPath string
	uploadDataSetPath      string
	uploadDataSetRenamed   string
	uploadDataSet          *DataSet
}

func (c *stubDataSetClient) CreateDataSet(dataSet *DataSet) (*DataSet, error) {
	c.createDataSetArg = dataSet
	return c.createDataSet, c.createDataSetError
}

func (c *stubDataSetClient) ListDataSetFiles(dataSetId int64, path string) ([]DataSetFile, error) {
	c.listDataSetFilesId = dataSetId
	c.listDataSetFilesPath = path
	return c.listDataSetFiles, c.listDataSetFilesError
}

func (c *stubDataSetClient) UploadDataSetFile(dataSetId int64, localPath, renamed, path string) (*DataSet, error) {
	c.uploadDataSetFileId = dataSetId
	c.uploadDataSetLocalPath = localPath
	c.uploadDataSetRenamed = renamed
	c.uploadDataSetPath = path
	return c.uploadDataSet, c.uploadDataSetError
}

func TestDataSetTasks_CreateTaskFactory(t *testing.T) {
	var testTask = DataSetTasks.CreateTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnIncomplete)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnPretend)
}

func TestDataSetTasks_ListFilesTaskFactory(t *testing.T) {
	var testTask = DataSetTasks.ListFilesTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnPretend)
}

func TestDataSetTasks_UploadTaskFactory(t *testing.T) {
	var testTask = DataSetTasks.UploadTaskFactory()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleDataSetCreateComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		CreateFields: dataset.CreateFields{
			Description: "myDesc",
		},
		NodeSetParams: &workspace.NodeSetParams{},
		NodeResolveParams: &workspace.NodeResolveParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var expectedDataSet = &DataSet{
		Id:   new(int64(42)),
		Name: "myDataSet",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			createDataSet: expectedDataSet,
		}, nil
	}

	assert.NoError(t, handleDataSetCreateComplete(testParams, testState))
	assert.Equal(t, expectedDataSet.Id, testParams.NodeSetParams.DataSetId)

	if actual, ok := testState.Internal.(DataSet); ok {
		assert.Equal(t, *expectedDataSet, actual)
	} else {
		assert.Fail(t, "expected a dataset")
	}
}

func Test_handleDataSetCreateComplete_CreateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		CreateFields: dataset.CreateFields{
			Description: "myDesc",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			createDataSetError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataSetCreateComplete(testParams, testState), expectedError)
}

func Test_handleDataSetCreateComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleDataSetCreateComplete(&DataSetCreateParams{}, &task.State{}), dataset.Errors.DataSetHandleRequired())
}

func Test_handleDataSetCreateComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetCreateComplete(testParams, testState), expectedError)
}

func Test_handleDataSetCreateContext(t *testing.T) {
	var testState = &task.State{}
	var testParams = &DataSetCreateParams{
		CreateFields: dataset.CreateFields{
			WorkspaceId: new(int64(57)),
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			NodeId: new(int64(73)),
		},
		NodeSetParams: &workspace.NodeSetParams{
			PortHandle: "port",
		},
	}
	var expectedHandle = fmt.Sprintf("%d_%s", *testParams.NodeResolveParams.NodeId, testParams.NodeSetParams.PortHandle)

	assert.NoError(t, handleDataSetCreateContext(testParams, testState))
	assert.Equal(t, expectedHandle, testState.Output)
}

func Test_handleDataSetCreateContext_InvalidDataSetHandle(t *testing.T) {
	var testParams = &DataSetCreateParams{
		CreateFields: dataset.CreateFields{
			WorkspaceId: new(int64(57)),
		},
	}

	assert.ErrorIs(t, handleDataSetCreateContext(testParams, &task.State{}), dataset.Errors.DataSetHandleRequired())
}

func Test_handleDataSetCreateContext_InvalidWorkspaceId(t *testing.T) {
	var testParams = &DataSetCreateParams{
		CreateFields: dataset.CreateFields{},
	}

	assert.ErrorIs(t, handleDataSetCreateContext(testParams, &task.State{}), dataset.Errors.DataSetWorkspaceRequired())
}

func Test_handleDataSetCreateContext_KnownDataSetId(t *testing.T) {
	var testParams = &DataSetCreateParams{
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}

	assert.ErrorIs(t, handleDataSetCreateContext(testParams, &task.State{}), dataset.Errors.DataSetIdKnown())
}

func Test_handleDataSetCreateContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleDataSetCreateContext(&DataSetCreateParams{}, testState))
}

func Test_handleDataSetCreateIncomplete(t *testing.T) {
	var testState = &task.State{}

	assert.NoError(t, handleDataSetCreateIncomplete(&DataSetCreateParams{}, testState))
	assert.True(t, testState.Completed)
}

func Test_handleDataSetCreateIncomplete_KnownDataSetId(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  dataset.Errors.DataSetIdKnown(),
		Logger: testLogger,
	}
	var testParams = &DataSetCreateParams{
		CreateFields: dataset.CreateFields{
			Name:        "name",
			Handle:      "handle",
			Description: "description",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(49)),
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleDataSetCreateIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
	assert.Equal(t, 4, len(testHook.Entries))
}

func Test_handleDataSetCreateIncomplete_StateError(t *testing.T) {
	var testState = &task.State{
		Error: errors.New("expected"),
	}

	assert.ErrorIs(t, handleDataSetCreateIncomplete(&DataSetCreateParams{}, testState), testState.Error)
	assert.True(t, testState.Completed)
}

func Test_handleDataSetCreatePretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		CreateFields: dataset.CreateFields{
			Description: "myDesc",
			Handle:      "myHandle",
			Name:        "myName",
		},
		NodeResolveParams: &workspace.NodeResolveParams{
			WorkspaceId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{}, nil
	}

	assert.NoError(t, handleDataSetCreatePretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.CreateFields.Description)
	assert.Contains(t, output, testParams.CreateFields.Handle)
	assert.Contains(t, output, testParams.CreateFields.Name)
	assert.Contains(t, output, cast.ToString(*testParams.NodeResolveParams.WorkspaceId))
}

func Test_handleDataSetCreatePretend_DataSetIdKnown(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  dataset.Errors.DataSetIdKnown(),
		Logger: testLogger,
	}
	var testParams = &DataSetCreateParams{
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(64)),
		},
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleDataSetCreatePretend(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
}

func Test_handleDataSetCreatePretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetCreateParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetCreatePretend(testParams, testState), expectedError)
}

func Test_handleDataSetCreatePretend_StateError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleDataSetCreatePretend(&DataSetCreateParams{}, testState), expectedError)
}

func Test_handleDataSetListFilesComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
		Internal: DataSet{
			Mimes: []string{"text"},
		},
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		ListFields: &dataset.ListFields{
			Path: filepath.Join(t.TempDir(), "file"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedFiles = []DataSetFile{
		{
			Name: "file.txt",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			listDataSetFiles: expectedFiles,
		}, nil
	}

	assert.NoError(t, handleDataSetListFilesComplete(testParams, testState))

	if actual, ok := testState.Internal.([]DataSetFile); !ok {
		assert.Fail(t, "expected a slice of data set files")
	} else {
		assert.Equal(t, expectedFiles, actual)
	}
}

func Test_handleDataSetListFilesComplete_ListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		ListFields: &dataset.ListFields{
			Path: filepath.Join(t.TempDir(), "file"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			listDataSetFilesError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataSetListFilesComplete(testParams, testState), expectedError)
}

func Test_handleDataSetListFilesComplete_MimesFormat(t *testing.T) {
	var expectedMimes = []string{"text", "application-json"}
	var expectedFormats = []string{"none", "none"}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
		Internal: DataSet{
			Mimes:   expectedMimes,
			Formats: expectedFormats,
		},
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		ListFields: &dataset.ListFields{
			Path: filepath.Join(t.TempDir(), "file"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedFiles = []DataSetFile{
		{
			Name: "file.txt",
		},
		{
			Name: "config.json",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			listDataSetFiles: expectedFiles,
		}, nil
	}

	assert.NoError(t, handleDataSetListFilesComplete(testParams, testState))

	if actual, ok := testState.Internal.([]DataSetFile); !ok {
		assert.Fail(t, "expected a slice of data set files")
	} else {
		assert.Equal(t, expectedFiles[0].Name, actual[0].Name)
		assert.Equal(t, expectedMimes[0], actual[0].Mime)
		assert.Equal(t, expectedFormats[0], actual[0].Format)
		assert.Equal(t, expectedFiles[1].Name, actual[1].Name)
		assert.Equal(t, expectedMimes[1], actual[1].Mime)
		assert.Equal(t, expectedFormats[1], actual[1].Format)
	}
}

func Test_handleDataSetListFilesComplete_NoDataSets(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		ListFields: &dataset.ListFields{
			Path: filepath.Join(t.TempDir(), "file"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedFiles = []DataSetFile{
		{
			Name: "file.txt",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			listDataSetFiles: expectedFiles,
		}, nil
	}

	assert.NoError(t, handleDataSetListFilesComplete(testParams, testState))

	if actual, ok := testState.Internal.([]DataSetFile); !ok {
		assert.Fail(t, "expected a slice of data set files")
	} else {
		assert.Equal(t, expectedFiles, actual)
	}
}

func Test_handleDataSetListFilesComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleDataSetListFilesComplete(&DataSetListFilesParams{}, &task.State{}), dataset.Errors.DataSetIdUnknown())
}

func Test_handleDataSetListFilesComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetListFilesComplete(testParams, testState), expectedError)
}

func Test_handleDataSetListFilesContext(t *testing.T) {
	var testParams = &DataSetListFilesParams{
		ListFields: &dataset.ListFields{
			DataSetId: new(int64(73)),
		},
	}

	assert.NoError(t, handleDataSetListFilesContext(testParams, &task.State{}))
}

func Test_handleDataSetListFilesContext_DataSetIdUnknown(t *testing.T) {
	assert.ErrorIs(t, handleDataSetListFilesContext(&DataSetListFilesParams{}, &task.State{}), dataset.Errors.DataSetIdUnknown())
}

func Test_handleDataSetListFilesContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleDataSetListFilesContext(&DataSetListFilesParams{}, testState))
}

func Test_handleDataSetListFilesPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(39)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{}, nil
	}

	assert.NoError(t, handleDataSetListFilesPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeSetParams.DataSetId))
}

func Test_handleDataSetListFilesPretend_FilePath(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		ListFields: &dataset.ListFields{
			DataSetId: new(int64(37)),
			Path:      "/confs",
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{}, nil
	}

	assert.NoError(t, handleDataSetListFilesPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.ListFields.DataSetId))
	assert.Contains(t, output, testParams.ListFields.Path)
}

func Test_handleDataSetListFilesPretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetListFilesParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetListFilesPretend(testParams, testState), expectedError)
}

func Test_handleDataSetUploadComplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
		Internal: []DataSetFile{
			{
				Name: "/not/the/file/you/are/looking/for",
			},
		},
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedDataSet = &DataSet{
		Id: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			uploadDataSet: expectedDataSet,
		}, nil
	}

	assert.NoError(t, handleDataSetUploadComplete(testParams, testState))

	if actual, ok := testState.Internal.(DataSet); ok {
		assert.Equal(t, *expectedDataSet, actual)
	} else {
		assert.Fail(t, "expected a data set")
	}
}

func Test_handleDataSetUploadComplete_DuplicateError(t *testing.T) {
	var expectedPath = "path"
	var expectedName = "name"
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
		Internal: []DataSetFile{
			{
				Name: "/" + filepath.Join(expectedPath, expectedName),
			},
		},
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		UploadFields: dataset.UploadFields{
			DestinationPath: "path",
			FileName:        "name",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedDataSet = &DataSet{
		Id: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			uploadDataSet: expectedDataSet,
		}, nil
	}

	assert.ErrorIs(t, handleDataSetUploadComplete(testParams, testState), dataset.Errors.DataSetFileExists())
}

func Test_handleDataSetUploadComplete_NoOutput(t *testing.T) {
	assert.ErrorIs(t, handleDataSetUploadComplete(&DataSetUploadParams{}, &task.State{}), dataset.Errors.DataSetIdUnknown())

}

func Test_handleDataSetUploadComplete_Overwriting(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var expectedDataSet = &DataSet{
		Id: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			uploadDataSet: expectedDataSet,
		}, nil
	}

	assert.NoError(t, handleDataSetUploadComplete(testParams, testState))

	if actual, ok := testState.Internal.(DataSet); ok {
		assert.Equal(t, *expectedDataSet, actual)
	} else {
		assert.Fail(t, "expected a data set")
	}
}

func Test_handleDataSetUploadComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetUploadComplete(testParams, testState), expectedError)
}

func Test_handleDataSetUploadComplete_UploadError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "output",
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		UploadFields: dataset.UploadFields{
			FilePath: filepath.Join(t.TempDir(), "testFile"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{
			uploadDataSetError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataSetUploadComplete(testParams, testState), expectedError)
}

func Test_handleDataSetUploadContext(t *testing.T) {
	var testState = &task.State{}
	var testParams = &DataSetUploadParams{
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(93)),
		},
		UploadFields: dataset.UploadFields{
			FilePath: filepath.Join(t.TempDir(), "file"),
		},
	}

	assert.NoError(t, handleDataSetUploadContext(testParams, testState))
	assert.Equal(t, testParams.UploadFields.FilePath, testState.Output)
}

func Test_handleDataSetUploadContext_DataSetIdUnknown(t *testing.T) {
	assert.ErrorIs(t, handleDataSetUploadContext(&DataSetUploadParams{}, &task.State{}), dataset.Errors.DataSetIdUnknown())
}

func Test_handleDataSetUploadContext_FileUnknown(t *testing.T) {
	var testParams = &DataSetUploadParams{
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(93)),
		},
	}

	assert.ErrorIs(t, handleDataSetUploadContext(testParams, &task.State{}), dataset.Errors.DataSetFileUnknown())
}

func Test_handleDataSetUploadContext_KnownOutput(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleDataSetUploadContext(&DataSetUploadParams{}, testState))
}

func Test_handleDataSetUploadPretend(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		UploadFields: dataset.UploadFields{
			FilePath: filepath.Join(t.TempDir(), "file.txt"),
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{}, nil
	}

	assert.NoError(t, handleDataSetUploadPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeSetParams.DataSetId))
	assert.Contains(t, output, testParams.FilePath)

}

func Test_handleDataSetUploadPretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	assert.ErrorIs(t, handleDataSetUploadPretend(testParams, testState), expectedError)
}

func Test_handleDataSetUploadPretend_Supplementary(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSetUploadParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		UploadFields: dataset.UploadFields{
			FilePath:        filepath.Join(t.TempDir(), "file.txt"),
			FileName:        "newName.txt",
			DestinationPath: "/path",
		},
		NodeSetParams: &workspace.NodeSetParams{
			DataSetId: new(int64(37)),
		},
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w
	defer func() {
		os.Stdout = stdoutRestore
	}()

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataSetClient{}, nil
	}

	assert.NoError(t, handleDataSetUploadPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, cast.ToString(*testParams.NodeSetParams.DataSetId))
	assert.Contains(t, output, testParams.FilePath)
	assert.Contains(t, output, testParams.FileName)
	assert.Contains(t, output, testParams.DestinationPath)
}
