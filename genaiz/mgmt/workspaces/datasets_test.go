package workspaces

import (
	"fmt"
	"path/filepath"
	"testing"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz/mgmt"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestUserWorkspaceDataSet_MarshalJSON(t *testing.T) {
	var testNode = &UserWorkspaceDataSet{
		Id:          new(int64(37)),
		Name:        "dsName",
		Handle:      "dsHandle",
		Description: "dsDescription",
		InputPort:   "port",
		TotalSize:   int64(1337),
		TotalFiles:  2,
		Digest:      "digest",
		RunId:       new(int64(42)),
		RunJobId:    new(int64(69)),
		Created:     "created",
		Modified:    "modified",
		Active:      false,
		Flags:       new(3),
	}
	var bytes []byte
	var err error

	if bytes, err = testNode.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", *testNode.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testNode.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testNode.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testNode.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"inputPort\":\"%s\"", testNode.InputPort))
		assert.Contains(t, actual, fmt.Sprintf("\"totalSize\":%d", testNode.TotalSize))
		assert.Contains(t, actual, fmt.Sprintf("\"totalFiles\":%d", testNode.TotalFiles))
		assert.Contains(t, actual, fmt.Sprintf("\"digest\":\"%s\"", testNode.Digest))
		assert.Contains(t, actual, fmt.Sprintf("\"runId\":%d", *testNode.RunId))
		assert.Contains(t, actual, fmt.Sprintf("\"runJobId\":%d", *testNode.RunJobId))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", testNode.Created))
		assert.Contains(t, actual, fmt.Sprintf("\"modified\":\"%s\"", testNode.Modified))
		assert.Contains(t, actual, "\"active\":false")
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testNode.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceDataSet_MarshalSlice(t *testing.T) {
	var testWorkspaceNode = &UserWorkspaceDataSet{
		Id:                  new(int64(37)),
		SmartFunctionHandle: "expectedHandle",
		InputPort:           "expectedPort",
		TotalFiles:          0,
		Modified:            "modified",
		Sealed:              "perhaps",
	}
	var values []string
	var err error

	if values, err = testWorkspaceNode.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, cast.ToString(*testWorkspaceNode.Id), values[0])
		assert.Equal(t, testWorkspaceNode.SmartFunctionHandle, values[1])
		assert.Equal(t, testWorkspaceNode.InputPort, values[2])
		assert.Equal(t, cast.ToString(testWorkspaceNode.TotalFiles), values[3])
		assert.Equal(t, testWorkspaceNode.Modified, values[4])
		assert.Equal(t, testWorkspaceNode.Sealed, values[5])
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceDataSet_MarshalSlice_NoDsId(t *testing.T) {
	var testWorkspaceNode = &UserWorkspaceDataSet{
		SmartFunctionHandle: "expectedHandle",
		InputPort:           "expectedPort",
		TotalFiles:          0,
		Modified:            "modified",
		Sealed:              "perhaps",
	}
	var values []string
	var err error

	if values, err = testWorkspaceNode.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, "-", values[0])
		assert.Equal(t, testWorkspaceNode.SmartFunctionHandle, values[1])
		assert.Equal(t, testWorkspaceNode.InputPort, values[2])
		assert.Equal(t, cast.ToString(testWorkspaceNode.TotalFiles), values[3])
		assert.Equal(t, testWorkspaceNode.Modified, values[4])
		assert.Equal(t, testWorkspaceNode.Sealed, values[5])
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceDataSet_Match(t *testing.T) {
	var testDataSet = &UserWorkspaceDataSet{
		Id:        new(int64(37)),
		Name:      "myName",
		InputPort: "aPort",
	}

	if actual := testDataSet.Match("3"); actual == nil {
		assert.Fail(t, "user workspace data set should have matched")
	} else {
		assert.Equal(t, testDataSet.Id, actual.Id)
		assert.Equal(t, testDataSet.InputPort, actual.InputPort)
		assert.Equal(t, testDataSet.Name, actual.Name)
		assert.Equal(t, cast.ToString(*testDataSet.Id), actual.matched)
	}
}

