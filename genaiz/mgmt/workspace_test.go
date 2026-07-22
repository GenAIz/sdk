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
