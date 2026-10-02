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

func TestNewDataStoreListTask(t *testing.T) {
	var testTask = NewDataStoreListTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func TestNewDataStoreResolveTask(t *testing.T) {
	var testTask = NewDataStoreResolveTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotNil(t, testTask.OnPrepare)
	assert.NotNil(t, testTask.OnComplete)
	assert.NotNil(t, testTask.OnPretend)
}

func Test_handleDataStoreListComplete(t *testing.T) {
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
			listDataStores: expectedInstance,
		}, nil
	}

	assert.NoError(t, handleDataStoreListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, expectedInstance, actual)
}

func Test_handleDataStoreListComplete_FilterOem(t *testing.T) {
	var expectedOem = "expectedOem"
	var expectedInstances = []DataLinkInstance{
		{
			Id: new(int64(37)),
			DataLink: DataLink{
				Oem:    expectedOem,
				Handle: "someHandle",
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
			listDataStores: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataStoreListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[0], actual[0])
}

func Test_handleDataStoreListComplete_FilterOemHandle(t *testing.T) {
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
			listDataStores: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataStoreListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[2], actual[0])
}

func Test_handleDataStoreListComplete_FilterOemHandleVersion(t *testing.T) {
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
			listDataStores: expectedInstances,
		}, nil
	}

	assert.NoError(t, handleDataStoreListComplete(testParams, testState))
	actual, ok := testState.Internal.([]DataLinkInstance)
	assert.True(t, ok)
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedInstances[2], actual[0])
}

func Test_handleDataStoreListComplete_ListError(t *testing.T) {
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
			listDataStoresError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataStoreListComplete(testParams, testState), expectedError)
	assert.Nil(t, testState.Internal)
}

func Test_handleDataStoreListComplete_SessionError(t *testing.T) {
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

	assert.ErrorIs(t, handleDataStoreListComplete(testParams, &task.State{}), expectedError)
}

func Test_handleDataStoreListContext(t *testing.T) {
	assert.NoError(t, handleDataStoreListContext(&DataInstanceListParams{}, &task.State{}))
}

func Test_handleDataStoreListContext_InvalidFilter(t *testing.T) {
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Handle: "handleWithNoOem",
		},
	}

	assert.Error(t, handleDataStoreListContext(testParams, &task.State{}), errorDataShareOemRequired)
}

func Test_handleDataStoreListContext_OutputKnown(t *testing.T) {
	var testState = &task.State{
		Output: "known",
	}

	assert.NoError(t, handleDataStoreListContext(&DataInstanceListParams{}, testState))
}

func Test_handleDataStoreListContext_WithLink(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataInstanceListParams{
		DataLink: &DataLink{
			Id:    new(int64(37)),
			Flags: new(DataLinkFlags.Active),
		},
	}

	assert.NoError(t, handleDataStoreListContext(testParams, testState))
	assert.Equal(t, cast.ToString(*testParams.DataLink.Id), testState.Output)
}

func Test_handleDataStoreListPretend(t *testing.T) {
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

	assert.NoError(t, handleDataStoreListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleDataStoreListPretend_WithId(t *testing.T) {
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

	assert.NoError(t, handleDataStoreListPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
	assert.Contains(t, output, cast.ToString(*testParams.DataLink.Id))
}

func Test_handleDataStoreListPretend_SessionError(t *testing.T) {
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

	assert.ErrorIs(t, handleDataStoreListPretend(testParams, &task.State{}), expectedError)
}

func Test_handleDataStoreResolveComplete(t *testing.T) {
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
	var testParams = &DataStoreResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataStoreName: "myName",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{
			listDataStores: testInstances,
		}, nil
	}

	assert.NoError(t, handleDataStoreResolveComplete(testParams, testState))
	assert.Equal(t, expectedId, *testParams.DataStoreId)
}

func Test_handleDataStoreResolveComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataStoreResolveParams{
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

	assert.ErrorIs(t, handleDataStoreResolveComplete(testParams, testState), expectedError)
}

func Test_handleDataStoreResolveComplete_ListError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataStoreResolveParams{
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
			listDataStoresError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataStoreResolveComplete(testParams, testState), expectedError)
}

func Test_handleDataStoreResolveComplete_NoResults(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Output: "someName",
	}
	var testParams = &DataStoreResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataStoreName: "myName",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubDataShareClient{}, nil
	}

	assert.NoError(t, handleDataStoreResolveComplete(testParams, testState))
	assert.Nil(t, testParams.DataStoreId)
}

