package mgmt

import (
	"errors"
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestUserWorkspace_Match(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:   int64(37),
		Name: "workspaceName",
	}

	if uw := testUserWorkspace.Match("work"); uw == nil {
		assert.Fail(t, "expected a user workspace match")
	} else {
		assert.Equal(t, testUserWorkspace.Name, uw.matched)
	}
}

func TestUserWorkspace_Match_Id(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:   int64(1337),
		Name: "aHandleValue",
	}

	if uw := testUserWorkspace.Match("13"); uw == nil {
		assert.Fail(t, "expected a user workspace match")
	} else {
		assert.Equal(t, cast.ToString(testUserWorkspace.Id), uw.matched)
	}
}

func TestUserWorkspace_Match_NoMatch(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:   int64(1337),
		Name: "aHandleValue",
	}

	if uw := testUserWorkspace.Match("test"); uw != nil {
		assert.Fail(t, "not expected a user workspace to match")
	}
}

func TestUserWorkspace_Matched(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:      int64(37),
		matched: "expected",
	}
	var expected = fmt.Sprintf("%s\t%d", testUserWorkspace.matched, testUserWorkspace.Id)

	assert.Equal(t, expected, testUserWorkspace.Matched())
}

func TestUserWorkspace_Matched_Id(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:      int64(37),
		matched: "42",
	}
	var expected = fmt.Sprintf("%s\t%s", testUserWorkspace.matched, testUserWorkspace.Name)

	assert.Equal(t, expected, testUserWorkspace.Matched())
}

func TestUserWorkspace_Matched_Nothing(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:      int64(37),
		Name:    "Expected",
		matched: "",
	}
	var expected = fmt.Sprintf("%d\t%s", testUserWorkspace.Id, testUserWorkspace.Name)

	assert.Equal(t, expected, testUserWorkspace.Matched())
}