func TestUserWorkspaceDataSet_Match_Name(t *testing.T) {
	var testDataSet = &UserWorkspaceDataSet{
		Id:        new(int64(37)),
		Name:      "myName",
		InputPort: "aPort",
	}

	if actual := testDataSet.Match("myN"); actual == nil {
		assert.Fail(t, "user workspace data set should have matched")
	} else {
		assert.Equal(t, testDataSet.Id, actual.Id)
		assert.Equal(t, testDataSet.InputPort, actual.InputPort)
		assert.Equal(t, testDataSet.Name, actual.Name)
		assert.Equal(t, testDataSet.Name, actual.matched)
	}
}

func TestUserWorkspaceDataSet_Match_NoMatch(t *testing.T) {
	var testDataSet = &UserWorkspaceDataSet{
		Id:        new(int64(37)),
		Name:      "myName",
		InputPort: "aPort",
	}

	if actual := testDataSet.Match("will never match"); actual != nil {
		assert.Fail(t, "user workspace data set should not match")
	}
}

func TestUserWorkspaceDataSet_Match_Port(t *testing.T) {
	var testDataSet = &UserWorkspaceDataSet{
		Id:        new(int64(37)),
		Name:      "myName",
		InputPort: "aPort",
	}

	if actual := testDataSet.Match(testDataSet.InputPort); actual == nil {
		assert.Fail(t, "user workspace data set should have matched")
	} else {
		assert.Equal(t, testDataSet.Id, actual.Id)
		assert.Equal(t, testDataSet.InputPort, actual.InputPort)
		assert.Equal(t, testDataSet.Name, actual.Name)
		assert.Equal(t, testDataSet.InputPort, actual.matched)
	}
}

func TestUserWorkspaceDataSet_Matched(t *testing.T) {
	var expectedId = int64(1337)
	var expectedPort = "expectedPort"
	var testDataSet = &UserWorkspaceDataSet{
		Id:        &expectedId,
		InputPort: expectedPort,
		matched:   expectedPort,
	}

	assert.Equal(t, fmt.Sprintf("%s\t%d", expectedPort, expectedId), testDataSet.Matched())
}

func TestUserWorkspaceDataSet_Matched_DsId(t *testing.T) {
	var expectedId = int64(1337)
	var testDataSet = &UserWorkspaceDataSet{
		Id:        &expectedId,
		InputPort: "expectedPort",
		matched:   cast.ToString(expectedId),
	}

	assert.Equal(t, fmt.Sprintf("%d\t%s", expectedId, testDataSet.InputPort), testDataSet.Matched())
}

func TestUserWorkspaceDataSet_Matched_IntName(t *testing.T) {
	var expectedId = int64(1337)
	var expectedName = cast.ToString(int64(31337))
	var testDataSet = &UserWorkspaceDataSet{
		Id:        &expectedId,
		Name:      expectedName,
		InputPort: "expectedPort",
		matched:   expectedName,
	}

	assert.Equal(t, fmt.Sprintf("%s\t%s", expectedName, testDataSet.InputPort), testDataSet.Matched())
}

func TestUserWorkspaceDataSet_Matched_NoMatch(t *testing.T) {
	var expectedPort = "expectedPort"
	var testDataSet = &UserWorkspaceDataSet{
		InputPort: expectedPort,
	}

	assert.Equal(t, expectedPort, testDataSet.Matched())
}

func TestUserWorkspaceDataSet_Matched_PortAlone(t *testing.T) {
	var expectedPort = "expectedPort"
	var testDataSet = &UserWorkspaceDataSet{
		InputPort: expectedPort,
		matched:   expectedPort,
	}

	assert.Equal(t, expectedPort, testDataSet.Matched())
}

