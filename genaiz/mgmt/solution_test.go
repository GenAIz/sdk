package mgmt

import (
	"fmt"
	"testing"
	"time"

	"github.com/sirupsen/logrus"
	"github.com/spf13/cast"
	"github.com/stretchr/testify/assert"

	"genaiz.com/genaiz-lib/lang/timez"
	"genaiz.com/genaiz/lang"
	"genaiz.com/genaiz/task"
	"genaiz.com/genaiz/task/broker"
)

func TestUserSolution_Match(t *testing.T) {
	var testUserSolution = &UserSolution{}

	// empty filter always match all solutions
	assert.True(t, testUserSolution.Match(""))
	testUserSolution.Oem = "oem"
	assert.True(t, testUserSolution.Match(testUserSolution.Oem))
	assert.True(t, testUserSolution.Match(testUserSolution.Oem[0:2]))
	testUserSolution.Handle = "handle"
	testUserSolution.Fqdn = fmt.Sprintf("%s/%s", testUserSolution.Oem, testUserSolution.Handle)
	assert.True(t, testUserSolution.Match(testUserSolution.Fqdn))
	assert.True(t, testUserSolution.Match(testUserSolution.Fqdn[0:5]))
	testUserSolution.Version = "version"
	testPartialVersion := fmt.Sprintf("%s:%s", testUserSolution.Fqdn, testUserSolution.Version)
	assert.True(t, testUserSolution.Match(testPartialVersion[0:14]))
}

func TestUserSolution_Match_Id(t *testing.T) {
	var testUserSolution = &UserSolution{Id: 237}

	assert.False(t, testUserSolution.Match("345"))
	assert.True(t, testUserSolution.Match("23"))
	assert.True(t, testUserSolution.Match(cast.ToString(testUserSolution.Id)))
}