func TestUserWorkspace_MarshalJSON(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:          37,
		Name:        "expectedName",
		Description: "expectedDescription",
		OwnerAppId:  42,
		OwnerUserId: 69,
		Visibility:  "expectedVisibility",
	}
	var bytes []byte
	var err error

	if bytes, err = testUserWorkspace.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserWorkspace.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserWorkspace.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserWorkspace.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"ownerAppId\":%d", testUserWorkspace.OwnerAppId))
		assert.Contains(t, actual, fmt.Sprintf("\"ownerUserId\":%d", testUserWorkspace.OwnerUserId))
		assert.Contains(t, actual, fmt.Sprintf("\"visibility\":\"%s\"", testUserWorkspace.Visibility))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspace_MarshalJSON_WithCreated(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserWorkspace = &UserWorkspace{
		Id:          37,
		Name:        "expectedName",
		Description: "expectedDescription",
		OwnerUserId: 69,
		Visibility:  "expectedVisibility",
		Created:     testCreated.UnixMilli(),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserWorkspace.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserWorkspace.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserWorkspace.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserWorkspace.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"ownerUserId\":%d", testUserWorkspace.OwnerUserId))
		assert.Contains(t, actual, fmt.Sprintf("\"visibility\":\"%s\"", testUserWorkspace.Visibility))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", expectedCreated))
		assert.NotContains(t, actual, "ownerAppId")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspace_MarshalJSON_WithModified(t *testing.T) {
	var testModified = time.Now()
	var expectedModified = timez.NewTodayFormatter().FormatMillis(testModified.UnixMilli())
	var testUserWorkspace = &UserWorkspace{
		Id:          37,
		Name:        "expectedName",
		Description: "expectedDescription",
		OwnerUserId: 69,
		Visibility:  "expectedVisibility",
		Modified:    testModified.UnixMilli(),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserWorkspace.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserWorkspace.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserWorkspace.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserWorkspace.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"ownerUserId\":%d", testUserWorkspace.OwnerUserId))
		assert.Contains(t, actual, fmt.Sprintf("\"visibility\":\"%s\"", testUserWorkspace.Visibility))
		assert.Contains(t, actual, fmt.Sprintf("\"modified\":\"%s\"", expectedModified))
		assert.NotContains(t, actual, "ownerAppId")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspace_MarshalSlice(t *testing.T) {
	var testCreated = time.Now().UnixMilli()
	var testUserWorkspace = &UserWorkspace{
		Id:        37,
		Name:      "expectedName",
		Created:   testCreated,
		RcEnabled: true,
		Active:    true,
	}
	var values []string
	var err error

	if values, err = testUserWorkspace.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserWorkspace.Id))
		assert.Equal(t, values[1], testUserWorkspace.Name)
		assert.Equal(t, values[2], createdFormatter.FormatMillis(testCreated))
		assert.Equal(t, values[3], "yes")
		assert.Equal(t, values[4], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspace_MarshalSlice_NoCreated(t *testing.T) {
	var testUserWorkspace = &UserWorkspace{
		Id:        37,
		Name:      "expectedName",
		RcEnabled: true,
		Active:    true,
	}
	var values []string
	var err error

	if values, err = testUserWorkspace.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserWorkspace.Id))
		assert.Equal(t, values[1], testUserWorkspace.Name)
		assert.Equal(t, values[2], "-")
		assert.Equal(t, values[3], "yes")
		assert.Equal(t, values[4], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspacesFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserWorkspacesFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspacesFacade_Provider(t *testing.T) {
	var testProvider = NewUserWorkspacesFacade().
		WithParams(&broker.WorkspaceListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspacesProvider_Get(t *testing.T) {
	var calledParams broker.WorkspaceListParams
	var testWorkspaces = []broker.Workspace{
		{
			Id:    37,
			Name:  "expected",
			Flags: new(10),
		},
	}
	var testParams = &broker.WorkspaceListParams{
		RcEnabled: true,
	}
	var testProvider = &userWorkspacesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                   testParams,
		workspaceListTaskFactory: newWorkspaceListTaskCompleteCapture(&calledParams, testWorkspaces),
	}
	var actual []UserWorkspace
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testWorkspaces[0].Id, actual[0].Id)
		assert.Equal(t, testWorkspaces[0].Name, actual[0].Name)
		assert.Equal(t, testWorkspaces[0].Flags, actual[0].Flags)
		assert.True(t, calledParams.RcEnabled)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspacesProvider_Get_Failure(t *testing.T) {
	var expectedError = errors.New("expected")
	var testProvider = &userWorkspacesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		workspaceListTaskFactory: newWorkspaceListTaskCompleteError(expectedError),
	}
	var actual []UserWorkspace
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Fail(t, "expected an error")
		return
	}

	assert.Empty(t, actual)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestUserWorkspacesProvider_Get_Filtered(t *testing.T) {
	var calledParams broker.WorkspaceListParams
	var testWorkspaces = []broker.Workspace{
		{
			Id:          37,
			Name:        "expected",
			Description: "description",
			OwnerAppId:  42,
			OwnerUserId: 69,
			RcEnabled:   false,
			Visibility:  "visible",
			Created:     100,
			Modified:    101,
		},
	}
	var testParams = &broker.WorkspaceListParams{
		RcEnabled: true,
	}
	var testProvider = &userWorkspacesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                   testParams,
		filter:                   "3",
		workspaceListTaskFactory: newWorkspaceListTaskCompleteCapture(&calledParams, testWorkspaces),
	}
	var actual []UserWorkspace
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testWorkspaces[0].Id, actual[0].Id)
		assert.Equal(t, testWorkspaces[0].Name, actual[0].Name)
		assert.Equal(t, testWorkspaces[0].Description, actual[0].Description)
		assert.Equal(t, testWorkspaces[0].OwnerAppId, actual[0].OwnerAppId)
		assert.Equal(t, testWorkspaces[0].OwnerUserId, actual[0].OwnerUserId)
		assert.False(t, actual[0].RcEnabled)
		assert.Equal(t, testWorkspaces[0].Visibility, actual[0].Visibility)
		assert.Equal(t, testWorkspaces[0].Created, actual[0].Created)
		assert.Equal(t, testWorkspaces[0].Modified, actual[0].Modified)
		assert.True(t, calledParams.RcEnabled)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceFlow_MarshalJSON(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:              37,
		Name:            "expectedName",
		Description:     "expectedDescription",
		SolutionId:      1337,
		SolutionOem:     "expectedOem",
		SolutionHandle:  "expectedHandle",
		SolutionVersion: "expectedVersion",
		WorkflowHandle:  "expectedHandle",
		WorkflowId:      42,
		Active:          true,
		Ready:           false,
		Flags:           new(broker.WorkspaceFlowFlags.Active),
	}
	var bytes []byte
	var err error

	if bytes, err = testWorkspaceFlow.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testWorkspaceFlow.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testWorkspaceFlow.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testWorkspaceFlow.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionId\":%d", testWorkspaceFlow.SolutionId))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionOem\":\"%s\"", testWorkspaceFlow.SolutionOem))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionHandle\":\"%s\"", testWorkspaceFlow.SolutionHandle))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionVersion\":\"%s\"", testWorkspaceFlow.SolutionVersion))
		assert.Contains(t, actual, fmt.Sprintf("\"workflowId\":%d", testWorkspaceFlow.WorkflowId))
		assert.Contains(t, actual, fmt.Sprintf("\"workflowHandle\":\"%s\"", testWorkspaceFlow.WorkflowHandle))
		assert.NotContains(t, actual, "created")
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", broker.WorkspaceFlowFlags.Active))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceFlow_MarshalJSON_WithCreated(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:              37,
		Name:            "expectedName",
		Description:     "expectedDescription",
		SolutionId:      1337,
		SolutionOem:     "expectedOem",
		SolutionHandle:  "expectedHandle",
		SolutionVersion: "expectedVersion",
		WorkflowHandle:  "expectedHandle",
		WorkflowId:      42,
		Created:         1,
		Active:          true,
		Ready:           false,
		Flags:           new(broker.WorkspaceFlowFlags.Active),
	}
	var bytes []byte
	var err error

	if bytes, err = testWorkspaceFlow.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testWorkspaceFlow.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testWorkspaceFlow.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testWorkspaceFlow.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionId\":%d", testWorkspaceFlow.SolutionId))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionOem\":\"%s\"", testWorkspaceFlow.SolutionOem))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionHandle\":\"%s\"", testWorkspaceFlow.SolutionHandle))
		assert.Contains(t, actual, fmt.Sprintf("\"solutionVersion\":\"%s\"", testWorkspaceFlow.SolutionVersion))
		assert.Contains(t, actual, fmt.Sprintf("\"workflowId\":%d", testWorkspaceFlow.WorkflowId))
		assert.Contains(t, actual, fmt.Sprintf("\"workflowHandle\":\"%s\"", testWorkspaceFlow.WorkflowHandle))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", createdFormatter.FormatMillis(testWorkspaceFlow.Created)))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", broker.WorkspaceFlowFlags.Active))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceFlow_MarshalSlice(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:              37,
		Name:            "expectedName",
		Description:     "expectedDescription",
		SolutionId:      1337,
		SolutionOem:     "expectedOem",
		SolutionHandle:  "expectedHandle",
		SolutionVersion: "expectedVersion",
		WorkflowHandle:  "expectedHandle",
		WorkflowId:      42,
		Active:          true,
		Ready:           true,
		Flags:           new(broker.WorkspaceFlowFlags.Ready | broker.WorkspaceFlowFlags.Active),
	}
	var values []string
	var err error

	if values, err = testWorkspaceFlow.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testWorkspaceFlow.Id))
		assert.Equal(t, values[1], testWorkspaceFlow.Name)
		assert.Equal(t, values[2], testWorkspaceFlow.SolutionOem)
		assert.Equal(t, values[3], testWorkspaceFlow.SolutionHandle)
		assert.Equal(t, values[4], testWorkspaceFlow.SolutionVersion)
		assert.Equal(t, values[5], testWorkspaceFlow.WorkflowHandle)
		assert.Equal(t, values[6], "-")
		assert.Equal(t, values[7], "yes")
		assert.Equal(t, values[8], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceNode_MarshalSlice_WithCreated(t *testing.T) {
	var testCreated = time.Now().UnixMilli()
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:              37,
		Name:            "expectedName",
		Description:     "expectedDescription",
		SolutionId:      1337,
		SolutionOem:     "expectedOem",
		SolutionHandle:  "expectedHandle",
		SolutionVersion: "expectedVersion",
		WorkflowHandle:  "expectedHandle",
		WorkflowId:      42,
		Created:         testCreated,
		Active:          true,
		Ready:           true,
		Flags:           new(broker.WorkspaceFlowFlags.Ready | broker.WorkspaceFlowFlags.Active),
	}
	var values []string
	var err error

	if values, err = testWorkspaceFlow.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testWorkspaceFlow.Id))
		assert.Equal(t, values[1], testWorkspaceFlow.Name)
		assert.Equal(t, values[2], testWorkspaceFlow.SolutionOem)
		assert.Equal(t, values[3], testWorkspaceFlow.SolutionHandle)
		assert.Equal(t, values[4], testWorkspaceFlow.SolutionVersion)
		assert.Equal(t, values[5], testWorkspaceFlow.WorkflowHandle)
		assert.Equal(t, values[6], createdFormatter.FormatMillis(testCreated))
		assert.Equal(t, values[7], "yes")
		assert.Equal(t, values[8], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceFlow_Match(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:             int64(37),
		WorkflowHandle: "handle",
	}

	if uw := testWorkspaceFlow.Match("han"); uw == nil {
		assert.Fail(t, "expected a workspace flow match")
	} else {
		assert.Equal(t, testWorkspaceFlow.WorkflowHandle, uw.matched)
	}
}

