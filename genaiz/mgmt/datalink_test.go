package mgmt

import (
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

func TestUserDataLink_Match(t *testing.T) {
	var testUserDataLink = &UserDataLink{}

	// empty filter always match all solutions
	assert.True(t, testUserDataLink.Match(""))
	testUserDataLink.Oem = "oem"
	assert.True(t, testUserDataLink.Match(testUserDataLink.Oem))
	assert.True(t, testUserDataLink.Match(testUserDataLink.Oem[0:2]))
	testUserDataLink.Handle = "handle"
	testUserDataLink.Fqdn = fmt.Sprintf("%s/%s", testUserDataLink.Oem, testUserDataLink.Handle)
	assert.True(t, testUserDataLink.Match(testUserDataLink.Fqdn))
	assert.True(t, testUserDataLink.Match(testUserDataLink.Fqdn[0:5]))
	testUserDataLink.Version = "version"
	testPartialVersion := fmt.Sprintf("%s:%s", testUserDataLink.Fqdn, testUserDataLink.Version)
	assert.True(t, testUserDataLink.Match(testPartialVersion[0:14]))
}

func TestUserDataLink_Match_Id(t *testing.T) {
	var testUserDataLink = &UserDataLink{Id: 237}

	assert.False(t, testUserDataLink.Match("345"))
	assert.True(t, testUserDataLink.Match("23"))
	assert.True(t, testUserDataLink.Match(cast.ToString(testUserDataLink.Id)))
}

func TestUserDataLink_MarshalJSON(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserDataLink = &UserDataLink{
		Id:          37,
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Fqdn:        "expectedFqdn",
		Name:        "expectedName",
		Description: "expectedDescription",
		Created:     testCreated.UnixMilli(),
		Flags:       new(1337),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserDataLink.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserDataLink.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"oem\":\"%s\"", testUserDataLink.Oem))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testUserDataLink.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"version\":\"%s\"", testUserDataLink.Version))
		assert.Contains(t, actual, fmt.Sprintf("\"fqdn\":\"%s\"", testUserDataLink.Fqdn))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserDataLink.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserDataLink.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"created\":\"%s\"", expectedCreated))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testUserDataLink.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserDataLink_MarshalJSON_Modified(t *testing.T) {
	var testModified = time.Now()
	var expectedModified = timez.NewTodayFormatter().FormatMillis(testModified.UnixMilli())
	var testUserDataLink = &UserDataLink{
		Id:          37,
		Oem:         "expectedOem",
		Handle:      "expectedHandle",
		Version:     "expectedVersion",
		Fqdn:        "expectedFqdn",
		Name:        "expectedName",
		Description: "expectedDescription",
		Modified:    testModified.UnixMilli(),
		Flags:       new(1337),
	}
	var bytes []byte
	var err error

	if bytes, err = testUserDataLink.MarshalJSON(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, bytes)
		actual := string(bytes)
		assert.Contains(t, actual, fmt.Sprintf("\"id\":%d", testUserDataLink.Id))
		assert.Contains(t, actual, fmt.Sprintf("\"oem\":\"%s\"", testUserDataLink.Oem))
		assert.Contains(t, actual, fmt.Sprintf("\"handle\":\"%s\"", testUserDataLink.Handle))
		assert.Contains(t, actual, fmt.Sprintf("\"version\":\"%s\"", testUserDataLink.Version))
		assert.Contains(t, actual, fmt.Sprintf("\"fqdn\":\"%s\"", testUserDataLink.Fqdn))
		assert.Contains(t, actual, fmt.Sprintf("\"name\":\"%s\"", testUserDataLink.Name))
		assert.Contains(t, actual, fmt.Sprintf("\"description\":\"%s\"", testUserDataLink.Description))
		assert.Contains(t, actual, fmt.Sprintf("\"modified\":\"%s\"", expectedModified))
		assert.Contains(t, actual, fmt.Sprintf("\"flags\":%d", *testUserDataLink.Flags))
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserDataLink_MarshalSlice(t *testing.T) {
	var testCreated = time.Now()
	var expectedCreated = timez.NewTodayFormatter().FormatMillis(testCreated.UnixMilli())
	var testUserDatalink = &UserDataLink{
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

	if values, err = testUserDatalink.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserDatalink.Id))
		assert.Equal(t, values[1], testUserDatalink.Fqdn)
		assert.Equal(t, values[2], testUserDatalink.Version)
		assert.Equal(t, values[3], testUserDatalink.Name)
		assert.Equal(t, values[4], expectedCreated)
		assert.Equal(t, values[5], "yes")
		assert.Equal(t, values[6], "yes")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserDataLink_MarshalSlice_NoCreate(t *testing.T) {
	var testUserDatalink = &UserDataLink{
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

	if values, err = testUserDatalink.MarshalSlice(); err == nil {
		assert.NoError(t, err)
		assert.NotEmpty(t, values)
		assert.Equal(t, values[0], cast.ToString(testUserDatalink.Id))
		assert.Equal(t, values[1], testUserDatalink.Fqdn)
		assert.Equal(t, values[2], testUserDatalink.Version)
		assert.Equal(t, values[3], testUserDatalink.Name)
		assert.Equal(t, values[4], "-")
		assert.Equal(t, values[5], "no")
		assert.Equal(t, values[6], "no")
	} else {
		assert.Fail(t, err.Error())
	}
}

func TestUserDataLink_Filtering(t *testing.T) {
	var expectedFilter = "filter"
	var testProvider = NewUserDataLinkFacade().
		WithLogger(logrus.New()).
		Filtering(expectedFilter)

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserDataLink_Provider(t *testing.T) {
	var testProvider = NewUserDataLinkFacade().
		WithParams(&broker.DataLinkListParams{}).
		Provider()

	// We can't unit test the provider here
	assert.NotNil(t, testProvider)
}

func TestUserDataLink_Get(t *testing.T) {
	var calledParams broker.DataLinkListParams
	var testDatalinks = []broker.DataLink{
		{
			Id:    new(int64(37)),
			Name:  "expected",
			Flags: new(10),
		},
	}
	var testParams = &broker.DataLinkListParams{}
	var testProvider = &userDataLinksProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                 testParams,
		listLinksTaskFactory:   newDataLinkListTaskCompleteCapture(&calledParams, testDatalinks),
		reduceLinksTaskFactory: newDataLinkListTaskCompleteCapture(&calledParams, testDatalinks),
	}
	var actual []UserDataLink
	var err error

	if actual, err = testProvider.Get(); err == nil {
		assert.Equal(t, 1, len(actual))
		assert.Equal(t, *testDatalinks[0].Id, actual[0].Id)
		assert.Equal(t, testDatalinks[0].Name, actual[0].Name)
		assert.Equal(t, testDatalinks[0].Flags, actual[0].Flags)
		return
	}

	assert.Fail(t, "expected a list of results")
}

func TestUserDataLink_Get_Failure(t *testing.T) {
	var expectedError = task.NewError("expected")
	var testParams = &broker.DataLinkListParams{
		Oem: "oem",
	}
	var testProvider = &userDataLinksProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		params:                 testParams,
		filter:                 "3",
		listLinksTaskFactory:   newDataLinkListTaskCompleteError(expectedError),
		reduceLinksTaskFactory: newDataLinkListTaskCompleteError(expectedError),
	}
	var err error

	if _, err = testProvider.Get(); err != nil {
		assert.ErrorIs(t, err, expectedError)
		return
	}

	assert.Fail(t, "expected an error")
}

func TestUserDataLink_Get_Filtered(t *testing.T) {
	var calledParams broker.DataLinkListParams
	var testSolutions = []broker.DataLink{
		{
			Id:      new(int64(37)),
			Fqdn:    new("oem/handle"),
			Name:    "expected",
			Flags:   new(10),
			Created: new(int64(9)),
		},
		{
			Id:   new(int64(1337)),
			Fqdn: new("not_oem/handle"),
		},
		{
			Id:       new(int64(42)),
			Fqdn:     new("oem/handle2"),
			Name:     "second",
			Created:  new(int64(10)),
			Modified: new(int64(11)),
		},
		{
			Id:       new(int64(69)),
			Fqdn:     new("oem/handle3"),
			Name:     "first",
			Created:  new(int64(8)),
			Modified: new(int64(14)),
		},
		{
			Id:       new(int64(56)),
			Fqdn:     new("oem/handle4"),
			Name:     "unchecked",
			Created:  new(int64(8)),
			Modified: new(int64(13)),
		},
	}
	var testParams = &broker.DataLinkListParams{}
	var testProvider = &userDataLinksProvider{
		Plan: task.Plan{
			Logger: logrus.New(),
		},
		filter:                 "oem",
		params:                 testParams,
		listLinksTaskFactory:   newDataLinkListTaskCompleteCapture(&calledParams, testSolutions),
		reduceLinksTaskFactory: newDataLinkListTaskCompleteCapture(&calledParams, testSolutions),
	}
	var actual []UserDataLink
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

func newDataLinkListTaskCompleteCapture(capture *broker.DataLinkListParams, seeded []broker.DataLink) func() *task.Task[broker.DataLinkListParams] {
	return func() *task.Task[broker.DataLinkListParams] {
		return &task.Task[broker.DataLinkListParams]{
			OnPrepare: func(params *broker.DataLinkListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkListParams, state *task.State) error {
				*capture = *params
				state.Internal = seeded
				return nil
			},
		}
	}
}

func newDataLinkListTaskCompleteError(err error) func() *task.Task[broker.DataLinkListParams] {
	return func() *task.Task[broker.DataLinkListParams] {
		return &task.Task[broker.DataLinkListParams]{
			OnPrepare: func(params *broker.DataLinkListParams, state *task.State) error {
				return nil
			},
			OnComplete: func(params *broker.DataLinkListParams, state *task.State) error {
				return err
			},
		}
	}
}
