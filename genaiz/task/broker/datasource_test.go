package broker

import (
	"errors"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
)

func TestNewDataSourceListTask(t *testing.T) {
	var testTask = NewDataSourceListTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewDataSourceResolveTask(t *testing.T) {
	var testTask = NewDataSourceResolveTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleDataSourceListComplete(t *testing.T) {
	var expectedInstance = []DataLinkInstance{
		{
			Id: new(int64(37)),
		},
	}
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testState = &task.State{}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSources: expectedInstance,
		}, nil
	}

	assert.NoError(t, handleDataSourceListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, expectedInstance, actual)
}

func Test_handleDataSourceListComplete_FilterOem(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedInstances = []DataLinkInstance{
		{
			Id: new(int64(37)),
			DataLink: DataLink{
				Oem: expectedOem,
			},
		},
		{
			Id: new(int64(42)),
			DataLink: DataLink{
				Oem: "notExpected",
			},
		},
	}
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem: expectedOem,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSources: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataSourceListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[0], actual[0])
}

func Test_handleDataSourceListComplete_FilterOemHandle(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedInstances = []DataLinkInstance{
		{
			Id: new(int64(37)),
			DataLink: DataLink{
				Oem:    expectedOem,
				Handle: "notExpected",
			},
		},
		{
			Id: new(int64(42)),
			DataLink: DataLink{
				Oem: "notExpected",
			},
		},
		{
			Id: new(int64(69)),
			DataLink: DataLink{
				Oem:    expectedOem,
				Handle: expectedHandle,
			},
		},
	}
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem:    expectedOem,
			Handle: expectedHandle,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSources: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataSourceListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[2], actual[0])
}

func Test_handleDataSourceListComplete_FilterOemHandleVersion(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedHandle = "expectedHandle"
	var expectedVersion = "expectedVersion"
	var expectedSequence = 1
	var expectedInstances = []DataLinkInstance{
		{
			Id: new(int64(37)),
			DataLink: DataLink{
				Oem:    expectedOem,
				Handle: "notExpected",
			},
		},
		{
			Id: new(int64(42)),
			DataLink: DataLink{
				Oem: "notExpected",
			},
		},
		{
			Id: new(int64(69)),
			DataLink: DataLink{
				Oem:     expectedOem,
				Handle:  expectedHandle,
				Version: expectedVersion,
				Seq:     &expectedSequence,
			},
		},
	}
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSources: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataSourceListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[2], actual[0])
}

func Test_handleDataSourceListComplete_ListError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testState = &task.State{}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSourcesError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataSourceListComplete(testParams, testState), expectedError)
	assert.Nil(t, testState.Internal)
}

func Test_handleDataSourceListComplete_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataInstanceListParams{
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

	assert.ErrorIs(t, handleDataSourceListComplete(testParams, &task.State{}), expectedError)
}

func Test_handleDataSourceListContext(t *testing.T) {
	assert.NoError(t, handleDataSourceListContext(&DataInstanceListParams{}, &task.State{}))
}

func Test_handleDataSourceListContext_InvalidFilter(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Handle: "handleWithNoOem",
		},
	}

	assert.Error(t, handleDataSourceListContext(testParams, &task.State{}), errorDataShareOemRequired)
}

func Test_handleDataSourceListContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "known",
	}

	assert.NoError(t, handleDataSourceListContext(&DataInstanceListParams{}, testState))
}

func Test_handleDataSourceListContext_WithLink(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Id:    new(int64(37)),
			Flags: new(DataLinkFlags.Active),
		},
	}

	assert.NoError(t, handleDataSourceListContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.DataLink.Id), testState.Output)
}