func TestUserWorkspaceFlow_Match_Id(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:             int64(37),
		WorkflowHandle: "handle",
	}

	if uw := testWorkspaceFlow.Match("3"); uw == nil {
		assert.Fail(t, "expected a workspace flow match")
	} else {
		assert.Equal(t, cast.ToString(testWorkspaceFlow.Id), uw.matched)
	}
}

func TestUserWorkspaceFlow_Match_NoMatch(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:             int64(37),
		WorkflowHandle: "handle",
	}

	if uw := testWorkspaceFlow.Match("test"); uw != nil {
		assert.Fail(t, "not expecting a workspace flow match")
	}
}

func TestUserWorkspaceFlow_Matched(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id: int64(37),

		matched: "expected",
	}
	var expected = fmt.Sprintf("%s\t%d", testWorkspaceFlow.matched, testWorkspaceFlow.Id)

	assert.Equal(t, expected, testWorkspaceFlow.Matched())
}

func TestUserWorkspaceFlow_Matched_Id(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:             int64(37),
		WorkflowHandle: "handle",

		matched: "37",
	}
	var expected = fmt.Sprintf("%s\t%s", testWorkspaceFlow.matched, testWorkspaceFlow.WorkflowHandle)

	assert.Equal(t, expected, testWorkspaceFlow.Matched())
}

