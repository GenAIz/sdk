package broker

import (
	"errors"
	"fmt"
	"io"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/task"
)

func TestDataLinkParams_ToDataLink(t *testing.T) {
	var testParams = &DataLinkParams{
		Description: "expectedDescription",
		Handle:      "expectedHandle",
		Name:        "expectedName",
		Oem:         "expectedOem",
		Version:     "expectedVersion",
	}

	actual := testParams.ToDataLink()
	assert.Equal(t, testParams.Description, actual.Description)
	assert.Equal(t, testParams.Handle, actual.Handle)
	assert.Equal(t, testParams.Name, actual.Name)
	assert.Equal(t, testParams.Oem, actual.Oem)
	assert.Equal(t, testParams.Version, actual.Version)
}

func TestDataLinkParams_ToString(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:     "expected.oem",
		Handle:  "expected-handle",
		Version: "expected.version.final",
	}

	assert.Equal(t, fmt.Sprintf("%s/%s:%s", testParams.Oem, testParams.Handle, testParams.Version), testParams.ToString())
}

func TestDataLinkParam_isEqual(t *testing.T) {
	var testParams = &DataLinkParams{}
	var testLink = &DataLink{}

	assert.True(t, testParams.isEqual(testLink))
	testParams.Oem = "oem"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Oem = testParams.Oem
	assert.True(t, testParams.isEqual(testLink))
	testParams.Handle = "handle"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Handle = testParams.Handle
	assert.True(t, testParams.isEqual(testLink))
	testParams.Version = "version"
	assert.False(t, testParams.isEqual(testLink))
	testLink.Version = testParams.Version
	assert.True(t, testParams.isEqual(testLink))
}

func TestDataLinkParam_isValid(t *testing.T) {
	var testParams = &DataLinkParams{}

	assert.False(t, testParams.isValid())
	testParams.Oem = "oem"
	assert.False(t, testParams.isValid())
	testParams.Handle = "handle"
	assert.False(t, testParams.isValid())
	testParams.Version = "version"
	assert.True(t, testParams.isValid())
}

func TestNewDataLinkFindTask(t *testing.T) {
	var testTask = NewDataLinkFindTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func TestNewDataLinkPublishTask(t *testing.T) {
	var testTask = NewDataLinkPublishTask()

	assert.NotEmpty(t, testTask.Name)
	assert.NotEmpty(t, testTask.OnPrepare)
	assert.NotEmpty(t, testTask.OnComplete)
	assert.NotEmpty(t, testTask.OnPretend)
}

func Test_handleDataLinkAvailableError_InternalEmpty(t *testing.T) {
	var testState = &task.State{
		Internal: []DataLink{},
		Logger:   logrus.New(),
	}
	var testParams = &DataLinkParams{
		Oem:    "expectedOem",
		Handle: "expectedHandle",
	}

	if actual := handleDataLinkAvailableError(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, testParams.Oem)
		assert.Contains(t, actualText, testParams.Handle)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkAvailableError_InternalInvalid(t *testing.T) {
	var testState = &task.State{
		Internal: "invalid",
		Logger:   logrus.New(),
	}
	var testParams = &DataLinkParams{
		Oem:    "expectedOem",
		Handle: "expectedHandle",
	}

	if actual := handleDataLinkAvailableError(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, testParams.Oem)
		assert.Contains(t, actualText, testParams.Handle)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkAvailableError_InternalNil(t *testing.T) {
	assert.NoError(t, handleDataLinkAvailableError(&DataLinkParams{}, &task.State{}))
}

func Test_handleDataLinkFindContext(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}

	assert.NoError(t, handleDataLinkFindContext(testParams, &task.State{}))
}

func Test_handleDataLinkFindContext_InternalSet(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var testTask = &task.State{
		Internal: []DataLink{},
	}

	assert.NoError(t, handleDataLinkFindContext(testParams, testTask))
}

func Test_handleDataLinkFindContext_NoValid(t *testing.T) {
	assert.Error(t, handleDataLinkFindContext(&DataLinkParams{}, &task.State{}))
}

func Test_handleDataLinkFindComplete(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   DataLinkFlags.Active,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Oem:     expectedOem,
		Handle:  expectedHandle,
		Version: expectedVersion,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			listLinkData: expectedLinks,
		}, nil
	}

	assert.NoError(t, handleDataLinkFindComplete(testParams, testState))
	assert.Equal(t, expectedLinks[0], testState.Internal)
}

func Test_handleDataLinkFindComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
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

	assert.ErrorIs(t, expectedError, handleDataLinkFindComplete(testParams, &task.State{}))
}

func Test_handleDataLinkFindComplete_ListLinkError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
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
		return &stubFunctionClient{
			listLinkError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataLinkFindComplete(testParams, testState), expectedError)
}