func Test_handleDataSourceListPretend(t *testing.T) {
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleDataSourceListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleDataSourceListPretend_WithId(t *testing.T) {
	var testParams = &DataInstanceListParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataLink: &DataLink{
			Id: new(int64(37)),
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore

	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			client: client{
				HostAddr: testParams.Broker.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleDataSourceListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
	assert.Contains(t, output, cast.ToString(*testParams.DataLink.Id))
}

func Test_handleDataSourceListPretend_SessionError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataInstanceListParams{
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

	assert.ErrorIs(t, handleDataSourceListPretend(testParams, &task.State{}), expectedError)
}

func Test_handleDataSourceResolveComplete(t *testing.T) {
	var expectedName = "myName"
	var expectedId = int64(73)
	var testInstances = []DataLinkInstance{
		{
			Id:   new(int64(37)),
			Name: "myname",
		},
		{
			Id:   new(expectedId),
			Name: expectedName,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataSourceResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataSourceName: "myName",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataSources: testInstances,
		}, nil
	}

	assert.NoError(t, handleDataSourceResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.DataSourceId)
}

func Test_handleDataSourceResolveComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataSourceResolveParams{
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

	assert.ErrorIs(t, handleDataSourceResolveComplete(testParams, testState), expectedError)
}

func Test_handleDataSourceResolveComplete_ListError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataSourceResolveParams{
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
		return &stubDataShareClient{
			listDataSourcesError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataSourceResolveComplete(testParams, testState), expectedError)
}

func Test_handleDataSourceResolveComplete_NoResults(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataSourceResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataSourceName: "myName",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{}, nil
	}

	assert.NoError(t, handleDataSourceResolveComplete(testParams, testState))
	assert.Nil(t, testParams.DataSourceId)
}

func Test_handleDataSourceResolveComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleDataSourceResolveComplete(&DataSourceResolveParams{}, &task.State{}), errorDataSourceRequired)
}

func Test_handleDataSourceResolveContext(t *testing.T) {
	var testParams = &DataSourceResolveParams{
		DataSourceName: "aName",
	}

	assert.NoError(t, handleDataSourceResolveContext(testParams, &task.State{}))
}

func Test_handleDataSourceResolveContext_IdKnown(t *testing.T) {
	var testParams = &DataSourceResolveParams{
		DataSourceId: new(int64(37)),
	}

	assert.ErrorIs(t, handleDataSourceResolveContext(testParams, &task.State{}), errorDataSourceKnown)
}

func Test_handleDataSourceResolveContext_NameRequiredError(t *testing.T) {
	assert.ErrorIs(t, handleDataSourceResolveContext(&DataSourceResolveParams{}, &task.State{}), errorDataSourceRequired)
}

func Test_handleDataSourceResolveContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleDataSourceResolveContext(&DataSourceResolveParams{}, testState))
}

func Test_handleDataSourceResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  errorDataSourceKnown,
	}
	var testParams = &DataSourceResolveParams{
		DataSourceId: new(int64(37)),
	}

	assert.NoError(t, handleDataSourceResolveIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleDataSourceResolveIncomplete_StateError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleDataSourceResolveIncomplete(&DataSourceResolveParams{}, testState), expectedError)
	assert.False(t, testState.Completed)
}

func Test_handleDataSourceResolvePretend(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataSourceResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataSourceId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get
	var stdoutRestore = os.Stdout
	var r, w, _ = os.Pipe()

	os.Stdout = w

	defer func() {
		clientFactory.Get = restoredFactory
		os.Stdout = stdoutRestore
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			client: client{
				HostAddr: testParams.HostAddr,
			},
		}, nil
	}

	assert.NoError(t, handleDataSourceResolvePretend(testParams, testState), expectedError)

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleDataSourceResolvePretend_ClientError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
	}
	var testParams = &DataSourceResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataSourceId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.ErrorIs(t, handleDataSourceResolvePretend(testParams, testState), expectedError)
	assert.Equal(t, 0, len(testHook.Entries))
}

func Test_handleDataSourceResolvePretend_KnownError(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  errorDataSourceKnown,
		Logger: testLogger,
	}
	var testParams = &DataSourceResolveParams{
		DataSourceId: new(int64(37)),
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleDataSourceResolvePretend(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
}

func Test_handleDataSourceResolvePretend_StateError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  expectedError,
		Logger: testLogger,
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.ErrorIs(t, handleDataSourceResolvePretend(&DataSourceResolveParams{}, testState), expectedError)
	assert.Equal(t, 0, len(testHook.Entries))
}