func TestUserWorkspaceFlow_Matched_NoMatch(t *testing.T) {
	var testWorkspaceFlow = &UserWorkspaceFlow{
		Id:             int64(37),
		WorkflowHandle: "handle",
	}
	var expected = fmt.Sprintf("%d\t%s", testWorkspaceFlow.Id, testWorkspaceFlow.WorkflowHandle)

	assert.Equal(t, expected, testWorkspaceFlow.Matched())
}

func TestUserWorkspaceFlowsFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserWorkspaceFlowsFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspaceFlowsFacade_Provider(t *testing.T) {
	var testProvider = NewUserWorkspaceFlowsFacade().
		WithParams(&broker.WorkspaceFlowListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspaceFlowsProvider_Get(t *testing.T) {
	var calledListParams broker.WorkspaceFlowListParams
	var calledResolveParams broker.WorkspaceFlowResolveParams
	var expectedWorkspaceId = int64(1337)
	var expectedWorkflowId = int64(42)
	var testWorkspaceFlows = []broker.WorkspaceFlow{
		{
			Id:         37,
			WorkflowId: expectedWorkflowId,
			Solution: &broker.Solution{
				Workflows: []broker.Workflow{
					{
						Id:     new(expectedWorkflowId),
						Handle: "expected",
					},
				},
			},
			Flags: new(10),
		},
	}
	var testParams = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &broker.WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
				WorkflowId: new(expectedWorkflowId),
			},
		},
	}
	var testProvider = &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		workspaceFlowListTaskFactory: newWorkspaceFlowListTaskCompleteCapture(&calledListParams, testWorkspaceFlows),
		workspaceResolveTaskFactory:  newWorkspaceFlowResolveTaskCompleteCapture(&calledResolveParams, expectedWorkspaceId),
	}
	var actual []UserWorkspaceFlow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testWorkspaceFlows[0].Id, actual[0].Id)
		assert.Equal(t, testWorkspaceFlows[0].Solution.Workflows[0].Handle, actual[0].WorkflowHandle)
		assert.Equal(t, testWorkspaceFlows[0].Flags, actual[0].Flags)
		assert.Equal(t, expectedWorkspaceId, *calledResolveParams.WorkspaceId)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceFlowsProvider_Get_Failure(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testProvider = &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params: &broker.WorkspaceFlowListParams{
			WorkspaceFlowResolveParams: &broker.WorkspaceFlowResolveParams{
				WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
					WorkspaceId: new(int64(37)),
					WorkflowId:  new(int64(42)),
				},
			},
		},
		workspaceFlowListTaskFactory: newWorkspaceFlowListTaskCompleteError(expectedError),
	}
	var actual []UserWorkspaceFlow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Fail(t, "expected an error")
		return
	}

	assert.Empty(t, actual)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestUserWorkspaceFlowsProvider_Get_Filtered(t *testing.T) {
	var calledListParams broker.WorkspaceFlowListParams
	var calledResolveParams broker.WorkspaceFlowResolveParams
	var expectedWorkspaceId = int64(1337)
	var expectedWorkflowId = int64(42)
	var testWorkspaceFlows = []broker.WorkspaceFlow{
		{
			Id:         37,
			WorkflowId: expectedWorkflowId,
			Solution: &broker.Solution{
				Workflows: []broker.Workflow{
					{
						Id:     new(expectedWorkflowId),
						Handle: "expected",
					},
				},
			},
			Flags: new(10),
		},
	}
	var testParams = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &broker.WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
				WorkflowId: new(expectedWorkflowId),
			},
		},
	}
	var testProvider = &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		filter:                       "exp",
		workspaceFlowListTaskFactory: newWorkspaceFlowListTaskCompleteCapture(&calledListParams, testWorkspaceFlows),
		workspaceResolveTaskFactory:  newWorkspaceFlowResolveTaskCompleteCapture(&calledResolveParams, expectedWorkspaceId),
	}
	var actual []UserWorkspaceFlow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, testWorkspaceFlows[0].Id, actual[0].Id)
		assert.Equal(t, testWorkspaceFlows[0].Solution.Workflows[0].Handle, actual[0].WorkflowHandle)
		assert.Equal(t, testWorkspaceFlows[0].Flags, actual[0].Flags)
		assert.Equal(t, expectedWorkspaceId, *calledResolveParams.WorkspaceId)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceFlowsProvider_Get_NoResults(t *testing.T) {
	var calledListParams broker.WorkspaceFlowListParams
	var testParams = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &broker.WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
				WorkspaceId: new(int64(1337)),
				WorkflowId:  new(int64(42)),
			},
		},
	}
	var testProvider = &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		filter:                       "exp",
		workspaceFlowListTaskFactory: newWorkspaceFlowListTaskCompleteCapture(&calledListParams, []broker.WorkspaceFlow{}),
	}
	var actual []UserWorkspaceFlow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Empty(t, actual)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceFlowsProvider_Get_NoSolution(t *testing.T) {
	var calledListParams broker.WorkspaceFlowListParams
	var calledResolveParams broker.WorkspaceFlowResolveParams
	var expectedWorkspaceId = int64(1337)
	var expectedWorkflowId = int64(42)
	var testWorkspaceFlows = []broker.WorkspaceFlow{
		{
			Id:         37,
			WorkflowId: expectedWorkflowId,
			Flags:      new(10),
		},
	}
	var testParams = &broker.WorkspaceFlowListParams{
		WorkspaceFlowResolveParams: &broker.WorkspaceFlowResolveParams{
			WorkspaceFlowCreateParams: &broker.WorkspaceFlowCreateParams{
				WorkflowId: new(expectedWorkflowId),
			},
		},
	}
	var testProvider = &userWorkspaceFlowsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		filter:                       "exp",
		workspaceFlowListTaskFactory: newWorkspaceFlowListTaskCompleteCapture(&calledListParams, testWorkspaceFlows),
		workspaceResolveTaskFactory:  newWorkspaceFlowResolveTaskCompleteCapture(&calledResolveParams, expectedWorkspaceId),
	}
	var actual []UserWorkspaceFlow
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Empty(t, actual)
		assert.Equal(t, expectedWorkspaceId, *calledResolveParams.WorkspaceId)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceNode_MarshalSlice(t *testing.T) {
	var testWorkspaceNode = &UserWorkspaceNode{
		Id:                   37,
		WorkspaceId:          42,
		WorkspaceFlowId:      1337,
		SmartFunctionId:      69,
		SmartFunctionOem:     "expectedOem",
		SmartFunctionHandle:  "expectedHandle",
		SmartFunctionVersion: "expectedVersion",
		WorkflowNodeId:       31337,
		WorkflowNodeHandle:   "expectedNodeHandle",
	}
	var values []string
	var err error

	if values, err = testWorkspaceNode.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testWorkspaceNode.Id))
		assert.Equal(t, values[1], testWorkspaceNode.SmartFunctionOem)
		assert.Equal(t, values[2], testWorkspaceNode.SmartFunctionHandle)
		assert.Equal(t, values[3], testWorkspaceNode.SmartFunctionVersion)
		assert.Equal(t, values[4], testWorkspaceNode.WorkflowNodeHandle)
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserWorkspaceNodesFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserWorkspaceNodesFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspaceNodesFacade_Provider(t *testing.T) {
	var testProvider = NewUserWorkspaceNodesFacade().
		WithParams(&broker.WorkspaceNodeListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserWorkspaceNodesProvider_Get(t *testing.T) {
	var calledParams broker.WorkspaceNodeListParams
	var testNodes = []broker.WorkspaceNode{
		{
			Id:              37,
			WorkspaceId:     73,
			WorkspaceFlowId: 1337,
			WorkflowNodeId:  42,
			SmartFunctionId: 69,
			Flags:           new(10),
		},
		{
			Id: 42,
			SmartFunction: &broker.Function{
				Oem:     "expectedOem",
				Handle:  "expectedHandle",
				Version: "expectedVersion",
			},
			Flags: new(10),
		},
	}
	var testParams = &broker.WorkspaceNodeListParams{}
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteCapture(&calledParams, testNodes),
	}
	var actual []UserWorkspaceNode
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 2, len(actual))
		assert.Equal(t, testNodes[0].Id, actual[0].Id)
		assert.Equal(t, testNodes[0].WorkspaceId, actual[0].WorkspaceId)
		assert.Equal(t, testNodes[0].WorkspaceFlowId, actual[0].WorkspaceFlowId)
		assert.Equal(t, testNodes[0].WorkflowNodeId, actual[0].WorkflowNodeId)
		assert.Equal(t, testNodes[0].SmartFunctionId, actual[0].SmartFunctionId)
		assert.Equal(t, testNodes[1].SmartFunction.Oem, actual[1].SmartFunctionOem)
		assert.Equal(t, testNodes[1].SmartFunction.Handle, actual[1].SmartFunctionHandle)
		assert.Equal(t, testNodes[1].SmartFunction.Version, actual[1].SmartFunctionVersion)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserWorkspaceNodesProvider_Get_Failure(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       &broker.WorkspaceNodeListParams{},
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteError(expectedError),
	}
	var actual []UserWorkspaceNode
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Fail(t, "expected an error")
		return
	}

	assert.Empty(t, actual)
	assert.Equal(t, err.Error(), expectedError.Error())
}