func TestUserWorkspaceDataSetsProvider_Get(t *testing.T) {
	var captureResolveParams broker.WorkspaceNodeResolveParams
	var captureSetListParams broker.WorkspaceNodeSetListParams
	var expectedParams = &broker.WorkspaceNodeResolveParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var expectedPort = []broker.WorkspaceNodePort{
		{
			DataSetId: new(int64(37)),
		},
		{
			PortHandle: "port2",
		},
		{
			DataSetId: new(int64(42)),
			DataSet: &broker.DataSet{
				Created:  new(int64(31337)),
				Modified: new(int64(31338)),
			},
			PortHandle: "port3",
		},
		{
			DataSetId:  new(int64(69)),
			DataSet:    &broker.DataSet{},
			PortHandle: "port4",
			SmartFunction: &broker.Function{
				Handle: "port4Handle",
			},
		},
	}
	var testProvider = &userWorkspaceDataSetsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		filter:                          new("port"),
		format:                          mgmt.DefaultTimeFormat,
		params:                          expectedParams,
		workspaceNodeResolveTaskFactory: newWorkspaceNodeResolveTaskCompleteCapture(&captureResolveParams),
		workspaceNodeSetListTaskFactory: newWorkspaceNodeSetListTaskCompleteCapture(&captureSetListParams, expectedPort),
	}

	actual, err := testProvider.Get()
	assert.Equal(t, 3, len(actual))
	assert.Equal(t, expectedPort[1].PortHandle, actual[0].InputPort)
	assert.Equal(t, expectedPort[2].DataSetId, actual[1].Id)
	assert.Equal(t, expectedPort[2].PortHandle, actual[1].InputPort)
	assert.Equal(t, testProvider.format.FormatMillis(*expectedPort[2].DataSet.Created), actual[1].Created)
	assert.Equal(t, testProvider.format.FormatMillis(*expectedPort[2].DataSet.Modified), actual[1].Modified)
	assert.Equal(t, expectedPort[3].DataSetId, actual[2].Id)
	assert.Equal(t, expectedPort[3].PortHandle, actual[2].InputPort)
	assert.Equal(t, expectedPort[3].SmartFunction.Handle, actual[2].SmartFunctionHandle)
	assert.NoError(t, err)
}

func TestUserWorkspaceDataSetsProvider_Get_Error(t *testing.T) {
	var captureSetListParams broker.WorkspaceNodeSetListParams
	var expectedError = task.NewError("expected")
	var expectedParams = &broker.WorkspaceNodeResolveParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var testProvider = &userWorkspaceDataSetsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                          expectedParams,
		workspaceNodeResolveTaskFactory: newWorkspaceNodeResolveTaskCompleteError(expectedError),
		workspaceNodeSetListTaskFactory: newWorkspaceNodeSetListTaskCompleteCapture(&captureSetListParams, []broker.WorkspaceNodePort{}),
	}

	actual, err := testProvider.Get()
	assert.Empty(t, actual)
	assert.ErrorIs(t, err, expectedError)
	assert.Empty(t, captureSetListParams)
}

func TestUserWorkspaceDataSetsProvider_Get_NoFilter(t *testing.T) {
	var captureResolveParams broker.WorkspaceNodeResolveParams
	var captureSetListParams broker.WorkspaceNodeSetListParams
	var expectedParams = &broker.WorkspaceNodeResolveParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var expectedPort = []broker.WorkspaceNodePort{
		{
			DataSetId: new(int64(37)),
		},
	}
	var testProvider = &userWorkspaceDataSetsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                          expectedParams,
		workspaceNodeResolveTaskFactory: newWorkspaceNodeResolveTaskCompleteCapture(&captureResolveParams),
		workspaceNodeSetListTaskFactory: newWorkspaceNodeSetListTaskCompleteCapture(&captureSetListParams, expectedPort),
	}

	actual, err := testProvider.Get()
	assert.Equal(t, 1, len(actual))
	assert.Equal(t, expectedPort[0].DataSetId, actual[0].Id)
	assert.NoError(t, err)
}

func TestUserWorkspaceDataSetsProvider_Get_NoResults(t *testing.T) {
	var captureResolveParams broker.WorkspaceNodeResolveParams
	var captureSetListParams broker.WorkspaceNodeSetListParams
	var expectedParams = &broker.WorkspaceNodeResolveParams{
		Broker: broker.Broker{
			AuthFile: filepath.Join(t.TempDir(), ".auth"),
		},
	}
	var testProvider = &userWorkspaceDataSetsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                          expectedParams,
		workspaceNodeResolveTaskFactory: newWorkspaceNodeResolveTaskCompleteCapture(&captureResolveParams),
		workspaceNodeSetListTaskFactory: newWorkspaceNodeSetListTaskCompleteCapture(&captureSetListParams, []broker.WorkspaceNodePort{}),
	}

	actual, err := testProvider.Get()
	assert.Empty(t, actual)
	assert.NoError(t, err)
}

func TestNewUserWorkspaceDataSetsFactory(t *testing.T) {
	var testFactory = NewUserWorkspaceDataSetsFactory()

	assert.NotNil(t, testFactory.Provider())
}