func Test_handleDataStoreResolveComplete_OutputError(t *testing.T) {
	assert.ErrorIs(t, handleDataStoreResolveComplete(&DataStoreResolveParams{}, &task.State{}), errorDataStoreRequired)
}

func Test_handleDataStoreResolveContext(t *testing.T) {
	var testParams = &DataStoreResolveParams{
		DataStoreName: "aName",
	}

	assert.NoError(t, handleDataStoreResolveContext(testParams, &task.State{}))
}

func Test_handleDataStoreResolveContext_IdKnown(t *testing.T) {
	var testParams = &DataStoreResolveParams{
		DataStoreId: new(int64(37)),
	}

	assert.ErrorIs(t, handleDataStoreResolveContext(testParams, &task.State{}), errorDataStoreKnown)
}

func Test_handleDataStoreResolveContext_NameRequiredError(t *testing.T) {
	assert.ErrorIs(t, handleDataStoreResolveContext(&DataStoreResolveParams{}, &task.State{}), errorDataStoreRequired)
}

func Test_handleDataStoreResolveContext_OutputKnown(t *testing.T) {
	var testState = &task.State{Output: "output"}

	assert.NoError(t, handleDataStoreResolveContext(&DataStoreResolveParams{}, testState))
}

func Test_handleDataStoreResolveIncomplete(t *testing.T) {
	var testState = &task.State{
		Logger: logrus.New(),
		Error:  errorDataStoreKnown,
	}
	var testParams = &DataStoreResolveParams{
		DataStoreId: new(int64(37)),
	}

	assert.NoError(t, handleDataStoreResolveIncomplete(testParams, testState))
	assert.True(t, testState.Completed)
}

func Test_handleDataStoreResolveIncomplete_StateError(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Error: expectedError,
	}

	assert.ErrorIs(t, handleDataStoreResolveIncomplete(&DataStoreResolveParams{}, testState), expectedError)
	assert.False(t, testState.Completed)
}

func Test_handleDataStoreResolvePretend(t *testing.T) {
	var expectedError = errors.New("error")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataStoreResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataStoreId: new(int64(37)),
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

	assert.NoError(t, handleDataStoreResolvePretend(testParams, testState), expectedError)

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)
	assert.Contains(t, output, testParams.Broker.HostAddr)
}

func Test_handleDataStoreResolvePretend_ClientError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Logger: testLogger,
	}
	var testParams = &DataStoreResolveParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		DataStoreId: new(int64(37)),
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return nil, expectedError
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.ErrorIs(t, handleDataStoreResolvePretend(testParams, testState), expectedError)
	assert.Equal(t, 0, len(testHook.Entries))
}

func Test_handleDataStoreResolvePretend_KnownError(t *testing.T) {
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  errorDataStoreKnown,
		Logger: testLogger,
	}
	var testParams = &DataStoreResolveParams{
		DataStoreId: new(int64(37)),
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.NoError(t, handleDataStoreResolvePretend(testParams, testState))
	assert.Equal(t, 1, len(testHook.Entries))
}

func Test_handleDataStoreResolvePretend_StateError(t *testing.T) {
	var expectedError = errors.New("error")
	var testLogger, testHook = test.NewNullLogger()
	var testState = &task.State{
		Error:  expectedError,
		Logger: testLogger,
	}

	testLogger.SetLevel(logrus.DebugLevel)
	assert.ErrorIs(t, handleDataStoreResolvePretend(&DataStoreResolveParams{}, testState), expectedError)
	assert.Equal(t, 0, len(testHook.Entries))
}