func TestUserWorkspaceNodesProvider_Get_NoResults(t *testing.T) {
	var calledParams broker.WorkspaceNodeListParams
	var testParams = &broker.WorkspaceNodeListParams{}
	var testProvider = &userWorkspaceNodesProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                       testParams,
		workspaceNodeListTaskFactory: newWorkspaceNodeListTaskCompleteCapture(&calledParams, []broker.WorkspaceNode{}),
	}
	var actual []UserWorkspaceNode
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Empty(t, actual)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func newWorkspaceListTaskCompleteCapture(capture *broker.WorkspaceListParams, seeded []broker.Workspace) func() *task.Task[broker.WorkspaceListParams] {
	return func() *task.Task[broker.WorkspaceListParams] {
		return &task.Task[broker.WorkspaceListParams]{
			OnPrepare: func(params *broker.WorkspaceListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newWorkspaceListTaskCompleteError(err error) func() *task.Task[broker.WorkspaceListParams] {
	return func() *task.Task[broker.WorkspaceListParams] {
		return &task.Task[broker.WorkspaceListParams]{
			OnPrepare: func(params *broker.WorkspaceListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceListParams, state *task.State) error {
				return err
			},
		}
	}
}

func newWorkspaceFlowListTaskCompleteCapture(capture *broker.WorkspaceFlowListParams, seeded []broker.WorkspaceFlow) func() *task.Task[broker.WorkspaceFlowListParams] {
	return func() *task.Task[broker.WorkspaceFlowListParams] {
		return &task.Task[broker.WorkspaceFlowListParams]{
			OnPrepare: func(params *broker.WorkspaceFlowListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newWorkspaceFlowListTaskCompleteError(expected task.Error) func() *task.Task[broker.WorkspaceFlowListParams] {
	return func() *task.Task[broker.WorkspaceFlowListParams] {
		return &task.Task[broker.WorkspaceFlowListParams]{
			OnPrepare: func(params *broker.WorkspaceFlowListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowListParams, state *task.State) error {
				return expected
			},
		}
	}
}

func newWorkspaceFlowResolveTaskCompleteCapture(capture *broker.WorkspaceFlowResolveParams, seededId int64) func() *task.Task[broker.WorkspaceFlowResolveParams] {
	return func() *task.Task[broker.WorkspaceFlowResolveParams] {
		return &task.Task[broker.WorkspaceFlowResolveParams]{
			OnPrepare: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceFlowResolveParams, state *task.State) error {
				params.WorkspaceId = &seededId
				*capture = *params
				return nil
			},
		}
	}
}

func newWorkspaceNodeListTaskCompleteCapture(capture *broker.WorkspaceNodeListParams, seeded []broker.WorkspaceNode) func() *task.Task[broker.WorkspaceNodeListParams] {
	return func() *task.Task[broker.WorkspaceNodeListParams] {
		return &task.Task[broker.WorkspaceNodeListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newWorkspaceNodeListTaskCompleteError(expected task.Error) func() *task.Task[broker.WorkspaceNodeListParams] {
	return func() *task.Task[broker.WorkspaceNodeListParams] {
		return &task.Task[broker.WorkspaceNodeListParams]{
			OnPrepare: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.WorkspaceNodeListParams, state *task.State) error {
				return expected
			},
		}
	}
}