func TestUserSolution_MarshalJSON(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testModified = time.Now()
	var expectedModified = timez.NewTodayFormatter().FormatMillis(testModified.UnixMilli())
	var expectedDigest = "123456789A"
	var testUserSolution = &UserSolution{
		Id:          37,
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Fqdn:        "expectedFqdn",
		Digest:      "sha256:" + expectedDigest,
		Name:        "expectedName",
		Description: "expectedDescription",
		Created:     testCreated.UnixMilli(),
		Modified:    testModified.UnixMilli(),
		Flags:       lang.Ref(1337),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserSolution.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserSolution.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"oem\":\"%s\"", testUserSolution.Oem))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testUserSolution.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"version\":\"%s\"", testUserSolution.Version))
		assert.Contains(t, actual, fmt.Sprintf("\"fqdn\":\"%s\"", testUserSolution.Fqdn))
		assert.Contains(t, actual, fmt.Sprintf("\"digest\":\"%s\"", expectedDigest))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserSolution.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserSolution.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", expectedCreated))
		assert.Contains(t, actual, fmt.Sprintf("\"modified\":\"%s\"", expectedModified))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testUserSolution.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSolution_MarshalJSON_InvalidDigest(t *testing.T) {
	var expectedDigest = "invalidHash"
	var testUserSolution = &UserSolution{
		Digest: expectedDigest,
	}
	var bytes []byte
	var err error

	if bytes, err = testUserSolution.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.NotContains(t, actual, "\"id\":")
		assert.Contains(t, actual, "\"oem\":")
		assert.Contains(t, actual, "\"handle\":")
		assert.Contains(t, actual, "\"version\":")
		assert.Contains(t, actual, "\"fqdn\":")
		assert.NotContains(t, actual, fmt.Sprintf("\"digest\":\"%s\"", expectedDigest))
		assert.NotContains(t, actual, "\"name\":")
		assert.NotContains(t, actual, "\"description\":")
		assert.NotContains(t, actual, "\"created\":")
		assert.NotContains(t, actual, "\"modified\":")
		assert.NotContains(t, actual, "\"flags\":")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSolution_MarshalJSON_NoDigest(t *testing.T) {
	var testUserSolution = &UserSolution{}
	var bytes []byte
	var err error

	if bytes, err = testUserSolution.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.NotContains(t, actual, "\"id\":")
		assert.Contains(t, actual, "\"oem\":")
		assert.Contains(t, actual, "\"handle\":")
		assert.Contains(t, actual, "\"version\":")
		assert.Contains(t, actual, "\"fqdn\":")
		assert.NotContains(t, actual, "\"digest\":")
		assert.NotContains(t, actual, "\"name\":")
		assert.NotContains(t, actual, "\"description\":")
		assert.NotContains(t, actual, "\"created\":")
		assert.NotContains(t, actual, "\"modified\":")
		assert.NotContains(t, actual, "\"flags\":")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSolution_MarshalSlice(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserSolution = &UserSolution{
		Id:          37,
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Fqdn:        "expectedFqdn",
		Name:        "expectedName",
		Description: "expectedDescription",
		Created:     testCreated.UnixMilli(),
		Local:       true,
		Released:    true,
	}
	var values []string
	var err error

	if values, err = testUserSolution.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserSolution.Id))
		assert.Equal(t, values[1], testUserSolution.Fqdn)
		assert.Equal(t, values[2], testUserSolution.Version)
		assert.Equal(t, values[3], testUserSolution.Name)
		assert.Equal(t, values[4], expectedCreated)
		assert.Equal(t, values[5], "yes")
		assert.Equal(t, values[6], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSolution_MarshalSlice_NoCreate(t *testing.T) {
	var testUserSolution = &UserSolution{
		Id:          37,
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Fqdn:        "expectedFqdn",
		Name:        "expectedName",
		Description: "expectedDescription",
		Local:       false,
		Released:    false,
	}
	var values []string
	var err error

	if values, err = testUserSolution.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserSolution.Id))
		assert.Equal(t, values[1], testUserSolution.Fqdn)
		assert.Equal(t, values[2], testUserSolution.Version)
		assert.Equal(t, values[3], testUserSolution.Name)
		assert.Equal(t, values[4], "-")
		assert.Equal(t, values[5], "no")
		assert.Equal(t, values[6], "no")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserSolutionsFacade_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserSolutionFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserSolutionsFacade_Provider(t *testing.T) {
	var testProvider = NewUserSolutionFacade().
		WithParams(&broker.SolutionListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserSolutionsProvider_Get(t *testing.T) {
	var calledParams broker.SolutionListParams
	var testSolutions = []broker.Solution{
		{
			Id:    lang.Ref(int64(37)),
			Name:  "expected",
			Flags: lang.Ref(10),
		},
	}
	var testParams = &broker.SolutionListParams{}
	var testProvider = &userSolutionsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                    testParams,
		solutionReduceTaskFactory: newSolutionListTaskCompleteCapture(&calledParams, testSolutions),
	}
	var actual []UserSolution
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, *testSolutions[0].Id, actual[0].Id)
		assert.Equal(t, testSolutions[0].Name, actual[0].Name)
		assert.Equal(t, testSolutions[0].Flags, actual[0].Flags)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserSolutionsProvider_Get_Failure(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testParams = &broker.SolutionListParams{
		Oem: "oem",
	}
	var testProvider = &userSolutionsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                    testParams,
		filter:                    "3",
		solutionListTaskFactory:   newSolutionListTaskCompleteError(expectedError),
		solutionReduceTaskFactory: newSolutionListTaskCompleteError(expectedError),
	}
	var err error

	if _, err = testProvider.Get(); err != nil {
		assert.ErrorIs(t, err, expectedError)
		return
	}

	assert.Fail(t, "expected an error")
}

func TestUserSolutionsProvider_Get_Filtered(t *testing.T) {
	var calledParams broker.SolutionListParams
	var testSolutions = []broker.Solution{
		{
			Id:      lang.Ref(int64(37)),
			Fqdn:    lang.Ref("oem/handle"),
			Name:    "expected",
			Flags:   lang.Ref(10),
			Created: lang.Ref(int64(9)),
		},
		{
			Id:   lang.Ref(int64(1337)),
			Fqdn: lang.Ref("not_oem/handle"),
		},
		{
			Id:       lang.Ref(int64(42)),
			Fqdn:     lang.Ref("oem/handle2"),
			Name:     "second",
			Created:  lang.Ref(int64(10)),
			Modified: lang.Ref(int64(11)),
		},
		{
			Id:       lang.Ref(int64(69)),
			Fqdn:     lang.Ref("oem/handle3"),
			Name:     "first",
			Created:  lang.Ref(int64(8)),
			Modified: lang.Ref(int64(14)),
		},
		{
			Id:       lang.Ref(int64(56)),
			Fqdn:     lang.Ref("oem/handle4"),
			Name:     "unchecked",
			Created:  lang.Ref(int64(8)),
			Modified: lang.Ref(int64(13)),
		},
	}
	var testParams = &broker.SolutionListParams{}
	var testProvider = &userSolutionsProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		filter:                    "oem",
		params:                    testParams,
		solutionListTaskFactory:   newSolutionListTaskCompleteCapture(&calledParams, testSolutions),
		solutionReduceTaskFactory: newSolutionListTaskCompleteCapture(&calledParams, testSolutions),
	}
	var actual []UserSolution
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 4, len(actual))
		assert.Equal(t, *testSolutions[2].Id, actual[0].Id)
		assert.Equal(t, *testSolutions[2].Fqdn, actual[0].Fqdn)
		assert.Equal(t, testSolutions[2].Name, actual[0].Name)
		assert.Equal(t, *testSolutions[2].Created, actual[0].Created)
		assert.Equal(t, *testSolutions[2].Modified, actual[0].Modified)
		assert.Equal(t, *testSolutions[0].Id, actual[1].Id)
		assert.Equal(t, *testSolutions[0].Fqdn, actual[1].Fqdn)
		assert.Equal(t, testSolutions[0].Name, actual[1].Name)
		assert.Equal(t, *testSolutions[0].Created, actual[1].Created)
		assert.Equal(t, testSolutions[0].Flags, actual[1].Flags)
		assert.Equal(t, *testSolutions[3].Id, actual[2].Id)
		assert.Equal(t, *testSolutions[3].Fqdn, actual[2].Fqdn)
		assert.Equal(t, testSolutions[3].Name, actual[2].Name)
		assert.Equal(t, *testSolutions[3].Created, actual[2].Created)
		assert.Equal(t, *testSolutions[3].Modified, actual[2].Modified)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func newSolutionListTaskCompleteCapture(capture *broker.SolutionListParams, seeded []broker.Solution) func() *task.Task[broker.SolutionListParams] {
	return func() *task.Task[broker.SolutionListParams] {
		return &task.Task[broker.SolutionListParams]{
			OnPrepare: func(params *broker.SolutionListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.SolutionListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newSolutionListTaskCompleteError(err error) func() *task.Task[broker.SolutionListParams] {
	return func() *task.Task[broker.SolutionListParams] {
		return &task.Task[broker.SolutionListParams]{
			OnPrepare: func(params *broker.SolutionListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.SolutionListParams, state *task.State) error {
				return err
			},
		}
	}
}