func Test_handleDataLinkFindComplete_ListLinkInactive(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   0,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			listLinkData: expectedLinks,
		}, nil
	}

	if actual := handleDataLinkFindComplete(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, expectedOem)
		assert.Contains(t, actualText, expectedHandle)
		assert.Contains(t, actualText, expectedVersion)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkFindComplete_ListLinkUnavailable(t *testing.T) {
	var expectedOem = "expected.oem"
	var expectedHandle = "expected-handle"
	var expectedVersion = "expected-version"
	var expectedLinks = []DataLink{
		{
			Id:      int64(37),
			Oem:     expectedOem,
			Handle:  expectedHandle,
			Version: expectedVersion,
			Flags:   DataLinkFlags.Active,
		},
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			listLinkData: expectedLinks,
		}, nil
	}

	if actual := handleDataLinkFindComplete(testParams, testState); actual != nil {
		actualText := actual.Error()
		assert.Contains(t, actualText, expectedOem)
		assert.Contains(t, actualText, expectedHandle)
		assert.Contains(t, actualText, expectedVersion)
	} else {
		assert.Fail(t, "should have an error")
	}
}

func Test_handleDataLinkFindComplete_SkipValidation(t *testing.T) {
	var testParams = &DataLinkParams{NoValidation: true}
	var testState = &task.State{
		Logger: logrus.New(),
	}

	assert.NoError(t, handleDataLinkFindComplete(testParams, testState))
}

func Test_handleDataLinkFindPretend(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testParams = &DataLinkParams{
		Broker: Broker{
			HostAddr: expectedHostAddr,
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
		var clt = &stubFunctionClient{}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.NoError(t, handleDataLinkFindPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.HostAddr)
	assert.Contains(t, output, expectedToken)
}

func Test_handleDataLinkFindPretend_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
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

	assert.ErrorIs(t, handleDataLinkFindPretend(testParams, &task.State{}), expectedError)
}

func Test_handleDataLinkPublishContext(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
		Name:    "name",
	}

	assert.NoError(t, handleDataLinkPublishContext(testParams, &task.State{}))
}

func Test_handleDataLinkPublishContext_InternalSet(t *testing.T) {
	var testState = &task.State{
		Internal: &DataLink{},
	}

	assert.NoError(t, handleDataLinkPublishContext(&DataLinkParams{}, testState))
}

func Test_handleDataLinkPublishContext_InvalidFindContext(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:    "oem",
		Handle: "handle",
		Name:   "name",
	}

	assert.Error(t, handleDataLinkPublishContext(testParams, &task.State{}))
}

func Test_handleDataLinkPublishContext_InvalidPublishContext(t *testing.T) {
	var testParams = &DataLinkParams{
		Oem:     "oem",
		Handle:  "handle",
		Version: "version",
	}

	assert.Error(t, handleDataLinkPublishContext(testParams, &task.State{}))
}

func Test_handleDataLinkPublishComplete(t *testing.T) {
	var expectedLink = &DataLink{
		Oem:     "expected.oem",
		Handle:  "expected-handle",
		Version: "expected-version",
		Name:    "expected-name",
	}
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
		Broker: Broker{
			AuthFile: "file",
			HostAddr: "hostAddr",
		},
		Oem:     expectedLink.Oem,
		Handle:  expectedLink.Handle,
		Version: expectedLink.Version,
		Name:    expectedLink.Name,
	}
	var restoredFactory = clientFactory.Get

	defer func() {
		clientFactory.Get = restoredFactory
	}()
	clientFactory.Get = func(authFile, addr string) (Client, error) {
		return &stubFunctionClient{
			publishLinkData: expectedLink,
		}, nil
	}

	assert.NoError(t, handleDataLinkPublishComplete(testParams, testState))
	assert.Equal(t, expectedLink, testState.Internal)
}

func Test_handleDataLinkPublishComplete_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
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

	assert.ErrorIs(t, expectedError, handleDataLinkPublishComplete(testParams, &task.State{}))
}

func Test_handleDataLinkPublishComplete_PublishLinkError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testState = &task.State{
		Logger: logrus.New(),
	}
	var testParams = &DataLinkParams{
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
		return &stubFunctionClient{
			publishLinkError: expectedError,
		}, nil
	}

	assert.ErrorIs(t, handleDataLinkPublishComplete(testParams, testState), expectedError)
}

func Test_handleDataLinkPublishPretend(t *testing.T) {
	var expectedToken = "expectedToken"
	var expectedHostAddr = "expectedHost"
	var testParams = &DataLinkParams{
		Broker: Broker{
			HostAddr: expectedHostAddr,
		},
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Name:        "expectedName",
		Description: "expectedDescription",
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
		var clt = &stubFunctionClient{}

		clt.AuthToken = expectedToken
		clt.HostAddr = expectedHostAddr
		return clt, nil
	}

	assert.NoError(t, handleDataLinkPublishPretend(testParams, testState))

	_ = w.Close()
	b, _ := io.ReadAll(r)
	output := string(b)

	assert.Contains(t, output, testParams.HostAddr)
	assert.Contains(t, output, expectedToken)
	assert.Contains(t, output, testParams.Oem)
	assert.Contains(t, output, testParams.Handle)
	assert.Contains(t, output, testParams.Version)
	assert.Contains(t, output, testParams.Name)
	assert.Contains(t, output, testParams.Description)
}

func Test_handleDataLinkPublishPretend_ClientError(t *testing.T) {
	var expectedError = errors.New("expected")
	var testParams = &DataLinkParams{
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

	assert.ErrorIs(t, handleDataLinkPublishPretend(testParams, &task.State{}), expectedError)
}
